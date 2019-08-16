package transaction

import (
	"bytes"
	"errors"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// NodeRegistration Implement service layer for (new) node registration's transaction
type NodeRegistration struct {
	ID                    int64
	Body                  *model.NodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
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
		NodeID:             tx.ID,
		LockedBalance:      tx.Body.LockedBalance,
		Height:             tx.Height,
		NodeAddress:        tx.Body.NodeAddress,
		RegistrationHeight: tx.Height,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		Queued:             true,
		AccountAddress:     tx.Body.AccountAddress,
	}
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.LockedBalance + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
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
func (tx *NodeRegistration) Validate() error {
	var (
		accountBalance model.AccountBalance
	)
	// check balance
	senderQ, senderArg := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(senderQ, senderArg)
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
	nodeAddress := uint32(len([]byte(tx.Body.NodeAddress)))
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.AccountAddressLength + constant.NodeAddressLength + constant.AccountAddress +
		constant.Balance + nodeAddress + poown
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*NodeRegistration) ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey := buffer.Next(int(constant.NodePublicKey))
	accountAddressLength := util.ConvertBytesToUint32(buffer.Next(int(constant.AccountAddressLength)))
	accountAddress := buffer.Next(int(accountAddressLength))
	nodeAddressLength := util.ConvertBytesToUint32(buffer.Next(int(constant.NodeAddressLength))) // uint32 length of next bytes to read
	nodeAddress := buffer.Next(int(nodeAddressLength))                                           // based on nodeAddressLength
	lockedBalance := util.ConvertBytesToUint64(buffer.Next(int(constant.Balance)))
	poown := util.ParseProofOfOwnershipBytes(buffer.Next(int(util.GetProofOfOwnershipSize(true))))
	txBody := &model.NodeRegistrationTransactionBody{
		NodePublicKey:  nodePublicKey,
		AccountAddress: string(accountAddress),
		NodeAddress:    string(nodeAddress),
		LockedBalance:  int64(lockedBalance),
		Poown:          poown,
	}
	return txBody
}

// GetBodyBytes translate tx body to bytes representation
func (tx *NodeRegistration) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.AccountAddress)))))
	buffer.Write([]byte(tx.Body.AccountAddress))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.NodeAddress)))))
	buffer.Write([]byte(tx.Body.NodeAddress))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.LockedBalance)))
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
}
