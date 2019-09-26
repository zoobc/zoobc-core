package service

import (
	"errors"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
	rows, err := abs.Executor.ExecuteSelect(abs.AccountBalanceQuery.GetAccountBalanceByAccountAddress(request.AccountAddress), false)

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
