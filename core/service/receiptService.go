package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptServiceInterface interface {
		SelectReceipts(
			blockTimestamp int64,
			numberOfReceipt uint32,
			lastBlockHeight uint32,
		) ([]*model.PublishedReceipt, error)
		GenerateReceiptsMerkleRoot() error
		ValidateReceipt(
			receipt *model.BatchReceipt,
		) error
		PruningNodeReceipts() error
		GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		GenerateBatchReceiptWithReminder(
			ct chaintype.ChainType,
			receivedDatumHash []byte,
			lastBlock *model.Block,
			senderPublicKey []byte,
			nodeSecretPhrase, receiptKey string,
			datumType uint32,
		) (*model.BatchReceipt, error)
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
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		ReceiptUtil             coreUtil.ReceiptUtilInterface
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
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
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
		PublishedReceiptQuery:   publishedReceiptQuery,
		ReceiptUtil:             receiptUtil,
	}
}

// SelectReceipts select list of receipts to be included in a block by prioritizing receipts that might
// increase the participation score of the node
func (rs *ReceiptService) SelectReceipts(
	blockTimestamp int64,
	numberOfReceipt, lastBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var (
		linkedReceiptList = make(map[string][]*model.Receipt)
		// this variable is to store picked receipt recipient to avoid duplicates
		pickedRecipients  = make(map[string]bool)
		lowerBlockHeight  uint32
		linkedReceiptTree = make(map[string][]byte)
	)

	if numberOfReceipt < 1 { // possible no connected node
		return []*model.PublishedReceipt{}, nil
	}
	// get the last merkle tree we have build so far
	if lastBlockHeight > constant.NodeReceiptExpiryBlockHeight {
		lowerBlockHeight = lastBlockHeight - constant.NodeReceiptExpiryBlockHeight
	}

	err := func() error {
		treeQ := rs.MerkleTreeQuery.SelectMerkleTree(
			lowerBlockHeight,
			lastBlockHeight,
			numberOfReceipt*constant.ReceiptBatchPickMultiplier)
		linkedTreeRows, err := rs.QueryExecutor.ExecuteSelect(treeQ, false)
		if err != nil {
			return err
		}
		defer linkedTreeRows.Close()

		linkedReceiptTree, err = rs.MerkleTreeQuery.BuildTree(linkedTreeRows)
		if err != nil {
			return err
		}
		return nil
	}()
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
			if len(results) >= int(numberOfReceipt) {
				break
			}
			err = rs.ValidateReceipt(rc.BatchReceipt)
			if err != nil {
				// skip invalid receipt
				continue
			}
			var intermediateHashes [][]byte
			rcByte := rs.ReceiptUtil.GetSignedBatchReceiptBytes(rc.BatchReceipt)
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
	if len(results) < int(numberOfReceipt) {
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
	numberOfReceipt uint32,
	pickedReceipts []*model.PublishedReceipt,
	pickedRecipients map[string]bool,
	lowerBlockHeight, upperBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var receipts []*model.Receipt
	receiptsQ := rs.NodeReceiptQuery.GetReceiptsWithUniqueRecipient(
		numberOfReceipt*constant.ReceiptBatchPickMultiplier, lowerBlockHeight, upperBlockHeight)
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
		if len(pickedReceipts) >= int(numberOfReceipt) {
			break
		}
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

	row, _ := rs.QueryExecutor.ExecuteSelectRow(
		query.GetTotalRecordOfSelect(getBatchReceiptsQ),
		false,
	)
	err = row.Scan(&count)
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
			hashedBatchReceipt := sha3.Sum256(rs.ReceiptUtil.GetSignedBatchReceiptBytes(b))
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
		lastBlockRow, _ := rs.QueryExecutor.ExecuteSelectRow(lastBlockQ, false)
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
	unsignedBytes := rs.ReceiptUtil.GetUnsignedBatchReceiptBytes(receipt)
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
	blockAtHeightRow, _ := rs.QueryExecutor.ExecuteSelectRow(blockAtHeightQ, false)
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
	senderNodeRow, _ := rs.QueryExecutor.ExecuteSelectRow(senderNodeQ, false, senderNodeArgs...)
	err = rs.NodeRegistrationQuery.Scan(&senderNodeRegistration, senderNodeRow)
	if err != nil {
		return err
	}

	// get recipient address at height
	recipientNodeQ, recipientNodeArgs := rs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(
		receipt.RecipientPublicKey,
		receipt.ReferenceBlockHeight,
	)
	recipientNodeRow, _ := rs.QueryExecutor.ExecuteSelectRow(recipientNodeQ, false, recipientNodeArgs...)
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
		removeReceiptArgs, removeMerkleArgs []interface{}
		removeReceiptQ, removeMerkleQ       string
		err, rollbackErr                    error
		lastBlock                           model.Block
		row                                 *sql.Row
	)

	row, _ = rs.QueryExecutor.ExecuteSelectRow(rs.BlockQuery.GetLastBlock(), false)
	err = rs.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return err
	}

	limiter := int(lastBlock.GetHeight()) - (constant.NodeReceiptExpiryBlockHeight + int(constant.MinRollbackBlocks))
	if limiter > 0 {
		removeReceiptQ, removeReceiptArgs = rs.NodeReceiptQuery.RemoveReceipts(
			uint32(limiter),
			constant.PruningChunkedSize,
		)
		removeMerkleQ, removeMerkleArgs = rs.MerkleTreeQuery.RemoveMerkleTrees(
			uint32(limiter),
			constant.PruningChunkedSize,
		)
		err = rs.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = rs.QueryExecutor.ExecuteTransaction(removeReceiptQ, removeReceiptArgs...)
		if err != nil {
			rollbackErr = rs.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
		err = rs.QueryExecutor.ExecuteTransaction(removeMerkleQ, removeMerkleArgs...)
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
	}
	return nil
}

