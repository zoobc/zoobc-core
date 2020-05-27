package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	TransactionHelperInterface interface {
		InsertTransaction(transaction *model.Transaction) error
	}
	TransactionHelper struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
)

func NewTransactionHelper(
	transactionQuery query.TransactionQueryInterface,
	queryExecutor query.ExecutorInterface,
) *TransactionHelper {
	return &TransactionHelper{
		TransactionQuery: transactionQuery,
		QueryExecutor:    queryExecutor,
	}
}

func (th *TransactionHelper) InsertTransaction(transaction *model.Transaction) error {
	insertTxQ, args := th.TransactionQuery.InsertTransaction(transaction)
	err := th.QueryExecutor.ExecuteTransaction(insertTxQ, args...)
	if err != nil {
		return err
	}
	return nil
}
