package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TransactionHandler handles requests related to transactions
type LiquidTransactionHandler struct {
	Service service.LiquidTransactionServiceInterface
}

// GetLiquidTransactions handles request to get data of a single Transaction
func (lts *LiquidTransactionHandler) GetLiquidTransactions(
	ctx context.Context,
	req *model.GetLiquidTransactionsRequest,
) (*model.GetLiquidTransactionsResponse, error) {
	var (
		response *model.GetLiquidTransactionsResponse
		err      error
	)

	pagination := req.GetPagination()
	if pagination == nil {
		pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       0,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if pagination.GetLimit() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	response, err = lts.Service.GetLiquidTransactions(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
