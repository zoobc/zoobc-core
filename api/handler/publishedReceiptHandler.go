package handler

import (
	"context"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
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
	response, err = prh.Service.GetPublishedReceipts(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
