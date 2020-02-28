package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	PendingSignatureServiceInterface interface {
		GetPendingSignatureByTransactionHash(txHash []byte) ([]*model.PendingSignature, error)
		AddPendingSignature(signature *model.PendingSignature, dbTx bool) error
	}

	PendingSignatureService struct {
		PendingSignatureQuery query.PendingSignatureQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewPendingSignatureService(
	pendingSignatureQuery query.PendingSignatureQueryInterface,
	queryExecutor query.ExecutorInterface,
) *PendingSignatureService {
	return &PendingSignatureService{
		PendingSignatureQuery: pendingSignatureQuery,
		QueryExecutor:         queryExecutor,
	}
}

func (pss *PendingSignatureService) GetPendingSignatureByTransactionHash(txHash []byte) ([]*model.PendingSignature, error) {
	var (
		result []*model.PendingSignature
	)
	q, args := pss.PendingSignatureQuery.GetPendingSignatureByHash(txHash)
	rows, err := pss.QueryExecutor.ExecuteSelect(q, false, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err = pss.PendingSignatureQuery.BuildModel(result, rows)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (pss *PendingSignatureService) AddPendingSignature(signature *model.PendingSignature, dbTx bool) error {
	var (
		err error
	)
	q, args := pss.PendingSignatureQuery.InsertPendingSignature(signature)
	if dbTx {

	}
	if dbTx {
		err = pss.QueryExecutor.ExecuteTransaction(q, args...)
	} else {
		_, err = pss.QueryExecutor.ExecuteStatement(q, args...)
	}
	if err != nil {
		return err
	}
	return nil
}
