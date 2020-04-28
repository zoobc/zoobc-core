package handler

import (
	"context"
	"errors"
	"fmt"

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

	fmt.Println("1. request.AccountAddresses::", request.AccountAddresses)
	if len(request.AccountAddresses) == 0 {
		return nil, errors.New("Error: at least 1 address is required")
	}

	accountBalances, err := abh.Service.GetAccountBalances(request)
	if err != nil {
		return nil, err
	}
	return accountBalances, nil
}
