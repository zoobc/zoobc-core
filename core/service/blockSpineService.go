package service

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
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
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		SpinePublicKeyQuery     query.SpinePublicKeyQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		Observer                *observer.Observer
		Logger                  *log.Logger
	}
)

// NewBlock generate new block
func (bs *BlockSpineService) NewBlock(
	version uint32,
	previousBlockHash,
	blockSeed, blockSmithPublicKey []byte,
	previousBlockHeight uint32,
	timestamp,
	totalAmount,
	totalFee,
	totalCoinBase int64,
	transactions []*model.Transaction,
	publishedReceipts []*model.PublishedReceipt,
	payloadHash []byte,
	payloadLength uint32,
	secretPhrase string,
) (*model.Block, error) {
	block := &model.Block{
		Version:             version,
		PreviousBlockHash:   previousBlockHash,
		BlockSeed:           blockSeed,
		BlocksmithPublicKey: blockSmithPublicKey,
		Height:              previousBlockHeight,
		Timestamp:           timestamp,
		TotalAmount:         totalAmount,
		TotalFee:            totalFee,
		TotalCoinBase:       totalCoinBase,
		Transactions:        transactions,
		PublishedReceipts:   publishedReceipts,
		PayloadHash:         payloadHash,
		PayloadLength:       payloadLength,
	}
	blockUnsignedByte, err := util.GetBlockByte(block, false)
	if err != nil {
		bs.Logger.Error(err.Error())
	}
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	blockHash, err := util.GetBlockHash(block)
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

// ChainWriteLock locks the chain
func (bs *BlockSpineService) ChainWriteLock(actionType int) {
	monitoring.IncrementStatusLockCounter(actionType)
	bs.Lock()
	monitoring.SetBlockchainStatus(bs.Chaintype.GetTypeInt(), actionType)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockSpineService) ChainWriteUnlock(actionType int) {
	monitoring.SetBlockchainStatus(bs.Chaintype.GetTypeInt(), constant.BlockchainStatusIdle)
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
	blockHash, err := util.GetBlockHash(block)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockSpineService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
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
	if coreUtil.GetBlockID(block) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidID")
	}
	// Verify Signature
	blockByte, err := commonUtils.GetBlockByte(block, false)
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
	previousBlockHash, err := util.GetBlockHash(previousLastBlock)
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
func (bs *BlockSpineService) PushBlock(previousBlock, block *model.Block, broadcast bool) error {
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
	if err := bs.insertSpinePublicKeys(block); err != nil {
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
	bs.BlocksmithStrategy.SortBlocksmiths(block)
	// broadcast block
	if broadcast {
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype.GetTypeInt(), block)
	return nil
}

// CoinbaseLotteryWinners get the current list of blocksmiths, duplicate it (to not change the original one)
// and sort it using the NodeOrder algorithm. The first n (n = constant.MaxNumBlocksmithRewards) in the newly ordered list
// are the coinbase lottery winner (the blocksmiths that will be rewarded for the current block)
func (bs *BlockSpineService) CoinbaseLotteryWinners(blocksmiths []*model.Blocksmith) ([]string, error) {
	var (
		selectedAccounts []string
	)
	// copy the pointer array to not change original order

	// sort blocksmiths by NodeOrder
	sort.SliceStable(blocksmiths, func(i, j int) bool {
		bi, bj := blocksmiths[i], blocksmiths[j]
		res := bi.NodeOrder.Cmp(bj.NodeOrder)
		if res == 0 {
			// compare node ID
			nodePKI := new(big.Int).SetUint64(uint64(bi.NodeID))
			nodePKJ := new(big.Int).SetUint64(uint64(bj.NodeID))
			res = nodePKI.Cmp(nodePKJ)
		}
		// ascending sort
		return res < 0
	})

	for idx, sortedBlockSmith := range blocksmiths {
		if idx > constant.MaxNumBlocksmithRewards-1 {
			break
		}
		// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
		qry, args := bs.NodeRegistrationQuery.GetNodeRegistrationByID(sortedBlockSmith.NodeID)
		rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		nr, err := bs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if (err != nil) || len(nr) == 0 {
			rows.Close()
			return nil, blocker.NewBlocker(blocker.DBErr, "CoinbaseLotteryNodeRegistrationNotFound")
		}
		selectedAccounts = append(selectedAccounts, nr[0].AccountAddress)
		rows.Close()
	}
	return selectedAccounts, nil
}

// RewardBlocksmithAccountAddresses accrue the block total fees + total coinbase to selected list of accounts
func (bs *BlockSpineService) RewardBlocksmithAccountAddresses(
	blocksmithAccountAddresses []string,
	totalReward int64,
	height uint32) error {
	queries := make([][]interface{}, 0)
	if len(blocksmithAccountAddresses) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NoAccountToBeRewarded")
	}
	blocksmithReward := totalReward / int64(len(blocksmithAccountAddresses))
	for _, blocksmithAccountAddress := range blocksmithAccountAddresses {
		accountBalanceRecipientQ := bs.AccountBalanceQuery.AddAccountBalance(
			blocksmithReward,
			map[string]interface{}{
				"account_address": blocksmithAccountAddress,
				"block_height":    height,
			},
		)
		queries = append(queries, accountBalanceRecipientQ...)
	}
	if err := bs.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}

// GetBlockByID return the last pushed block
func (bs *BlockSpineService) GetBlockByID(id int64) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByID(id), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	var blocks []*model.Block
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	if len(blocks) > 0 {
		transactions, err := bs.GetTransactionsByBlockID(blocks[0].ID)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		blocks[0].Transactions = transactions
		return blocks[0], nil
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

	spinePublicKeys, err := bs.getSpinePublicKeysByHeightInterval(0, lastBlock.Height)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	lastBlock.SpinePublicKeys = spinePublicKeys
	return lastBlock, nil
}

// GetTransactionsByBlockID in spine blocks this is a dummy method and returns an empty slice
func (bs *BlockSpineService) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (bs *BlockSpineService) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := bs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err := bs.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = bs.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockSpineService) GetBlockByHeight(height uint32) (*model.Block, error) {
	block, err := commonUtils.GetBlockByHeight(height, bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	spinePublicKeys, err := bs.getSpinePublicKeysByHeightInterval(0, height)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.SpinePublicKeys = spinePublicKeys

	return block, nil
}

// GetGenesis return the last pushed block
func (bs *BlockSpineService) GetGenesisBlock() (*model.Block, error) {
	var (
		lastBlock model.Block
		row, _    = bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetGenesisBlock(), false)
	)
	if row == nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
	}
	err := bs.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
	}
	return &lastBlock, nil
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

