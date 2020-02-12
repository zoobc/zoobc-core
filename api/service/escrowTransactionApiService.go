package service

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// EscrowTransactionServiceInterface interface that contain methods of escrow transaction
	EscrowTransactionServiceInterface interface {
		GetEscrowTransactions(request *model.GetEscrowTransactionsRequest) (*model.GetEscrowTransactionsResponse, error)
	}
	// EscrowTransactionService struct that contain fields that needed
	escrowTransactionService struct {
		Query query.ExecutorInterface
	}
)

// NewEscrowTransactionService will create EscrowTransactionServiceInterface instance
func NewEscrowTransactionService(
	query query.ExecutorInterface,
) EscrowTransactionServiceInterface {
	return &escrowTransactionService{
		Query: query,
	}
}

// GetEscrowTransactions to get escrow transactions list
func (es *escrowTransactionService) GetEscrowTransactions(
	params *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	var (
		escrowQuery = query.NewEscrowTransactionQuery()
		countQuery  string
		escrows     []*model.Escrow
		rows        *sql.Rows
		count       int64
		row         *sql.Row
		err         error
	)

	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(escrowQuery.TableName, escrowQuery.Fields...)
	if params.GetApproverAddress() != "" {
		caseQuery.Where(caseQuery.Equal("approver_address", params.GetApproverAddress()))
	}
	if len(params.GetStatuses()) > 0 {
		caseQuery.And(caseQuery.In("status", params.GetStatuses()))
	}

	blockHeightStart := params.GetBlockHeightStart()
	blockHeightEnd := params.GetBlockHeightEnd()
	if blockHeightStart > 0 {
		caseQuery.Where(caseQuery.Between("block_height", blockHeightStart, blockHeightEnd))
	}

	// count first
	selectQuery, args := caseQuery.Build()
	countQuery = query.GetTotalRecordOfSelect(selectQuery)

	row, err = es.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = row.Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// select records
	page := params.GetPagination()
	if page.GetOrderField() == "" {
		caseQuery.OrderBy("id", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())

	escrowQ, escrowArgs := caseQuery.Build()
	rows, err = es.Query.ExecuteSelect(escrowQ, false, escrowArgs...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	escrows, err = escrowQuery.BuildModels(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetEscrowTransactionsResponse{
		Total:   uint64(count),
		Escrows: escrows,
	}, nil
}
