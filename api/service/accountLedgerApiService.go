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
	// AccountLedgerServiceInterface interface that has account ledger api service methods collection
	AccountLedgerServiceInterface interface {
		GetAccountLedgers(request *model.GetAccountLedgersRequest) (*model.GetAccountLedgersResponse, error)
	}
	// AccountLedgerService struct fields of AccountLedgerService
	AccountLedgerService struct {
		Query query.ExecutorInterface
	}
)

// NewAccountLedgerService create instance of AccountLedgerService
func NewAccountLedgerService(executorInterface query.ExecutorInterface) *AccountLedgerService {
	return &AccountLedgerService{
		Query: executorInterface,
	}
}

// GetAccountLedgers method of account ledger service that api purpose
func (al *AccountLedgerService) GetAccountLedgers(
	request *model.GetAccountLedgersRequest,
) (*model.GetAccountLedgersResponse, error) {
	var (
		response    *model.GetAccountLedgersResponse
		ledgers     []*model.AccountLedger
		args        []interface{}
		rows        *sql.Rows
		row         *sql.Row
		err         error
		count       uint64
		selectQuery string
	)

	ledgerQuery := query.NewAccountLedgerQuery()
	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(ledgerQuery.TableName, ledgerQuery.Fields...)
	if request.GetAccountAddress() != nil {
		caseQuery.Where(caseQuery.Equal("account_address", request.GetAccountAddress()))
	}

	if request.GetTransactionID() > 0 {
		caseQuery.Where(caseQuery.Equal("transaction_id", request.GetTransactionID()))
	}
	if request.GetTimestampEnd() > 0 {
		caseQuery.Where(caseQuery.Between("timestamp", request.GetTimestampStart(), request.GetTimestampEnd()))
	}
	if request.GetEventType() != model.EventType_EventAny {
		caseQuery.Where(caseQuery.Equal("event_type", request.GetEventType()))
	}

	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)

	row, err = al.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = row.Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	page := request.GetPagination()
	if page.GetOrderField() != "" {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())

	selectQuery, args = caseQuery.Build()
	rows, err = al.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	ledgers, err = ledgerQuery.BuildModel(ledgers, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response = &model.GetAccountLedgersResponse{
		Total:          count,
		AccountLedgers: ledgers,
	}
	return response, nil
}
