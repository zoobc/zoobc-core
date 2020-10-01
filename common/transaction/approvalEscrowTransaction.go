package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// ApprovalEscrowTransaction field
	ApprovalEscrowTransaction struct {
		ID                   int64
		Fee                  int64
		SenderAddress        string
		Height               uint32
		Body                 *model.ApprovalEscrowTransactionBody
		Escrow               *model.Escrow
		BlockQuery           query.BlockQueryInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		QueryExecutor        query.ExecutorInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
		EscrowFee            fee.FeeModelInterface
		NormalFee            fee.FeeModelInterface
	}
	// EscrowTypeAction is escrow transaction type methods collection
	EscrowTypeAction interface {
		// EscrowApplyConfirmed perhaps this method called with QueryExecutor.BeginTX() because inside this process has separated QueryExecutor.Execute
		EscrowApplyConfirmed(blockTimestamp int64) error
		EscrowApplyUnconfirmed() error
		EscrowUndoApplyUnconfirmed() error
		EscrowValidate(dbTx bool) error
		// EscrowApproval handle approval an escrow transaction, execute tasks that was skipped on EscrowApplyConfirmed.
		EscrowApproval(
			blockTimestamp int64,
			txBody *model.ApprovalEscrowTransactionBody,
		) error
	}
)

// SkipMempoolTransaction to filter out current Approval escrow transaction when
// this tx already expired based on new block height
func (tx *ApprovalEscrowTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	var (
		err = tx.checkEscrowValidity(false, newBlockHeight)
	)
	if err != nil {
		return true, err
	}
	return false, nil
}

// GetSize of approval transaction body bytes
func (*ApprovalEscrowTransaction) GetSize() uint32 {
	return constant.EscrowApprovalBytesLength
}

func (tx *ApprovalEscrowTransaction) GetMinimumFee() (int64, error) {
	if tx.Escrow.ApproverAddress != "" {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

// GetAmount return Amount from TransactionBody
func (tx *ApprovalEscrowTransaction) GetAmount() int64 {
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
Validate is func that for validating to Transaction type.
Check transaction fields, spendable balance and more
*/
func (tx *ApprovalEscrowTransaction) Validate(dbTx bool) error {
	var (
		err    error
		enough bool
	)
	err = tx.checkEscrowValidity(dbTx, tx.Height)
	if err != nil {
		return err
	}
	// check existing account & balance

	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotEnough")
	}

	return nil
}

func (tx *ApprovalEscrowTransaction) checkEscrowValidity(dbTx bool, blockHeight uint32) error {
	var (
		latestEscrow        model.Escrow
		row                 *sql.Row
		err                 error
		escrowQ, escrowArgs = tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Body.GetTransactionID())
	)
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
	if blockHeight >= latestEscrow.GetBlockHeight()+uint32(latestEscrow.Timeout) {
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowTimeout")
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
	return nil
}

/*
ApplyUnconfirmed exec before Confirmed
*/
func (tx *ApprovalEscrowTransaction) ApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -tx.Fee)
}

/*
UndoApplyUnconfirmed func exec before confirmed
*/
func (tx *ApprovalEscrowTransaction) UndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
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
	transaction.Height = tx.Height
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
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-tx.Fee,
		model.EventType_EventApprovalEscrowTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
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
