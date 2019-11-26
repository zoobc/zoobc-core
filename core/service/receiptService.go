package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
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
		ValidateReceipt(
			receipt *model.BatchReceipt,
		) error
		PruningNodeReceipts() error
	}

	ReceiptService struct {
		NodeReceiptQuery        query.NodeReceiptQueryInterface
		BatchReceiptQuery       query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		Signature               crypto.SignatureInterface
	}
)

func NewReceiptService(
	nodeReceiptQuery query.NodeReceiptQueryInterface,
	batchReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	blockQuery query.BlockQueryInterface,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	signature crypto.SignatureInterface,
) *ReceiptService {
	return &ReceiptService{
		NodeReceiptQuery:        nodeReceiptQuery,
		BatchReceiptQuery:       batchReceiptQuery,
		MerkleTreeQuery:         merkleTreeQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		BlockQuery:              blockQuery,
		KVExecutor:              kvExecutor,
		QueryExecutor:           queryExecutor,
		NodeRegistrationService: nodeRegistrationService,
		Signature:               signature,
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

	if numberOfReceipt < 1 { // possible no connected node
		return []*model.PublishedReceipt{}, nil
	}
	// get the last merkle tree we have build so far
	if lastBlockHeight > constant.NodeReceiptExpiryBlockHeight {
		lowerBlockHeight = lastBlockHeight - constant.NodeReceiptExpiryBlockHeight
	}
	treeQ := rs.MerkleTreeQuery.SelectMerkleTree(
		lowerBlockHeight,
		lastBlockHeight,
		uint32(numberOfReceipt)*constant.ReceiptBatchPickMultiplier)
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
		var nodeReceipts []*model.Receipt
		nodeReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptByRoot([]byte(linkedRoot))
		rows, err := rs.QueryExecutor.ExecuteSelect(nodeReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}

		nodeReceipts, err = rs.NodeReceiptQuery.BuildModel(nodeReceipts, rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		for _, rc := range nodeReceipts {
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
			err = rs.ValidateReceipt(rc.BatchReceipt)
			if err != nil {
				// skip invalid receipt
				continue
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
	// select non-linked receipt
	if len(results) < numberOfReceipt {
		rmrLinkedReceipts, err := rs.pickReceipts(
			numberOfReceipt, results, pickedRecipients, lowerBlockHeight, lastBlockHeight)
		if err != nil {
			return nil, err
		}
		results = rmrLinkedReceipts
	}

	return results, nil
}

func (rs *ReceiptService) pickReceipts(
	numberOfReceipt int,
	pickedReceipts []*model.PublishedReceipt,
	pickedRecipients map[string]bool,
	lowerBlockHeight, upperBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var receipts []*model.Receipt
	receiptsQ := rs.NodeReceiptQuery.GetReceiptsWithUniqueRecipient(
		uint32(numberOfReceipt)*constant.ReceiptBatchPickMultiplier, lowerBlockHeight, upperBlockHeight)
	rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	receipts, err = rs.NodeReceiptQuery.BuildModel(receipts, rows)
	if err != nil {
		return nil, err
	}
	for _, rc := range receipts {
		errValid := rs.ValidateReceipt(rc.BatchReceipt)
		if errValid != nil {
			// skipped invalid receipt
			continue
		}
		if !pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] {
			pickedReceipts = append(pickedReceipts, &model.PublishedReceipt{
				BatchReceipt:       rc.BatchReceipt,
				IntermediateHashes: nil,
				ReceiptIndex:       rc.RMRIndex,
			})
			pickedRecipients[string(rc.BatchReceipt.RecipientPublicKey)] = true
		}
	}
	return pickedReceipts, nil
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
		lastBlock         model.Block
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
			insertNodeReceiptQ, insertNodeReceiptArgs := rs.NodeReceiptQuery.InsertReceipt(receipt)
			queries[k] = append([]interface{}{insertNodeReceiptQ}, insertNodeReceiptArgs...)
			removeBatchReceiptQ, removeBatchReceiptArgs := rs.BatchReceiptQuery.RemoveBatchReceipt(br.DatumType, br.DatumHash)
			queries[(constant.ReceiptBatchMaximum)+uint32(k)] = append([]interface{}{removeBatchReceiptQ}, removeBatchReceiptArgs...)
		}
		lastBlockQ := rs.BlockQuery.GetLastBlock()
		lastBlockRow := rs.QueryExecutor.ExecuteSelectRow(lastBlockQ, false)
		err = rs.BlockQuery.Scan(&lastBlock, lastBlockRow)
		if err != nil {
			return err
		}
		insertMerkleTreeQ, insertMerkleTreeArgs := rs.MerkleTreeQuery.InsertMerkleTree(
			rootMerkle, treeMerkle, time.Now().Unix(), lastBlock.Height)
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

func (rs *ReceiptService) ValidateReceipt(
	receipt *model.BatchReceipt,
) error {
	var (
		blockAtHeight model.Block
		err           error
	)
	unsignedBytes := util.GetUnsignedBatchReceiptBytes(receipt)
	if !rs.Signature.VerifyNodeSignature(
		unsignedBytes,
		receipt.RecipientSignature,
		receipt.RecipientPublicKey,
	) {
		// rollback
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidReceiptSignature",
		)
	}
	blockAtHeightQ := rs.BlockQuery.GetBlockByHeight(receipt.ReferenceBlockHeight)
	blockAtHeightRow := rs.QueryExecutor.ExecuteSelectRow(blockAtHeightQ)
	err = rs.BlockQuery.Scan(&blockAtHeight, blockAtHeightRow)
	if err != nil {
		return err
	}
	// check block hash
	if !bytes.Equal(blockAtHeight.BlockHash, receipt.ReferenceBlockHash) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidReceiptBlockHash")
	}
	err = rs.validateReceiptSenderRecipient(receipt)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReceiptService) validateReceiptSenderRecipient(
	receipt *model.BatchReceipt,
) error {
	var (
		senderNodeRegistration    model.NodeRegistration
		recipientNodeRegistration model.NodeRegistration
		err                       error
	)
	// get sender address at height
	senderNodeQ, senderNodeArgs := rs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(
		receipt.SenderPublicKey,
		receipt.ReferenceBlockHeight,
	)
	senderNodeRow := rs.QueryExecutor.ExecuteSelectRow(senderNodeQ, senderNodeArgs...)
	err = rs.NodeRegistrationQuery.Scan(&senderNodeRegistration, senderNodeRow)
	if err != nil {
		return err
	}

	// get recipient address at height
	recipientNodeQ, recipientNodeArgs := rs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(
		receipt.RecipientPublicKey,
		receipt.ReferenceBlockHeight,
	)
	recipientNodeRow := rs.QueryExecutor.ExecuteSelectRow(recipientNodeQ, recipientNodeArgs...)
	err = rs.NodeRegistrationQuery.Scan(&recipientNodeRegistration, recipientNodeRow)
	if err != nil {
		return err
	}
	recipientFullAddress := fmt.Sprintf(
		"%s:%d", recipientNodeRegistration.NodeAddress.Address, recipientNodeRegistration.NodeAddress.Port)
	// get or build scrambled nodes at height
	scrambledNodes, err := rs.NodeRegistrationService.GetScrambleNodesByHeight(receipt.ReferenceBlockHeight)
	if err != nil {
		return err
	}
	// get priority peer of sender from scrambledNodes
	peers, err := p2pUtil.GetPriorityPeersByNodeFullAddress(
		fmt.Sprintf("%s:%d", senderNodeRegistration.NodeAddress.Address, senderNodeRegistration.NodeAddress.Port),
		scrambledNodes,
	)
	if err != nil {
		return err
	}
	// check if recipient is in sender.Peers list
	for _, peer := range peers {
		if p2pUtil.GetFullAddressPeer(peer) == recipientFullAddress {
			// valid recipient and sender
			return nil
		}
	}
	return blocker.NewBlocker(blocker.ValidationErr, "InvalidReceiptSenderOrRecipient")
}

/*
PruningNodeReceipts will pruning the receipts that was expired by block_height + minimum rollback block, affected:
	1. NodeReceipt
	2. MerkleTree
*/
func (rs *ReceiptService) PruningNodeReceipts() error {
	var (
		removeReceiptQ, removeMerkleQ string
		err, rollbackErr              error
		lastBlock                     model.Block
		row                           *sql.Row
	)

	row = rs.QueryExecutor.ExecuteSelectRow(rs.BlockQuery.GetLastBlock())
	err = rs.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return err
	}

	removeReceiptQ = rs.NodeReceiptQuery.RemoveReceipts(
		lastBlock.GetHeight()+constant.NodeReceiptExpiryBlockHeight+constant.MinRollbackBlocks,
		constant.MinRollbackBlocks,
	)
	removeMerkleQ = rs.MerkleTreeQuery.RemoveMerkleTrees(
		lastBlock.GetHeight()+constant.NodeReceiptExpiryBlockHeight+constant.MinRollbackBlocks,
		constant.MinRollbackBlocks,
	)
	err = rs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	err = rs.QueryExecutor.ExecuteTransaction(removeReceiptQ)
	if err != nil {
		rollbackErr = rs.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	err = rs.QueryExecutor.ExecuteTransaction(removeMerkleQ)
	if err != nil {
		rollbackErr = rs.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	err = rs.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}
