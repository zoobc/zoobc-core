package service

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"math"
	"sort"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptServiceInterface interface {
		Initialize() error
		SelectReceipts(previousBlock *model.Block) ([]*model.PublishedReceipt, []*model.PublishedReceipt, error)
		GenerateReceiptsMerkleRoot() error
		// ValidateReceipt to validating *model.BatchReceipt when send block or send transaction and also when want to publishing receipt
		ValidateReceipt(
			receipt *model.Receipt,
		) error
		GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		// GenerateReceiptWithReminder generating batch receipt and store to reminder also
		GenerateReceipt(
			ct chaintype.ChainType,
			receivedDatumHash []byte,
			lastBlock *storage.BlockCacheObject,
			senderPublicKey []byte,
			nodeSecretPhrase string,
			datumType uint32,
		) (*model.Receipt, error)
		// CheckDuplication to check duplication of *model.BatchReceipt when get response from send block and send transaction
		CheckDuplication(publicKey []byte, datumHash []byte) (err error)
		StoreReceipt(receipt *model.Receipt, chaintype chaintype.ChainType) error
		ClearCache()
		SaveReceiptAndMerkle(receiptBatchObject storage.ReceiptBatchObject) error
		GetReceiptFromPool(hash []byte) ([]model.Receipt, error)
	}

	ReceiptService struct {
		BatchReceiptQuery            query.BatchReceiptQueryInterface
		MerkleTreeQuery              query.MerkleTreeQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		BlockQuery                   query.BlockQueryInterface
		QueryExecutor                query.ExecutorInterface
		NodeRegistrationService      NodeRegistrationServiceInterface
		Signature                    crypto.SignatureInterface
		PublishedReceiptQuery        query.PublishedReceiptQueryInterface
		ReceiptUtil                  coreUtil.ReceiptUtilInterface
		MainBlockStateStorage        storage.CacheStorageInterface
		ScrambleNodeService          ScrambleNodeServiceInterface
		ProvedReceiptReminderStorage storage.CacheStorageInterface
		ReceiptPoolCacheStorage      storage.CacheStorageInterface
		ReceiptBatchStorage          storage.CacheStackStorageInterface
		MainBlocksStorage            storage.CacheStackStorageInterface
		randomNumberGenerator        *crypto.RandomNumberGenerator
		// local cache
		LastMerkleRoot []byte
	}
)

func NewReceiptService(
	nodeReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	blockQuery query.BlockQueryInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	signature crypto.SignatureInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	mainBlockStateStorage, provedReceiptReminderStorage, receiptPoolCacheStorage storage.CacheStorageInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
	mainBlocksStorage, receiptBatchStorage storage.CacheStackStorageInterface,
	randomNumberGenerator *crypto.RandomNumberGenerator,
) *ReceiptService {
	return &ReceiptService{
		BatchReceiptQuery:            nodeReceiptQuery,
		MerkleTreeQuery:              merkleTreeQuery,
		NodeRegistrationQuery:        nodeRegistrationQuery,
		BlockQuery:                   blockQuery,
		QueryExecutor:                queryExecutor,
		NodeRegistrationService:      nodeRegistrationService,
		Signature:                    signature,
		PublishedReceiptQuery:        publishedReceiptQuery,
		ReceiptUtil:                  receiptUtil,
		MainBlockStateStorage:        mainBlockStateStorage,
		ScrambleNodeService:          scrambleNodeService,
		ProvedReceiptReminderStorage: provedReceiptReminderStorage,
		ReceiptPoolCacheStorage:      receiptPoolCacheStorage,
		ReceiptBatchStorage:          receiptBatchStorage,
		MainBlocksStorage:            mainBlocksStorage,
		randomNumberGenerator:        randomNumberGenerator,
		LastMerkleRoot:               nil,
	}
}

func (rs *ReceiptService) Initialize() error {
	lastRmrQ := rs.MerkleTreeQuery.GetLastMerkleRoot()
	row, _ := rs.QueryExecutor.ExecuteSelectRow(lastRmrQ, false)

	lastMerkleRoot, err := rs.MerkleTreeQuery.ScanRoot(row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}
	rs.LastMerkleRoot = lastMerkleRoot
	return nil
}

