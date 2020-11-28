package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/storage"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// BlockServiceSpineInterface interface that contains methods specific of BlockSpineService
	BlockServiceSpineInterface interface {
		ValidateSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error
	}

	BlockSpineService struct {
		sync.RWMutex
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		Signature                 crypto.SignatureInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlocksmithService         BlocksmithServiceInterface
		SnapshotMainBlockService  SnapshotBlockServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlocksStorage             storage.CacheStackStorageInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		MainBlockService          BlockServiceInterface
	}
)

func NewBlockSpineService(
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	spineBlockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	obsr *observer.Observer,
	blocksmithStrategy strategy.BlocksmithStrategyInterface,
	logger *log.Logger,
	megablockQuery query.SpineBlockManifestQueryInterface,
	blocksmithService BlocksmithServiceInterface,
	snapshotMainblockService SnapshotBlockServiceInterface,
	blockStateStorage storage.CacheStorageInterface,
	blockchainStatusService BlockchainStatusServiceInterface,
	spinePublicKeyService BlockSpinePublicKeyServiceInterface,
	mainBlockService BlockServiceInterface,
	blocksStorage storage.CacheStackStorageInterface,
) *BlockSpineService {
	return &BlockSpineService{
		Chaintype:             ct,
		QueryExecutor:         queryExecutor,
		BlockQuery:            spineBlockQuery,
		Signature:             signature,
		BlocksmithStrategy:    blocksmithStrategy,
		Observer:              obsr,
		Logger:                logger,
		SpinePublicKeyService: spinePublicKeyService,
		SpineBlockManifestService: NewSpineBlockManifestService(
			queryExecutor,
			megablockQuery,
			spineBlockQuery,
			logger,
		),
		BlocksmithService:        blocksmithService,
		SnapshotMainBlockService: snapshotMainblockService,
		BlockStateStorage:        blockStateStorage,
		BlocksStorage:            blocksStorage,
		BlockchainStatusService:  blockchainStatusService,
		MainBlockService:         mainBlockService,
	}
}

