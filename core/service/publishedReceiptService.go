package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
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

// ProcessPublishedReceipts takes published receipts in a block and validate them, this function will run in a db transaction
// so ensure queryExecutor.Begin() is called before calling this function.
func (ps *PublishedReceiptService) ProcessPublishedReceipts(block *model.Block) (int, error) {
	var (
		linkedCount int
		err         error
	)
	for index, rc := range block.GetPublishedReceipts() {
		// validate sender and recipient of receipt
		err = ps.ReceiptService.ValidateReceipt(rc.BatchReceipt)
		if err != nil {
			return 0, err
		}
		// check if linked
		if rc.IntermediateHashes != nil && len(rc.IntermediateHashes) > 0 {
			merkle := &commonUtils.MerkleRoot{}
			rcByte := ps.ReceiptUtil.GetSignedBatchReceiptBytes(rc.BatchReceipt)
			rcHash := sha3.Sum256(rcByte)
			root, err := merkle.GetMerkleRootFromIntermediateHashes(
				rcHash[:],
				rc.ReceiptIndex,
				merkle.RestoreIntermediateHashes(rc.IntermediateHashes),
			)
			if err != nil {
				return 0, err
			}
			// look up root in published_receipt table
			_, err = ps.PublishedReceiptUtil.GetPublishedReceiptByLinkedRMR(root)
			if err != nil {
				return 0, err
			}
			// add to linked receipt count for calculation later
			linkedCount++
		}
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)

		err := ps.PublishedReceiptUtil.InsertPublishedReceipt(rc, true)
		if err != nil {
			return 0, err
		}
	}
	return linkedCount, nil
}
