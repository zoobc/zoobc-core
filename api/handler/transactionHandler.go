package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

// TransactionHandler handles requests related to transactions
type TransactionHandler struct {
	Service service.TransactionServiceInterface
}

// GetTransaction handles request to get data of a single Transaction
func (th *TransactionHandler) GetTransaction(ctx context.Context, req *model.GetTransactionRequest) (*model.Transaction, error) {
	var transaction *model.Transaction
	var err error
	chainType := chaintype.GetChainType(0)
	transaction, err = th.Service.GetTransaction(chainType, req)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransactions handles request to get data of a single Transaction
func (th *TransactionHandler) GetTransactions(ctx context.Context,
	req *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	var response *model.GetTransactionsResponse
	var err error
	chainType := chaintype.GetChainType(0)
	response, err = th.Service.GetTransactions(chainType, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