// NewSpineBlock generate new spinechain block
func (bs *BlockSpineService) NewSpineBlock(
	version uint32,
	previousBlockHash,
	blockSeed, blockSmithPublicKey, merkleRoot, merkleTree []byte,
	previousBlockHeight, referenceBlockHeight uint32,
	timestamp int64,
	secretPhrase string,
	spinePublicKeys []*model.SpinePublicKey,
	spineBlockManifests []*model.SpineBlockManifest,
) (*model.Block, error) {
	var (
		payloadLength uint32
		err           error
	)
	block := &model.Block{
		Version:              version,
		PreviousBlockHash:    previousBlockHash,
		BlockSeed:            blockSeed,
		BlocksmithPublicKey:  blockSmithPublicKey,
		Height:               previousBlockHeight,
		Timestamp:            timestamp,
		PayloadLength:        payloadLength,
		SpinePublicKeys:      spinePublicKeys,
		SpineBlockManifests:  spineBlockManifests,
		ReferenceBlockHeight: referenceBlockHeight,
		MerkleRoot:           merkleRoot,
		MerkleTree:           merkleTree,
	}

	// compute block's payload hash and length and add it to block struct
	if block.PayloadHash, block.PayloadLength, err = bs.GetPayloadHashAndLength(block); err != nil {
		return nil, err
	}

	blockUnsignedByte, err := commonUtils.GetBlockByte(block, false, bs.Chaintype)
	if err != nil {
		bs.Logger.Error(err.Error())
	}
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	blockHash, err := commonUtils.GetBlockHash(block, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// GetChainType returns the chaintype
func (bs *BlockSpineService) GetChainType() chaintype.ChainType {
	return bs.Chaintype
}

func (bs *BlockSpineService) GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface {
	return bs.BlocksmithStrategy
}

// ChainWriteLock locks the chain
func (bs *BlockSpineService) ChainWriteLock(actionType int) {
	monitoring.IncrementStatusLockCounter(bs.Chaintype, actionType)
	bs.Lock()
	monitoring.SetBlockchainStatus(bs.Chaintype, actionType)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockSpineService) ChainWriteUnlock(actionType int) {
	bs.Unlock()
	monitoring.DecrementStatusLockCounter(bs.Chaintype, actionType)
	monitoring.SetBlockchainStatus(bs.Chaintype, constant.BlockchainStatusIdle)
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockSpineService) NewGenesisBlock(
	version uint32,
	previousBlockHash, blockSeed, blockSmithPublicKey, merkleRoot, merkleTree []byte,
	previousBlockHeight, referenceBlockHeight uint32,
	timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction,
	spinePublicKeys []*model.SpinePublicKey,
	payloadHash []byte,
	payloadLength uint32,
	cumulativeDifficulty *big.Int,
	genesisSignature []byte,
) (*model.Block, error) {
	var (
		block = &model.Block{
			Version:              version,
			PreviousBlockHash:    previousBlockHash,
			BlockSeed:            blockSeed,
			BlocksmithPublicKey:  blockSmithPublicKey,
			Height:               previousBlockHeight,
			Timestamp:            timestamp,
			TotalAmount:          totalAmount,
			TotalFee:             totalFee,
			TotalCoinBase:        totalCoinBase,
			Transactions:         transactions,
			SpinePublicKeys:      spinePublicKeys,
			PayloadLength:        payloadLength,
			PayloadHash:          payloadHash,
			CumulativeDifficulty: cumulativeDifficulty.String(),
			BlockSignature:       genesisSignature,
			ReferenceBlockHeight: referenceBlockHeight,
			MerkleRoot:           merkleRoot,
			MerkleTree:           merkleTree,
		}

		blockHash, err = commonUtils.GetBlockHash(block, bs.Chaintype)
	)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// ValidatePayloadHash validate (computed) block's payload data hash against block's payload hash
func (bs *BlockSpineService) ValidatePayloadHash(block *model.Block) error {
	hash, length, err := bs.GetPayloadHashAndLength(block)
	if err != nil {
		return err
	}
	if length != block.GetPayloadLength() || !bytes.Equal(hash, block.GetPayloadHash()) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockPayload")
	}
	return nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockSpineService) ValidateBlock(block, previousLastBlock *model.Block) error {
	// validate block's payload data
	if err := bs.ValidatePayloadHash(block); err != nil {
		return err
	}

	err := bs.BlocksmithStrategy.IsBlockValid(previousLastBlock, block)
	if err != nil {
		return err
	}
	if coreUtil.GetBlockID(block, bs.Chaintype) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidID")
	}
	// Verify Signature
	blockByte, err := commonUtils.GetBlockByte(block, false, bs.Chaintype)
	if err != nil {
		return err
	}

	if !bs.Signature.VerifyNodeSignature(
		blockByte,
		block.BlockSignature,
		block.BlocksmithPublicKey,
	) {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSignature")
	}
	// Verify previous block hash
	previousBlockHash, err := commonUtils.GetBlockHash(previousLastBlock, bs.Chaintype)
	if err != nil {
		return err
	}
	if !bytes.Equal(previousBlockHash, block.PreviousBlockHash) {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidPreviousBlockHash")
	}
	// if the same block height is already in the database compare cummulative difficulty.
	if err := bs.validateBlockHeight(block); err != nil {
		return err
	}
	// check included main block
	err = bs.validateIncludedMainBlock(previousLastBlock, block)
	if err != nil {
		return err
	}
	return nil
}

// validateBlockAtHeight Check if the same block height is already in the database compare cummulative difficulty.
// and return error if current block's cumulative difficulty is lower than the one in db
func (bs *BlockSpineService) validateBlockHeight(block *model.Block) error {
	var (
		bl                                                 []*model.Block
		refCumulativeDifficulty, blockCumulativeDifficulty *big.Int
		ok                                                 bool
	)
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByHeight(block.Height), false)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	bl, err = bs.BlockQuery.BuildModel(bl, rows)
	if err != nil {
		return err
	}
	if len(bl) > 0 {
		refBlock := bl[0]
		if refCumulativeDifficulty, ok = new(big.Int).SetString(refBlock.CumulativeDifficulty, 10); !ok {
			return err
		}
		if blockCumulativeDifficulty, ok = new(big.Int).SetString(block.CumulativeDifficulty, 10); !ok {
			return err
		}

		// if cumulative difficulty of the reference block is > of the one of the (new) block, new block is invalid
		if refCumulativeDifficulty.Cmp(blockCumulativeDifficulty) > 0 {
			return blocker.NewBlocker(blocker.BlockErr, "InvalidCumulativeDifficulty")
		}
	}
	return nil
}

