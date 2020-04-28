package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountBalanceHandler struct {
		Service service.AccountBalanceServiceInterface
	}
)

func (abh *AccountBalanceHandler) GetAccountBalance(ctx context.Context,
	request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	accountBalance, err := abh.Service.GetAccountBalance(request)
	if err != nil {
		return nil, err
	}
	return accountBalance, nil
}

func (abh *AccountBalanceHandler) GetAccountBalances(ctx context.Context,
	request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error) {

	accountBalances, err := abh.Service.GetAccountBalances(request)
	if err != nil {
		return nil, err
	}
	return accountBalances, nil
}
