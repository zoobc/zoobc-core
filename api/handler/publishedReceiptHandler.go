package handler

import (
	"context"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PublishedReceiptHandler handles requests related to published receipts
type PublishedReceiptHandler struct {
	Service service.PublishedReceiptServiceInterface
}

// GetTransactions handles request to get data of a single Transaction
func (prh *PublishedReceiptHandler) GetPublishedReceipts(
	ctx context.Context,
	req *model.GetPublishedReceiptsRequest,
) (*model.GetPublishedReceiptsResponse, error) {
	var (
		response *model.GetPublishedReceiptsResponse
		err      error
	)
	if req.GetFromHeight() > req.GetToHeight() {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"ToHeight should bigger than FromHeight",
		)
	}
	if req.GetToHeight()-req.GetFromHeight() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}
	response, err = prh.Service.GetPublishedReceipts(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