// PushBlock push block into blockchain, to broadcast the block after pushing to own node, switch the
// broadcast flag to `true`, and `false` otherwise
func (bs *BlockSpineService) PushBlock(previousBlock, block *model.Block, broadcast, persist bool) error {
	var (
		err error
	)
	if !coreUtil.IsGenesis(previousBlock.GetID(), block) {
		block.Height = previousBlock.GetHeight() + 1
		block.CumulativeDifficulty, err = bs.BlocksmithStrategy.CalculateCumulativeDifficulty(previousBlock, block)
		if err != nil {
			return blocker.NewBlocker(blocker.BlockErr, fmt.Sprintf("CalculateCumulativeDifficulty:%v", err))
		}
	}
	// start db transaction here
	err = bs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err = bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}

	// add new spine public keys (pub keys included in this spine block) into spinePublicKey table
	if err := bs.SpinePublicKeyService.InsertSpinePublicKeys(block); err != nil {
		bs.Logger.Error(err.Error())
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}

	// if present, add new spine block manifests into spineBlockManifest table
	for _, spineBlockManifest := range block.SpineBlockManifests {
		if err := bs.SpineBlockManifestService.InsertSpineBlockManifest(spineBlockManifest); err != nil {
			bs.Logger.Error(err.Error())
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}

	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	// cache last block state
	// Note: Make sure every time calling query insert & rollback block, calling this SetItem too
	err = bs.UpdateLastBlockCache(block)
	if err != nil {
		return err
	}
	// cache last block into blocks cache storage
	err = bs.BlocksStorage.Push(commonUtils.BlockConvertToCacheFormat(block))
	if err != nil {
		bs.Logger.Warnf("FailedPushBlocksStorageCache-%v", err)
		_ = bs.InitializeBlocksCache()
	}
	bs.Logger.Debugf("%s Block Pushed ID: %d", bs.Chaintype.GetName(), block.GetID())
	// broadcast block
	if broadcast {
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)

	bs.BlockchainStatusService.SetLastBlock(block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype, block)
	return nil
}

// GetBlockByID return a block by its ID
// withAttachedData if true returns extra attached data for the block (transactions)
func (bs *BlockSpineService) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	if id == 0 {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "block ID 0 is not found")
	}
	var (
		block    model.Block
		row, err = bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetBlockByID(id), false)
	)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if err = bs.BlockQuery.Scan(&block, row); err != nil {
		if err == sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, err.Error())
		}
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	if block.ID == 0 {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block %v is not found", id))
	}
	if withAttachedData {
		err := bs.PopulateBlockData(&block)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
	}
	return &block, nil
}

// GetBlocksFromHeight get all blocks from a given height till last block (or a given limit is reached).
// Note: this only returns main block data, it doesn't populate attached data (spinePublicKeys)
func (bs *BlockSpineService) GetBlocksFromHeight(startHeight, limit uint32, withAttachedData bool) ([]*model.Block, error) {
	bs.ChainWriteLock(constant.BlockchainStatusGettingBlocks)
	defer bs.ChainWriteUnlock(constant.BlockchainStatusGettingBlocks)

	var blocks []*model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockFromHeight(startHeight, limit), false)
	if err != nil {
		return []*model.Block{}, err
	}
	defer rows.Close()
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}
	if withAttachedData {
		for _, block := range blocks {
			err := bs.PopulateBlockData(block)
			if err != nil {
				return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
			}
		}
	}

	return blocks, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockSpineService) GetLastBlock() (*model.Block, error) {
	var (
		lastBlock model.Block
		err       = bs.BlockStateStorage.GetItem(nil, &lastBlock)
	)
	if err != nil {
		return nil, err
	}
	return &lastBlock, nil
}

func (bs *BlockSpineService) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	var (
		lastBlock storage.BlockCacheObject
		err       = bs.BlocksStorage.GetTop(&lastBlock)
	)
	if err != nil {
		return nil, err
	}
	return &lastBlock, nil
}

// GetBlockHash return block's hash (makes sure always include spine public keys)
func (bs *BlockSpineService) GetBlockHash(block *model.Block) ([]byte, error) {
	err := bs.PopulateBlockData(block)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return commonUtils.GetBlockHash(block, bs.GetChainType())

}

// GetBlockByHeight return the last pushed block
func (bs *BlockSpineService) GetBlockByHeight(height uint32) (*model.Block, error) {
	block, err := commonUtils.GetBlockByHeight(height, bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = bs.PopulateBlockData(block)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return block, nil
}

func (bs *BlockSpineService) GetBlockByHeightCacheFormat(height uint32) (*storage.BlockCacheObject, error) {
	return commonUtils.GetBlockByHeightUseBlocksCache(
		height,
		bs.QueryExecutor,
		bs.BlockQuery,
		bs.BlocksStorage,
	)
}

// GetGenesisBlock return the genesis block
func (bs *BlockSpineService) GetGenesisBlock() (*model.Block, error) {
	var (
		genesisBlock model.Block
		row, _       = bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetGenesisBlock(), false)
	)
	if row == nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
	}
	err := bs.BlockQuery.Scan(&genesisBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "cannot parse genesis block db entity")
	}
	genesisBlock.SpineBlockManifests = make([]*model.SpineBlockManifest, 0)
	return &genesisBlock, nil
}

