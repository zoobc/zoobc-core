package service

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
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
	}
)

func NewBlockSpineService(
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	spineBlockQuery query.BlockQueryInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	signature crypto.SignatureInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	obsr *observer.Observer,
	blocksmithStrategy strategy.BlocksmithStrategyInterface,
	logger *log.Logger,
	megablockQuery query.SpineBlockManifestQueryInterface,
) *BlockSpineService {
	return &BlockSpineService{
		Chaintype:          ct,
		QueryExecutor:      queryExecutor,
		BlockQuery:         spineBlockQuery,
		Signature:          signature,
		BlocksmithStrategy: blocksmithStrategy,
		Observer:           obsr,
		Logger:             logger,
		SpinePublicKeyService: &BlockSpinePublicKeyService{
			Logger:                logger,
			NodeRegistrationQuery: nodeRegistrationQuery,
			QueryExecutor:         queryExecutor,
			Signature:             signature,
			SpinePublicKeyQuery:   spinePublicKeyQuery,
		},
		SpineBlockManifestService: NewSpineBlockManifestService(
			queryExecutor,
			megablockQuery,
			spineBlockQuery,
			logger,
		),
	}
}

// NewSpineBlock generate new spinechain block
func (bs *BlockSpineService) NewSpineBlock(
	version uint32,
	previousBlockHash,
	blockSeed, blockSmithPublicKey []byte,
	previousBlockHeight uint32,
	timestamp int64,
	payloadHash []byte,
	payloadLength uint32,
	secretPhrase string,
	spinePublicKeys []*model.SpinePublicKey,
	spineBlockManifests []*model.SpineBlockManifest,
) (*model.Block, error) {
	block := &model.Block{
		Version:             version,
		PreviousBlockHash:   previousBlockHash,
		BlockSeed:           blockSeed,
		BlocksmithPublicKey: blockSmithPublicKey,
		Height:              previousBlockHeight,
		Timestamp:           timestamp,
		PayloadHash:         payloadHash,
		PayloadLength:       payloadLength,
		SpinePublicKeys:     spinePublicKeys,
		SpineBlockManifests: spineBlockManifests,
	}
	blockUnsignedByte, err := util.GetBlockByte(block, false, bs.Chaintype)
	if err != nil {
		bs.Logger.Error(err.Error())
	}
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	blockHash, err := util.GetBlockHash(block, bs.Chaintype)
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
	monitoring.IncrementStatusLockCounter(actionType)
	bs.Lock()
	monitoring.SetBlockchainStatus(bs.Chaintype, actionType)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockSpineService) ChainWriteUnlock(actionType int) {
	monitoring.SetBlockchainStatus(bs.Chaintype, constant.BlockchainStatusIdle)
	monitoring.DecrementStatusLockCounter(actionType)
	bs.Unlock()
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockSpineService) NewGenesisBlock(
	version uint32,
	previousBlockHash, blockSeed, blockSmithPublicKey []byte,
	previousBlockHeight uint32,
	timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction,
	publishedReceipts []*model.PublishedReceipt,
	spinePublicKeys []*model.SpinePublicKey,
	payloadHash []byte,
	payloadLength uint32,
	cumulativeDifficulty *big.Int,
	genesisSignature []byte,
) (*model.Block, error) {
	block := &model.Block{
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
		PublishedReceipts:    publishedReceipts,
		PayloadLength:        payloadLength,
		PayloadHash:          payloadHash,
		CumulativeDifficulty: cumulativeDifficulty.String(),
		BlockSignature:       genesisSignature,
	}
	blockHash, err := util.GetBlockHash(block, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockSpineService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	// TODO: should we validate the received spineblcokManifests against the one that have been generated locally?
	//		 what if they have been deleted?

	// todo: validate previous time
	if block.GetTimestamp() > curTime+constant.GenerateBlockTimeoutSec {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidTimestamp")
	}
	// check if blocksmith can smith at the time
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousLastBlock)
	blocksmithIndex := blocksmithsMap[string(block.BlocksmithPublicKey)]
	if blocksmithIndex == nil {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidBlocksmith")
	}
	blocksmithTime := bs.BlocksmithStrategy.GetSmithTime(*blocksmithIndex, previousLastBlock)
	if blocksmithTime > block.GetTimestamp() {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSmithTime")
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
	previousBlockHash, err := util.GetBlockHash(previousLastBlock, bs.Chaintype)
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
		sortedBlocksmithMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousBlock)
		blocksmithIndex := sortedBlocksmithMap[string(block.GetBlocksmithPublicKey())]
		if blocksmithIndex == nil {
			return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInSmithingList")
		}
		blockCumulativeDifficulty, err := coreUtil.CalculateCumulativeDifficulty(
			previousBlock, *blocksmithIndex,
		)
		if err != nil {
			return err
		}
		block.CumulativeDifficulty = blockCumulativeDifficulty
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

	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	bs.Logger.Debugf("%s Block Pushed ID: %d", bs.Chaintype.GetName(), block.GetID())
	// sort blocksmiths for next block
	bs.BlocksmithStrategy.SortBlocksmiths(block, true)
	// broadcast block
	if broadcast {
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype, block)
	return nil
}

