package service

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	ReceiptServiceInterface interface {
		SelectReceipts(blockTimestamp int64) ([]*model.Receipt, error)
	}

	ReceiptService struct {
		ReceiptQuery    query.ReceiptQueryInterface
		MerkleTreeQuery query.MerkleTreeQueryInterface
		KVExecutor      kvdb.KVExecutorInterface
		QueryExecutor   query.ExecutorInterface
	}
)

func NewReceiptService(
	receiptQuery query.ReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
) *ReceiptService {
	return &ReceiptService{
		ReceiptQuery:    receiptQuery,
		MerkleTreeQuery: merkleTreeQuery,
		KVExecutor:      kvExecutor,
		QueryExecutor:   queryExecutor,
	}
}

// SelectReceipts select list of receipts to be included in a block by prioritizing receipts that might
// increase the participation score of the node
func (rs *ReceiptService) SelectReceipts(blockTimestamp int64) ([]*model.Receipt, error) {
	// get linked rmr that has been included in previously published blocks
	rmrFilters, err := rs.KVExecutor.GetByPrefix(constant.TableBlockReminderKey)
	if err != nil {
		return nil, err
	}
	var receiptList = make(map[string][]*model.Receipt)
	var receiptTree = make(map[string][]byte)
	// use the rmr as filter to fetch node receipt
	for k, _ := range rmrFilters {
		var receipts []*model.Receipt
		root := []byte(k)[len([]byte(constant.TableBlockReminderKey)):]
		// look up the tree, todo: use join query (with receipts) instead later - andy-shi88
		treeQ, treeArgs := rs.MerkleTreeQuery.GetMerkleTreeByRoot(root)
		row := rs.QueryExecutor.ExecuteSelectRow(treeQ, false, treeArgs)
		if err != nil {
			return nil, err
		}
		tree, err := rs.MerkleTreeQuery.ScanTree(row)
		if err != nil {
			return nil, err
		}
		// look up rmr in table
		receiptsQ, rootArgs := rs.ReceiptQuery.GetReceiptByRoot(root)
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false, rootArgs)
		if err != nil {
			return nil, err
		}
		receipts = rs.ReceiptQuery.BuildModel(receipts, rows)
		// store fetched receipts and tree
		receiptList[string(root)] = receipts
		receiptTree[string(root)] = tree
	}
	// limit the selected portion to 20 receipts
	// filter the selected receipts on second phase
	//var i int
	//for rcRoot, rcReceipt := range receiptList {
	//	if i >= 20 {
	//		break
	//	}
	//
	//	merkleRoot := util.MerkleRoot{}
	//	i++
	//}
	// get intermediate hashes of every linked receipt
	return nil, nil
}
