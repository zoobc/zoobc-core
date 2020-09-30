package transaction

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

// NodeRegistration Implement service layer for (new) node registration's transaction
type NodeRegistration struct {
	ID                       int64
	Fee                      int64
	SenderAddress            []byte
	Height                   uint32
	Body                     *model.NodeRegistrationTransactionBody
	Escrow                   *model.Escrow
	AccountBalanceQuery      query.AccountBalanceQueryInterface
	NodeRegistrationQuery    query.NodeRegistrationQueryInterface
	BlockQuery               query.BlockQueryInterface
	ParticipationScoreQuery  query.ParticipationScoreQueryInterface
	QueryExecutor            query.ExecutorInterface
	AuthPoown                auth.NodeAuthValidationInterface
	AccountLedgerQuery       query.AccountLedgerQueryInterface
	EscrowQuery              query.EscrowTransactionQueryInterface
	PendingNodeRegistryCache storage.TransactionalCache
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *NodeRegistration) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
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

// ApplyConfirmed method for confirmed the transaction and store into database
func (tx *NodeRegistration) ApplyConfirmed(blockTimestamp int64) error {
	var (
		queries                                                     [][]interface{}
		registrationStatus                                          uint32
		prevNodeRegistrationByPubKey, prevNodeRegistrationByAccount model.NodeRegistration
		nodeAccountAddress                                          []byte
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
	queries = append(queries, accountBalanceSenderQ...)

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.Body.LockedBalance + tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	nodeRow, _ := tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	err := tx.NodeRegistrationQuery.Scan(&prevNodeRegistrationByPubKey, nodeRow)
	prevNodeFound := true
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		prevNodeFound = false
	}

	if prevNodeFound {
		if prevNodeRegistrationByPubKey.RegistrationStatus != uint32(model.NodeRegistrationState_NodeDeleted) {
			// there can't be two nodes registered with the same pub key
			return errors.New("NodePublicKeyAlreadyInRegistry")
		}
		// if there is a previously deleted node registration, set its latest status to false, to avoid duplicates
		clearDeletedNodeRegistrationQ := tx.NodeRegistrationQuery.ClearDeletedNodeRegistration(&prevNodeRegistrationByPubKey)
		queries = append(queries, clearDeletedNodeRegistrationQ...)
	} else {
		// check if this account previously deleted a registered node. in that case, set the 'deleted' one's latest to 0
		// check for account address duplication (accounts can register one node at the time)
		qryNodeByAccount, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(nodeAccountAddress)
		nodeRow2, _ := tx.QueryExecutor.ExecuteSelectRow(qryNodeByAccount, false, args...)
		err := tx.NodeRegistrationQuery.Scan(&prevNodeRegistrationByAccount, nodeRow2)
		prevNodeFound := true
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			prevNodeFound = false
		}
		// in case a node with same account address has been previosly deleted, set its latest status to false
		// to avoid having duplicates (multiple node registrations with same account address)
		if prevNodeFound && prevNodeRegistrationByAccount.RegistrationStatus ==
			uint32(model.NodeRegistrationState_NodeDeleted) {
			clearDeletedNodeRegistrationQ := tx.NodeRegistrationQuery.ClearDeletedNodeRegistration(&prevNodeRegistrationByAccount)
			queries = append(queries, clearDeletedNodeRegistrationQ...)
		}
	}

	// if a node with this public key has been previously deleted, update its owner to the new registerer
	nodeRegistration := &model.NodeRegistration{
		NodeID:             tx.ID,
		LockedBalance:      tx.Body.LockedBalance,
		Height:             tx.Height,
		RegistrationHeight: tx.Height,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: registrationStatus,
		AccountAddress:     nodeAccountAddress,
	}

	updateNodeRegistrationQ := tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries = append(queries, updateNodeRegistrationQ...)

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
		// update node registry cache (in transaction) and resort
		err = tx.PendingNodeRegistryCache.TxSetItem(nil, storage.NodeRegistry{
			Node:               *nodeRegistration,
			ParticipationScore: ps.Score,
		})
		if err != nil {
			return err
		}
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
		nodeRegByNodePub, nodeRegByAccAddress model.NodeRegistration
		accountBalance                        model.AccountBalance
		row                                   *sql.Row
		err                                   error
	)

	// no need to validate node registration transaction for genesis block
	if tx.SenderAddress == constant.MainchainGenesisAccountAddress {
		return nil
	}

	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}

	// validate poown
	if err := tx.AuthPoown.ValidateProofOfOwnership(tx.Body.Poown, tx.Body.NodePublicKey, tx.QueryExecutor, tx.BlockQuery); err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	// check balance
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, "AccountNotFound")
	}

	if accountBalance.GetSpendableBalance() < tx.Body.GetLockedBalance()+tx.Fee {
		return blocker.NewBlocker(blocker.AppErr, "UserBalanceNotEnough")
	}

	// check for public key duplication
	row, err = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), dbTx, tx.Body.GetNodePublicKey())
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegByNodePub, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
	} else {
		// in case a node with same pub key exists, validation must pass only if that node is tagged as deleted
		// if any other state validation should fail
		if nodeRegByNodePub.GetRegistrationStatus() != uint32(model.NodeRegistrationState_NodeDeleted) {
			return blocker.NewBlocker(blocker.AuthErr, "NodeAlreadyRegistered")
		}
	}

	// check for account address duplication (accounts can register one node at the time)
	qryNodeByAccount, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.Body.AccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qryNodeByAccount, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegByAccAddress, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		if nodeRegByAccAddress.GetRegistrationStatus() != uint32(model.NodeRegistrationState_NodeDeleted) {
			return blocker.NewBlocker(blocker.AuthErr, "AccountAlreadyNodeOwner")
		}
	}

	return nil
}

func (tx *NodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (*NodeRegistration) GetMinimumFee() (int64, error) {
	return 0, nil
}

func (tx *NodeRegistration) GetSize() uint32 {
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.AccountAddressLength + constant.AccountAddress +
		constant.Balance + poown
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
		LockedBalance:  int64(lockedBalance),
		Poown:          poown,
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *NodeRegistration) GetBodyBytes() []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.AccountAddress)))))
	buffer.Write([]byte(tx.Body.AccountAddress))
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
	for _, genesisEntry := range constant.GenesisConfig {
		if bytes.Equal(tx.Body.NodePublicKey, genesisEntry.NodePublicKey) {
			return genesisEntry.ParticipationScore
		}
	}
	return constant.DefaultParticipationScore
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *NodeRegistration) Escrowable() (EscrowTypeAction, bool) {

	return nil, false
}

/**
Escrow Part
1. ApplyUnconfirmed
2. UndoApplyUnconfirmed
3. ApplyConfirmed
4. Validate
*/

// EscrowValidate special validation for escrow's transaction
func (tx *NodeRegistration) EscrowValidate(dbTx bool) error {

	return nil
}

/*
EscrowApplyUnconfirmed is applyUnconfirmed specific for Escrow's transaction
similar with ApplyUnconfirmed and Escrow.Commission
*/
func (tx *NodeRegistration) EscrowApplyUnconfirmed() error {

	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *NodeRegistration) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed func that for applying Transaction SendMoney type.
*/
func (tx *NodeRegistration) EscrowApplyConfirmed(int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *NodeRegistration) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {
	return nil
}