// GetBlockByID return a block by its ID
// withAttachedData if true returns extra attached data for the block (transactions)
func (bs *BlockSpineService) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	var (
		block model.Block
	)
	row, err := bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetBlockByID(id), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if err = bs.BlockQuery.Scan(&block, row); err != nil {
		if err == sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, err.Error())
		}
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	if block.ID != 0 {
		if withAttachedData {
			err := bs.PopulateBlockData(&block)
			if err != nil {
				return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
			}
		}
		return &block, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block %v is not found", id))
}

// GetBlocksFromHeight get all blocks from a given height till last block (or a given limit is reached).
// Note: this only returns main block data, it doesn't populate attached data (spinePublicKeys)
func (bs *BlockSpineService) GetBlocksFromHeight(startHeight, limit uint32) ([]*model.Block, error) {
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

	return blocks, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockSpineService) GetLastBlock() (*model.Block, error) {
	lastBlock, err := commonUtils.GetLastBlock(bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = bs.PopulateBlockData(lastBlock)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return lastBlock, nil
}

// GetBlockHash return block's hash (makes sure always include spine public keys)
func (bs *BlockSpineService) GetBlockHash(block *model.Block) ([]byte, error) {
	err := bs.PopulateBlockData(block)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return commonUtils.GetBlockHash(block, bs.GetChainType())

}

// GetLastBlock return the last pushed block
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

// GetGenesis return the genesis block
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
	spineBlockManifests, err := bs.SpineBlockManifestService.GetSpineBlockManifestsForSpineBlock(block.Height, block.Timestamp)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "error getting block spineBlockManifests")
	}
	block.SpineBlockManifests = spineBlockManifests

	return nil
}

