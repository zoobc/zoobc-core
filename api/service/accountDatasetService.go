package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// AccountDatasetServiceInterface a methods collection for AccountDataset
	AccountDatasetServiceInterface interface {
		GetAccountDatasets(request *model.GetAccountDatasetsRequest) (*model.GetAccountDatasetsResponse, error)
		GetAccountDataset(request *model.GetAccountDatasetRequest) (*model.AccountDataset, error)
	}
	// AccountDatasetService contain fields that needed for AccountDatasetServiceInterface
	AccountDatasetService struct {
		AccountDatasetQuery *query.AccountDatasetsQuery
		QueryExecutor       query.ExecutorInterface
	}
)

func NewAccountDatasetService(
	accountDatasetQuery *query.AccountDatasetsQuery,
	queryExecutor query.ExecutorInterface,
) AccountDatasetServiceInterface {
	return &AccountDatasetService{
		AccountDatasetQuery: accountDatasetQuery,
		QueryExecutor:       queryExecutor,
	}
}

// GetAccountDatasets a method service that use for GetAccountDatasets Handler
func (ads *AccountDatasetService) GetAccountDatasets(
	request *model.GetAccountDatasetsRequest,
) (*model.GetAccountDatasetsResponse, error) {
	var (
		accDatasets []*model.AccountDataset
		rowCount    *sql.Row
		caseQ       = query.NewCaseQuery()
		count       uint64
		rows        *sql.Rows
		err         error
	)

	caseQ.Select(ads.AccountDatasetQuery.TableName, ads.AccountDatasetQuery.GetFields()...)
	if request.GetProperty() != "" {
		caseQ.Where(caseQ.Equal("property", request.GetProperty()))
	}
	if request.GetValue() != "" {
		caseQ.Where(caseQ.Equal("value", request.GetValue()))
	}
	if request.GetRecipientAccountAddress() != "" {
		caseQ.Where(caseQ.Equal("recipient_account_address", request.GetRecipientAccountAddress()))
	}
	if request.GetSetterAccountAddress() != "" {
		caseQ.Where(caseQ.Equal("setter_account_address", request.GetSetterAccountAddress()))
	}
	if request.GetHeight() > 0 {
		caseQ.Where(caseQ.Equal("height", request.GetHeight()))
	}
	caseQ.And(caseQ.Equal("latest", true))

	countQ, countArgs := caseQ.Build()
	rowCount, _ = ads.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(countQ), false, countArgs...)
	if err = rowCount.Scan(&count); err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "Something wrong happened")
		}

		return nil, status.Error(codes.NotFound, "Record not found")
	}

	pagination := request.GetPagination()
	if pagination == nil {
		pagination = &model.Pagination{
			OrderField: "height",
			OrderBy:    model.OrderBy_ASC,
			Page:       0,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if pagination.GetLimit() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	if pagination.GetOrderField() != "" {
		caseQ.OrderBy(pagination.GetOrderField(), pagination.GetOrderBy())
	} else {
		caseQ.OrderBy("height", pagination.GetOrderBy())
	}
	caseQ.Paginate(pagination.GetLimit(), pagination.GetPage())

	selectQ, args := caseQ.Build()
	rows, err = ads.QueryExecutor.ExecuteSelect(selectQ, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	accDatasets, err = ads.AccountDatasetQuery.BuildModel([]*model.AccountDataset{}, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, "Something wrong happened")
	}

	return &model.GetAccountDatasetsResponse{
		Total:           count,
		AccountDatasets: accDatasets,
	}, nil
}

func (ads *AccountDatasetService) GetAccountDataset(
	request *model.GetAccountDatasetRequest,
) (*model.AccountDataset, error) {
	var (
		err        error
		accDataset model.AccountDataset
		row        *sql.Row
		caseQ      = query.NewCaseQuery()
	)

	if request.GetRecipientAccountAddress() == "" && request.GetProperty() == "" {
		return nil, status.Error(codes.InvalidArgument, "Request must have Property or RecipientAccountAddress")
	}

	caseQ.Select(ads.AccountDatasetQuery.TableName, ads.AccountDatasetQuery.GetFields()...)
	if request.GetProperty() != "" {
		caseQ.Where(caseQ.Equal("property", request.GetProperty()))
	}
	if request.GetRecipientAccountAddress() != "" {
		caseQ.Where(caseQ.Equal("recipient_account_address", request.GetRecipientAccountAddress()))
	}
	caseQ.And(caseQ.Equal("latest", 1))

	selectQ, args := caseQ.Build()
	row, _ = ads.QueryExecutor.ExecuteSelectRow(selectQ, false, args...)
	err = ads.AccountDatasetQuery.Scan(&accDataset, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "There is something wrong")
		}

		return nil, status.Error(codes.NotFound, "Record not found")
	}

	return &accDataset, nil
}
