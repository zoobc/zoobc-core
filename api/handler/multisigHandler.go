package handler

import (
	"context"
	"fmt"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	MultisigHandler struct {
		MultisigService service.MultisigServiceInterface
	}
)

func (msh *MultisigHandler) GetPendingTransactions(
	ctx context.Context,
	req *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}

	var notAllowedStatuses = map[model.PendingTransactionStatus]bool{
		model.PendingTransactionStatus_PendingTransactionPending: false,
		model.PendingTransactionStatus_PendingTransactionExpired: false,
	}
	if s, ok := notAllowedStatuses[req.GetStatus()]; ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Getting by status: %v denied", s))
	}
	return msh.MultisigService.GetPendingTransactions(req)
}

func (msh *MultisigHandler) GetPendingTransactionDetailByTransactionHash(
	ctx context.Context,
	req *model.GetPendingTransactionDetailByTransactionHashRequest,
) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	result, err := msh.MultisigService.GetPendingTransactionDetailByTransactionHash(req)
	return result, err
}

func (msh *MultisigHandler) GetMultisignatureInfo(
	ctx context.Context,
	req *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	if req.GetPagination().GetPage() > 30 {
		return nil, status.Error(codes.InvalidArgument, "LimitCannotBeMoreThan30")
	}
	result, err := msh.MultisigService.GetMultisignatureInfo(req)
	return result, err
}
