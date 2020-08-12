package transaction

import (
	"bytes"
	"errors"
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// LiquidPaymentTransaction is Transaction Type that implemented TypeAction
	LiquidPaymentTransaction struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 string
		RecipientAddress              string
		Height                        uint32
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		AccountLedgerHelper           AccountLedgerHelperInterface
		NormalFee                     fee.FeeModelInterface
	}
	LiquidPaymentTransactionInterface interface {
		CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error
	}
)

func (tx *LiquidPaymentTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		queries [][]interface{}
		err     error
	)

	// update sender
	err = tx.AccountBalanceHelper.AddAccountBalance(tx.SenderAddress,
		-(tx.Body.Amount + tx.Fee), tx.Height)

	if err != nil {
		return err
	}

	// sender ledger
	err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.GetAmount() + tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventLiquidPaymentTransaction,
		Timestamp:      uint64(blockTimestamp),
	})

	if err != nil {
		return err
	}

	// create the Liquid payment record
	liquidPaymentTransaction := &model.LiquidPayment{
		ID:               tx.ID,
		SenderAddress:    tx.SenderAddress,
		RecipientAddress: tx.RecipientAddress,
		Amount:           tx.Body.GetAmount(),
		AppliedTime:      blockTimestamp,
		CompleteMinutes:  tx.Body.GetCompleteMinutes(),
		Status:           model.LiquidPaymentStatus_LiquidPaymentPending,
		BlockHeight:      tx.Height,
	}
	liquidPaymentTransactionQ := tx.LiquidPaymentTransactionQuery.InsertLiquidPaymentTransaction(liquidPaymentTransaction)
	queries = append(queries, liquidPaymentTransactionQ...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)

	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) ApplyUnconfirmed() error {
	// update sender
	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Body.Amount + tx.Fee))
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) UndoApplyUnconfirmed() error {
	// update sender
	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Body.Amount+tx.Fee)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		err            error
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.SenderAddress == "" {
		return errors.New("transaction must have a valid sender account id")
	}
	if tx.RecipientAddress == "" {
		return errors.New("transaction must have a valid recipient account id")
	}

	// check existing & balance account sender
	err = tx.AccountBalanceHelper.GetBalanceByAccountID(&accountBalance, tx.SenderAddress, dbTx)
	if err != nil {
		return err
	}

	if accountBalance.SpendableBalance < (tx.Body.GetAmount() + tx.Fee) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"UserBalanceNotEnough",
		)
	}
	return nil
}

func (tx *LiquidPaymentTransaction) GetMinimumFee() (int64, error) {
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, nil)
}

func (tx *LiquidPaymentTransaction) GetAmount() int64 {
	return tx.Body.Amount
}

func (tx *LiquidPaymentTransaction) GetSize() uint32 {
	// only amount
	return constant.Balance + constant.LiquidPaymentCompleteMinutesLength
}

func (tx *LiquidPaymentTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	_, err := util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(tx.GetSize()))
	if err != nil {
		return nil, err
	}
	// read body bytes
	bufferBytes := bytes.NewBuffer(txBodyBytes)
	amount := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.Balance)))
	completeMinutes := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.LiquidPaymentCompleteMinutesLength)))
	return &model.LiquidPaymentTransactionBody{
		Amount:          int64(amount),
		CompleteMinutes: completeMinutes,
	}, nil
}

func (tx *LiquidPaymentTransaction) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	buffer.Write(util.ConvertUint64ToBytes(tx.Body.CompleteMinutes))
	return buffer.Bytes()
}

func (tx *LiquidPaymentTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_LiquidPaymentTransactionBody{
		LiquidPaymentTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction filter out of the mempool tx under specific condition
func (tx *LiquidPaymentTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *LiquidPaymentTransaction) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}

func (tx *LiquidPaymentTransaction) CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error {
	var (
		queries                                           [][]interface{}
		err                                               error
		recipientBalanceIncrement, senderBalanceIncrement int64
		blockTimestampTime                                = time.Unix(blockTimestamp, 0)
		firstAppliedTimestampTime                         = time.Unix(firstAppliedTimestamp, 0)
		durationPassed                                    = blockTimestampTime.Sub(firstAppliedTimestampTime).Minutes()
	)

	if durationPassed < 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "blockTimestamp is less than firstAppliedTimestamp")
	}

	durationRate := durationPassed / float64(tx.Body.GetCompleteMinutes())
	if durationRate > 1 {
		recipientBalanceIncrement = tx.Body.GetAmount()
	} else {
		recipientBalanceIncrement = int64(math.Ceil(durationRate * float64(tx.Body.GetAmount())))
		senderBalanceIncrement = tx.Body.GetAmount() - recipientBalanceIncrement
	}

	// transfer the money to the recipient pro-rate wise
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.RecipientAddress, recipientBalanceIncrement, blockHeight)
	if err != nil {
		return err
	}

	// recipient ledger
	err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
		AccountAddress: tx.RecipientAddress,
		BalanceChange:  recipientBalanceIncrement,
		TransactionID:  tx.ID,
		BlockHeight:    blockHeight,
		EventType:      model.EventType_EventLiquidPaymentPaidTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	if err != nil {
		return err
	}

	if senderBalanceIncrement > 0 {
		// returning the remaining payment to the sender
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress, senderBalanceIncrement, blockHeight)
		if err != nil {
			return err
		}

		// sender ledger
		err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
			AccountAddress: tx.SenderAddress,
			BalanceChange:  senderBalanceIncrement,
			TransactionID:  tx.ID,
			BlockHeight:    blockHeight,
			EventType:      model.EventType_EventLiquidPaymentPaidTransaction,
			Timestamp:      uint64(blockTimestamp),
		})
		if err != nil {
			return err
		}
	}

	// update the status of the liquid payment
	liquidPaymentStatusUpdateQ := tx.LiquidPaymentTransactionQuery.CompleteLiquidPaymentTransaction(tx.ID,
		map[string]interface{}{"block_height": blockHeight})
	queries = append(queries, liquidPaymentStatusUpdateQ...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)

	if err != nil {
		return err
	}
	return nil
}