// GetBlocks return all pushed blocks
func (bs *BlockSpineService) GetBlocks() ([]*model.Block, error) {
	var (
		blocks    []*model.Block
		rows, err = bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlocks(0, 100), false)
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// PopulateBlockData add spine public keys to model.Block instance
func (bs *BlockSpineService) PopulateBlockData(block *model.Block) error {
	spinePublicKeys, err := bs.SpinePublicKeyService.GetSpinePublicKeysByBlockHeight(block.Height)
	if err != nil {
		bs.Logger.Errorln(err)
		return blocker.NewBlocker(blocker.BlockErr, "error getting block spine public keys")
	}
	block.SpinePublicKeys = spinePublicKeys
	spineBlockManifests, err := bs.SpineBlockManifestService.GetSpineBlockManifestBySpineBlockHeight(block.Height)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "error getting block spineBlockManifests")
	}
	block.SpineBlockManifests = spineBlockManifests
	return nil
}

// UpdateLastBlockCache to update the state of last block cache
func (bs *BlockSpineService) UpdateLastBlockCache(block *model.Block) error {
	var err error
	// direct update storage cache if block is not nil
	// Note: make sure block already populate their data before cache
	if block != nil {
		err = bs.BlockStateStorage.SetItem(nil, *block)
		if err != nil {
			return err
		}
		return nil
	}

	// getting last Block from DB when incoming block nil
	var lastBlock *model.Block
	lastBlock, err = commonUtils.GetLastBlock(bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = bs.PopulateBlockData(lastBlock)
	if err != nil {
		return err
	}
	err = bs.BlockStateStorage.SetItem(nil, *lastBlock)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BlockSpineService) InitializeBlocksCache() error {
	var err = bs.BlocksStorage.Clear()
	if err != nil {
		return err
	}
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return err
	}
	var firstBlocksHeightCache uint32 = 0
	if lastBlock.Height > constant.MaxBlocksCacheStorage {
		firstBlocksHeightCache = lastBlock.Height - constant.MaxBlocksCacheStorage
	}
	var (
		blocks []*model.Block
	)
	blocks, err = bs.GetBlocksFromHeight(firstBlocksHeightCache, constant.MaxBlocksCacheStorage, false)
	if err != nil {
		return err
	}
	for i := 0; i < len(blocks); i++ {
		err = bs.BlocksStorage.Push(
			commonUtils.BlockConvertToCacheFormat(blocks[i]),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPayloadHashAndLength compute and return the block's payload hash
func (bs *BlockSpineService) GetPayloadHashAndLength(block *model.Block) (payloadHash []byte, payloadLength uint32, err error) {
	var (
		digest = sha3.New256()
	)
	for _, spinePubKey := range block.GetSpinePublicKeys() {
		spinePubKeyBytes := commonUtils.GetSpinePublicKeyBytes(spinePubKey)
		if _, err := digest.Write(spinePubKeyBytes); err != nil {
			return nil, 0, err
		}
		payloadLength += uint32(len(spinePubKeyBytes))

	}
	// compute the block payload length and hash by parsing all file chunks db entities into their bytes representation
	for _, spineBlockManifest := range block.GetSpineBlockManifests() {
		spineBlockManifestBytes := bs.SpineBlockManifestService.GetSpineBlockManifestBytes(spineBlockManifest)
		if _, err := digest.Write(spineBlockManifestBytes); err != nil {
			return nil, 0, err
		}
		payloadLength += uint32(len(spineBlockManifestBytes))
	}
	payloadHash = digest.Sum([]byte{})
	return
}

// GenerateBlock generate block from transactions in mempool
func (bs *BlockSpineService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
	_ bool,
) (*model.Block, error) {
	var (
		err                         error
		spinePublicKeys             []*model.SpinePublicKey
		spineBlockManifests         []*model.SpineBlockManifest
		includedMainBlocks          []*model.Block
		blockSmithPublicKey         = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
		newBlockHeight              = previousBlock.Height + 1
		newIncludedFirstBlockHeight = previousBlock.ReferenceBlockHeight + 1
		newReferenceBlockHeight     uint32
	)
	// select main block to be include in spine block
	lastMainBlock, err := bs.MainBlockService.GetLastBlockCacheFormat()
	if err != nil {
		return nil, err
	}
	// check last main block height still higher from SpineReferenceBlockHeightOffset
	if lastMainBlock.Height > constant.SpineReferenceBlockHeightOffset {
		newReferenceBlockHeight = lastMainBlock.Height - constant.SpineReferenceBlockHeightOffset
	}
	// make sure new reference block height is greater than previous Reference Block Height
	if newReferenceBlockHeight > previousBlock.ReferenceBlockHeight {
		limit := newReferenceBlockHeight - previousBlock.ReferenceBlockHeight
		includedMainBlocks, err = bs.MainBlockService.GetBlocksFromHeight(newIncludedFirstBlockHeight, limit, false)
		if err != nil {
			return nil, err
		}
	}
	if len(includedMainBlocks) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "NoNewMainBlocks")
	}
	// generate merkle root & merkle tree for new spnine block
	var (
		merkleRoot      commonUtils.MerkleRoot
		hashedMainBlock []*bytes.Buffer
	)
	for i := 0; i < len(includedMainBlocks); i++ {
		hashedMainBlock = append(hashedMainBlock, bytes.NewBuffer(includedMainBlocks[i].BlockHash))
	}
	_, err = merkleRoot.GenerateMerkleRoot(hashedMainBlock)
	if err != nil {
		return nil, err
	}
	mRoot, mTree := merkleRoot.ToBytes()

	// compute spine pub keys from mainchain node registrations
	spinePublicKeys, err = bs.SpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry(
		newIncludedFirstBlockHeight,
		newReferenceBlockHeight,
		newBlockHeight,
	)
	if err != nil {
		return nil, err
	}

	// retrieve all spineBlockManifests at current spine height (complete with file chunks entities)
	spineBlockManifests, err = bs.SpineBlockManifestService.GetSpineBlockManifestsForSpineBlock(newBlockHeight, timestamp)
	if err != nil {
		return nil, err
	}
	// assign spine block height to every manifests
	for _, spm := range spineBlockManifests {
		spm.ManifestSpineBlockHeight = newBlockHeight
	}
	digest := sha3.New256()
	digest.Reset() // reset the digest
	if _, err := digest.Write(previousBlock.GetBlockSeed()); err != nil {
		return nil, err
	}
	previousSeedHash := digest.Sum([]byte{})
	blockSeed := bs.Signature.GenerateBlockSeed(previousSeedHash, secretPhrase)
	digest.Reset() // reset the digest
	// compute the previous block hash
	previousBlockHash, err := commonUtils.GetBlockHash(previousBlock, bs.Chaintype)
	if err != nil {
		return nil, err
	}

	block, err := bs.NewSpineBlock(
		1,
		previousBlockHash,
		blockSeed,
		blockSmithPublicKey,
		mRoot,
		mTree,
		newBlockHeight,
		newReferenceBlockHeight,
		timestamp,
		secretPhrase,
		spinePublicKeys,
		spineBlockManifests,
	)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GenerateGenesisBlock generate and return genesis block from a given template (see constant/genesis.go)
func (bs *BlockSpineService) GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinBase int64
		blockTransactions                    []*model.Transaction
		spineChainPublicKeys                 []*model.SpinePublicKey
		payloadBytes                         []byte
		payloadLength                        uint32
		digest                               = sha3.New256()
		err                                  error
	)

	// add spine public keys from mainchain genesis configuration to spine genesis block
	spineChainPublicKeys, err = bs.getGenesisSpinePublicKeys(genesisEntries)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(spineChainPublicKeys, func(i, j int) bool {
		intI := new(big.Int).SetBytes(spineChainPublicKeys[i].NodePublicKey)
		intJ := new(big.Int).SetBytes(spineChainPublicKeys[j].NodePublicKey)
		res := intI.Cmp(intJ)
		// Ascending sort
		return res < 0
	})

	payloadBytes = bs.getGenesisSpinePayloadBytes(spineChainPublicKeys)
	payloadLength = uint32(len(payloadBytes))
	if _, err := digest.Write(payloadBytes); err != nil {
		return nil, err
	}

	payloadHash := digest.Sum([]byte{})
	mainGenesisBlock, err := bs.MainBlockService.GenerateGenesisBlock(constant.GenesisConfig)
	if err != nil {
		return nil, err
	}
	var merkleRoot commonUtils.MerkleRoot
	_, err = merkleRoot.GenerateMerkleRoot([]*bytes.Buffer{bytes.NewBuffer(mainGenesisBlock.BlockHash)})
	if err != nil {
		return nil, err
	}
	mRoot, mTree := merkleRoot.ToBytes()

	block, err := bs.NewGenesisBlock(1, nil, bs.Chaintype.GetGenesisBlockSeed(), bs.Chaintype.GetGenesisNodePublicKey(), mRoot, mTree, 0, 0, bs.Chaintype.GetGenesisBlockTimestamp(), totalAmount, totalFee, totalCoinBase, blockTransactions, spineChainPublicKeys, payloadHash, payloadLength, big.NewInt(0), bs.Chaintype.GetGenesisBlockSignature())
	if err != nil {
		return nil, err
	}
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block, bs.Chaintype)
	if block.ID == 0 {
		return nil, blocker.NewBlocker(blocker.BlockErr, fmt.Sprintf("Invalid %s Genesis Block ID", bs.Chaintype.GetName()))
	}
	return block, nil
}

