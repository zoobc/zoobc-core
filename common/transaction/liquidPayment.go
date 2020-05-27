package transaction

import (
	"bytes"
	"database/sql"
	"errors"
	"math"

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
		AccountBalanceQuery           query.AccountBalanceQueryInterface
		AccountLedgerQuery            query.AccountLedgerQueryInterface
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
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.GetAmount() + tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventLiquidPaymentTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

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
	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.Amount + tx.Fee),
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

func (tx *LiquidPaymentTransaction) UndoApplyUnconfirmed() error {
	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.Amount+tx.Fee,
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

func (tx *LiquidPaymentTransaction) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		row            *sql.Row
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

	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return err
	}

	if accountBalance.SpendableBalance < (tx.Body.GetAmount() + tx.Fee) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"balance not enough",
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
	return constant.Balance + constant.LiquidPaymentCompleteMinutes
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
	completeMinutes := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.LiquidPaymentCompleteMinutes)))
	return &model.LiquidPaymentTransactionBody{
		Amount:          int64(amount),
		CompleteMinutes: uint64(completeMinutes),
	}, nil
}

func (tx *LiquidPaymentTransaction) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.CompleteMinutes)))
	return buffer.Bytes()
}

func (tx *LiquidPaymentTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_LiquidPaymentTransactionBody{
		LiquidPaymentTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction filter out of the mempool tx under specific condition
func (tx *LiquidPaymentTransaction) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
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
	)

	if blockTimestamp-firstAppliedTimestamp < 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "blockTimestamp is less than firstAppliedTimestamp")
	}

	durationRate := (float64(blockTimestamp-firstAppliedTimestamp) / 60) / float64(tx.Body.GetCompleteMinutes())
	if durationRate > 1 {
		recipientBalanceIncrement = tx.Body.GetAmount()
	} else {
		recipientBalanceIncrement = int64(math.Ceil(durationRate * float64(tx.Body.GetAmount())))
		senderBalanceIncrement = tx.Body.GetAmount() - recipientBalanceIncrement
	}

	// transfer the money to the recipient pro-rate wise
	accountBalanceRecipientQ := tx.AccountBalanceQuery.AddAccountBalance(
		recipientBalanceIncrement,
		map[string]interface{}{
			"account_address": tx.RecipientAddress,
			"block_height":    blockHeight,
		},
	)
	queries = append(queries, accountBalanceRecipientQ...)

	// recipient ledger
	recipientAccountLedgerQ, recipientAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.RecipientAddress,
		BalanceChange:  recipientBalanceIncrement,
		TransactionID:  tx.ID,
		BlockHeight:    blockHeight,
		EventType:      model.EventType_EventLiquidPaymentPaidTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	recipientAccountLedgerArgs = append([]interface{}{recipientAccountLedgerQ}, recipientAccountLedgerArgs...)
	queries = append(queries, recipientAccountLedgerArgs)

	if senderBalanceIncrement > 0 {
		// returning the remaining payment to the sender
		accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
			senderBalanceIncrement,
			map[string]interface{}{
				"account_address": tx.SenderAddress,
				"block_height":    blockHeight,
			},
		)
		queries = append(queries, accountBalanceSenderQ...)

		// sender ledger
		senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.SenderAddress,
			BalanceChange:  senderBalanceIncrement,
			TransactionID:  tx.ID,
			BlockHeight:    blockHeight,
			EventType:      model.EventType_EventLiquidPaymentPaidTransaction,
			Timestamp:      uint64(blockTimestamp),
		})
		senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
		queries = append(queries, senderAccountLedgerArgs)
	}

	// update the status of the liquid payment
	liquidPaymentStatusUpdateQ := tx.LiquidPaymentTransactionQuery.CompleteLiquidPaymentTransaction(tx.ID, map[string]interface{}{"block_height": blockHeight})
	queries = append(queries, liquidPaymentStatusUpdateQ...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)

	if err != nil {
		return err
	}
	return nil
}
