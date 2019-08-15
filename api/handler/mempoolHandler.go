package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MempoolTransactionHandler struct {
		Service service.MempoolTransactionServiceInterface
	}
)

func (uth *MempoolTransactionHandler) GetMempoolTransaction(
	context.Context,
	*model.GetMempoolTransactionRequest,
) (*model.GetMempoolTransactionResponse, error) {
	// TODO: Implement GetMempoolTransaction
	panic("implement me")
}

func (uth *MempoolTransactionHandler) GetMempoolTransactions(
	ctx context.Context,
	req *model.GetMempoolTransactionsRequest,
) (*model.GetMempoolTransactionsResponse, error) {
	var (
		response *model.GetMempoolTransactionsResponse
		err      error
	)

	chainType := chaintype.GetChainType(0)
	response, err = uth.Service.GetMempoolTransactions(chainType, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