// RemoveMempoolTransactions removes a list of transactions tx from mempool given their Ids
func (bs *BlockSpineService) RemoveMempoolTransactions(transactions []*model.Transaction) error {
	var idsStr []string
	for _, tx := range transactions {
		idsStr = append(idsStr, "'"+strconv.FormatInt(tx.ID, 10)+"'")
	}
	err := bs.QueryExecutor.ExecuteTransaction(bs.MempoolQuery.DeleteMempoolTransactions(idsStr))
	if err != nil {
		return err
	}
	bs.Logger.Infof("mempool transaction with IDs = %s deleted", idsStr)
	return nil
}

// GenerateBlock generate block from transactions in mempool
func (bs *BlockSpineService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinbase int64
		payloadLength                        uint32
		// only for mainchain
		sortedTransactions  []*model.Transaction
		publishedReceipts   []*model.PublishedReceipt
		payloadHash         []byte
		err                 error
		digest              = sha3.New256()
		blockSmithPublicKey = util.GetPublicKeyFromSeed(secretPhrase)
	)
	newBlockHeight := previousBlock.Height + 1
	//TODO: @iltoga add all spine public keys since previous spine block (all nodes added and removed
	//		to the registry since previousBlock.height)
	payloadHash = digest.Sum([]byte{})
	// loop through transaction to build block hash
	digest.Reset() // reset the digest
	if _, err := digest.Write(previousBlock.GetBlockSeed()); err != nil {
		return nil, err
	}

	previousSeedHash := digest.Sum([]byte{})
	blockSeed := bs.Signature.SignByNode(previousSeedHash, secretPhrase)
	digest.Reset() // reset the digest
	previousBlockHash, err := util.GetBlockHash(previousBlock)
	if err != nil {
		return nil, err
	}
	block, err := bs.NewBlock(
		1,
		previousBlockHash,
		blockSeed,
		blockSmithPublicKey,
		newBlockHeight,
		timestamp,
		totalAmount,
		totalFee,
		totalCoinbase,
		sortedTransactions,
		publishedReceipts,
		payloadHash,
		payloadLength,
		secretPhrase,
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
	payloadBytes = bs.getSpinePayloadBytes()
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
	block.ID = coreUtil.GetBlockID(block)
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
	err = bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, false)
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
) (*model.BatchReceipt, error) {
	// make sure block has previous block hash
	if block.GetPreviousBlockHash() == nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"last block hash does not exist",
		)
	}
	receiptKey, err := commonUtils.GetReceiptKey(
		block.GetBlockHash(), senderPublicKey,
	)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			err.Error(),
		)
	}
	//  check equality last block hash with previous block hash from received block
	if !bytes.Equal(lastBlock.GetBlockHash(), block.GetPreviousBlockHash()) {
		// check if incoming block is of higher quality
		if bytes.Equal(lastBlock.GetPreviousBlockHash(), block.PreviousBlockHash) &&
			block.Timestamp < lastBlock.Timestamp {
			err := func() error {
				bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
				defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)
				previousBlock, err := commonUtils.GetBlockByHeight(lastBlock.Height-1, bs.QueryExecutor, bs.BlockQuery)
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
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false)
					if errPushBlock != nil {
						bs.Logger.Errorf("pushing back popped off block fail: %v", errPushBlock)
						return status.Error(codes.InvalidArgument, "InvalidBlock")
					}

					bs.Logger.Info("pushing back popped off block")
					return status.Error(codes.InvalidArgument, "InvalidBlock")
				}
				err = bs.PushBlock(previousBlock, block, true)
				if err != nil {
					errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false)
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
		// check if already broadcast receipt to this node
		_, err := bs.KVExecutor.Get(constant.KVdbTableBlockReminderKey + string(receiptKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				blockHash, err := commonUtils.GetBlockHash(block)
				if err != nil {
					return nil, err
				}
				if !bytes.Equal(blockHash, lastBlock.GetBlockHash()) {
					// invalid block hash don't send receipt to client
					return nil, status.Error(codes.InvalidArgument, "InvalidBlockHash")
				}
				batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
					block.GetBlockHash(),
					lastBlock,
					senderPublicKey,
					nodeSecretPhrase,
					constant.KVdbTableBlockReminderKey+string(receiptKey),
					constant.ReceiptDatumTypeBlock,
					bs.Signature,
					bs.QueryExecutor,
					bs.KVExecutor,
				)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
				return batchReceipt, nil
			}
			return nil, status.Error(codes.Internal, err.Error())
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
		lastBlock, err = commonUtils.GetLastBlock(bs.QueryExecutor, bs.BlockQuery)
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
		err = bs.PushBlock(lastBlock, block, true)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	// TODO: ask @ali @andy @barton if we need to add the chaintype to receipts
	// generate receipt and return as response
	batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
		block.GetBlockHash(),
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		constant.KVdbTableBlockReminderKey+string(receiptKey),
		constant.ReceiptDatumTypeBlock,
		bs.Signature,
		bs.QueryExecutor,
		bs.KVExecutor,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return batchReceipt, nil
}

