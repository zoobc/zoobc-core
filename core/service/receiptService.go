package service

import (
	"bytes"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	ReceiptServiceInterface interface {
		SelectReceipts(blockTimestamp int64, numberOfReceipt int) ([]*model.PublishedReceipt, error)
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
func (rs *ReceiptService) SelectReceipts(blockTimestamp int64, numberOfReceipt int) ([]*model.PublishedReceipt, error) {
	var (
		linkedReceiptList = make(map[string][]*model.Receipt)
		// this variable is to store picked receipt recipient to avoid duplicates
		pickedRecipients = make(map[string]bool)
	)

	// get the last merkle tree we have build so far
	treeQ := rs.MerkleTreeQuery.SelectMerkleTree(0, 1000, uint32(numberOfReceipt))
	rows, err := rs.QueryExecutor.ExecuteSelect(treeQ, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	linkedReceiptTree, err := rs.MerkleTreeQuery.BuildTree(rows)
	if err != nil {
		return nil, err
	}
	for linkedRoot := range linkedReceiptTree {
		var receipts []*model.Receipt
		receiptsQ, rootArgs := rs.ReceiptQuery.GetReceiptByRoot([]byte(linkedRoot))
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		receipts = rs.ReceiptQuery.BuildModel(receipts, rows)
		for _, rc := range receipts {
			if !pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] {
				pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] = true
				linkedReceiptList[linkedRoot] = append(linkedReceiptList[linkedRoot], rc)
			}
		}
	}
	// limit the selected portion to `numberOfReceipt` receipts
	// filter the selected receipts on second phase
	var (
		results []*model.PublishedReceipt
	)
	for rcRoot, rcReceipt := range linkedReceiptList {
		merkle := util.MerkleRoot{}
		merkle.HashTree = merkle.FromBytes(linkedReceiptTree[rcRoot], []byte(rcRoot))
		for _, rc := range rcReceipt {
			if len(results) >= numberOfReceipt {
				break
			}
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
				&model.PublishedReceipt{
					BatchReceipt:       rc.BatchReceipt,
					IntermediateHashes: merkle.FlattenIntermediateHashes(intermediateHashes),
					ReceiptIndex:       rc.RMRIndex,
				},
			)
		}
	}
	// prioritize those receipts with rmr_linked != nil
	if len(results) < numberOfReceipt {
		var receipts []*model.Receipt
		// look up rmr in table | todo: randomize selection
		receiptsQ := rs.ReceiptQuery.GetReceiptsWithUniqueRecipient(uint32(numberOfReceipt-len(results)), 0, true)
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		receipts = rs.ReceiptQuery.BuildModel(receipts, rows)
		for _, rc := range receipts {
			if !pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] {
				results = append(results, &model.PublishedReceipt{
					BatchReceipt:       rc.BatchReceipt,
					IntermediateHashes: nil,
					ReceiptIndex:       rc.RMRIndex,
				})
				pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] = true
			}
		}
	}
	// fill in unlinked receipts if the limit has not been reached
	if len(results) < numberOfReceipt { // get unlinked receipts randomly, in future additional filter may be added
		var receipts []*model.Receipt
		// look up rmr in table | todo: randomize selection
		receiptsQ := rs.ReceiptQuery.GetReceiptsWithUniqueRecipient(uint32(numberOfReceipt-len(results)), 0, false)
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		receipts = rs.ReceiptQuery.BuildModel(receipts, rows)
		for _, rc := range receipts {
			if !pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] {
				results = append(results, &model.PublishedReceipt{
					BatchReceipt:       rc.BatchReceipt,
					IntermediateHashes: nil,
					ReceiptIndex:       rc.RMRIndex,
				})
			}
		}
	}
	return results, nil
}
