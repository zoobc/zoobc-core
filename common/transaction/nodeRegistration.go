package transaction

import (
	"bytes"
	"errors"
	"net"
	"net/url"
	"strconv"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// NodeRegistration Implement service layer for (new) node registration's transaction
type NodeRegistration struct {
	ID                      int64
	Body                    *model.NodeRegistrationTransactionBody
	Fee                     int64
	SenderAddress           string
	Height                  uint32
	AccountBalanceQuery     query.AccountBalanceQueryInterface
	NodeRegistrationQuery   query.NodeRegistrationQueryInterface
	BlockQuery              query.BlockQueryInterface
	ParticipationScoreQuery query.ParticipationScoreQueryInterface
	QueryExecutor           query.ExecutorInterface
	AuthPoown               auth.ProofOfOwnershipValidationInterface
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *NodeRegistration) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	authorizedType := map[model.TransactionType]bool{
		model.TransactionType_ClaimNodeRegistrationTransaction:  true,
		model.TransactionType_UpdateNodeRegistrationTransaction: true,
		model.TransactionType_RemoveNodeRegistrationTransaction: true,
	}
	for _, sel := range selectedTransactions {
		// if we find another node registration tx in currently selected transactions, filter current one out of selection
		if _, ok := authorizedType[model.TransactionType(sel.GetTransactionType())]; ok && tx.SenderAddress == sel.SenderAccountAddress {
			return true, nil
		}
	}
	return false, nil
}