// GenerateBlock generate block from transactions in mempool
func (bs *BlockSpineService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
) (*model.Block, error) {
	var (
		payloadLength             uint32
		spinePublicKeys           []*model.SpinePublicKey
		payloadBytes, payloadHash []byte
		err                       error
		digest                    = sha3.New256()
		blockSmithPublicKey       = util.GetPublicKeyFromSeed(secretPhrase)
		fromTimestamp             = previousBlock.Timestamp
		spineBlockManifests       []*model.SpineBlockManifest
	)
	newBlockHeight := previousBlock.Height + 1
	// compute spine pub keys from mainchain node registrations
	// Note: since spine blocks are not in sync with main blocks and they are unaware of the height (on mainchain) where to retrieve
	// node registration's public keys, we use timestamps for now
	if fromTimestamp == bs.GetChainType().GetGenesisBlockTimestamp() {
		fromTimestamp++
	}
	spinePublicKeys, err = bs.SpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry(fromTimestamp, timestamp, newBlockHeight)
	for _, spinePubKey := range spinePublicKeys {
		payloadBytes = append(payloadBytes, commonUtils.GetSpinePublicKeyBytes(spinePubKey)...)
	}
	if err != nil {
		return nil, err
	}

	// retrieve all spineBlockManifests at current spine height (complete with file chunks entities)
	spineBlockManifests, err = bs.SpineBlockManifestService.GetSpineBlockManifestsForSpineBlock(newBlockHeight, timestamp)
	if err != nil {
		return nil, err
	}
	// compute the block payload length and hash by parsing all file chunks db entities into their bytes representation
	if len(spineBlockManifests) > 0 {
		for _, spineBlockManifest := range spineBlockManifests {
			megablockBytes := bs.SpineBlockManifestService.GetSpineBlockManifestBytes(spineBlockManifest)
			payloadBytes = append(payloadBytes, megablockBytes...)
		}
	}

	if _, err := digest.Write(payloadBytes); err != nil {
		return nil, err
	}
	payloadHash = digest.Sum([]byte{})
	payloadLength = uint32(len(payloadBytes))
	// loop through transaction to build block hash
	digest.Reset() // reset the digest
	if _, err := digest.Write(previousBlock.GetBlockSeed()); err != nil {
		return nil, err
	}

	previousSeedHash := digest.Sum([]byte{})
	blockSeed := bs.Signature.SignByNode(previousSeedHash, secretPhrase)
	digest.Reset() // reset the digest
	// compute the previous block hash
	previousBlockHash, err := util.GetBlockHash(previousBlock, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block, err := bs.NewSpineBlock(
		1,
		previousBlockHash,
		blockSeed,
		blockSmithPublicKey,
		newBlockHeight,
		timestamp,
		payloadHash,
		payloadLength,
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
	)

	// add spine public keys from mainchain genesis configuration to spine genesis block
	spineChainPublicKeys = bs.getGenesisSpinePublicKeys(genesisEntries)
	payloadBytes = bs.getGenesisSpinePayloadBytes(spineChainPublicKeys)
	payloadLength = uint32(len(payloadBytes))
	if _, err := digest.Write(payloadBytes); err != nil {
		return nil, err
	}

	payloadHash := digest.Sum([]byte{})
	block, err := bs.NewGenesisBlock(
		1,
		nil,
		bs.Chaintype.GetGenesisBlockSeed(),
		bs.Chaintype.GetGenesisNodePublicKey(),
		0,
		bs.Chaintype.GetGenesisBlockTimestamp(),
		totalAmount,
		totalFee,
		totalCoinBase,
		blockTransactions,
		[]*model.PublishedReceipt{},
		spineChainPublicKeys,
		payloadHash,
		payloadLength,
		big.NewInt(0),
		bs.Chaintype.GetGenesisBlockSignature(),
	)
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
		bs.Logger.Fatal("PushGenesisBlock:fail ", err)
	}
	return nil
}

// CheckGenesis check if genesis has been added
func (bs *BlockSpineService) CheckGenesis() bool {
	genesisBlock, err := bs.GetGenesisBlock()
	if err != nil { // Genesis is not in the blockchain yet
		return false
	}
	if genesisBlock.ID != bs.Chaintype.GetGenesisBlockID() {
		bs.Logger.Fatalf("Genesis ID does not match, expect: %d, get: %d", bs.Chaintype.GetGenesisBlockID(), genesisBlock.ID)
	}
	return true
}

// ReceiveBlock handle the block received from connected peers
// argument lastBlock is the lastblock in this node
// argument block is the in coming block from peer
func (bs *BlockSpineService) ReceiveBlock(
	senderPublicKey []byte,
	lastBlock, block *model.Block,
	nodeSecretPhrase string,
	peer *model.Peer,
) (*model.BatchReceipt, error) {
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
				err = bs.ValidateBlock(block, previousBlock, time.Now().Unix())
				if err != nil {
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false, true)
					if errPushBlock != nil {
						bs.Logger.Errorf("pushing back popped off block fail: %v", errPushBlock)
						return status.Error(codes.InvalidArgument, "InvalidBlock")
					}

					bs.Logger.Info("pushing back popped off block")
					return status.Error(codes.InvalidArgument, "InvalidBlock")
				}
				err = bs.PushBlock(previousBlock, block, true, true)
				if err != nil {
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false, true)
					if errPushBlock != nil {
						bs.Logger.Errorf("pushing back popped off block fail: %v", errPushBlock)
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
		err = bs.ValidateBlock(block, lastBlock, time.Now().Unix())
		if err != nil {
			return status.Error(codes.InvalidArgument, "InvalidBlock")
		}
		err = bs.PushBlock(lastBlock, block, true, true)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	// spine blocks don't return any receipts
	// TODO: @iltoga make sure to manage nil in calling function
	return nil, nil
}

