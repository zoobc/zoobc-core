package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zoobc/zoobc-core/api/service"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MultisigHandler struct {
		MultisigService service.MultisigServiceInterface
	}
)

func (msh *MultisigHandler) GetPendingTransactionByAddress(
	ctx context.Context,
	req *model.GetPendingTransactionByAddressRequest,
) (*model.GetPendingTransactionByAddressResponse, error) {
	if req.GetSenderAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "SenderAddressNotProvided")
	}
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	result, err := msh.MultisigService.GetPendingTransactionByAddress(req)
	return result, err
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
