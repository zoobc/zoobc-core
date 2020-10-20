package util

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// PublishedReceiptUtilInterface act as interface for data getter on published_receipt entity
	PublishedReceiptUtilInterface interface {
		GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		GetPublishedReceiptsByBlockHeightRange(fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error)
		GetPublishedReceiptByLinkedRMR(root []byte) (*model.PublishedReceipt, error)
		InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt, tx bool) error
	}
	PublishedReceiptUtil struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewPublishedReceiptUtil(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	queryExecutor query.ExecutorInterface,
) *PublishedReceiptUtil {
	return &PublishedReceiptUtil{
		PublishedReceiptQuery: publishedReceiptQuery,
		QueryExecutor:         queryExecutor,
	}
}

// GetPublishedReceiptByBlockHeight get data from published_receipt table by the block height they were published / broadcasted
func (psu *PublishedReceiptUtil) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := psu.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err := psu.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = psu.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

// GetPublishedReceiptByBlockHeightRange get data from published_receipt table by the block height they were published / broadcasted
func (psu *PublishedReceiptUtil) GetPublishedReceiptsByBlockHeightRange(fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := psu.PublishedReceiptQuery.GetPublishedReceiptByBlockHeightRange(
		fromBlockHeight, toBlockHeight,
	)
	rows, err := psu.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = psu.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

func (psu *PublishedReceiptUtil) GetPublishedReceiptByLinkedRMR(root []byte) (*model.PublishedReceipt, error) {
	var (
		publishedReceipt = &model.PublishedReceipt{
			Receipt:            &model.Receipt{},
			IntermediateHashes: nil,
			BlockHeight:        0,
			ReceiptIndex:       0,
		}
		err error
	)
	// look up root in published_receipt table
	rcQ, rcArgs := psu.PublishedReceiptQuery.GetPublishedReceiptByLinkedRMR(root)
	row, _ := psu.QueryExecutor.ExecuteSelectRow(rcQ, false, rcArgs...)
	err = psu.PublishedReceiptQuery.Scan(publishedReceipt, row)
	if err != nil {
		return nil, err
	}
	return publishedReceipt, nil
}

func (psu *PublishedReceiptUtil) InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt, tx bool) error {
	var err error
	insertPublishedReceiptQ, insertPublishedReceiptArgs := psu.PublishedReceiptQuery.InsertPublishedReceipt(
		publishedReceipt,
	)
	if tx {
		err = psu.QueryExecutor.ExecuteTransaction(insertPublishedReceiptQ, insertPublishedReceiptArgs...)
	} else {
		_, err = psu.QueryExecutor.ExecuteStatement(insertPublishedReceiptQ, insertPublishedReceiptArgs...)
	}
	if err != nil {
		return err
	}
	return nil
}