// AddGenesis generate and add (push) genesis block to db
func (bs *BlockSpineService) AddGenesis() error {
	block, err := bs.GenerateGenesisBlock(constant.GenesisConfig)
	if err != nil {
		return err
	}
	err = bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, false, true)
	if err != nil {
		bs.Logger.Fatal("PushGenesisBlock:fail ", blocker.NewBlocker(blocker.PushSpineBlockErr, err.Error(), block.GetID()))
	}
	return nil
}

// CheckGenesis check if genesis has been added
func (bs *BlockSpineService) CheckGenesis() (bool, error) {
	genesisBlock, err := bs.GetGenesisBlock()
	if err != nil { // Genesis is not in the blockchain yet
		return false, nil
	}
	if genesisBlock.ID != bs.Chaintype.GetGenesisBlockID() {
		return false, fmt.Errorf("genesis ID does not match, expect: %d, get: %d", bs.Chaintype.GetGenesisBlockID(), genesisBlock.ID)
	}
	return true, nil
}

// ReceiveBlock handle the block received from connected peers
// argument lastBlock is the lastblock in this node
// argument block is the in coming block from peer
func (bs *BlockSpineService) ReceiveBlock(
	senderPublicKey []byte,
	lastBlock, block *model.Block,
	nodeSecretPhrase string,
	peer *model.Peer,
) (*model.Receipt, error) {
	var (
		err error
	)
	// make sure block has previous block hash
	if block.PreviousBlockHash == nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"last block hash does not exist",
		)
	}
	//  check equality last block hash with previous block hash from received block
	if !bytes.Equal(lastBlock.BlockHash, block.PreviousBlockHash) {
		// check if incoming block is of higher quality
		if bytes.Equal(lastBlock.PreviousBlockHash, block.PreviousBlockHash) &&
			block.Timestamp < lastBlock.Timestamp {
			err := func() error {
				bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
				defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)
				previousBlock, err := bs.GetBlockByHeight(lastBlock.Height - 1)
				if err != nil {
					return status.Error(codes.Internal,
						"fail to get last block",
					)
				}
				if !bytes.Equal(previousBlock.GetBlockHash(), block.PreviousBlockHash) {
					return status.Error(codes.InvalidArgument,
						"blockchain changed, ignore the incoming block",
					)
				}
				lastBlocks, err := bs.PopOffToBlock(previousBlock)
				if err != nil {
					return err
				}
				err = bs.ValidateBlock(block, previousBlock)
				if err != nil {
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false, true)
					if errPushBlock != nil {
						bs.Logger.Errorf("ReceiveBlock:pushing back popped off block fail: %v",
							blocker.NewBlocker(blocker.PushSpineBlockErr, err.Error(), block.GetID(), lastBlock.GetID()))
						return status.Error(codes.InvalidArgument, "InvalidBlock")
					}

					bs.Logger.Info("pushing back popped off block")
					return status.Error(codes.InvalidArgument, "InvalidBlock")
				}
				err = bs.PushBlock(previousBlock, block, true, true)
				if err != nil {
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false, true)
					if errPushBlock != nil {
						bs.Logger.Errorf("ReceiveBlock:pushing back popped off block fail: %v",
							blocker.NewBlocker(blocker.PushSpineBlockErr, err.Error(), block.GetID(), lastBlock.GetID()))
						return status.Error(codes.InvalidArgument, "InvalidBlock")
					}
					bs.Logger.Info("pushing back popped off block")
					return status.Error(codes.InvalidArgument, "InvalidBlock")
				}
				return nil
			}()
			if err != nil {
				return nil, err
			}
		}
		return nil, status.Error(codes.InvalidArgument,
			"previousBlockHashDoesNotMatchWithLastBlockHash",
		)
	}
	err = func() error {
		// pushBlock closure to release lock as soon as block pushed
		// Securing receive block process
		bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
		defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)
		// making sure get last block after paused process
		lastBlock, err = bs.GetLastBlock()
		if err != nil {
			return status.Error(codes.Internal,
				"fail to get last block",
			)
		}
		// Validate incoming block
		err = bs.ValidateBlock(block, lastBlock)
		if err != nil {
			return status.Error(codes.InvalidArgument, "InvalidBlock")
		}
		err = bs.PushBlock(lastBlock, block, true, true)
		if err != nil {
			bs.Logger.Errorf(
				"[ReceiveBlock] pushBlock fail: %v",
				blocker.NewBlocker(blocker.PushSpineBlockErr, err.Error(), block.GetID(), lastBlock.GetID()),
			)
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	// spine blocks don't return any receipts
	return nil, nil
}

