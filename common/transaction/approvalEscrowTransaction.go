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

type (
	// ApprovalEscrowTransaction field
	ApprovalEscrowTransaction struct {
		ID                  int64
		Fee                 int64
		SenderAddress       string
		Height              uint32
		Body                *model.ApprovalEscrowTransactionBody
		Escrow              *model.Escrow
		BlockQuery          query.BlockQueryInterface
		EscrowQuery         query.EscrowTransactionQueryInterface
		QueryExecutor       query.ExecutorInterface
		TransactionQuery    query.TransactionQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
		TypeActionSwitcher  TypeActionSwitcher
	}
	// EscrowTypeAction is escrow transaction type methods collection
	EscrowTypeAction interface {
		EscrowApplyConfirmed(blockTimestamp int64) error
		EscrowApplyUnconfirmed() error
		EscrowUndoApplyUnconfirmed() error
		EscrowValidate(dbTx bool) error
		EscrowApproval(
			blockTimestamp int64,
			txBody *model.ApprovalEscrowTransactionBody,
		) error
	}
)

// SkipMempoolTransaction this tx type has no mempool filter
func (*ApprovalEscrowTransaction) SkipMempoolTransaction([]*model.Transaction) (bool, error) {
	return false, nil
}

// GetSize of approval transaction body bytes
func (*ApprovalEscrowTransaction) GetSize() uint32 {
	return constant.EscrowApprovalBytesLength
}

// GetAmount return Amount from TransactionBody
func (*ApprovalEscrowTransaction) GetAmount() int64 {
	return 0
}

// GetBodyBytes translate tx body to bytes representation
func (tx *ApprovalEscrowTransaction) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(tx.Body.GetApproval())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.GetTransactionID())))
	return buffer.Bytes()
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *ApprovalEscrowTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_ApprovalEscrowTransactionBody{
		ApprovalEscrowTransactionBody: tx.Body,
	}
}

// ParseBodyBytes validate and parse body bytes to TransactionBody interface
func (tx *ApprovalEscrowTransaction) ParseBodyBytes(
	bodyBytes []byte,
) (model.TransactionBodyInterface, error) {
	var (
		buffer  = bytes.NewBuffer(bodyBytes)
		chunked []byte
		err     error
	)

	chunked, err = util.ReadTransactionBytes(buffer, int(constant.EscrowApproval))
	if err != nil {
		return nil, err
	}
	approvalInt := util.ConvertBytesToUint32(chunked)

	chunked, err = util.ReadTransactionBytes(buffer, int(constant.EscrowID))
	if err != nil {
		return nil, err
	}
	escrowID := util.ConvertBytesToUint64(chunked)

	return &model.ApprovalEscrowTransactionBody{
		Approval:      model.EscrowApproval(approvalInt),
		TransactionID: int64(escrowID),
	}, nil
}

/*
Validate is func that for validating to Transaction SendMoney type
That specs:
	- If Genesis, sender and recipient allowed not exists,
	- If Not Genesis,  sender and recipient must be exists, `sender.spendable_balance` must bigger than amount
*/
func (tx *ApprovalEscrowTransaction) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		latestEscrow   model.Escrow
		row            *sql.Row
		err            error
	)

	escrowQ, escrowArgs := tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Body.GetTransactionID())
	row, err = tx.QueryExecutor.ExecuteSelectRow(escrowQ, dbTx, escrowArgs...)
	if err != nil {
		return err
	}

	// Check escrow exists
	err = tx.EscrowQuery.Scan(&latestEscrow, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowNotExists")
	}

	// Check escrow status still pending before allow to apply
	if latestEscrow.GetStatus() != model.EscrowStatus_Pending {
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowTargetNotValidByStatus")
	}

	// Check sender, should be approver address
	if latestEscrow.GetApproverAddress() != tx.SenderAddress {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSenderAddress")
	}

	// check transaction id is valid
	if latestEscrow.GetID() != tx.Body.GetTransactionID() {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidTransactionID")
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
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidAccountSender")
	}
	if accountBalance.SpendableBalance < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "UserBalanceNotEnough")
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SendMoney` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ApprovalEscrowTransaction) ApplyUnconfirmed() error {
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-tx.Fee,
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
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *ApprovalEscrowTransaction) UndoApplyUnconfirmed() error {
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

/*
ApplyConfirmed func that for applying Transaction SendMoney type.
If Genesis perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
If Not Genesis, perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
`sender.balance` = current - amount
*/
func (tx *ApprovalEscrowTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		latestEscrow model.Escrow
		transaction  model.Transaction
		queries      [][]interface{}
		txType       TypeAction
		row          *sql.Row
		err          error
	)

	// Get escrow by reference transaction ID
	escrowQ, escrowArgs := tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Body.GetTransactionID())
	row, err = tx.QueryExecutor.ExecuteSelectRow(escrowQ, false, escrowArgs...)
	if err != nil {
		return err
	}

	err = tx.EscrowQuery.Scan(&latestEscrow, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "EscrowNotFound")
	}

	// get what transaction type it is, and switch to specific approval
	transactionQ := tx.TransactionQuery.GetTransaction(latestEscrow.GetID())
	row, err = tx.QueryExecutor.ExecuteSelectRow(transactionQ, false)
	if err != nil {
		return err
	}
	err = tx.TransactionQuery.Scan(&transaction, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "TransactionNotFound")

	}
	transaction.Escrow = &latestEscrow

	txType, err = tx.TypeActionSwitcher.GetTransactionType(&transaction)
	if err != nil {
		return err
	}

	// now only send money has EscrowApproval method
	escrowable, ok := txType.Escrowable()
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "ExpectEscrowableTransaction")
	}
	err = escrowable.EscrowApproval(blockTimestamp, tx.Body)
	if err != nil {
		return blocker.NewBlocker(blocker.AppErr, "EscrowApprovalFailed")
	}

	// Update sender
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// Sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventApprovalEscrowTransaction,
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
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *ApprovalEscrowTransaction) Escrowable() (EscrowTypeAction, bool) {

	if tx.Escrow.GetApproverAddress() != "" {
		return EscrowTypeAction(tx), true
	}
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
func (tx *ApprovalEscrowTransaction) EscrowValidate(dbTx bool) error {

	return nil
}

/*
EscrowApplyUnconfirmed is applyUnconfirmed specific for Escrow's transaction
similar with ApplyUnconfirmed and Escrow.Commission
*/
func (tx *ApprovalEscrowTransaction) EscrowApplyUnconfirmed() error {

	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *ApprovalEscrowTransaction) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed func that for applying Transaction SendMoney type.
*/
func (tx *ApprovalEscrowTransaction) EscrowApplyConfirmed(int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *ApprovalEscrowTransaction) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {
	return nil
}
