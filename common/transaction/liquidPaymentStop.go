package transaction

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// LiquidPaymentStopTransaction is Transaction Type that implemented TypeAction
	LiquidPaymentStopTransaction struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		RecipientAddress              []byte
		Height                        uint32
		Body                          *model.LiquidPaymentStopTransactionBody
		Escrow                        *model.Escrow
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		NormalFee                     fee.FeeModelInterface
		EscrowFee                     fee.FeeModelInterface
		TypeActionSwitcher            TypeActionSwitcher
		EscrowQuery                   query.EscrowTransactionQueryInterface
	}
)

func (tx *LiquidPaymentStopTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		row           *sql.Row
		err           error
		liquidPayment model.LiquidPayment
		transaction   model.Transaction
		txType        TypeAction
	)

	// update sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-tx.Fee,
		model.EventType_EventLiquidPaymentStopTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	// processing the liquid payment transaction
	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(
		tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(liquidPaymentQ, true, liquidPaymentArgs...)
	if err != nil {
		return err
	}
	err = tx.LiquidPaymentTransactionQuery.Scan(&liquidPayment, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentNotExists")
		}
		return err
	}

	// handle multiple stop transaction
	if liquidPayment.Status == model.LiquidPaymentStatus_LiquidPaymentCompleted {
		return nil
	}

	// get what transaction type it is, and switch to specific approval
	transactionQ := tx.TransactionQuery.GetTransaction(tx.Body.GetTransactionID())
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

	txType, err = tx.TypeActionSwitcher.GetTransactionType(&transaction)
	if err != nil {
		return err
	}
	liquidPaymentTransaction, ok := txType.(LiquidPaymentTransactionInterface)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "Wrong type of transaction")
	}
	err = liquidPaymentTransaction.CompletePayment(tx.Height, blockTimestamp, liquidPayment.AppliedTime)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) ApplyUnconfirmed(applyInCache bool) error {
	if applyInCache {
		return tx.AccountBalanceHelper.AddAccountSpendableBalanceInCache(tx.SenderAddress, -(tx.Fee))
	}
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Fee))
}

func (tx *LiquidPaymentStopTransaction) UndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
	if err != nil {
		return err
	}
	// update existing spendable balance in cache storage
	return tx.AccountBalanceHelper.UpdateAccountSpendableBalanceInCache(tx.SenderAddress, tx.Fee)
}

func (tx *LiquidPaymentStopTransaction) Validate(dbTx bool) error {
	var (
		row           *sql.Row
		err           error
		liquidPayment model.LiquidPayment
		enough        bool
	)
	if tx.SenderAddress == nil {
		return errors.New("transaction must have a valid sender account id")
	}

	if tx.Body.TransactionID == 0 {
		return errors.New("transaction must have a valid transaction id")
	}

	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(
		tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(liquidPaymentQ, dbTx, liquidPaymentArgs...)
	if err != nil {
		return err
	}
	err = tx.LiquidPaymentTransactionQuery.Scan(&liquidPayment, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentNotExists")
		}
		return err
	}

	if !bytes.Equal(liquidPayment.SenderAddress, tx.SenderAddress) && !bytes.Equal(liquidPayment.RecipientAddress, tx.SenderAddress) {
		return blocker.NewBlocker(blocker.ValidationErr, "Only sender or recipient of the payment can stop the payment")
	}

	if liquidPayment.Status == model.LiquidPaymentStatus_LiquidPaymentCompleted {
		return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentHasPreviouslyCompleted")
	}

	// check existing & balance account sender
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

func (tx *LiquidPaymentStopTransaction) GetMinimumFee() (int64, error) {
	if tx.Escrow != nil && tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

func (tx *LiquidPaymentStopTransaction) GetAmount() int64 {
	return tx.Fee
}

func (tx *LiquidPaymentStopTransaction) GetSize() (uint32, error) {
	// only TransactionID
	return constant.TransactionID, nil
}

func (tx *LiquidPaymentStopTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	txSize, err := tx.GetSize()
	if err != nil {
		return nil, err
	}
	_, err = util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(txSize))
	if err != nil {
		return nil, err
	}
	// read body bytes
	bufferBytes := bytes.NewBuffer(txBodyBytes)
	txID := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.Balance)))
	return &model.LiquidPaymentStopTransactionBody{
		TransactionID: int64(txID),
	}, nil
}

func (tx *LiquidPaymentStopTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.TransactionID)))
	return buffer.Bytes(), nil
}

func (tx *LiquidPaymentStopTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_LiquidPaymentStopTransactionBody{
		LiquidPaymentStopTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction filter out of the mempool tx under specific condition
func (tx *LiquidPaymentStopTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *LiquidPaymentStopTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
			Instruction:     tx.Escrow.GetInstruction(),
		}
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *LiquidPaymentStopTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee + tx.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) EscrowApplyUnconfirmed(applyInCache bool) error {
	var addedSpendable = -(tx.Fee + tx.Escrow.GetCommission())
	if applyInCache {
		return tx.AccountBalanceHelper.AddAccountSpendableBalanceInCache(tx.SenderAddress, addedSpendable)
	}
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, addedSpendable)

}

func (tx *LiquidPaymentStopTransaction) EscrowUndoApplyUnconfirmed() error {
	var (
		addedSpendable = tx.Fee + tx.Escrow.GetCommission()
		err            = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, addedSpendable)
	)
	if err != nil {
		return err
	}
	// update existing spendable balance in cache storage
	return tx.AccountBalanceHelper.UpdateAccountSpendableBalanceInCache(tx.SenderAddress, addedSpendable)
}

func (tx *LiquidPaymentStopTransaction) EscrowValidate(dbTx bool) (err error) {
	var enough bool

	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionRequired")
	}
	if tx.Escrow.GetTimeout() > uint64(constant.MinRollbackBlocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutLimitExceeded")
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
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

func (tx *LiquidPaymentStopTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
		err = tx.ApplyConfirmed(blockTimestamp)
		if err != nil {
			return err
		}
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}
