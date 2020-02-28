package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	PendingTransactionServiceInterface interface {
		GetPendingTransactionByHash(txHash []byte) (*model.PendingTransaction, error)
		AddPendingTransaction(tx *model.PendingTransaction, dbTx bool) error
	}

	PendingTransactionService struct {
		PendingTransactionQuery query.PendingTransactionQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
)

func NewPendingTransactionService(
	pendingTransactionQuery query.PendingTransactionQueryInterface,
	queryExecutor query.ExecutorInterface,
) *PendingTransactionService {
	return &PendingTransactionService{
		PendingTransactionQuery: pendingTransactionQuery,
		QueryExecutor:           queryExecutor,
	}
}

func (*PendingTransactionService) GetPendingTransactionByHash(txHash []byte) (*model.PendingTransaction, error) {
	return nil, nil
}

func (pts *PendingTransactionService) AddPendingTransaction(tx *model.PendingTransaction, dbTx bool) error {
	var (
		err error
	)
	q, args := pts.PendingTransactionQuery.InsertPendingTransaction(tx)
	if dbTx {

	}
	if dbTx {
		err = pts.QueryExecutor.ExecuteTransaction(q, args...)
	} else {
		_, err = pts.QueryExecutor.ExecuteStatement(q, args...)
	}
	if err != nil {
		return err
	}
	return nil
}
