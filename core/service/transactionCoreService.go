package service

import (
	"database/sql"

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
		ExpiringEscrowTransactions(blockHeight uint32) error
	}

	TransactionCoreService struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
)

func NewTransactionCoreService(
	queryExecutor query.ExecutorInterface,
	transactionQuery query.TransactionQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
) TransactionCoreServiceInterface {
	return &TransactionCoreService{
		TransactionQuery:       transactionQuery,
		EscrowTransactionQuery: escrowTransactionQuery,
		QueryExecutor:          queryExecutor,
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
func (tg *TransactionCoreService) ExpiringEscrowTransactions(blockHeight uint32) error {
	var (
		escrows []*model.Escrow
		rows    *sql.Rows
		err     error
	)

	escrowQ, escrowArgs := tg.EscrowTransactionQuery.GetEscrowTransactions(map[string]interface{}{
		"timeout": blockHeight,
		"status":  model.EscrowStatus_Pending,
		"latest":  1,
	})
	rows, err = tg.QueryExecutor.ExecuteSelect(escrowQ, false, escrowArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	escrows, err = tg.EscrowTransactionQuery.BuildModels(rows)
	if err != nil {
		return err
	}
	if len(escrows) > 0 {
		err = tg.QueryExecutor.BeginTx()
		if err != nil {
			return err
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
				return err
			}
		}

		err = tg.QueryExecutor.CommitTx()
		if err != nil {
			if errRollback := tg.QueryExecutor.RollbackTx(); errRollback != nil {
				return err
			}
		}
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
