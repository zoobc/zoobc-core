package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TransactionHandler handles requests related to transactions
type TransactionHandler struct {
	Service service.TransactionServiceInterface
}

// GetTransaction handles request to get data of a single Transaction
func (th *TransactionHandler) GetTransaction(
	ctx context.Context,
	req *model.GetTransactionRequest,
) (*model.Transaction, error) {
	var (
		transaction *model.Transaction
		err         error
	)
	chainType := chaintype.GetChainType(0)
	transaction, err = th.Service.GetTransaction(chainType, req)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetTransactions handles request to get data of a single Transaction
func (th *TransactionHandler) GetTransactions(
	ctx context.Context,
	req *model.GetTransactionsRequest,
) (*model.GetTransactionsResponse, error) {
	var (
		response *model.GetTransactionsResponse
		err      error
	)

	pagination := req.GetPagination()
	if pagination == nil {
		pagination = &model.Pagination{
			OrderField: "timestamp",
			OrderBy:    model.OrderBy_DESC,
			Page:       0,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if pagination.GetLimit() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	chainType := chaintype.GetChainType(0)
	response, err = th.Service.GetTransactions(chainType, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// PostTransaction handle transaction submitted by client
func (th *TransactionHandler) PostTransaction(
	ctx context.Context,
	req *model.PostTransactionRequest,
) (*model.PostTransactionResponse, error) {
	chainType := &chaintype.MainChain{}
	transaction, err := th.Service.PostTransaction(chainType, req)
	if err != nil {
		return nil, err
	}
	return &model.PostTransactionResponse{
		Transaction: transaction,
	}, nil
}

// GetTransactionMinimumFee handles request to get transaction's minimum fee
func (th *TransactionHandler) GetTransactionMinimumFee(
	ctx context.Context,
	req *model.GetTransactionMinimumFeeRequest,
) (*model.GetTransactionMinimumFeeResponse, error) {
	var (
		transactionFee *model.GetTransactionMinimumFeeResponse
		err            error
	)
	transactionFee, err = th.Service.GetTransactionMinimumFee(req)
	if err != nil {
		return nil, err
	}

	return transactionFee, nil
}
