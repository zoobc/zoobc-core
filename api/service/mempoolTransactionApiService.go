package service

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	MempoolTransactionServiceInterface interface {
		GetMempoolTransaction(
			chainType chaintype.ChainType,
			params *model.GetMempoolTransactionRequest,
		) (*model.GetMempoolTransactionResponse, error)
		GetMempoolTransactions(
			chainType chaintype.ChainType,
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
	chainType chaintype.ChainType,
	params *model.GetMempoolTransactionRequest,
) (*model.GetMempoolTransactionResponse, error) {
	var (
		err error
		row *sql.Row
		tx  model.MempoolTransaction
	)

	txQuery := query.NewMempoolQuery(chainType)
	row, _ = ut.Query.ExecuteSelectRow(txQuery.GetMempoolTransaction(), false, params.GetID())
	if row == nil {
		return nil, status.Error(codes.NotFound, "transaction not found in mempool")
	}

	err = txQuery.Scan(&tx, row)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(tx.GetTransactionBytes()) == 0 {
		return nil, status.Error(codes.NotFound, "tx byte is empty")
	}
	return &model.GetMempoolTransactionResponse{
		Transaction: &tx,
	}, nil
}

func (ut *MempoolTransactionService) GetMempoolTransactions(
	chainType chaintype.ChainType,
	params *model.GetMempoolTransactionsRequest,
) (*model.GetMempoolTransactionsResponse, error) {
	var (
		err                     error
		count                   uint64
		selectQuery, countQuery string
		rowCount                *sql.Row
		rows2                   *sql.Rows
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
	if address != nil {
		caseQuery.And(caseQuery.Equal("sender_account_address", address)).
			Or(caseQuery.Equal("recipient_account_address", address))

	}

	// count first
	selectQuery, args = caseQuery.Build()
	countQuery = query.GetTotalRecordOfSelect(selectQuery)

	rowCount, err = ut.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = rowCount.Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// select records
	page := params.GetPagination()
	if page.GetOrderField() == "" {
		caseQuery.OrderBy("arrival_timestamp", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())

	selectQuery, args = caseQuery.Build()
	rows2, err = ut.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows2.Close()

	txs, err = txQuery.BuildModel(txs, rows2)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response = &model.GetMempoolTransactionsResponse{
		MempoolTransactions: txs,
		Total:               count,
	}
	return response, nil
}