// GetPublishedReceiptsByHeight that handling database connection to get published receipts by height
func (rs *ReceiptService) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var (
		publishedReceipts []*model.PublishedReceipt
		rows              *sql.Rows
		err               error
	)

	qStr, qArgs := rs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err = rs.QueryExecutor.ExecuteSelect(qStr, false, qArgs...)
	if err != nil {
		return publishedReceipts, err
	}
	defer rows.Close()

	publishedReceipts, err = rs.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return publishedReceipts, err
	}
	return publishedReceipts, nil
}

func (rs *ReceiptService) GenerateBatchReceiptWithReminder(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *model.Block,
	senderPublicKey []byte,
	nodeSecretPhrase, receiptKey string,
	datumType uint32,
) (*model.BatchReceipt, error) {
	var (
		rmrLinked     []byte
		batchReceipt  *model.BatchReceipt
		err           error
		merkleQuery   = query.NewMerkleTreeQuery()
		nodePublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSecretPhrase)
		lastRmrQ      = merkleQuery.GetLastMerkleRoot()
		row, _        = rs.QueryExecutor.ExecuteSelectRow(lastRmrQ, false)
	)

	rmrLinked, err = merkleQuery.ScanRoot(row)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// generate receipt
	batchReceipt, err = rs.ReceiptUtil.GenerateBatchReceipt(
		ct,
		lastBlock,
		senderPublicKey,
		nodePublicKey,
		receivedDatumHash,
		rmrLinked,
		datumType,
	)
	if err != nil {
		return nil, err
	}
	batchReceipt.RecipientSignature = rs.Signature.SignByNode(
		rs.ReceiptUtil.GetUnsignedBatchReceiptBytes(batchReceipt),
		nodeSecretPhrase,
	)
	// store the generated batch receipt hash for reminder
	err = rs.KVExecutor.Insert(receiptKey, receivedDatumHash, constant.KVdbExpiryReceiptReminder)
	if err != nil {
		return nil, err
	}
	return batchReceipt, nil
}