// validateIncludedMainBlock to validate included main block in spine block
func (bs *BlockSpineService) validateIncludedMainBlock(lastBlock, incomingBlock *model.Block) error {
	if incomingBlock.GetReferenceBlockHeight() == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "NoIncludedMainBlock")
	}
	if incomingBlock.ReferenceBlockHeight <= lastBlock.ReferenceBlockHeight {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidReferenceBlockHeight")
	}
	var mainLastBlock, err = bs.MainBlockService.GetLastBlockCacheFormat()
	if err != nil {
		return err
	}
	// no need validate merkle root when reference block height is higher than current last main block
	if incomingBlock.ReferenceBlockHeight > mainLastBlock.Height {
		return nil
	}
	var referenceBlock = mainLastBlock
	if mainLastBlock.Height != incomingBlock.ReferenceBlockHeight {
		referenceBlock, err = bs.MainBlockService.GetBlockByHeightCacheFormat(incomingBlock.ReferenceBlockHeight)
		if err != nil {
			return err
		}
	}
	var (
		merkleRoot         commonUtils.MerkleRoot
		rootHash           []byte
		leafIndex          = (incomingBlock.ReferenceBlockHeight - lastBlock.ReferenceBlockHeight) - 1
		intermediateHashes [][]byte
	)
	merkleRoot.HashTree = merkleRoot.FromBytes(incomingBlock.MerkleTree, incomingBlock.MerkleRoot)
	intermediateHashesBuffer := merkleRoot.GetIntermediateHashes(bytes.NewBuffer(referenceBlock.BlockHash), int32(leafIndex))
	for _, buf := range intermediateHashesBuffer {
		intermediateHashes = append(intermediateHashes, buf.Bytes())
	}
	rootHash, err = merkleRoot.GetMerkleRootFromIntermediateHashes(
		referenceBlock.BlockHash,
		leafIndex,
		intermediateHashes,
	)
	if err != nil {
		return err
	}
	if !bytes.Equal(incomingBlock.MerkleRoot, rootHash) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidMerkleRootBlock")
	}
	return nil
}

