package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// RemoveNodeRegistration Implement service layer for (new) node registration's transaction
type RemoveNodeRegistration struct {
	ID                    int64
	Fee                   int64
	SenderAddress         string
	Height                uint32
	Body                  *model.RemoveNodeRegistrationTransactionBody
	Escrow                *model.Escrow
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	QueryExecutor         query.ExecutorInterface
	AccountLedgerQuery    query.AccountLedgerQueryInterface
	EscrowQuery           query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *RemoveNodeRegistration) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
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
func (tx *RemoveNodeRegistration) ApplyConfirmed(blockTimestamp int64) error {

	var (
		nodeQueries       [][]interface{}
		nodeRegistrations []*model.NodeRegistration
	)

	nodeRow, err := tx.QueryExecutor.ExecuteSelect(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false, tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	nodeRegistrations, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, nodeRow)
	if (err != nil) || len(nodeRegistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	prevNodeRegistration := nodeRegistrations[0]
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

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  prevNodeRegistration.GetLockedBalance() - tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

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

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *RemoveNodeRegistration) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow != nil {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
		}

		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *RemoveNodeRegistration) EscrowValidate(dbTx bool) error {
	var (
		nodeRegistration model.NodeRegistration
		row              *sql.Row
		err              error
	)
	if tx.Escrow.GetApproverAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}

	// check for duplication
	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		dbTx,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "NodeNotRegistered")
	}

	// sender must be node owner
	if tx.SenderAddress != nodeRegistration.GetAccountAddress() {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountNotNodeOwner")
	}
	if nodeRegistration.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyDeleted")
	}
	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveNodeRegistration` type.
Perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *RemoveNodeRegistration) EscrowApplyUnconfirmed() error {

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee + tx.Escrow.GetCommission()),
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

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *RemoveNodeRegistration) EscrowUndoApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Fee+tx.Escrow.GetCommission(),
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

// EscrowApplyConfirmed method for confirmed the transaction and store into database
func (tx *RemoveNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {

	var (
		nodeRegistration model.NodeRegistration
		queries          [][]interface{}
		row              *sql.Row
		err              error
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// Rebuild node registration
	nodeRegistration.Height = tx.Height
	nodeRegistration.LockedBalance = 0
	nodeRegistration.NodeAddress = nil
	nodeRegistration.NodePublicKey = tx.Body.GetNodePublicKey()
	nodeRegistration.Latest = true
	nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)

	// sender balance by refunding the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		nodeRegistration.GetLockedBalance()-(tx.Fee+tx.Escrow.GetCommission()),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  nodeRegistration.GetLockedBalance() - (tx.Fee + tx.Escrow.GetCommission()),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *RemoveNodeRegistration) EscrowApproval(blockTimestamp int64) error {
	var (
		nodeRegistration model.NodeRegistration
		queries          [][]interface{}
		row              *sql.Row
		err              error
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// Rebuild node registration
	nodeRegistration.Height = tx.Height
	nodeRegistration.LockedBalance = 0
	nodeRegistration.NodeAddress = nil
	nodeRegistration.NodePublicKey = tx.Body.GetNodePublicKey()
	nodeRegistration.Latest = true
	nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)

	// Node registration
	queries = append(queries, tx.NodeRegistrationQuery.UpdateNodeRegistration(&nodeRegistration)...)

	// approver balance
	approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
		tx.Escrow.GetCommission(),
		map[string]interface{}{
			"account_address": tx.Escrow.GetApproverAddress(),
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, approverBalanceQ...)
	// approver ledger
	approverLedgerQ, approverLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.Escrow.GetApproverAddress(),
		BalanceChange:  tx.Escrow.GetCommission(),
		BlockHeight:    tx.Height,
		TransactionID:  tx.ID,
		Timestamp:      uint64(blockTimestamp),
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
	})
	approverLedgerArgs = append([]interface{}{approverLedgerQ}, approverLedgerArgs...)
	queries = append(queries, approverLedgerArgs)

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}