// GetBlockExtendedInfo spine blocks have no extended info so far, so we just return the 'basic' block info (from model.Block)
func (bs *BlockSpineService) GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error) {
	var (
		blExt = &model.BlockExtendedInfo{}
	)
	blExt.Block = block
	// block extra (computed) info
	return blExt, nil
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
	minRollbackHeight := util.GetMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork process
		bs.Logger.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return []*model.Block{}, nil
	}

	_, err = bs.GetBlockByID(commonBlock.ID, false)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var poppedBlocks []*model.Block
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
			_ = bs.QueryExecutor.RollbackTx()
			return []*model.Block{}, err
		}
	}

	err = bs.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}

	return poppedBlocks, nil
}

func (bs *BlockSpineService) getGenesisSpinePayloadBytes(spinePublicKeys []*model.SpinePublicKey) (spinePublicKeysBytes []byte) {
	spinePublicKeysBytes = make([]byte, 0)
	for _, spinePublicKey := range spinePublicKeys {
		spinePublicKeysBytes = append(spinePublicKeysBytes, util.GetSpinePublicKeyBytes(spinePublicKey)...)
	}
	return spinePublicKeysBytes
}

// getGenesisSpinePublicKeys returns spine block's genesis payload, as an array of model.SpinePublicKey and in bytes,
// based on nodes registered at genesis
func (bs *BlockSpineService) getGenesisSpinePublicKeys(
	genesisEntries []constant.GenesisConfigEntry,
) (spinePublicKeys []*model.SpinePublicKey) {
	spinePublicKeys = make([]*model.SpinePublicKey, 0)
	for _, mainchainGenesisEntry := range genesisEntries {
		if mainchainGenesisEntry.NodePublicKey == nil {
			continue
		}
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   mainchainGenesisEntry.NodePublicKey,
			PublicKeyAction: model.SpinePublicKeyAction_AddKey,
			MainBlockHeight: 0,
			Height:          0,
			Latest:          true,
		}
		spinePublicKeys = append(spinePublicKeys, spinePublicKey)
	}
	return spinePublicKeys
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

func (bs *BlockSpineService) WillSmith(
	blocksmith *model.Blocksmith,
	blockchainProcessorLastBlockID int64,
) (int64, error) {
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return blockchainProcessorLastBlockID, blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}
	// caching: only calculate smith time once per new block
	if lastBlock.GetID() != blockchainProcessorLastBlockID {
		blockchainProcessorLastBlockID = lastBlock.GetID()
		blockSmithStrategy := bs.GetBlocksmithStrategy()
		blockSmithStrategy.SortBlocksmiths(lastBlock, true)
		// check if eligible to create block in this round
		blocksmithsMap := blockSmithStrategy.GetSortedBlocksmithsMap(lastBlock)
		if blocksmithsMap[string(blocksmith.NodePublicKey)] == nil {
			return blockchainProcessorLastBlockID,
				blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
		}
		// calculate blocksmith score for the block type
		// FIXME: ask @barton how to compute score for spine blocksmiths, since we don't have participation score and receipts attached to them?
		blocksmithScore := constant.DefaultParticipationScore
		err = blockSmithStrategy.CalculateSmith(
			lastBlock,
			*(blocksmithsMap[string(blocksmith.NodePublicKey)]),
			blocksmith,
			blocksmithScore,
		)
		if err != nil {
			return blockchainProcessorLastBlockID, err
		}
		monitoring.SetBlockchainSmithTime(bs.GetChainType(), blocksmith.SmithTime-lastBlock.Timestamp)
	}
	return blockchainProcessorLastBlockID, nil
}
