package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	MultisigInfoServiceInterface interface {
		AddMultisigInfo(info *model.MultiSignatureInfo, dbTx bool) error
		GetMultisigInfoByAddress(address string) (*model.MultiSignatureInfo, error)
	}
	MultisigInfoService struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
)

func NewMultisigInfoService(
	multisignatureInfoQuery query.MultisignatureInfoQueryInterface,
	queryExecutor query.ExecutorInterface,
) *MultisigInfoService {
	return &MultisigInfoService{
		MultisignatureInfoQuery: multisignatureInfoQuery,
		QueryExecutor:           queryExecutor,
	}
}

func (mss *MultisigInfoService) AddMultisigInfo(info *model.MultiSignatureInfo, dbTx bool) error {
	var err error
	q, args := mss.MultisignatureInfoQuery.InsertMultisignatureInfo(info)
	if dbTx {
		err = mss.QueryExecutor.ExecuteTransaction(q, args...)
	} else {
		_, err = mss.QueryExecutor.ExecuteStatement(q, args...)
	}
	if err != nil {
		return err
	}
	return nil
}

func (mss *MultisigInfoService) GetMultisigInfoByAddress(address string) (*model.MultiSignatureInfo, error) {
	var (
		err    error
		result *model.MultiSignatureInfo
	)
	q, args := mss.MultisignatureInfoQuery.GetMultisignatureInfoByAddress(address)
	row, _ := mss.QueryExecutor.ExecuteSelectRow(q, false, args...)
	err = mss.MultisignatureInfoQuery.Scan(result, row)
	if err != nil {
		return nil, err
	}
	return result, nil
}
