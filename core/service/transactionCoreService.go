package service

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	TransactionCoreServiceInterface interface {
		GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error)
		GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error)
		ValidateTransaction(txAction transaction.TypeAction, useTX bool) error
		ApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		UndoApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		ApplyConfirmedTransaction(txAction transaction.TypeAction, blockTimestamp int64) error
		ExpiringEscrowTransactions(blockHeight uint32, useTX bool) error
		ExpiringPendingTransactions(blockHeight uint32, useTX bool) error
	}

	TransactionCoreService struct {
		Log                     *logrus.Logger
		QueryExecutor           query.ExecutorInterface
		TypeActionSwitcher      transaction.TypeActionSwitcher
		TransactionUtil         transaction.UtilInterface
		TransactionQuery        query.TransactionQueryInterface
		EscrowTransactionQuery  query.EscrowTransactionQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
	}
)

func NewTransactionCoreService(
	log *logrus.Logger,
	queryExecutor query.ExecutorInterface,
	typeActionSwitcher transaction.TypeActionSwitcher,
	transactionUtil transaction.UtilInterface,
	transactionQuery query.TransactionQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	pendingTransactionQuery query.PendingTransactionQueryInterface,
) TransactionCoreServiceInterface {
	return &TransactionCoreService{
		Log:                     log,
		QueryExecutor:           queryExecutor,
		TypeActionSwitcher:      typeActionSwitcher,
		TransactionUtil:         transactionUtil,
		TransactionQuery:        transactionQuery,
		EscrowTransactionQuery:  escrowTransactionQuery,
		PendingTransactionQuery: pendingTransactionQuery,
	}
}

func (tg *TransactionCoreService) GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error) {
	var (
		rows *sql.Rows
		err  error
	)
	txQuery, _ := tg.TransactionQuery.GetTransactionsByIds(transactionIds)
	rows, err = tg.QueryExecutor.ExecuteSelect(txQuery, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	transactions, err = tg.TransactionQuery.BuildModel(transactions, rows)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsByBlockID get transactions of the block
func (tg *TransactionCoreService) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction

	// get transaction of the block
	transactionQ, transactionArg := tg.TransactionQuery.GetTransactionsByBlockID(blockID)
	rows, err := tg.QueryExecutor.ExecuteSelect(transactionQ, false, transactionArg...)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	return tg.TransactionQuery.BuildModel(transactions, rows)
}

// ExpiringEscrowTransactions push an observer event that is ExpiringEscrowTransactions,
// will set status to be expired caused by current block height
// query lock from outside (PushBlock)
func (tg *TransactionCoreService) ExpiringEscrowTransactions(blockHeight uint32, useTX bool) error {
	var (
		escrows []*model.Escrow
		rows    *sql.Rows
		err     error
	)

	err = func() error {
		escrowQ, escrowArgs := tg.EscrowTransactionQuery.GetEscrowTransactions(map[string]interface{}{
			"timeout": blockHeight,
			"status":  model.EscrowStatus_Pending,
			"latest":  1,
		})
		rows, err = tg.QueryExecutor.ExecuteSelect(escrowQ, useTX, escrowArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		escrows, err = tg.EscrowTransactionQuery.BuildModels(rows)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	if len(escrows) > 0 {
		if !useTX {
			err = tg.QueryExecutor.BeginTx()
			if err != nil {
				return err
			}
		}
		for _, escrow := range escrows {
			/**
			SET Escrow
			1. block height = current block height
			2. status = expired
			*/
			nEscrow := escrow
			nEscrow.BlockHeight = blockHeight
			nEscrow.Status = model.EscrowStatus_Expired
			q := tg.EscrowTransactionQuery.InsertEscrowTransaction(escrow)
			err = tg.QueryExecutor.ExecuteTransactions(q)
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
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}
		}
	}
	return nil
}

// ExpiringPendingTransactions will set status to be expired caused by current block height
func (tg *TransactionCoreService) ExpiringPendingTransactions(blockHeight uint32, useTX bool) error {
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
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}
		}
		return err
	}
	return nil
}

func (tg *TransactionCoreService) ValidateTransaction(txAction transaction.TypeAction, useTX bool) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		return escrowAction.EscrowValidate(useTX)
	default:
		return txAction.Validate(useTX)
	}
}

func (tg *TransactionCoreService) ApplyUnconfirmedTransaction(txAction transaction.TypeAction) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowApplyUnconfirmed()
		return err
	default:
		err := txAction.ApplyUnconfirmed()
		return err
	}
}

func (tg *TransactionCoreService) UndoApplyUnconfirmedTransaction(txAction transaction.TypeAction) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowUndoApplyUnconfirmed()
		return err
	default:
		err := txAction.UndoApplyUnconfirmed()
		return err
	}
}

func (tg *TransactionCoreService) ApplyConfirmedTransaction(
	txAction transaction.TypeAction,
	blockTimestamp int64,
) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowApplyConfirmed(blockTimestamp)
		return err
	default:
		err := txAction.ApplyConfirmed(blockTimestamp)
		return err
	}
}
