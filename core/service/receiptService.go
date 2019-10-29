package service

import (
	"bytes"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptServiceInterface interface {
		SelectReceipts(
			blockTimestamp int64,
			numberOfReceipt int,
			lastBlockHeight uint32,
		) ([]*model.PublishedReceipt, error)
		GenerateReceiptsMerkleRoot() error
	}

	ReceiptService struct {
		ReceiptQuery      query.ReceiptQueryInterface
		BatchReceiptQuery query.BatchReceiptQueryInterface
		MerkleTreeQuery   query.MerkleTreeQueryInterface
		KVExecutor        kvdb.KVExecutorInterface
		QueryExecutor     query.ExecutorInterface
	}
)

func NewReceiptService(
	receiptQuery query.ReceiptQueryInterface,
	batchReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,

) *ReceiptService {
	return &ReceiptService{
		ReceiptQuery:      receiptQuery,
		BatchReceiptQuery: batchReceiptQuery,
		MerkleTreeQuery:   merkleTreeQuery,
		KVExecutor:        kvExecutor,
		QueryExecutor:     queryExecutor,
	}
}

// SelectReceipts select list of receipts to be included in a block by prioritizing receipts that might
// increase the participation score of the node
func (rs *ReceiptService) SelectReceipts(
	blockTimestamp int64,
	numberOfReceipt int,
	lastBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var (
		linkedReceiptList = make(map[string][]*model.Receipt)
		// this variable is to store picked receipt recipient to avoid duplicates
		pickedRecipients = make(map[string]bool)
		lowerBlockHeight uint32
	)

	// get the last merkle tree we have build so far
	if lastBlockHeight > constant.ReceiptNumberOfBlockToPick {
		lowerBlockHeight = lastBlockHeight - constant.ReceiptNumberOfBlockToPick
	}
	treeQ := rs.MerkleTreeQuery.SelectMerkleTree(lowerBlockHeight, lastBlockHeight, uint32(numberOfReceipt))
	linkedTreeRows, err := rs.QueryExecutor.ExecuteSelect(treeQ, false)
	if err != nil {
		return nil, err
	}
	defer linkedTreeRows.Close()

	linkedReceiptTree, err := rs.MerkleTreeQuery.BuildTree(linkedTreeRows)
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

		receipts, err = rs.ReceiptQuery.BuildModel(receipts, rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		for _, rc := range receipts {
			if !pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] {
				pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] = true
				linkedReceiptList[linkedRoot] = append(linkedReceiptList[linkedRoot], rc)
			}
		}
		if rows != nil {
			rows.Close()
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
		uniqueReceiptRows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		defer uniqueReceiptRows.Close()

		receipts, err = rs.ReceiptQuery.BuildModel(receipts, uniqueReceiptRows)
		if err != nil {
			return nil, err
		}
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
		uniqueReceiptRandRows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		defer uniqueReceiptRandRows.Close()

		receipts, err = rs.ReceiptQuery.BuildModel(receipts, uniqueReceiptRandRows)
		if err != nil {
			return nil, err
		}

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

// GenerateReceiptsMerkleRoot generate merkle root of some batch receipts
// generating will do when number of collected receipts(batch receipts) already same with number of required
func (rs *ReceiptService) GenerateReceiptsMerkleRoot() error {
	var (
		err               error
		count             uint32
		queries           [][]interface{}
		batchReceipts     []*model.BatchReceipt
		receipt           *model.Receipt
		hashedReceipts    []*bytes.Buffer
		merkleRoot        util.MerkleRoot
		getBatchReceiptsQ string
	)

	getBatchReceiptsQ = rs.BatchReceiptQuery.GetBatchReceipts(model.Pagination{
		Limit:      constant.ReceiptBatchMaximum,
		OrderField: "reference_block_height",
		OrderBy:    model.OrderBy_ASC,
	})

	err = rs.QueryExecutor.ExecuteSelectRow(
		query.GetTotalRecordOfSelect(getBatchReceiptsQ),
	).Scan(&count)
	if err != nil {
		return err
	}

	if count >= constant.ReceiptBatchMaximum {
		rows, err := rs.QueryExecutor.ExecuteSelect(getBatchReceiptsQ, false)
		if err != nil {
			return err
		}
		defer rows.Close()

		queries = make([][]interface{}, (constant.ReceiptBatchMaximum*2)+1)
		batchReceipts, err = rs.BatchReceiptQuery.BuildModel(batchReceipts, rows)
		if err != nil {
			return err
		}

		for _, b := range batchReceipts {
			// hash the receipts
			hashedBatchReceipt := sha3.Sum256(util.GetSignedBatchReceiptBytes(b))
			hashedReceipts = append(
				hashedReceipts,
				bytes.NewBuffer(hashedBatchReceipt[:]),
			)
		}
		_, err = merkleRoot.GenerateMerkleRoot(hashedReceipts)
		if err != nil {
			return err
		}
		rootMerkle, treeMerkle := merkleRoot.ToBytes()

		for k, r := range batchReceipts {
			var (
				br       = r
				rmrIndex = uint32(k)
			)

			receipt = &model.Receipt{
				BatchReceipt: br,
				RMR:          rootMerkle,
				RMRIndex:     rmrIndex,
			}
			insertReceiptQ, insertReceiptArgs := rs.ReceiptQuery.InsertReceipt(receipt)
			queries[k] = append([]interface{}{insertReceiptQ}, insertReceiptArgs...)
			removeBatchReceiptQ, removeBatchReceiptArgs := rs.BatchReceiptQuery.RemoveBatchReceipt(br.DatumType, br.DatumHash)
			queries[(constant.ReceiptBatchMaximum)+uint32(k)] = append([]interface{}{removeBatchReceiptQ}, removeBatchReceiptArgs...)
		}

		insertMerkleTreeQ, insertMerkleTreeArgs := rs.MerkleTreeQuery.InsertMerkleTree(rootMerkle, treeMerkle, time.Now().Unix())
		queries[len(queries)-1] = append([]interface{}{insertMerkleTreeQ}, insertMerkleTreeArgs...)

		err = rs.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = rs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			_ = rs.QueryExecutor.RollbackTx()
			return err
		}
		err = rs.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}
