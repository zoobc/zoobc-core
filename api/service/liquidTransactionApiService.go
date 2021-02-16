package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	LiquidTransactionServiceInterface interface {
		GetLiquidTransactions(request *model.GetLiquidTransactionsRequest) (*model.GetLiquidTransactionsResponse, error)
	}

	LiquidTransactionService struct {
		LiquidPaymentTransactionQuery *query.LiquidPaymentTransactionQuery
		QueryExecutor                 query.ExecutorInterface
	}
)

func NewLiquidTransactionService(executor query.ExecutorInterface,
	liquidPaymentTransactionQuery *query.LiquidPaymentTransactionQuery) *LiquidTransactionService {
	return &LiquidTransactionService{
		LiquidPaymentTransactionQuery: liquidPaymentTransactionQuery,
		QueryExecutor:                 executor,
	}
}

func (lts *LiquidTransactionService) GetLiquidTransactions(
	request *model.GetLiquidTransactionsRequest,
) (*model.GetLiquidTransactionsResponse, error) {
	var (
		liquidTransactions []*model.LiquidPayment
		caseQ              = query.NewCaseQuery()
		rows               *sql.Rows
		rowCount           *sql.Row
		count              uint64
		err                error
	)

	caseQ.Select(lts.LiquidPaymentTransactionQuery.TableName, lts.LiquidPaymentTransactionQuery.Fields...)
	id := request.GetID()
	if id != 0 {
		caseQ.And(caseQ.Equal("id", id))
	}
	senderAddress := request.GetSenderAddress()
	if senderAddress != nil {
		caseQ.And(caseQ.Equal("sender_address", senderAddress))
	}
	recipientAddress := request.GetRecipientAddress()
	if recipientAddress != nil {
		caseQ.And(caseQ.Equal("secipient_address", recipientAddress))
	}
	lpStatus := request.GetStatus()
	if lpStatus != -1 {
		caseQ.And(caseQ.Equal("status", lpStatus))
	}

	latest := true
	caseQ.And(caseQ.Equal("latest", latest))

	// count first
	selectQuery, args := caseQ.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rowCount, err = lts.QueryExecutor.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = rowCount.Scan(
		&count,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// pagination
	page := request.GetPagination()
	if page.GetOrderField() != "" {
		caseQ.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}

	caseQ.Paginate(page.GetLimit(), page.GetPage())

	selectQ, args := caseQ.Build()
	rows, err = lts.QueryExecutor.ExecuteSelect(selectQ, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	liquidTransactions, err = lts.LiquidPaymentTransactionQuery.BuildModels(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetLiquidTransactionsResponse{
		Total:              count,
		LiquidTransactions: liquidTransactions,
	}, nil
}
