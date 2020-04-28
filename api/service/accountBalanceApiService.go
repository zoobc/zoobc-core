package service

import (
	"database/sql"
	"errors"

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
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Executor            query.ExecutorInterface
	}
)

func NewAccountBalanceService(executor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface) *AccountBalanceService {
	return &AccountBalanceService{
		AccountBalanceQuery: accountBalanceQuery,
		Executor:            executor,
	}
}

func (abs *AccountBalanceService) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	var (
		accountBalance model.AccountBalance
		row            *sql.Row
		err            error
	)

	qry, args := abs.AccountBalanceQuery.GetAccountBalanceByAccountAddress(request.AccountAddress)
	row, _ = abs.Executor.ExecuteSelectRow(qry, false, args...)
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

func (abs *AccountBalanceService) GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error) {

	if len(request.AccountAddresses) == 0 {
		return nil, errors.New("error: at least 1 address is required")
	}

	var (
		accountBalance  model.AccountBalance
		accountBalances []*model.AccountBalance
		row             *sql.Row
		err             error
	)

	for _, accountAddress := range request.AccountAddresses {
		qry, args := abs.AccountBalanceQuery.GetAccountBalanceByAccountAddress(accountAddress)
		row, _ = abs.Executor.ExecuteSelectRow(qry, false, args...)
		err = abs.AccountBalanceQuery.Scan(&accountBalance, row)
		if err != nil {
			accountBalance = model.AccountBalance{AccountAddress: accountAddress}
		}

		accountBalances = append(accountBalances, &accountBalance)
	}

	return &model.GetAccountBalancesResponse{
		AccountBalance: accountBalances,
	}, nil
}
