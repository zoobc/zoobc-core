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
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		AccountLedgerHelper           AccountLedgerHelperInterface
		NormalFee                     fee.FeeModelInterface
		TypeActionSwitcher            TypeActionSwitcher
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
		tx.SenderAddress, -tx.Fee, tx.Height)
	if err != nil {
		return err
	}

	// sender ledger
	err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventLiquidPaymentStopTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	if err != nil {
		return err
	}

	// processing the liquid payment transaction
	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending)
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

func (tx *LiquidPaymentStopTransaction) ApplyUnconfirmed() error {
	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -tx.Fee)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) UndoApplyUnconfirmed() error {
	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) Validate(dbTx bool) error {
	var (
		row            *sql.Row
		err            error
		liquidPayment  model.LiquidPayment
		accountBalance model.AccountBalance
	)
	if tx.SenderAddress == nil {
		return errors.New("transaction must have a valid sender account id")
	}

	if tx.Body.TransactionID == 0 {
		return errors.New("transaction must have a valid transaction id")
	}

	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending)
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
	err = tx.AccountBalanceHelper.GetBalanceByAccountID(&accountBalance, tx.SenderAddress, dbTx)
	if err != nil {
		return err
	}

	if accountBalance.SpendableBalance < tx.Fee {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"UserBalanceNotEnough",
		)
	}

	return nil
}

func (tx *LiquidPaymentStopTransaction) GetMinimumFee() (int64, error) {
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, nil)
}

func (tx *LiquidPaymentStopTransaction) GetAmount() int64 {
	return tx.Fee
}

func (tx *LiquidPaymentStopTransaction) GetSize() uint32 {
	// only TransactionID
	return constant.TransactionID
}

func (tx *LiquidPaymentStopTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	_, err := util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(tx.GetSize()))
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

func (tx *LiquidPaymentStopTransaction) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.TransactionID)))
	return buffer.Bytes()
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
	return nil, false
}
