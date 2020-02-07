package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	TransactionCoreServiceInterface interface {
		GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error)
		ValidateTransaction(txAction transaction.TypeAction, useTX bool) error
		ApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		UndoApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		ApplyConfirmedTransaction(txAction transaction.TypeAction, blockTimestamp int64) error
	}

	TransactionCoreService struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
)

func NewTransactionCoreService(
	transactionQuery query.TransactionQueryInterface,
	queryExecutor query.ExecutorInterface,
) TransactionCoreServiceInterface {
	return &TransactionCoreService{
		TransactionQuery: transactionQuery,
		QueryExecutor:    queryExecutor,
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

func (tg *TransactionCoreService) ApplyConfirmedTransaction(txAction transaction.TypeAction, blockTimestamp int64) error {

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
