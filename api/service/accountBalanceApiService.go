package service

import (
	"errors"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// AccountBalanceServiceInterface represents interface for BlockService
	AccountBalanceServiceInterface interface {
		GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error)
		GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error)
	}

	// AccountBalanceService represents struct of BlockService
	AccountBalanceService struct {
		Query               query.ExecutorInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
	}
)

var accountBalanceServiceInstance *AccountBalanceService

// NewBlockService create a singleton instance of BlockService
func NewAccountBalanceService(queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface) *AccountBalanceService {
	if accountBalanceServiceInstance == nil {
		accountBalanceServiceInstance = &AccountBalanceService{
			Query:               queryExecutor,
			AccountBalanceQuery: accountBalanceQuery,
		}
	}
	return accountBalanceServiceInstance
}

func (abs *AccountBalanceService) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	var (
		accountBalances []*model.AccountBalance
	)
	accountID := util.CreateAccountIDFromAddress(request.AccountType, request.AccountAddress)
	accountBalanceQ, accountBalanceArg := abs.AccountBalanceQuery.GetAccountBalanceByAccountID(accountID)
	rows, err := abs.Query.ExecuteSelect(accountBalanceQ, accountBalanceArg)
	if err != nil {
		return nil, errors.New("error fetching account balance")
	}
	defer rows.Close()

	accountBalances = abs.AccountBalanceQuery.BuildModel(accountBalances, rows)
	if len(accountBalances) == 0 {
		return nil, errors.New("account balance with provided address and type not found")
	}
	return &model.GetAccountBalanceResponse{
		AccountBalance: accountBalances[0],
	}, nil
}

func (abs *AccountBalanceService) GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error) {
	return nil, nil
}
