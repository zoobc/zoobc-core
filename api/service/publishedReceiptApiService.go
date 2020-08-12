package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	util2 "github.com/zoobc/zoobc-core/core/util"
)

type (
	// PublishedReceiptServiceInterface represents interface for PublishedReceiptService
	PublishedReceiptServiceInterface interface {
		GetPublishedReceipts(*model.GetPublishedReceiptsRequest) (*model.GetPublishedReceiptsResponse, error)
	}

	// PublishedReceiptService represents struct of published receipt service
	PublishedReceiptService struct {
		PublishedReceiptUtil util2.PublishedReceiptUtilInterface
	}
)

// NewPublishedReceiptService creates an instance of PublishedReceiptService
func NewPublishedReceiptService(
	publishedReceiptUtil util2.PublishedReceiptUtilInterface,
) *PublishedReceiptService {
	return &PublishedReceiptService{
		PublishedReceiptUtil: publishedReceiptUtil,
	}
}

// GetPublishedReceipts fetches a published receipts within range of `fromHeight` - `toHeight`
func (prs *PublishedReceiptService) GetPublishedReceipts(
	params *model.GetPublishedReceiptsRequest,
) (*model.GetPublishedReceiptsResponse, error) {
	publishedReceipts, err := prs.PublishedReceiptUtil.GetPublishedReceiptsByBlockHeightRange(
		params.GetFromHeight(), params.GetToHeight(),
	)
	if err != nil {
		return nil, err
	}

	return &model.GetPublishedReceiptsResponse{
		PublishedReceipts: publishedReceipts,
	}, nil
}