func (bs *BlockSpineService) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	var (
		err error
	)
	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return []*model.Block{}, err
	}
	minRollbackHeight := commonUtils.GetMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork process
		bs.Logger.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return []*model.Block{}, nil
	}

	_, err = bs.GetBlockByID(commonBlock.ID, false)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var (
		poppedBlocks    []*model.Block
		poppedManifests []*model.SpineBlockManifest
	)
	block := lastBlock

	for block.ID != commonBlock.ID && block.ID != bs.Chaintype.GetGenesisBlockID() {
		poppedBlocks = append(poppedBlocks, block)
		block, err = bs.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
	}

	derivedQueries := query.GetDerivedQuery(bs.Chaintype)
	err = bs.QueryExecutor.BeginTx()
	if err != nil {
		return []*model.Block{}, err
	}

	for _, dQuery := range derivedQueries {
		queries := dQuery.Rollback(commonBlock.Height)
		err = bs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			rollbackErr := bs.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				bs.Logger.Warnf("spineblock-rollback-err: %v", rollbackErr)
			}
			return []*model.Block{}, err
		}
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}

	// cache last block state.
	// Note: Make sure every time calling query insert & rollback block, calling this SetItem too
	err = bs.UpdateLastBlockCache(nil)
	if err != nil {
		return nil, err
	}
	err = bs.BlocksStorage.PopTo(commonBlock.Height)
	if err != nil {
		return nil, err
	}

	go func() {
		// post rollback action:
		// - clean snapshot data
		poppedManifests, err = bs.SpineBlockManifestService.GetSpineBlockManifestsFromSpineBlockHeight(commonBlock.Height)
		if err != nil {
			rollbackErr := bs.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				bs.Logger.Warn(rollbackErr)
			}
		}
		for _, manifest := range poppedManifests {
			// ignore error, file deletion can fail
			deleteErr := bs.SnapshotMainBlockService.DeleteFileByChunkHashes(manifest.FileChunkHashes)
			if deleteErr != nil {
				log.Warnf("fail deleting snapshot during rollback: %v\n", deleteErr)
			}
		}
	}()

	// Need to sort ascending since was descended in above by Height
	sort.Slice(poppedBlocks, func(i, j int) bool {
		return poppedBlocks[i].GetHeight() < poppedBlocks[j].GetHeight()
	})

	return poppedBlocks, nil
}

