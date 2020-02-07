package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TransactionCoreServiceInterface interface {
		GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error)
	}

	TransactionCoreService struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
)

func NewTransactionCoreService(transactionQuery query.TransactionQueryInterface,
	queryExecutor query.ExecutorInterface) TransactionCoreServiceInterface {
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
