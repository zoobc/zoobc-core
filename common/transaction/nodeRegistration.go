package transaction

import (
	"bytes"
	"errors"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// NodeRegistration Implement service layer for (new) node registration's transaction
type NodeRegistration struct {
	Body                  *model.NodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	SenderAccountType     uint32
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	AccountQuery          query.AccountQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	QueryExecutor         query.ExecutorInterface
}

func (tx *NodeRegistration) ApplyConfirmed() error {
	var (
		queries [][]interface{}
	)
	if err := tx.Validate(); err != nil {
		return err
	}

	if tx.Height > 0 {
		err := tx.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}
	nodeRegistration := &model.NodeRegistration{
		LockedBalance:      tx.Body.LockedBalance,
		Height:             tx.Height,
		NodeAddress:        tx.Body.NodeAddress,
		RegistrationHeight: tx.Height,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		Queued:             true,
		AccountId:          util.CreateAccountIDFromAddress(tx.Body.AccountType, tx.Body.AccountAddress),
	}
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.LockedBalance + tx.Fee),
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
			"block_height": tx.Height,
		},
	)
	insertNodeQ, insertNodeArg := tx.NodeRegistrationQuery.InsertNodeRegistration(nodeRegistration)
	queries = append(append([][]interface{}{}, accountBalanceSenderQ...),
		append([]interface{}{insertNodeQ}, insertNodeArg...),
	)
	// add row to node_registry table
	err := tx.QueryExecutor.ExecuteTransactions(queries)
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

	if err := tx.Validate(); err != nil {
		return err
	}

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.LockedBalance + tx.Fee),
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
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
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *NodeRegistration) Validate() error {
	var (
		accountBalance model.AccountBalance
	)
	// check balance
	senderID := util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress)
	senderQ, senderArg := tx.AccountBalanceQuery.GetAccountBalanceByAccountID(senderID)
	rows, err := tx.QueryExecutor.ExecuteSelect(senderQ, senderArg)
	if err != nil {
		return err
	} else if rows.Next() {
		_ = rows.Scan(
			&accountBalance.AccountID,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest,
		)
	}
	defer rows.Close()

	if accountBalance.SpendableBalance < tx.Body.LockedBalance+tx.Fee {
		return errors.New("UserBalanceNotEnough")
	}
	// check for duplication
	nodeQuery, nodeArg := tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey)
	nodeRow, err := tx.QueryExecutor.ExecuteSelect(nodeQuery, nodeArg...)
	if err != nil {
		return err
	}
	defer nodeRow.Close()
	if nodeRow.Next() {
		return errors.New("node already registered")
	}
	return nil
}

func (tx *NodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (tx *NodeRegistration) GetSize() uint32 {
	nodePublicKey := 32
	accountType := 2
	//TODO: this is valid for account type = 0
	accountAddress := 44
	nodeAddressLength := 1
	nodeAddress := tx.Body.NodeAddressLength
	lockedBalance := 8
	//TODO: return bytes of ProofOfOwnership (message + signature) when implemented
	poown := 256
	return uint32(nodePublicKey+accountType+nodeAddressLength+accountAddress+lockedBalance+poown) + nodeAddress
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*NodeRegistration) ParseBodyBytes(txBodyBytes []byte) *model.NodeRegistrationTransactionBody {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey := buffer.Next(32)
	accountTypeBytes := buffer.Next(2)
	accountType := util.ConvertBytesToUint32([]byte{accountTypeBytes[0], accountTypeBytes[1], 0, 0})
	accountAddressBytes := buffer.Next(44)
	nodeAddressLength := util.ConvertBytesToUint32([]byte{buffer.Next(1)[0], 0, 0, 0}) // uint32 length of next bytes to read
	nodeAddress := buffer.Next(int(nodeAddressLength))                                 // based on nodeAddressLength
	lockedBalance := util.ConvertBytesToUint64(buffer.Next(8))
	//TODO: parse ProofOfOwnership (message + signature) bytes when implemented
	poown := new(model.ProofOfOwnership)
	return &model.NodeRegistrationTransactionBody{
		NodePublicKey:     nodePublicKey,
		AccountType:       accountType,
		AccountAddress:    string(accountAddressBytes),
		NodeAddressLength: nodeAddressLength,
		NodeAddress:       string(nodeAddress),
		LockedBalance:     int64(lockedBalance),
		Poown:             poown,
	}
}

// GetBodyBytes translate tx body to bytes representation
func (*NodeRegistration) GetBodyBytes(txBody *model.NodeRegistrationTransactionBody) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(txBody.NodePublicKey)
	buffer.Write(util.ConvertUint32ToBytes(txBody.AccountType)[:2])
	buffer.Write([]byte(txBody.AccountAddress))
	addressLengthBytes := util.ConvertUint32ToBytes(txBody.NodeAddressLength)
	buffer.Write([]byte{addressLengthBytes[0]})
	buffer.Write([]byte(txBody.NodeAddress))
	buffer.Write(util.ConvertUint64ToBytes(uint64(txBody.LockedBalance)))
	//TODO: convert ProofOfOwnership (message + signature) to bytes
	buffer.Write([]byte{})
	return buffer.Bytes()
}