// GetParticipationScore handle received block from another node
func (bs *BlockSpineService) GetParticipationScore(nodePublicKey []byte) (int64, error) {
	var (
		participationScores []*model.ParticipationScore
	)
	participationScoreQ, args := bs.ParticipationScoreQuery.GetParticipationScoreByNodePublicKey(nodePublicKey)
	rows, err := bs.QueryExecutor.ExecuteSelect(participationScoreQ, false, args...)
	if err != nil {
		return 0, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	participationScores, err = bs.ParticipationScoreQuery.BuildModel(participationScores, rows)
	// if there aren't participation scores for this address/node, return 0
	if (err != nil) || len(participationScores) == 0 {
		return 0, nil
	}
	return participationScores[0].Score, nil
}

// GetParticipationScore handle received block from another node
func (bs *BlockSpineService) GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error) {
	var (
		blExt                         = &model.BlockExtendedInfo{}
		skippedBlocksmiths            []*model.SkippedBlocksmith
		publishedReceipts             []*model.PublishedReceipt
		nodeRegistryAtHeight          []*model.NodeRegistration
		linkedPublishedReceiptCount   uint32
		unLinkedPublishedReceiptCount uint32
		err                           error
	)
	blExt.Block = block
	// block extra (computed) info
	if block.Height > 0 {
		blExt.BlocksmithAccountAddress, err = bs.GetBlocksmithAccountAddress(block)
		if err != nil {
			return nil, err
		}
	} else {
		blExt.BlocksmithAccountAddress = constant.MainchainGenesisAccountAddress
	}
	skippedBlocksmithsQuery := bs.SkippedBlocksmithQuery.GetSkippedBlocksmithsByBlockHeight(block.Height)
	skippedBlocksmithsRows, err := bs.QueryExecutor.ExecuteSelect(skippedBlocksmithsQuery, false)
	if err != nil {
		return nil, err
	}
	defer skippedBlocksmithsRows.Close()
	blExt.SkippedBlocksmiths, err = bs.SkippedBlocksmithQuery.BuildModel(skippedBlocksmiths, skippedBlocksmithsRows)
	if err != nil {
		return nil, err
	}
	publishedReceiptQ, publishedReceiptArgs := bs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(block.Height)
	publishedReceiptRows, err := bs.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArgs...)
	if err != nil {
		return nil, err
	}
	defer publishedReceiptRows.Close()
	publishedReceipts, err = bs.PublishedReceiptQuery.BuildModel(publishedReceipts, publishedReceiptRows)
	if err != nil {
		return nil, err
	}
	blExt.TotalReceipts = int64(len(publishedReceipts))
	for _, pr := range publishedReceipts {
		if pr.IntermediateHashes != nil {
			linkedPublishedReceiptCount++
		} else {
			unLinkedPublishedReceiptCount++
		}
	}
	nodeRegistryAtHeightQ := bs.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.Height)
	nodeRegistryAtHeightRows, err := bs.QueryExecutor.ExecuteSelect(nodeRegistryAtHeightQ, false)
	if err != nil {
		return nil, err
	}
	defer nodeRegistryAtHeightRows.Close()
	nodeRegistryAtHeight, err = bs.NodeRegistrationQuery.BuildModel(nodeRegistryAtHeight, nodeRegistryAtHeightRows)
	if err != nil {
		return nil, err
	}
	blExt.ReceiptValue = commonUtils.GetReceiptValue(linkedPublishedReceiptCount, unLinkedPublishedReceiptCount)
	blExt.PopChange, err = util.CalculateParticipationScore(
		linkedPublishedReceiptCount,
		unLinkedPublishedReceiptCount,
		coreUtil.GetNumberOfMaxReceipts(len(nodeRegistryAtHeight)),
	)
	if err != nil {
		return nil, err
	}

	if includeReceipts {
		blExt.Block.PublishedReceipts = publishedReceipts
	}

	return blExt, nil
}