func (tx *NodeRegistration) ApplyConfirmed() error {
	var (
		queries            [][]interface{}
		registrationStatus uint32
		nodeRegistrations  []*model.NodeRegistration
		nodeAccountAddress string
	)
	if tx.Height > 0 {
		registrationStatus = uint32(model.NodeRegistrationState_NodeQueued)
		nodeAccountAddress = tx.SenderAddress
	} else {
		registrationStatus = uint32(model.NodeRegistrationState_NodeRegistered)
		nodeAccountAddress = tx.Body.AccountAddress
	}

	// update sender balance by reducing his spendable balance of the tx fee and locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.LockedBalance + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)

	nodeRow, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	defer nodeRow.Close()

	nodeRegistrations, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, nodeRow)
	if err != nil {
		return err
	}
	// if a node with this public key has been previously deleted, update its owner to the new registerer
	nodeRegistration := &model.NodeRegistration{
		NodeID:             tx.ID,
		LockedBalance:      tx.Body.LockedBalance,
		Height:             tx.Height,
		NodeAddress:        tx.Body.NodeAddress,
		RegistrationHeight: tx.Height,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: registrationStatus,
		AccountAddress:     nodeAccountAddress,
	}
	if len(nodeRegistrations) > 0 {
		if nodeRegistrations[0].RegistrationStatus == uint32(model.NodeRegistrationState_NodeDeleted) {
			queries = tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
			queries = append(queries, accountBalanceSenderQ...)
		} else {
			// this can happen if there are two node register tx with same node pub key submitted together,
			// racing to be included in the same block. Only the first one will make it through
			return errors.New("NodeAlreadyInRegistry")
		}
	} else {
		insertNodeQ, insertNodeArg := tx.NodeRegistrationQuery.InsertNodeRegistration(nodeRegistration)
		queries = append(append([][]interface{}{}, accountBalanceSenderQ...),
			append([]interface{}{insertNodeQ}, insertNodeArg...),
		)
	}

	// insert default participation score for nodes that are registered at genesis height
	if tx.Height == 0 {
		ps := &model.ParticipationScore{
			NodeID: tx.ID,
			Score:  tx.getDefaultParticipationScore(),
			Latest: true,
			Height: 0,
		}
		insertParticipationScoreQ, insertParticipationScoreArg := tx.ParticipationScoreQuery.InsertParticipationScore(ps)
		newQ := append([]interface{}{insertParticipationScoreQ}, insertParticipationScoreArg...)
		queries = append(queries, newQ)
	}

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `NodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *NodeRegistration) ApplyUnconfirmed() error {

	var (
		err error
	)

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.LockedBalance + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	// add row to node_registry table
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (tx *NodeRegistration) UndoApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.LockedBalance+tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *NodeRegistration) Validate(dbTx bool) error {

	var (
		accountBalance                        model.AccountBalance
		nodeRegistrations, nodeRegistrations2 []*model.NodeRegistration
	)

	// no need to validate node registration transaction for genesis block
	if tx.SenderAddress == constant.MainchainGenesisAccountAddress {
		return nil
	}

	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}
	if tx.Body.GetNodeAddress() == nil {
		return blocker.NewBlocker(blocker.RequestParameterErr, "NodeAddressRequired")
	}

	// validate poown
	if err := tx.AuthPoown.ValidateProofOfOwnership(tx.Body.Poown, tx.Body.NodePublicKey, tx.QueryExecutor, tx.BlockQuery); err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	// check balance
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(qry, dbTx, args...)
	if err != nil {
		return err
	} else if rows.Next() {
		_ = rows.Scan(
			&accountBalance.AccountAddress,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest,
		)
	}
	defer rows.Close()

	if accountBalance.SpendableBalance < tx.Body.LockedBalance+tx.Fee {
		return blocker.NewBlocker(blocker.AppErr, "UserBalanceNotEnough")
	}
	// check for public key duplication
	nodeRow, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		dbTx, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	defer nodeRow.Close()
	nodeRegistrations, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, nodeRow)
	if err != nil {
		return err
	}
	// in case a node with same pub key exists, validation must pass only if that node is tagged as deleted
	// if any other state validation should fail
	if len(nodeRegistrations) > 0 && nodeRegistrations[0].RegistrationStatus != uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "NodeAlreadyRegistered")
	}

	// check for account address duplication (accounts can register one node at the time)
	qryNodeByAccount, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.Body.AccountAddress)
	nodeRow2, err := tx.QueryExecutor.ExecuteSelect(qryNodeByAccount, dbTx, args...)
	if err != nil {
		return err
	}
	defer nodeRow2.Close()
	nodeRegistrations2, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations2, nodeRow2)
	if err != nil {
		return err
	}
	// in case a node with same account address, validation must pass only if that node is tagged as deleted
	// if any other state validation should fail
	if len(nodeRegistrations2) > 0 && nodeRegistrations2[0].RegistrationStatus != uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "AccountAlreadyNodeOwner")
	}

	// validate node address
	nodeAddress := tx.Body.GetNodeAddress()
	if nodeAddress == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAddressEmpty")
	}
	_, err = url.ParseRequestURI(tx.NodeRegistrationQuery.ExtractNodeAddress(
		nodeAddress,
	))
	if err != nil {
		if ip := net.ParseIP(nodeAddress.GetAddress()); ip == nil {
			return blocker.NewBlocker(blocker.ValidationErr, "InvalidNodeAddress:IP")
		}
		port := int(nodeAddress.GetPort())
		if _, err := strconv.ParseUint(strconv.Itoa(port), 10, 16); err != nil {
			return blocker.NewBlocker(blocker.ValidationErr, "InvalidNodeAddress:Port")
		}
	}

	return nil
}

func (tx *NodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (tx *NodeRegistration) GetSize() uint32 {
	nodeAddress := uint32(len([]byte(tx.NodeRegistrationQuery.ExtractNodeAddress(
		tx.Body.GetNodeAddress(),
	))))
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.AccountAddressLength + constant.NodeAddressLength + constant.AccountAddress +
		constant.Balance + nodeAddress + poown
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *NodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	accountAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	accountAddressLength := util.ConvertBytesToUint32(accountAddressLengthBytes)
	accountAddress, err := util.ReadTransactionBytes(buffer, int(accountAddressLength))
	if err != nil {
		return nil, err
	}
	nodeAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.NodeAddressLength))
	if err != nil {
		return nil, err
	}
	nodeAddressLength := util.ConvertBytesToUint32(nodeAddressLengthBytes)        // uint32 length of next bytes to read
	nodeAddress, err := util.ReadTransactionBytes(buffer, int(nodeAddressLength)) // based on nodeAddressLength
	if err != nil {
		return nil, err
	}
	lockedBalanceBytes, err := util.ReadTransactionBytes(buffer, int(constant.Balance))
	if err != nil {
		return nil, err
	}
	lockedBalance := util.ConvertBytesToUint64(lockedBalanceBytes)
	poownBytes, err := util.ReadTransactionBytes(buffer, int(util.GetProofOfOwnershipSize(true)))
	if err != nil {
		return nil, err
	}
	poown, err := util.ParseProofOfOwnershipBytes(poownBytes)
	if err != nil {
		return nil, err
	}

	txBody := &model.NodeRegistrationTransactionBody{
		NodePublicKey:  nodePublicKey,
		AccountAddress: string(accountAddress),
		NodeAddress:    tx.NodeRegistrationQuery.BuildNodeAddress(string(nodeAddress)),
		LockedBalance:  int64(lockedBalance),
		Poown:          poown,
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *NodeRegistration) GetBodyBytes() []byte {

	var fullNodeAddress = tx.NodeRegistrationQuery.ExtractNodeAddress(tx.Body.GetNodeAddress())

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.AccountAddress)))))
	buffer.Write([]byte(tx.Body.AccountAddress))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(
		fullNodeAddress,
	)))))
	buffer.Write([]byte(
		fullNodeAddress,
	))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.LockedBalance)))
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
}

func (tx *NodeRegistration) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_NodeRegistrationTransactionBody{
		NodeRegistrationTransactionBody: tx.Body,
	}
}

func (tx *NodeRegistration) getDefaultParticipationScore() int64 {
	for _, genesisEntry := range constant.MainChainGenesisConfig {
		if bytes.Equal(tx.Body.NodePublicKey, genesisEntry.NodePublicKey) {
			return genesisEntry.ParticipationScore
		}
	}
	return constant.DefaultParticipationScore
}
