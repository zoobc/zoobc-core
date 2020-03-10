package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

// EscrowTransactionHandler to handle request related to Escrow Transaction grpc handler from client
type EscrowTransactionHandler struct {
	Service service.EscrowTransactionServiceInterface
}

// GetEscrowTransactions get escrow transactions with filter fields params
func (eh *EscrowTransactionHandler) GetEscrowTransactions(
	_ context.Context,
	req *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	return eh.Service.GetEscrowTransactions(req)
}

func (eh *EscrowTransactionHandler) GetEscrowTransaction(
	_ context.Context,
	req *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	return eh.Service.GetEscrowTransaction(req)
}
