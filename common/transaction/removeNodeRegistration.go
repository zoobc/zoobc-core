package transaction

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
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

// FilterMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *RemoveNodeRegistration) FilterMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	for _, sel := range selectedTransactions {
		// if we find another node registration tx in currently selected transactions, filter current one out of selection
		buf := util.ConvertUint32ToBytes(sel.GetTransactionType())
		if buf[0] == 2 {
			return true, nil
		}
	}
	// no tx found matching filter conditions, so do not filter (we should never get here anyways)
	return false, nil
}

func (tx *RemoveNodeRegistration) ApplyConfirmed() error {

	var (
		nodeQueries       [][]interface{}
		nodereGistrations []*model.NodeRegistration
	)

	nodeRow, err := tx.QueryExecutor.ExecuteSelect(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false, tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	nodereGistrations, err = tx.NodeRegistrationQuery.BuildModel(nodereGistrations, nodeRow)
	if (err != nil) || len(nodereGistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	prevNodeRegistration := nodereGistrations[0]
	// tag the node as deleted
	nodeRegistration := &model.NodeRegistration{
		NodeID:             prevNodeRegistration.NodeID,
		LockedBalance:      0,
		Height:             tx.Height,
		NodeAddress:        nil,
		RegistrationHeight: prevNodeRegistration.RegistrationHeight,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: prevNodeRegistration.AccountAddress,
	}
	// update sender balance by refunding the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		prevNodeRegistration.LockedBalance-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	nodeQueries = tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries := append(accountBalanceSenderQ, nodeQueries...)
	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
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
		nodeRegistrations []*model.NodeRegistration
	)
	// check for duplication
	nodeRow, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), dbTx, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	nodeRegistrations, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, nodeRow)
	if (err != nil) || len(nodeRegistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}
	nr := nodeRegistrations[0]
	// sender must be node owner
	if tx.SenderAddress != nr.AccountAddress {
		return blocker.NewBlocker(blocker.AuthErr, "AccountNotNodeOwner")
	}
	if nr.RegistrationStatus == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "NodeAlreadyDeleted")
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
func (tx *RemoveNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	txBody := &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveNodeRegistration) GetBodyBytes() []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	return buffer.Bytes()
}

func (tx *RemoveNodeRegistration) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_RemoveNodeRegistrationTransactionBody{
		RemoveNodeRegistrationTransactionBody: tx.Body,
	}
}