func (bs *BlockSpineService) GetBlocksmithAccountAddress(block *model.Block) (string, error) {
	var (
		nr []*model.NodeRegistration
	)
	// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
	qry, args := bs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(block.BlocksmithPublicKey, block.Height)
	rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return "", blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nr, err = bs.NodeRegistrationQuery.BuildModel(nr, rows)
	if (err != nil) || len(nr) == 0 {
		return "", blocker.NewBlocker(blocker.DBErr, "VersionedNodeRegistrationNotFound")
	}
	return nr[0].AccountAddress, nil
}

// GetCoinbase since spine blocks have no coinbase reward, we return zero
func (*BlockSpineService) GetCoinbase() int64 {
	return 0
}

func (bs *BlockSpineService) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	var (
		mempoolsBackupBytes *bytes.Buffer
		mempoolsBackup      []*model.MempoolTransaction
		err                 error
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

	_, err = bs.GetBlockByID(commonBlock.ID)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var poppedBlocks []*model.Block
	block := lastBlock

	// TODO:
	// Need to refactor this codes with better solution in the future
	// https://github.com/zoobc/zoobc-core/pull/514#discussion_r355297318
	publishedReceipts, err := bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.PublishedReceipts = publishedReceipts

	for block.ID != commonBlock.ID && block.ID != bs.Chaintype.GetGenesisBlockID() {
		poppedBlocks = append(poppedBlocks, block)
		block, err = bs.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		publishedReceipts, err := bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		block.PublishedReceipts = publishedReceipts
	}

	// Backup existing transactions from mempool before rollback
	mempoolsBackup, err = bs.MempoolService.GetMempoolTransactionsWantToBackup(commonBlock.Height)
	if err != nil {
		return nil, err
	}
	bs.Logger.Warnf("mempool tx backup %d in total with block_height %d", len(mempoolsBackup), commonBlock.GetHeight())
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

	mempoolsBackupBytes = bytes.NewBuffer([]byte{})

	for _, mempool := range mempoolsBackup {
		var (
			tx     *model.Transaction
			txType transaction.TypeAction
		)
		tx, err := transaction.ParseTransactionBytes(mempool.GetTransactionBytes(), true)
		if err != nil {
			return nil, err
		}
		txType, err = bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}

		err = txType.UndoApplyUnconfirmed()
		if err != nil {
			return nil, err
		}

		/*
			mempoolsBackupBytes format is
			[...{4}byteSize,{bytesSize}transactionBytes]
		*/
		sizeMempool := uint32(len(mempool.GetTransactionBytes()))
		mempoolsBackupBytes.Write(util.ConvertUint32ToBytes(sizeMempool))
		mempoolsBackupBytes.Write(mempool.GetTransactionBytes())
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}

	if mempoolsBackupBytes.Len() > 0 {
		err = bs.KVExecutor.Insert(constant.KVDBMempoolsBackup, mempoolsBackupBytes.Bytes(), int(constant.KVDBMempoolsBackupExpiry))
		if err != nil {
			return nil, err
		}
	}
	// remove peer memoization
	bs.NodeRegistrationService.ResetScrambledNodes()
	return poppedBlocks, nil
}

