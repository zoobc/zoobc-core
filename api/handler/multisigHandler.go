package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
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
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if req.GetPagination().GetPage() < 1 {
		return nil, status.Error(codes.InvalidArgument, "PageCannotBeLessThanOne")
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}

	return msh.MultisigService.GetPendingTransactions(req)
}

func (msh *MultisigHandler) GetPendingTransactionDetailByTransactionHash(
	_ context.Context,
	req *model.GetPendingTransactionDetailByTransactionHashRequest,
) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	result, err := msh.MultisigService.GetPendingTransactionDetailByTransactionHash(req)
	return result, err
}

func (msh *MultisigHandler) GetMultisignatureInfo(
	_ context.Context,
	req *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
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

func (msh *MultisigHandler) GetMultisigAddressByParticipantAddress(
	_ context.Context,
	req *model.GetMultisigAddressByParticipantAddressRequest,
) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	if req.Pagination == nil {
		req.Pagination = &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
		}
	}
	if req.GetPagination().GetOrderField() == "" {
		req.Pagination.OrderField = "block_height"
		req.Pagination.OrderBy = model.OrderBy_DESC
	}
	result, err := msh.MultisigService.GetMultisigAddressByParticipantAddress(req)
	return result, err
}