func (bs *BlockSpineService) getGenesisSpinePayloadBytes(spinePublicKeys []*model.SpinePublicKey) (spinePublicKeysBytes []byte) {
	spinePublicKeysBytes = make([]byte, 0)
	for _, spinePublicKey := range spinePublicKeys {
		spinePublicKeysBytes = append(spinePublicKeysBytes, commonUtils.GetSpinePublicKeyBytes(spinePublicKey)...)
	}
	return spinePublicKeysBytes
}

// getGenesisSpinePublicKeys returns spine block's genesis payload, as an array of model.SpinePublicKey and in bytes,
// based on nodes registered at genesis
func (bs *BlockSpineService) getGenesisSpinePublicKeys(
	genesisEntries []constant.GenesisConfigEntry,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	spinePublicKeys = make([]*model.SpinePublicKey, 0)
	for _, mainchainGenesisEntry := range genesisEntries {
		if mainchainGenesisEntry.NodePublicKey == nil {
			continue
		}
		// pass to genesis the fullAddress (accountType + accountPublicKey) in bytes
		accountFullAddress, err := accounttype.ParseEncodedAccountToAccountAddress(
			int32(model.AccountType_ZbcAccountType),
			mainchainGenesisEntry.AccountAddress,
		)
		if err != nil {
			return nil, err
		}
		genesisNodeRegistrationTx, err := GetGenesisNodeRegistrationTx(
			accountFullAddress,
			mainchainGenesisEntry.Message,
			mainchainGenesisEntry.LockedBalance,
			mainchainGenesisEntry.NodePublicKey,
		)
		if err != nil {
			return nil, err
		}
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   mainchainGenesisEntry.NodePublicKey,
			NodeID:          genesisNodeRegistrationTx.ID,
			PublicKeyAction: model.SpinePublicKeyAction_AddKey,
			MainBlockHeight: 0,
			Height:          0,
			Latest:          true,
		}
		spinePublicKeys = append(spinePublicKeys, spinePublicKey)
	}
	return spinePublicKeys, nil
}

func (bs *BlockSpineService) ReceivedValidatedBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsInterface interface{}, args ...interface{}) {},
	}
}

func (bs *BlockSpineService) BlockTransactionsRequestedListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsIdsInterface interface{}, args ...interface{}) {},
	}
}

func (bs *BlockSpineService) ValidateSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error {
	var (
		block model.Block
		found bool
	)
	qry := bs.BlockQuery.GetBlockFromTimestamp(spineBlockManifest.GetExpirationTimestamp(), 1)
	row, _ := bs.QueryExecutor.ExecuteSelectRow(qry, false)
	if err := bs.BlockQuery.Scan(&block, row); err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSpineBlockManifestTimestamp")
	}
	if err := bs.PopulateBlockData(&block); err != nil {
		return err
	}

	// first check if spineBlockManifest is included in block data
	spineBlockManifestBytes := bs.SpineBlockManifestService.GetSpineBlockManifestBytes(spineBlockManifest)
	for _, blSpineBlockManifest := range block.GetSpineBlockManifests() {
		blSpineBlockManifestBytes := bs.SpineBlockManifestService.GetSpineBlockManifestBytes(blSpineBlockManifest)
		if bytes.Equal(spineBlockManifestBytes, blSpineBlockManifestBytes) {
			found = true
			break
		}
	}
	if !found {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSpineBlockManifestData")
	}

	// now validate against block payload hash
	computedHash, computedLength, err := bs.GetPayloadHashAndLength(&block)
	if err != nil {
		return err
	}
	if !bytes.Equal(computedHash, block.GetPayloadHash()) || computedLength != block.PayloadLength {
		// in this case it could be that one or more spine block manifest entries have been manually added to db after the block
		// has been pushed to db
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidComputedSpineBlockPayloadHash")
	}

	return nil
}
