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
		Body                *model.ApprovalEscrowTransactionBody
		Escrow              *model.Escrow
		AccountBalanceQuery query.AccountBalanceQueryInterface
		QueryExecutor       query.ExecutorInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		EscrowQuery         query.EscrowTransactionQueryInterface
	}
	// EscrowTypeAction is escrow transaction type methods collection
	EscrowTypeAction interface {
		EscrowApplyConfirmed(blockTimestamp int64) error
		EscrowApplyUnconfirmed() error
		EscrowUndoApplyUnconfirmed() error
		EscrowValidate(dbTx bool) error
		EscrowApproval(int64) error
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
		err      error
		approval model.EscrowApproval
		id       int64
	)
	_, err = util.ReadTransactionBytes(bytes.NewBuffer(bodyBytes), int(tx.GetSize()))
	if err != nil {
		return nil, err
	}

	approval, id, err = ParseEscrowApprovalBytes(bodyBytes)
	if err != nil {
		return nil, err
	}
	return &model.ApprovalEscrowTransactionBody{
		Approval:      approval,
		TransactionID: id,
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
		latestEscrow model.Escrow
		row          *sql.Row
		err          error
	)

	escrowQ, escrowArgs := tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Escrow.GetID())
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

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SendMoney` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ApprovalEscrowTransaction) ApplyUnconfirmed() error {

	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *ApprovalEscrowTransaction) UndoApplyUnconfirmed() error {

	return nil
}

/*
ApplyConfirmed func that for applying Transaction SendMoney type.
		- If Genesis perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
		- If Not Genesis, perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
		`sender.balance` = current - amount
*/
func (tx *ApprovalEscrowTransaction) ApplyConfirmed(blockTimestamp int64) error {

	return nil
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *ApprovalEscrowTransaction) Escrowable() (EscrowTypeAction, bool) {

	return nil, false
}
