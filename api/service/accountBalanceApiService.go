package service

import (
	"errors"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	AccountBalanceServiceInterface interface {
		GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error)
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
		err             error
		accountBalances []*model.AccountBalance
	)
	accountID := util.CreateAccountIDFromAddress(request.AccountType, request.AccountAddress)
	accountBalanceQuery, arg := abs.AccountBalanceQuery.GetAccountBalanceByAccountID(accountID)
	rows, err := abs.Executor.ExecuteSelect(accountBalanceQuery, arg)

	if err != nil {
		return nil, err
	}

	accountBalances = abs.AccountBalanceQuery.BuildModel(accountBalances, rows)

	if len(accountBalances) == 0 {
		return nil, errors.New("error: account not found")
	}

	return &model.GetAccountBalanceResponse{
		AccountBalance: accountBalances[0],
	}, nil
}