// getGenesisSpinePublicKeys returns spine block's genesis payload, as an array of model.SpinePublicKey and in bytes,
// based on nodes registered at genesis
func (bs *BlockSpineService) getGenesisSpinePublicKeys(
	genesisEntries []constant.GenesisConfigEntry,
) (spinePublicKeys []*model.SpinePublicKey) {
	spinePublicKeys = make([]*model.SpinePublicKey, 0)
	for _, mainchainGenesisEntry := range genesisEntries {
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   mainchainGenesisEntry.NodePublicKey,
			PublicKeyAction: model.SpinePublicKeyAction_AddKey,
			Height:          0,
			Latest:          true,
		}
		spinePublicKeys = append(spinePublicKeys, spinePublicKey)
	}
	return spinePublicKeys
}

func (bs *BlockSpineService) getSpinePublicKeysByHeightInterval(
	fromHeigth,
	toHeigth uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.SpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(fromHeigth, toHeigth), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	spinePublicKeys, err = bs.SpinePublicKeyQuery.BuildModel(spinePublicKeys, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spinePublicKeys, nil
}

func (bs *BlockSpineService) getSpinePayloadBytes() (spinePublicKeysBytes []byte) {
	spinePublicKeysBytes = make([]byte, 0)
	for _, mainchainGenesisEntry := range constant.GenesisConfig {
		spinePublicKey := &model.SpinePublicKey{
			NodePublicKey:   mainchainGenesisEntry.NodePublicKey,
			PublicKeyAction: model.SpinePublicKeyAction_AddKey,
			Height:          0,
			Latest:          true,
		}
		spinePublicKeysBytes = append(spinePublicKeysBytes, coreUtil.GetSpinePublicKeyBytes(spinePublicKey)...)
	}
	return spinePublicKeysBytes
}

// insertSpinePublicKeys insert all spine block publicKeys into spinePublicKey table
// Note: at this stage the spine pub keys have already been parsed into their model struct
func (bs *BlockSpineService) insertSpinePublicKeys(block *model.Block) error {
	queries := make([][]interface{}, 0)
	for _, spinePublicKey := range block.SpinePublicKeys {
		insertSpkQry := bs.SpinePublicKeyQuery.InsertSpinePublicKey(spinePublicKey)
		queries = append(queries, insertSpkQry...)
	}
	if err := bs.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}
