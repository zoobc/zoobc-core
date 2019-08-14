package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	UnconfirmedTransactionHandler struct {
		Service service.UnconfirmedTransactionServiceInterface
	}
)

func (uth *UnconfirmedTransactionHandler) GetUnconfirmedTransactions(
	ctx context.Context,
	req *model.GetMempoolTransactionsRequest,
) (*model.GetMempoolTransactionsResponse, error) {
	var (
		response *model.GetMempoolTransactionsResponse
		err      error
	)

	chainType := chaintype.GetChainType(0)
	response, err = uth.Service.GetUnconfirmedTransactions(chainType, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
