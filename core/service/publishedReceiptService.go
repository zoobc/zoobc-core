package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// PublishedReceiptServiceInterface act as interface for processing the published receipt data
	PublishedReceiptServiceInterface interface {
		ProcessPublishedReceipts(block *model.Block) (int, error)
	}

	PublishedReceiptService struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		ReceiptUtil           util.ReceiptUtilInterface
		PublishedReceiptUtil  util.PublishedReceiptUtilInterface
		ReceiptService        ReceiptServiceInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewPublishedReceiptService(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil util.ReceiptUtilInterface,
	publishedReceiptUtil util.PublishedReceiptUtilInterface,
	receiptService ReceiptServiceInterface,
	queryExecutor query.ExecutorInterface,
) *PublishedReceiptService {
	return &PublishedReceiptService{
		PublishedReceiptQuery: publishedReceiptQuery,
		ReceiptUtil:           receiptUtil,
		PublishedReceiptUtil:  publishedReceiptUtil,
		ReceiptService:        receiptService,
		QueryExecutor:         queryExecutor,
	}
}

// ProcessPublishedReceipts takes published receipts in a block and validate
// them, this function will run in a db transaction so ensure
// queryExecutor.Begin() is called before calling this function.
func (ps *PublishedReceiptService) ProcessPublishedReceipts(block *model.Block) (int, error) {
	var (
		linkedCount int
		err         error
	)
	for index, rc := range block.GetFreeReceipts() {
		// validate sender and recipient of receipt
		rcCopy := *rc
		err = ps.ReceiptService.ValidateReceipt(rc.GetReceipt())
		if err != nil {
			return 0, err
		}
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)
		err := ps.PublishedReceiptUtil.InsertPublishedReceipt(&rcCopy, true)
		if err != nil {
			return 0, err
		}
	}
	return linkedCount, nil
}