// SelectReceipts select list of receipts to be included in a block by, the receipt will be separated to 2 categories
// - free receipts, whatever receipts goes in (current.height - 40) blocks, by selecting random batch to include
// - proved receipts, receipts that
func (rs *ReceiptService) SelectReceipts(previousBlock *model.Block) ([]*model.PublishedReceipt, []*model.PublishedReceipt, error) {
	var (
		freeBatch                    storage.ReceiptBatchObject
		freeReceipts, provedReceipts = make([]*model.PublishedReceipt, 0), make([]*model.PublishedReceipt, 0)
		err                          error
	)
	if previousBlock.GetHeight() <= constant.MaxReceiptBatchCacheRound {
		return freeReceipts, provedReceipts, err
	}
	err = rs.randomNumberGenerator.Reset(constant.ReceiptSelectionSeedPrefix, previousBlock.GetBlockSeed())
	if err != nil {
		return freeReceipts, provedReceipts, err
	}
	randomNumber := rs.randomNumberGenerator.Next()
	batchIndex := int(math.Floor(float64(randomNumber) / float64(len(freeBatch.ReceiptBatch))))
	// choose free receipts
	err = rs.ReceiptBatchStorage.GetAtIndex(uint32(0), &freeBatch)
	if err != nil {
		return freeReceipts, provedReceipts, err
	}
	var (
		merkleRoot util.MerkleRoot
	)
	merkleRoot.HashTree = merkleRoot.FromBytes(freeBatch.MerkleTree, freeBatch.MerkleRoot)
	getReceiptIntermediateHashes := func(rc model.Receipt, leafIndex int32) []byte {
		var (
			intermediateHashes [][]byte
		)
		intermediateHashesBuffer := merkleRoot.GetIntermediateHashes(bytes.NewBuffer(rc.DatumHash), leafIndex)
		for _, buf := range intermediateHashesBuffer {
			intermediateHashes = append(intermediateHashes, buf.Bytes())
		}
		return merkleRoot.FlattenIntermediateHashes(intermediateHashes)
	}
	for i := 0; i < len(freeBatch.ReceiptBatch[batchIndex]); i++ {
		bReceipt := freeBatch.ReceiptBatch[batchIndex][i]
		leafIndex := (batchIndex * i) + i
		pReceipt := &model.PublishedReceipt{
			Receipt:                   &bReceipt,
			IntermediateHashes:        getReceiptIntermediateHashes(bReceipt, int32(leafIndex)),
			BlockHeight:               previousBlock.GetHeight() + 1,
			BatchReferenceBlockHeight: bReceipt.ReferenceBlockHeight,
			ReceiptIndex:              uint32(leafIndex),
			PublishedIndex:            uint32(i),
			PublishedReceiptType:      model.PublishedReceiptType_FreeReceipt,
		}
		freeReceipts = append(freeReceipts, pReceipt)
	}
	// choose proved receipts
	var (
		provedReceiptReminders = make(map[uint32]model.PublishedReceipt)
	)
	err = rs.ProvedReceiptReminderStorage.GetAllItems(&provedReceiptReminders)
	if err != nil {
		return freeReceipts, provedReceipts, err
	}
	// fetch proved reminders
	return freeReceipts, provedReceipts, err
}

func (rs *ReceiptService) GetReceiptFromPool(hash []byte) ([]model.Receipt, error) {
	var (
		result []model.Receipt
		err    error
	)
	hashHex := hex.EncodeToString(hash)
	err = rs.ReceiptPoolCacheStorage.GetItem(hashHex, &result)
	if result == nil {
		result = make([]model.Receipt, 0)
	}
	return result, err
}

