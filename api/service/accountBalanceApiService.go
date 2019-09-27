package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	accountBalanceQuery, arg := abs.AccountBalanceQuery.GetAccountBalanceByAccountAddress(request.AccountAddress)
	rows, err := abs.Executor.ExecuteSelect(accountBalanceQuery, false, arg)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	accountBalances = abs.AccountBalanceQuery.BuildModel(accountBalances, rows)

	if len(accountBalances) == 0 {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	return &model.GetAccountBalanceResponse{
		AccountBalance: accountBalances[0],
	}, nil
}
