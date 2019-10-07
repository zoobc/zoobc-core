package service

import (
	"bytes"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	ReceiptServiceInterface interface {
		SelectReceipts(blockTimestamp int64, numberOfReceipt int) ([]*model.BlockReceipt, error)
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
func (rs *ReceiptService) SelectReceipts(blockTimestamp int64, numberOfReceipt int) ([]*model.BlockReceipt, error) {
	// get linked rmr that has been included in previously published blocks
	rmrFilters, err := rs.KVExecutor.GetByPrefix(constant.TableBlockReminderKey)
	if err != nil {
		return nil, err
	}
	var (
		linkedReceiptList = make(map[string][]*model.Receipt)
		linkedReceiptTree = make(map[string][]byte)
	)
	// use the rmr as filter to fetch node receipt
	for k := range rmrFilters {
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
		linkedReceiptList[string(root)] = receipts
		linkedReceiptTree[string(root)] = tree
	}
	// limit the selected portion to `numberOfReceipt` receipts
	// filter the selected receipts on second phase
	var (
		i       int
		results []*model.BlockReceipt
	)
	for rcRoot, rcReceipt := range linkedReceiptList {
		if len(results) >= numberOfReceipt {
			break
		}
		merkle := util.MerkleRoot{}
		merkle.HashTree = merkle.FromBytes(linkedReceiptTree[rcRoot], []byte(rcRoot))
		for _, rc := range rcReceipt {
			var intermediateHashes [][]byte
			rcByte := util.GetSignedBatchReceiptBytes(rc.BatchReceipt)
			rcHash := sha3.Sum256(rcByte)

			intermediateHashesBuffer := merkle.GetIntermediateHashes(
				bytes.NewBuffer(rcHash[:]),
				int32(rc.RMRIndex),
			)
			for _, buf := range intermediateHashesBuffer {
				intermediateHashes = append(intermediateHashes, buf.Bytes())
			}
			results = append(
				results,
				&model.BlockReceipt{
					ReceiptHash:        rcHash[:],
					IntermediateHashes: intermediateHashes,
				},
			)
		}
		i++
	}
	// fill in unlinked receipts if the limit has not been reached
	if len(results) < numberOfReceipt { // get unlinked receipts randomly, in future additional filter may be added
		var receipts []*model.Receipt
		// look up rmr in table | todo: randomize selection
		receiptsQ := rs.ReceiptQuery.GetReceipts(uint32(numberOfReceipt-len(results)), 0)
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		receipts = rs.ReceiptQuery.BuildModel(receipts, rows)
		for _, rc := range receipts {
			rcByte := util.GetSignedBatchReceiptBytes(rc.BatchReceipt)
			rcHash := sha3.Sum256(rcByte)

			results = append(results, &model.BlockReceipt{
				ReceiptHash:        rcHash[:],
				IntermediateHashes: nil,
			})
		}
	}
	return results, nil
}
