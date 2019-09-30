package transaction

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/util"

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
		nodeQueries       [][]interface{}
		nodereGistrations []*model.NodeRegistration
	)

	nodeRow, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	nodereGistrations = tx.NodeRegistrationQuery.BuildModel(nodereGistrations, nodeRow)
	if len(nodereGistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
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
	nodeRegistrations = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, nodeRow)
	if len(nodeRegistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}
	// sender must be node owner
	if tx.SenderAddress != nodeRegistrations[0].AccountAddress {
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