// SaveReceiptAndMerkle save receipts and its generated merkle root to database and memory
func (rs *ReceiptService) SaveReceiptAndMerkle(receiptBatchObject storage.ReceiptBatchObject) error {
	if len(receiptBatchObject.ReceiptBatch) == 0 || len(receiptBatchObject.ReceiptBatch[0]) == 0 {
		// we don't need to process empty receipts
		return nil
	}
	var (
		merkleRoot   util.MerkleRoot
		receiptCount = len(receiptBatchObject.ReceiptBatch) * len(receiptBatchObject.ReceiptBatch[0])
		merkleLeafs  = make([]*bytes.Buffer, 0, receiptCount)
		queries      = make([][]interface{}, receiptCount+1)
		err          error
	)
	for i := 0; i < len(receiptBatchObject.ReceiptBatch); i++ {
		for j := 0; j < len(receiptBatchObject.ReceiptBatch[i]); j++ {
			rcHash := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(&(receiptBatchObject.ReceiptBatch[i][j])))
			merkleLeafs = append(merkleLeafs, bytes.NewBuffer(rcHash[:]))
		}
	}
	_, err = merkleRoot.GenerateMerkleRoot(merkleLeafs)
	if err != nil {
		return err
	}
	root, merkleTree := merkleRoot.ToBytes()
	receiptBatchObject.MerkleRoot = root
	for i := 0; i < len(receiptBatchObject.ReceiptBatch); i++ {
		for j := 0; j < len(receiptBatchObject.ReceiptBatch[i]); j++ {
			batchReceipt := &model.BatchReceipt{
				Receipt:  &(receiptBatchObject.ReceiptBatch[i][j]),
				RMR:      receiptBatchObject.MerkleRoot,
				RMRIndex: uint32(i*j + j),
			}
			insertNodeReceiptQ, insertNodeReceiptArgs := rs.BatchReceiptQuery.InsertReceipt(batchReceipt)
			queries[i*j+j] = append([]interface{}{insertNodeReceiptQ}, insertNodeReceiptArgs...)
		}
	}
	insertMerkleTreeQ, insertMerkleTreeArgs := rs.MerkleTreeQuery.InsertMerkleTree(
		receiptBatchObject.MerkleRoot,
		merkleTree,
		time.Now().Unix(),
		receiptBatchObject.BlockHeight,
	)
	queries[len(queries)-1] = append([]interface{}{insertMerkleTreeQ}, insertMerkleTreeArgs...)

	err = rs.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	rs.LastMerkleRoot = receiptBatchObject.MerkleRoot // update local cache
	receiptBatchObject.MerkleTree = merkleTree
	return rs.ReceiptBatchStorage.Push(receiptBatchObject)
}

// GenerateReceiptsMerkleRoot generate merkle root of some batch receipts and also remove from cache
// generating will do when number of collected receipts(batch receipts) already <= the number of required
func (rs *ReceiptService) GenerateReceiptsMerkleRoot() error {
	var (
		receiptsCached, receipts []model.Receipt
		hashedReceipts           []*bytes.Buffer
		merkleRoot               util.MerkleRoot
		queries                  [][]interface{}
		batchReceipt             *model.BatchReceipt
		block                    model.Block
		err                      error
	)

	err = rs.ReceiptPoolCacheStorage.GetAllItems(&receiptsCached)
	if err != nil {
		return err
	}

	if len(receiptsCached) >= int(constant.ReceiptBatchMaximum) {
		// Need to sorting before do next
		sort.SliceStable(receiptsCached, func(i, j int) bool {
			return receiptsCached[i].ReferenceBlockHeight < receiptsCached[j].ReferenceBlockHeight
		})

		var cacheCount int
		for _, receipt := range receiptsCached {
			if len(receipts) == int(constant.ReceiptBatchMaximum) {
				break
			}
			b := receipt
			err = rs.ValidateReceipt(&b)
			if err == nil {
				receipts = append(receipts, b)
				hashedReceipt := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(&b))
				hashedReceipts = append(hashedReceipts, bytes.NewBuffer(hashedReceipt[:]))
			}
			cacheCount++
		}
		receiptsCached = receiptsCached[cacheCount:]

		_, err = merkleRoot.GenerateMerkleRoot(hashedReceipts)
		if err != nil {
			return err
		}
		rootMerkle, treeMerkle := merkleRoot.ToBytes()

		queries = make([][]interface{}, len(hashedReceipts)+1)
		for k, receipt := range receipts {
			b := receipt
			batchReceipt = &model.BatchReceipt{
				Receipt:  &b,
				RMR:      rootMerkle,
				RMRIndex: uint32(k),
			}
			insertNodeReceiptQ, insertNodeReceiptArgs := rs.BatchReceiptQuery.InsertReceipt(batchReceipt)
			queries[k] = append([]interface{}{insertNodeReceiptQ}, insertNodeReceiptArgs...)
		}
		err = rs.MainBlockStateStorage.GetItem(nil, &block)
		if err != nil {
			return err
		}
		insertMerkleTreeQ, insertMerkleTreeArgs := rs.MerkleTreeQuery.InsertMerkleTree(
			rootMerkle,
			treeMerkle,
			time.Now().Unix(),
			block.Height,
		)
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
		rs.LastMerkleRoot = rootMerkle // update local cache
		return rs.ReceiptPoolCacheStorage.SetItems(receiptsCached)
	}

	return nil
}

