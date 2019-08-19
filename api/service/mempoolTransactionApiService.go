package service

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	MempoolTransactionServiceInterface interface {
		GetMempoolTransaction(
			chainType contract.ChainType,
			params *model.GetMempoolTransactionRequest,
		) (*model.GetMempoolTransactionResponse, error)
		GetMempoolTransactions(
			chainType contract.ChainType,
			params *model.GetMempoolTransactionsRequest,
		) (*model.GetMempoolTransactionsResponse, error)
	}
	MempoolTransactionService struct {
		Query query.ExecutorInterface
	}
)

func NewMempoolTransactionsService(
	queryExecutor query.ExecutorInterface,
) *MempoolTransactionService {
	return &MempoolTransactionService{
		Query: queryExecutor,
	}
}

func (ut *MempoolTransactionService) GetMempoolTransaction(
	chainType contract.ChainType,
	params *model.GetMempoolTransactionRequest,
) (*model.GetMempoolTransactionResponse, error) {
	var (
		err error
		row *sql.Row
		tx  model.MempoolTransaction
	)

	txQuery := query.NewMempoolQuery(chainType)
	row = ut.Query.ExecuteSelectRow(txQuery.GetMempoolTransaction(), params.GetID())
	if row == nil {
		return nil, err
	}

	err = txQuery.Scan(&tx, row)
	if err != nil {
		return nil, err
	}

	if len(tx.GetTransactionBytes()) == 0 {
		return nil, errors.New("record not found")
	}
	return &model.GetMempoolTransactionResponse{
		Transaction: &tx,
	}, nil
}

func (ut *MempoolTransactionService) GetMempoolTransactions(
	chainType contract.ChainType,
	params *model.GetMempoolTransactionsRequest,
) (*model.GetMempoolTransactionsResponse, error) {
	var (
		err                     error
		count                   uint64
		selectQuery, countQuery string
		rows                    *sql.Rows
		txs                     []*model.MempoolTransaction
		response                *model.GetMempoolTransactionsResponse
		args                    []interface{}
	)

	txQuery := query.NewMempoolQuery(chainType)
	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	timestampStart := params.GetTimestampStart()
	timestampEnd := params.GetTimestampEnd()
	if timestampStart > 0 {
		caseQuery.Where(caseQuery.Between("arrival_timestamp", timestampStart, timestampEnd))
	}

	address := params.GetAddress()
	if address != "" {
		if timestampStart > 0 {
			caseQuery.And(caseQuery.Equal("sender_account_address", address)).
				Or(caseQuery.Equal("recipient_account_address", address))
		} else {
			caseQuery.Where(caseQuery.Equal("sender_account_address", address)).
				Or(caseQuery.Equal("recipient_account_address", address))
		}
	}

	// count first
	selectQuery, args = caseQuery.Build()
	countQuery = query.GetTotalRecordOfSelect(selectQuery)

	rows, err = ut.Query.ExecuteSelect(countQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return response, err
		}
	}

	// select records
	caseQuery.Paginate(params.GetLimit(), params.GetPage())
	selectQuery, args = caseQuery.Build()

	rows, err = ut.Query.ExecuteSelect(selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs = txQuery.BuildModel(txs, rows)

	response = &model.GetMempoolTransactionsResponse{
		MempoolTransactions: txs,
		Total:               count,
	}
	return response, nil
}
