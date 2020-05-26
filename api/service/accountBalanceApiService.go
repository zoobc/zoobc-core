package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	AccountBalanceServiceInterface interface {
		GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error)
		GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error)
	}

	AccountBalanceService struct {
		AccountBalanceQuery *query.AccountBalanceQuery
		QueryExecutor       query.ExecutorInterface
	}
)

func NewAccountBalanceService(executor query.ExecutorInterface,
	accountBalanceQuery *query.AccountBalanceQuery) *AccountBalanceService {
	return &AccountBalanceService{
		AccountBalanceQuery: accountBalanceQuery,
		QueryExecutor:       executor,
	}
}

func (abs *AccountBalanceService) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	var (
		accountBalance model.AccountBalance
		row            *sql.Row
		err            error
	)

	qry, args := abs.AccountBalanceQuery.GetAccountBalanceByAccountAddress(request.AccountAddress)
	row, _ = abs.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	err = abs.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, "account not found")

	}

	return &model.GetAccountBalanceResponse{
		AccountBalance: &accountBalance,
	}, nil
}

func (abs *AccountBalanceService) GetAccountBalances(
	request *model.GetAccountBalancesRequest,
) (*model.GetAccountBalancesResponse, error) {
	var (
		accountBalances []*model.AccountBalance
		caseQ           = query.NewCaseQuery()
		rows            *sql.Rows
		err             error
	)

	caseQ.Select(abs.AccountBalanceQuery.TableName, abs.AccountBalanceQuery.Fields...)
	var accountAddresses []interface{}
	for _, v := range request.AccountAddresses {
		accountAddresses = append(accountAddresses, v)
	}
	caseQ.And(caseQ.In("account_address", accountAddresses...))
	caseQ.And(caseQ.Equal("latest", true))

	selectQ, args := caseQ.Build()
	rows, err = abs.QueryExecutor.ExecuteSelect(selectQ, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	accountBalances, err = abs.AccountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetAccountBalancesResponse{
		AccountBalances: accountBalances,
	}, nil
}