// CheckDuplication check existing batch receipt in cache storage
func (rs *ReceiptService) CheckDuplication(publicKey, datumHash []byte) (err error) {
	var (
		receiptKey []byte
		cType      chaintype.ChainType
	)
	if len(publicKey) == 0 && len(datumHash) == 0 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"EmptyParams",
		)
	}
	receiptKey, err = rs.ReceiptUtil.GetReceiptKey(datumHash, publicKey)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			err.Error(),
		)
	}

	err = rs.ProvedReceiptReminderStorage.GetItem(hex.EncodeToString(receiptKey), &cType)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailedGetReceiptKey",
		)
	}
	if cType != nil {
		return blocker.NewBlocker(blocker.DuplicateReceiptErr, "ReceiptExistsOnReminder")
	}
	return nil
}

func (rs *ReceiptService) ValidateReceipt(
	receipt *model.Receipt,
) error {
	var (
		blockAtHeight *storage.BlockCacheObject
		err           error
	)
	unsignedBytes := rs.ReceiptUtil.GetUnsignedReceiptBytes(receipt)
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
	blockAtHeight, err = util.GetBlockByHeightUseBlocksCache(
		receipt.ReferenceBlockHeight,
		rs.QueryExecutor,
		rs.BlockQuery,
		rs.MainBlocksStorage,
	)
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
	receipt *model.Receipt,
) error {
	var (
		err   error
		peers map[string]*model.Peer
	)
	// get or build scrambled nodes at height
	scrambledNode, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(receipt.ReferenceBlockHeight)
	if err != nil {
		return err
	}
	// get sender address at height
	senderNodeID, ok := scrambledNode.NodePublicKeyToIDMap[hex.EncodeToString(receipt.GetSenderPublicKey())]
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptSenderNotInScrambleList")
	}
	// get recipient address at height
	recipientNodeID, ok := scrambledNode.NodePublicKeyToIDMap[hex.EncodeToString(receipt.GetRecipientPublicKey())]
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptRecipientNotInScrambleList")
	}
	if peers, err = p2pUtil.GetPriorityPeersByNodeID(
		senderNodeID,
		scrambledNode,
	); err != nil {
		return err
	}

	// check if recipient is in sender.Peers list
	for _, peer := range peers {
		if peer.GetInfo().ID == recipientNodeID {
			// valid recipient and sender
			return nil
		}
	}
	return blocker.NewBlocker(blocker.ValidationErr, "ReceiptRecipientNotInPriorityList")
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

// GenerateReceiptWithReminder generate batch receipt at last block and store into batch receipt storage
func (rs *ReceiptService) GenerateReceipt(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *storage.BlockCacheObject,
	senderPublicKey []byte,
	nodeSecretPhrase string,
	datumType uint32,
) (*model.Receipt, error) {
	var (
		rmrLinked     = rs.LastMerkleRoot
		receipt       *model.Receipt
		err           error
		nodePublicKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSecretPhrase)
	)

	// generate receipt
	receipt, err = rs.ReceiptUtil.GenerateReceipt(
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
	receipt.RecipientSignature = rs.Signature.SignByNode(
		rs.ReceiptUtil.GetUnsignedReceiptBytes(receipt),
		nodeSecretPhrase,
	)
	return receipt, err
}

func (rs *ReceiptService) StoreReceipt(receipt *model.Receipt, chaintype chaintype.ChainType) error {
	b := *receipt
	err := rs.ReceiptPoolCacheStorage.SetItem(hex.EncodeToString(receipt.DatumHash), b)
	if err != nil {
		return err
	}
	err = rs.ProvedReceiptReminderStorage.SetItem(hex.EncodeToString(receipt.DatumHash), chaintype)
	return err
}

func (rs *ReceiptService) ClearCache() {
	_ = rs.ReceiptPoolCacheStorage.ClearCache()
}
