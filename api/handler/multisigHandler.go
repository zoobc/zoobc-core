package handler

import (
	"context"

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
