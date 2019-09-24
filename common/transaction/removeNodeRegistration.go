package transaction

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

// RemoveNodeRegistration Implement service layer for (new) node registration's transaction
type RemoveNodeRegistration struct {
	Body                  *model.RemoveNodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	QueryExecutor         query.ExecutorInterface
}

func (tx *RemoveNodeRegistration) ApplyConfirmed() error {
	var (
		queries           [][]interface{}
		nodereGistrations []*model.NodeRegistration
	)

	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey)
	nodeRow, err := tx.QueryExecutor.ExecuteSelect(qry, false, args)
	if err != nil {
		return err
	}
	nodereGistrations = tx.NodeRegistrationQuery.BuildModel(nodereGistrations, nodeRow)
	if len(nodereGistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}
	if err := tx.QueryExecutor.BeginTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotInitiated")
	}

	prevNodeRegistration := nodereGistrations[0]
	nodeRegistration := &model.NodeRegistration{
		NodeID:             prevNodeRegistration.NodeID,
		LockedBalance:      0,
		Height:             tx.Height,
		NodeAddress:        "",
		RegistrationHeight: prevNodeRegistration.RegistrationHeight,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		Queued:             true,
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: "00000000000000000000000000000000000000000000",
	}
	// update sender balance by refunding the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		(prevNodeRegistration.LockedBalance - tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	insertNodeQ, insertNodeArg := tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries = append(append([][]interface{}{}, accountBalanceSenderQ...),
		append([]interface{}{insertNodeQ}, insertNodeArg...),
	)
	// add row to node_registry table
	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	if err := tx.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *RemoveNodeRegistration) ApplyUnconfirmed() error {
	var (
		err error
	)

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (tx *RemoveNodeRegistration) UndoApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Fee,
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
func (tx *RemoveNodeRegistration) Validate(dbTx bool) error {
	var (
		nodereGistrations []*model.NodeRegistration
	)
	// check for duplication
	nodeQuery, nodeArg := tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey)
	nodeRow, err := tx.QueryExecutor.ExecuteSelect(nodeQuery, dbTx, nodeArg...)
	if err != nil {
		return err
	}
	nodereGistrations = tx.NodeRegistrationQuery.BuildModel(nodereGistrations, nodeRow)
	if len(nodereGistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}
	// sender must be node owner
	if tx.SenderAddress != nodereGistrations[0].AccountAddress {
		return blocker.NewBlocker(blocker.AuthErr, "AccountNotNodeOwner")
	}
	return nil
}

func (tx *RemoveNodeRegistration) GetAmount() int64 {
	return 0
}

func (tx *RemoveNodeRegistration) GetSize() uint32 {
	return constant.NodePublicKey
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*RemoveNodeRegistration) ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey := buffer.Next(int(constant.NodePublicKey))
	txBody := &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
	}
	return txBody
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveNodeRegistration) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	return buffer.Bytes()
}
