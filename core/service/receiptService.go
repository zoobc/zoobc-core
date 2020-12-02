package service

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
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
		SelectReceipts(
			previousBlock *model.Block,
			currentBlockSeed []byte,
			maxReceipt int,
		) ([]*model.PublishedReceipt, []*model.PublishedReceipt, error)
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
		StoreReceipt(receipt *model.Receipt) error
		ClearCache()
		SaveReceiptAndMerkle(receiptBatchObject storage.ReceiptBatchObject) error
		GetReceiptFromPool(payloadHash []byte) ([]model.Receipt, error)
		GetReceipByRootAndDatumHash(merkleRoot []byte, datumHash []byte) ([]*model.BatchReceipt, error)
		IsProvedReceiptEmpty(receipt *model.PublishedReceipt) bool
		GetMerkleRootFromReceiptIntermediateHash(
			receipt *model.Receipt,
			intermediateHashes []byte,
		) ([]byte, error)
	}

	ReceiptService struct {
		BatchReceiptQuery            query.BatchReceiptQueryInterface
		MerkleTreeQuery              query.MerkleTreeQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		BlockQuery                   query.BlockQueryInterface
		QueryExecutor                query.ExecutorInterface
		TransactionCoreService       TransactionCoreServiceInterface
		NodeRegistrationService      NodeRegistrationServiceInterface
		NodeConfigurationService     NodeConfigurationServiceInterface
		Signature                    crypto.SignatureInterface
		PublishedReceiptQuery        query.PublishedReceiptQueryInterface
		ReceiptUtil                  coreUtil.ReceiptUtilInterface
		MainBlockStateStorage        storage.CacheStorageInterface
		ScrambleNodeService          ScrambleNodeServiceInterface
		ProvedReceiptReminderStorage storage.CacheStackStorageInterface
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
	transactionCoreService TransactionCoreServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	signature crypto.SignatureInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	mainBlockStateStorage, receiptPoolCacheStorage storage.CacheStorageInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
	nodeConfigurationService NodeConfigurationServiceInterface,
	mainBlocksStorage, receiptBatchStorage, provedReceiptReminderStorage storage.CacheStackStorageInterface,
	randomNumberGenerator *crypto.RandomNumberGenerator,
) *ReceiptService {
	return &ReceiptService{
		BatchReceiptQuery:            nodeReceiptQuery,
		MerkleTreeQuery:              merkleTreeQuery,
		NodeRegistrationQuery:        nodeRegistrationQuery,
		BlockQuery:                   blockQuery,
		QueryExecutor:                queryExecutor,
		TransactionCoreService:       transactionCoreService,
		NodeRegistrationService:      nodeRegistrationService,
		Signature:                    signature,
		PublishedReceiptQuery:        publishedReceiptQuery,
		ReceiptUtil:                  receiptUtil,
		MainBlockStateStorage:        mainBlockStateStorage,
		ScrambleNodeService:          scrambleNodeService,
		NodeConfigurationService:     nodeConfigurationService,
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

func (rs *ReceiptService) GetReceipByRootAndDatumHash(merkleRoot []byte, datumHash []byte) ([]*model.BatchReceipt, error) {
	var (
		batchReceipts = make([]*model.BatchReceipt, 0)
		err           error
	)
	// fetch batch_receipt where merkle_root == provedReceiptRO.MerkleRoot
	qry, args := rs.BatchReceiptQuery.GetReceiptByRootAndDatumHash(merkleRoot, datumHash)
	rows, err := rs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return batchReceipts, err
	}
	defer rows.Close()
	return rs.BatchReceiptQuery.BuildModel(batchReceipts, rows)
}

func (rs *ReceiptService) getFreeReceipts(previousBlock *model.Block, currentBlockSeed []byte) ([]*model.PublishedReceipt, error) {
	var (
		allBatch = make([]storage.ReceiptBatchObject, 0)
		result   = make([]*model.PublishedReceipt, 0)
		err      error
	)

	err = rs.randomNumberGenerator.Reset(constant.BlocksmithSelectionFreeReceiptSeedPrefix, currentBlockSeed)
	if err != nil {
		return result, err
	}
	err = rs.ReceiptBatchStorage.GetAll(&allBatch)
	if err != nil {
		return result, err
	}

	// todo: delete this dummy section for logging
	var arr = make([]int, 0)
	for _, batch := range allBatch {
		arr = append(arr, len(batch.ReceiptBatch[0]))
	}
	fmt.Printf("height: %d - freeReceiptCount: %d: [%v]\n", previousBlock.GetHeight(), len(allBatch), arr)
	// todo: delete this dummy section for logging

	if len(allBatch) < constant.MaxReceiptBatchCacheRound {
		return result, blocker.NewBlocker(blocker.CacheEmpty,
			fmt.Sprintf("NoEnoughBatchReceipt-minimum: %d\tsupplied: %d\n", constant.MaxReceiptBatchCacheRound,
				len(allBatch)))
	}
	// choose free receipts

	fmt.Printf("height: %d - freeReceipt-%d-count: %d\n", previousBlock.Height, previousBlock.GetHeight()-4, len(allBatch[0].ReceiptBatch))
	if len(allBatch[0].ReceiptBatch) == 0 {
		return result, blocker.NewBlocker(blocker.CacheEmpty, "NoBatchReceipt")
	}
	randomNumber := rs.randomNumberGenerator.Next()

	batchIndex := rs.randomNumberGenerator.ConvertRandomNumberToIndex(randomNumber, int64(len(allBatch[0].ReceiptBatch)))

	for i := 0; i < len(allBatch[0].ReceiptBatch[batchIndex]); i++ {
		bReceipt := allBatch[0].ReceiptBatch[batchIndex][i]
		leafIndex := (int(batchIndex) * i) + i
		pReceipt := &model.PublishedReceipt{
			Receipt:                   &bReceipt,
			IntermediateHashes:        nil, // no intermediate hashes for free receipts
			BlockHeight:               previousBlock.GetHeight() + 1,
			BatchReferenceBlockHeight: bReceipt.ReferenceBlockHeight,
			ReceiptIndex:              uint32(leafIndex),
			PublishedIndex:            uint32(i),
			PublishedReceiptType:      model.PublishedReceiptType_FreeReceipt,
		}
		result = append(result, pReceipt)
	}
	return result, nil
}

func (rs *ReceiptService) getProvedReceipts(
	previousBlock *model.Block,
	currentBlockSeed []byte,
	maxReceipt int,
) ([]*model.PublishedReceipt, error) {
	// choose proved receipts
	var (
		result                 = make([]*model.PublishedReceipt, 0)
		provedReceiptReminders = make([]storage.ProvedReceiptReminderObject, 0)
		err                    error
	)
	err = rs.ProvedReceiptReminderStorage.GetAll(&provedReceiptReminders)
	fmt.Printf("getProvedReceipt: err: %v\tcount: %v\n", err, len(provedReceiptReminders))

	if err != nil {
		return result, err
	}
	if len(provedReceiptReminders) < maxReceipt {
		return result, blocker.NewBlocker(blocker.InsufficientError,
			fmt.Sprintf("SelectReceipts-InsufficientProvedReceipt - required: %d\thave: %d",
				maxReceipt,
				len(provedReceiptReminders)),
		)
	}

	fetchMerkleTree := func(merkleRoot []byte) ([]byte, error) {
		root, args := rs.MerkleTreeQuery.GetMerkleTreeByRoot(merkleRoot)
		row, err := rs.QueryExecutor.ExecuteSelectRow(root, false, args...)
		if err != nil {
			return nil, err
		}
		return rs.MerkleTreeQuery.ScanTree(row)
	}
	rng := crypto.NewRandomNumberGenerator()
	rng.Reset(constant.BlocksmithSelectionProvedReceiptSeedPrefix, currentBlockSeed)
	emptyProvedReceipt := &model.PublishedReceipt{
		Receipt:                   nil,
		IntermediateHashes:        []byte{},
		BlockHeight:               previousBlock.GetHeight() + 1,
		BatchReferenceBlockHeight: 0,
		ReceiptIndex:              0,
		PublishedIndex:            uint32(len(result)),
		PublishedReceiptType:      model.PublishedReceiptType_ProvedReceipt,
	}
	hostID, err := rs.NodeConfigurationService.GetHostID()
	if err != nil {
		return result, err
	}
	// fetch proved reminders
	var count int
	for _, provedReceiptRO := range provedReceiptReminders {
		count++
		// generate random number (consensus safe) as to which receipt to pick
		itemIndexRandomNumber := rng.Next()
		leafIndexRandomNumber := rng.Next()
		// if provedReceiptRO.MerkleRoot = []byte{} / empty bytes, then it means we are in the scramble at the height
		// but not getting reference receipt published, so skipped
		if len(provedReceiptRO.MerkleRoot) == 0 {
			// keep filling to proved receipt list even if we don't have it, this is to keep the rng in consensus
			// to the receipt list index
			result = append(result, emptyProvedReceipt)
			continue
		}

		blockAtHeight, err := util.GetBlockByHeightUseBlocksCache(
			provedReceiptRO.ReferenceBlockHeight-1,
			rs.QueryExecutor,
			rs.BlockQuery,
			rs.MainBlocksStorage,
		)
		if err != nil {
			fmt.Printf("erorr: getblockByMerkleRoot: %v\n\n\n\n", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		txsAtHeight, err := rs.TransactionCoreService.GetTransactionsByBlockHeight(provedReceiptRO.ReferenceBlockHeight)
		if err != nil {
			fmt.Printf("erorr: GetTransactionsByBlockHeight: %v\n\n\n\n", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		itemIndex := rng.ConvertRandomNumberToIndex(itemIndexRandomNumber, int64(len(txsAtHeight)+1))
		// pick receipt and fetch its intermediate hashes
		var (
			itemHash []byte
		)
		if itemIndex == 0 {
			itemHash = blockAtHeight.BlockHash
		} else {
			itemHash = txsAtHeight[itemIndex-1].TransactionHash
		}
		merkleItems, err := rs.GetReceipByRootAndDatumHash(provedReceiptRO.MerkleRoot, itemHash)
		if err != nil || len(merkleItems) == 0 {
			// log error
			if err == nil {
				fmt.Printf(
					"no receipt with root:datum hash: %v:%v\n",
					hex.EncodeToString(provedReceiptRO.MerkleRoot),
					hex.EncodeToString(itemHash),
				)
			}
			fmt.Printf("%v", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		scrambleAtHeight, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(blockAtHeight.Height)
		if err != nil {
			fmt.Printf("erorr: GetScrambleNodesByHeight: %v\n\n\n\n", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		_, sortedPriorityAtHeight, err := p2pUtil.GetPriorityPeersByNodeID(hostID, scrambleAtHeight)
		if err != nil {
			fmt.Printf("%v", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		receiverIndex := rng.ConvertRandomNumberToIndex(leafIndexRandomNumber, int64(len(sortedPriorityAtHeight)))
		if int(receiverIndex) >= len(merkleItems) {
			fmt.Printf("erorr: ConvertRandomNumberToIndex: receiverIndex > merkleItems %v > %v\n\n\n\n",
				receiverIndex, len(merkleItems),
			)

			result = append(result, emptyProvedReceipt)
			continue
		}
		leaf := merkleItems[receiverIndex]

		tree, err := fetchMerkleTree(provedReceiptRO.MerkleRoot)
		if err != nil {
			fmt.Printf("err: fetchMerkleTree: %v\n", err)
			result = append(result, emptyProvedReceipt)
			continue
		}
		intermediateHashes := rs.getReceiptIntermediateHash(
			*leaf.GetReceipt(),
			int32(receiverIndex),
			provedReceiptRO.MerkleRoot,
			tree,
		)
		pReceipt := &model.PublishedReceipt{
			Receipt:                   leaf.GetReceipt(),
			IntermediateHashes:        intermediateHashes,
			BlockHeight:               previousBlock.GetHeight() + 1,
			BatchReferenceBlockHeight: blockAtHeight.Height,
			ReceiptIndex:              uint32(receiverIndex),
			PublishedIndex:            uint32(len(result)),
			PublishedReceiptType:      model.PublishedReceiptType_ProvedReceipt,
		}
		result = append(result, pReceipt)
	}
	return result, nil
}

func (rs *ReceiptService) getReceiptIntermediateHash(rc model.Receipt, leafIndex int32, root, merkleTree []byte) []byte {
	var (
		intermediateHashes [][]byte
		merkleRoot         util.MerkleRoot
	)
	merkleRoot.HashTree = merkleRoot.FromBytes(merkleTree, root)
	rcHash := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(&rc))
	intermediateHashesBuffer := merkleRoot.GetIntermediateHashes(bytes.NewBuffer(rcHash[:]), leafIndex)
	for _, buf := range intermediateHashesBuffer {
		intermediateHashes = append(intermediateHashes, buf.Bytes())
	}
	return merkleRoot.FlattenIntermediateHashes(intermediateHashes)
}

// SelectReceipts select list of receipts to be included in a block by, the receipt will be separated to 2 categories
// - free receipts, whatever receipts goes in (current.height - 40) blocks, by selecting random batch to include
// - proved receipts, receipts that can be linked back to past blocksmith's block including one / more of our receipt's
// merkle root.
// failure in this function should not stop the node to generate block, since it's not consensus rule but independent node's
// receipt collection.
func (rs *ReceiptService) SelectReceipts(
	previousBlock *model.Block,
	currentBlockSeed []byte,
	maxReceipt int,
) ([]*model.PublishedReceipt, []*model.PublishedReceipt, error) {
	var (
		freeReceipts, provedReceipts = make([]*model.PublishedReceipt, 0), make([]*model.PublishedReceipt, 0)
		err                          error
	)
	if previousBlock.GetHeight() <= constant.MaxReceiptBatchCacheRound {
		return freeReceipts, provedReceipts, err
	}
	freeReceipts, err = rs.getFreeReceipts(previousBlock, currentBlockSeed)
	if err != nil {
		// todo: log only, continue looking for proved receipt
		freeReceipts = make([]*model.PublishedReceipt, 0)
	}
	provedReceipts, err = rs.getProvedReceipts(previousBlock, currentBlockSeed, maxReceipt)
	if err != nil {
		// todo: log only, continue returning empty receipt
		provedReceipts = make([]*model.PublishedReceipt, 0)
	}
	return freeReceipts, provedReceipts, err
}

func (rs *ReceiptService) GetReceiptFromPool(payloadHash []byte) ([]model.Receipt, error) {
	var (
		result []model.Receipt
		err    error
	)
	payloadHashHex := hex.EncodeToString(payloadHash)
	err = rs.ReceiptPoolCacheStorage.GetItem(payloadHashHex, &result)
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
		merkleTree   []byte
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
	receiptBatchObject.MerkleRoot, merkleTree = merkleRoot.ToBytes()
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
	fmt.Printf("merkle-root: %v\nmerkle-tree: %v\nmerkle-leafs: %v\n\n\n",
		hex.EncodeToString(receiptBatchObject.MerkleRoot),
		len(merkleTree),
		len(receiptBatchObject.ReceiptBatch[0]))
	err = rs.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	rs.LastMerkleRoot = receiptBatchObject.MerkleRoot // update local cache
	receiptBatchObject.MerkleTree = merkleTree
	return rs.ReceiptBatchStorage.Push(receiptBatchObject)
}

// CheckDuplication check existing batch receipt in cache storage
func (rs *ReceiptService) CheckDuplication(publicKey, datumHash []byte) (err error) {
	// var (
	// 	receiptKey []byte
	// 	cType      chaintype.ChainType
	// )
	// if len(publicKey) == 0 && len(datumHash) == 0 {
	// 	return blocker.NewBlocker(
	// 		blocker.ValidationErr,
	// 		"EmptyParams",
	// 	)
	// }
	// receiptKey, err = rs.ReceiptUtil.GetReceiptKey(datumHash, publicKey)
	// if err != nil {
	// 	return blocker.NewBlocker(
	// 		blocker.ValidationErr,
	// 		err.Error(),
	// 	)
	// }
	//
	// err = rs.ReceiptPoolCacheStorage.GetItem(hex.EncodeToString(receiptKey), &cType)
	// if err != nil {
	// 	return blocker.NewBlocker(
	// 		blocker.ValidationErr,
	// 		"FailedGetReceiptPool",
	// 	)
	// }
	// if cType != nil {
	// 	return blocker.NewBlocker(blocker.DuplicateReceiptErr, "ReceiptExistsOnReminder")
	// }
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
	if peers, _, err = p2pUtil.GetPriorityPeersByNodeID(
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

func (rs *ReceiptService) StoreReceipt(receipt *model.Receipt) error {
	b := *receipt
	err := rs.ReceiptPoolCacheStorage.SetItem(hex.EncodeToString(receipt.DatumHash), b)
	return err
}

func (*ReceiptService) IsProvedReceiptEmpty(receipt *model.PublishedReceipt) bool {
	if receipt.GetReceipt() == nil {
		return true
	}
	return false
}

func (rs *ReceiptService) ClearCache() {
	_ = rs.ReceiptPoolCacheStorage.ClearCache()
	_ = rs.ReceiptBatchStorage.Clear()
	_ = rs.ProvedReceiptReminderStorage.Clear()
}

func (rs *ReceiptService) GetMerkleRootFromReceiptIntermediateHash(
	receipt *model.Receipt,
	intermediateHashes []byte,
) ([]byte, error) {
	var (
		root []byte
		err  error
	)
	hash := sha3.New256()
	rcHash := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(receipt))
	hash.Write(rcHash[:])
	for i := 0; i < len(intermediateHashes)/constant.ReceiptHashSize; i++ {
		iHash := intermediateHashes[i*constant.ReceiptHashSize : (i+1)*constant.ReceiptHashSize]
		hash.Write(iHash)
	}
	root = hash.Sum([]byte{})
	return root, err
}
