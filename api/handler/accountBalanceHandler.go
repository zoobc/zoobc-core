package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if len(request.AccountAddresses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "At least 1 address is required")
	}

	return abh.Service.GetAccountBalances(request)
}
