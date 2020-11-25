package service

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	PendingTransactionServiceInterface interface {
		ExpiringPendingTransactions(blockHeight uint32, useTX bool) error
	}

	PendingTransactionService struct {
		Log                     *logrus.Logger
		QueryExecutor           query.ExecutorInterface
		TypeActionSwitcher      transaction.TypeActionSwitcher
		TransactionUtil         transaction.UtilInterface
		TransactionQuery        query.TransactionQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
	}
)

func NewPendingTransactionService(
	log *logrus.Logger,
	queryExecutor query.ExecutorInterface,
	typeActionSwitcher transaction.TypeActionSwitcher,
	transactionUtil transaction.UtilInterface,
	transactionQuery query.TransactionQueryInterface,
	pendingTransactionQuery query.PendingTransactionQueryInterface,
) PendingTransactionServiceInterface {
	return &PendingTransactionService{
		Log:                     log,
		QueryExecutor:           queryExecutor,
		TypeActionSwitcher:      typeActionSwitcher,
		TransactionUtil:         transactionUtil,
		TransactionQuery:        transactionQuery,
		PendingTransactionQuery: pendingTransactionQuery,
	}
}

// ExpiringPendingTransactions will set status to be expired caused by current block height
func (tg *PendingTransactionService) ExpiringPendingTransactions(blockHeight uint32, useTX bool) error {
	var (
		pendingTransactions []*model.PendingTransaction
		innerTransaction    *model.Transaction
		typeAction          transaction.TypeAction
		rows                *sql.Rows
		err                 error
	)

	err = func() error {
		qy, qArgs := tg.PendingTransactionQuery.GetPendingTransactionsExpireByHeight(blockHeight)
		rows, err = tg.QueryExecutor.ExecuteSelect(qy, useTX, qArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		pendingTransactions, err = tg.PendingTransactionQuery.BuildModel(pendingTransactions, rows)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	if len(pendingTransactions) > 0 {
		if !useTX {
			err = tg.QueryExecutor.BeginTx()
			if err != nil {
				return err
			}
		}
		for _, pendingTransaction := range pendingTransactions {

			/**
			SET PendingTransaction
			1. block height = current block height
			2. status = expired
			*/
			nPendingTransaction := pendingTransaction
			nPendingTransaction.BlockHeight = blockHeight
			nPendingTransaction.Status = model.PendingTransactionStatus_PendingTransactionExpired
			q := tg.PendingTransactionQuery.InsertPendingTransaction(nPendingTransaction)
			err = tg.QueryExecutor.ExecuteTransactions(q)
			if err != nil {
				break
			}
			// Do UndoApplyConfirmed
			innerTransaction, err = tg.TransactionUtil.ParseTransactionBytes(nPendingTransaction.GetTransactionBytes(), false)
			if err != nil {
				break
			}
			typeAction, err = tg.TypeActionSwitcher.GetTransactionType(innerTransaction)
			if err != nil {
				break
			}
			err = typeAction.UndoApplyUnconfirmed()
			if err != nil {
				break
			}
		}

		if !useTX {
			/*
				Check the latest error is not nil, otherwise need to aborting the whole query transactions safety with rollBack.
				And automatically unlock mutex
			*/
			if err != nil {
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}
			err = tg.QueryExecutor.CommitTx()
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}
