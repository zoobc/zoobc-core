package service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
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
	// BlockServiceMainInterface interface that contains methods specific of BlockService
	BlockServiceMainInterface interface {
		NewMainBlock(
			version uint32,
			previousBlockHash, blockSeed, blockSmithPublicKey []byte,
			previousBlockHeight uint32,
			timestamp, totalAmount, totalFee, totalCoinBase int64,
			transactions []*model.Transaction,
			blockReceipts []*model.PublishedReceipt,
			payloadHash []byte,
			payloadLength uint32,
			secretPhrase string,
		) (*model.Block, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
		ReceivedValidatedBlockTransactionsListener() observer.Listener
		BlockTransactionsRequestedListener() observer.Listener
		ScanBlockPool() error
	}

	// TODO: rename to BlockMainService
	BlockService struct {
		sync.RWMutex
		Chaintype                   chaintype.ChainType
		KVExecutor                  kvdb.KVExecutorInterface
		QueryExecutor               query.ExecutorInterface
		BlockQuery                  query.BlockQueryInterface
		MempoolQuery                query.MempoolQueryInterface
		TransactionQuery            query.TransactionQueryInterface
		PublishedReceiptQuery       query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		Signature                   crypto.SignatureInterface
		MempoolService              MempoolServiceInterface
		ReceiptService              ReceiptServiceInterface
		NodeRegistrationService     NodeRegistrationServiceInterface
		BlocksmithService           BlocksmithServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		BlocksmithStrategy          strategy.BlocksmithStrategyInterface
		BlockIncompleteQueueService BlockIncompleteQueueServiceInterface
		BlockPoolService            BlockPoolServiceInterface
		Observer                    *observer.Observer
		Logger                      *log.Logger
		TransactionUtil             transaction.UtilInterface
		ReceiptUtil                 coreUtil.ReceiptUtilInterface
		PublishedReceiptUtil        coreUtil.PublishedReceiptUtilInterface
		TransactionCoreService      TransactionCoreServiceInterface
		CoinbaseService             CoinbaseServiceInterface
		ParticipationScoreService   ParticipationScoreServiceInterface
		PublishedReceiptService     PublishedReceiptServiceInterface
	}
)

func NewBlockMainService(
	ct chaintype.ChainType,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	receiptService ReceiptServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	obsr *observer.Observer,
	blocksmithStrategy strategy.BlocksmithStrategyInterface,
	logger *log.Logger,
	accountLedgerQuery query.AccountLedgerQueryInterface,
	blockIncompleteQueueService BlockIncompleteQueueServiceInterface,
	transactionUtil transaction.UtilInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	publishedReceiptUtil coreUtil.PublishedReceiptUtilInterface,
	transactionCoreService TransactionCoreServiceInterface,
	blockPoolService BlockPoolServiceInterface,
	blocksmithService BlocksmithServiceInterface,
	coinbaseService CoinbaseServiceInterface,
	participationScoreService ParticipationScoreServiceInterface,
	publishedReceiptService PublishedReceiptServiceInterface,
) *BlockService {
	return &BlockService{
		Chaintype:                   ct,
		KVExecutor:                  kvExecutor,
		QueryExecutor:               queryExecutor,
		BlockQuery:                  blockQuery,
		MempoolQuery:                mempoolQuery,
		TransactionQuery:            transactionQuery,
		SkippedBlocksmithQuery:      skippedBlocksmithQuery,
		Signature:                   signature,
		MempoolService:              mempoolService,
		ReceiptService:              receiptService,
		NodeRegistrationService:     nodeRegistrationService,
		ActionTypeSwitcher:          txTypeSwitcher,
		AccountBalanceQuery:         accountBalanceQuery,
		ParticipationScoreQuery:     participationScoreQuery,
		NodeRegistrationQuery:       nodeRegistrationQuery,
		BlocksmithStrategy:          blocksmithStrategy,
		Observer:                    obsr,
		Logger:                      logger,
		AccountLedgerQuery:          accountLedgerQuery,
		BlockIncompleteQueueService: blockIncompleteQueueService,
		TransactionUtil:             transactionUtil,
		ReceiptUtil:                 receiptUtil,
		PublishedReceiptUtil:        publishedReceiptUtil,
		TransactionCoreService:      transactionCoreService,
		BlockPoolService:            blockPoolService,
		BlocksmithService:           blocksmithService,
		CoinbaseService:             coinbaseService,
		ParticipationScoreService:   participationScoreService,
		PublishedReceiptService:     publishedReceiptService,
	}
}

// NewMainBlock generate new mainchain block
func (bs *BlockService) NewMainBlock(
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
	secretPhrase string,
) (*model.Block, error) {
	var (
		err error
	)

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
	}

	// compute block's payload hash and length and add it to block struct
	if block.PayloadHash, block.PayloadLength, err = bs.GetPayloadHashAndLength(block); err != nil {
		return nil, err
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
func (bs *BlockService) GetChainType() chaintype.ChainType {
	return bs.Chaintype
}

func (bs *BlockService) GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface {
	return bs.BlocksmithStrategy
}

// ChainWriteLock locks the chain
func (bs *BlockService) ChainWriteLock(actionType int) {
	monitoring.IncrementStatusLockCounter(bs.Chaintype, actionType)
	bs.Lock()
	monitoring.SetBlockchainStatus(bs.Chaintype, actionType)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockService) ChainWriteUnlock(actionType int) {
	monitoring.SetBlockchainStatus(bs.Chaintype, constant.BlockchainStatusIdle)
	monitoring.DecrementStatusLockCounter(bs.Chaintype, actionType)
	bs.Unlock()
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockService) NewGenesisBlock(
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

// ValidatePayloadHash validate (computed) block's payload data hash against block's payload hash
func (bs *BlockService) ValidatePayloadHash(block *model.Block) error {
	hash, length, err := bs.GetPayloadHashAndLength(block)
	if err != nil {
		return err
	}
	if length != block.GetPayloadLength() || !bytes.Equal(hash, block.GetPayloadHash()) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockPayload")
	}
	return nil
}

// PreValidateBlock validate block without it's transactions
func (bs *BlockService) PreValidateBlock(block, previousLastBlock *model.Block) error {
	// check if blocksmith can smith at the time
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousLastBlock)
	blocksmithIndex := blocksmithsMap[string(block.BlocksmithPublicKey)]
	if blocksmithIndex == nil {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidBlocksmith")
	}
	// check smithtime
	blocksmithTime := bs.BlocksmithStrategy.GetSmithTime(*blocksmithIndex, previousLastBlock)
	if blocksmithTime > block.GetTimestamp() {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSmithTime")
	}
	return nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	if err := bs.ValidatePayloadHash(block); err != nil {
		return err
	}

	// check block timestamp
	if block.GetTimestamp() > curTime+constant.GenerateBlockTimeoutSec {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidTimestamp")
	}

	err := bs.PreValidateBlock(block, previousLastBlock)
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
func (bs *BlockService) validateBlockHeight(block *model.Block) error {
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
func (bs *BlockService) PushBlock(previousBlock, block *model.Block, broadcast, persist bool) error {
	var (
		blocksmithIndex *int64
		err             error
	)

	if !coreUtil.IsGenesis(previousBlock.GetID(), block) {
		block.Height = previousBlock.GetHeight() + 1
		sortedBlocksmithMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousBlock)
		blocksmithIndex = sortedBlocksmithMap[string(block.GetBlocksmithPublicKey())]
		if blocksmithIndex == nil {
			return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInSmithingList")
		}
		// check for duplicate in block pool
		blockPool := bs.BlockPoolService.GetBlock(*blocksmithIndex)
		if blockPool != nil && !persist {
			return blocker.NewBlocker(
				blocker.BlockErr, "DuplicateBlockPool",
			)
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

	// Respecting Expiring escrow before push block process
	err = bs.TransactionCoreService.ExpiringEscrowTransactions(block.GetHeight(), true)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, err.Error())
	}

	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err = bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	var transactionIDs = make([]int64, len(block.GetTransactions()))
	// apply transactions and remove them from mempool
	for index, tx := range block.GetTransactions() {
		// assign block id and block height to tx
		tx.BlockID = block.ID
		tx.Height = block.Height
		tx.TransactionIndex = uint32(index) + 1
		transactionIDs[index] = tx.GetID()
		// validate tx here
		// check if is in mempool : if yes, undo unconfirmed
		rows, err := bs.QueryExecutor.ExecuteSelect(bs.MempoolQuery.GetMempoolTransaction(), false, tx.ID)
		if err != nil {
			rows.Close()
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			rows.Close()
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}

		if rows.Next() {
			err = bs.TransactionCoreService.UndoApplyUnconfirmedTransaction(txType)
			if err != nil {
				rows.Close()
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		}
		rows.Close()
		if block.Height > 0 {
			err = bs.TransactionCoreService.ValidateTransaction(txType, true)
			if err != nil {
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		}
		// validate tx body and apply/perform transaction-specific logic
		err = bs.TransactionCoreService.ApplyConfirmedTransaction(txType, block.GetTimestamp())
		if err == nil {
			transactionInsertQuery, transactionInsertValue := bs.TransactionQuery.InsertTransaction(tx)
			err := bs.QueryExecutor.ExecuteTransaction(transactionInsertQuery, transactionInsertValue...)
			if err != nil {
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		} else {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	if block.Height != 0 {
		if errRemoveMempool := bs.RemoveMempoolTransactions(block.GetTransactions()); errRemoveMempool != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return errRemoveMempool
		}
	}
	linkedCount, err := bs.PublishedReceiptService.ProcessPublishedReceipts(block)
	if err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}

	// Mainchain specific:
	// - Compute and update popscore
	// - Block reward
	// - Admit/Expel nodes to/from registry
	// - Build scrambled node registry
	if block.Height > 0 {
		// this is to manage the edge case when the blocksmith array has not been initialized yet:
		// when start smithing from a block with height > 0, since SortedBlocksmiths are computed  after a block is pushed,
		// for the first block that is pushed, we don't know who are the blocksmith to be rewarded
		// sort blocksmiths for current block
		popScore, err := commonUtils.CalculateParticipationScore(
			uint32(linkedCount),
			uint32(len(block.GetPublishedReceipts())-linkedCount),
			bs.ReceiptUtil.GetNumberOfMaxReceipts(len(bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))),
		)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		err = bs.updatePopScore(popScore, previousBlock, block)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}

		// selecting multiple account to be rewarded and split the total coinbase + totalFees evenly between them
		totalReward := block.TotalFee + block.TotalCoinBase
		lotteryAccounts, err := bs.CoinbaseService.CoinbaseLotteryWinners(
			bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock),
		)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		if err := bs.BlocksmithService.RewardBlocksmithAccountAddresses(
			lotteryAccounts,
			totalReward,
			block.GetTimestamp(),
			block.Height,
		); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	// admit nodes from registry at genesis and regular intervals
	// expel nodes from node registry as soon as they reach zero participation score
	if err := bs.expelNodes(block); err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	if block.Height == 0 || block.Height%bs.NodeRegistrationService.GetNodeAdmittanceCycle() == 0 {
		if err := bs.admitNodes(block); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	// building scrambled node registry
	if block.GetHeight() == bs.NodeRegistrationService.GetBlockHeightToBuildScrambleNodes(block.GetHeight()) {
		err = bs.NodeRegistrationService.BuildScrambledNodes(block)
		if err != nil {
			bs.Logger.Error(err.Error())
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	// persist flag will only be turned off only when generate or receive block broadcasted by another peer
	if !persist { // block content are validated
		// get blocksmith index
		blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousBlock)
		blocksmithIndex = blocksmithsMap[string(block.BlocksmithPublicKey)]
		// handle if is first index
		if *blocksmithIndex > 0 {
			// check if current block is in pushable window
			if !bs.canPersistBlock(*blocksmithIndex, previousBlock) {
				// insert into block pool
				bs.BlockPoolService.InsertBlock(block, *blocksmithIndex)
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				if broadcast {
					// add transactionIDs and remove transaction before broadcast
					block.TransactionIDs = transactionIDs
					block.Transactions = []*model.Transaction{}
					bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
				}
				return nil
			}
			// if canPersistBlock return true ignore the passed `persist` flag
		}
		// block is in first place continue to persist block to database ignoring the `persist` flag
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	bs.Logger.Debugf("%s Block Pushed ID: %d", bs.Chaintype.GetName(), block.GetID())
	// sort blocksmiths for next block
	bs.BlocksmithStrategy.SortBlocksmiths(block, true)
	// clear the block pool
	bs.BlockPoolService.ClearBlockPool()
	// broadcast block
	if broadcast && !persist && *blocksmithIndex == 0 {
		// add transactionIDs and remove transaction before broadcast
		block.TransactionIDs = transactionIDs
		block.Transactions = []*model.Transaction{}
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype, block)
	return nil
}

// ScanBlockPool scan the whole block pool to check if there are any block that's legal to be pushed yet
func (bs *BlockService) ScanBlockPool() error {
	previousBlock, err := bs.GetLastBlock()
	if err != nil {
		return err
	}
	blocks := bs.BlockPoolService.GetBlocks()
	for index, block := range blocks {
		if bs.canPersistBlock(index, previousBlock) {
			bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
			err := bs.PushBlock(previousBlock, block, true, true)
			bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)
			if err != nil {
				return blocker.NewBlocker(
					blocker.BlockErr, "ScanBlockPool:PushBlockFail",
				)
			}
		}
	}
	return nil
}

// canPersistBlock check if the blocksmith can push the block based on previous block's blocksmiths order
// this function must only run when receiving / generating block, not on download block since it uses the current machine
// time as comparison
// todo: will move this to block pool service + write the test when refactoring the block service
func (bs *BlockService) canPersistBlock(blocksmithIndex int64, previousBlock *model.Block) bool {
	if blocksmithIndex < 1 {
		return true
	}
	var (
		currentTime = time.Now().Unix()
	)
	blocksmithAllowedBeginTime := bs.BlocksmithStrategy.GetSmithTime(blocksmithIndex, previousBlock)
	blocksmithExpiredPersistTime := blocksmithAllowedBeginTime +
		constant.SmithingBlockCreationTime + constant.SmithingNetworkTolerance
	previousBlocksmithAllowedBeginTime := blocksmithAllowedBeginTime - constant.SmithingBlocksmithTimeGap
	blocksmithAllowedPersistTime := previousBlocksmithAllowedBeginTime +
		constant.SmithingBlockCreationTime + constant.SmithingNetworkTolerance
	// allowed time window = lastBlocksmithExpiredTime < current_time <= currentBlocksmithExpiredTime
	if previousBlock.GetHeight() == 0 {
		return currentTime > blocksmithAllowedPersistTime
	}
	return currentTime >= blocksmithAllowedPersistTime && currentTime <= blocksmithExpiredPersistTime
}

// adminNodes seelct and admit nodes from node registry
func (bs *BlockService) admitNodes(block *model.Block) error {
	// select n (= MaxNodeAdmittancePerCycle) queued nodes with the highest locked balance from node registry
	nodeRegistrations, err := bs.NodeRegistrationService.SelectNodesToBeAdmitted(constant.MaxNodeAdmittancePerCycle)
	if err != nil {
		return err
	}
	if len(nodeRegistrations) > 0 {
		err = bs.NodeRegistrationService.AdmitNodes(nodeRegistrations, block.Height)
		if err != nil {
			return err
		}
	}
	return nil
}

// expelNodes seelct and expel nodes from node registry
func (bs *BlockService) expelNodes(block *model.Block) error {
	nodeRegistrations, err := bs.NodeRegistrationService.SelectNodesToBeExpelled()
	if err != nil {
		return err
	}
	if len(nodeRegistrations) > 0 {
		err = bs.NodeRegistrationService.ExpelNodes(nodeRegistrations, block.Height)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockService) updatePopScore(popScore int64, previousBlock, block *model.Block) error {
	var (
		blocksmithNode  *model.Blocksmith
		blocksmithIndex = -1
		err             error
	)
	for i, bsm := range bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock) {
		if reflect.DeepEqual(block.BlocksmithPublicKey, bsm.NodePublicKey) {
			blocksmithIndex = i
			blocksmithNode = bsm
			break
		}
	}
	if blocksmithIndex < 0 {
		return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInBlocksmithList")
	}
	// punish the skipped (index earlier than current blocksmith) blocksmith
	for i, bsm := range (bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))[:blocksmithIndex] {
		skippedBlocksmith := &model.SkippedBlocksmith{
			BlocksmithPublicKey: bsm.NodePublicKey,
			POPChange:           constant.ParticipationScorePunishAmount,
			BlockHeight:         block.Height,
			BlocksmithIndex:     int32(i),
		}
		// store to skipped_blocksmith table
		qStr, args := bs.SkippedBlocksmithQuery.InsertSkippedBlocksmith(
			skippedBlocksmith,
		)
		err = bs.QueryExecutor.ExecuteTransaction(qStr, args...)
		if err != nil {
			return err
		}
		// punish score
		_, err = bs.NodeRegistrationService.AddParticipationScore(bsm.NodeID, constant.ParticipationScorePunishAmount, block.Height, true)
		if err != nil {
			return err
		}
	}
	_, err = bs.NodeRegistrationService.AddParticipationScore(blocksmithNode.NodeID, popScore, block.Height, true)
	if err != nil {
		return err
	}
	return nil
}

// GetBlockByID return a block by its ID
// withAttachedData if true returns extra attached data for the block (transactions)
func (bs *BlockService) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
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
			transactions, err := bs.TransactionCoreService.GetTransactionsByBlockID(block.ID)
			if err != nil {
				return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
			}
			block.Transactions = transactions
		}
		return &block, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block %v is not found", id))
}

// GetBlocksFromHeight get all blocks from a given height till last block (or a given limit is reached).
// Note: this only returns main block data, it doesn't populate attached data (transactions, receipts)
func (bs *BlockService) GetBlocksFromHeight(startHeight, limit uint32, withAttachedData bool) ([]*model.Block, error) {
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
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	var (
		transactions []*model.Transaction
		lastBlock    *model.Block
		err          error
	)

	lastBlock, err = commonUtils.GetLastBlock(bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	transactions, err = bs.TransactionCoreService.GetTransactionsByBlockID(lastBlock.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	lastBlock.Transactions = transactions
	return lastBlock, nil
}

// GetBlockHash return block's hash (makes sure always include transactions)
func (bs *BlockService) GetBlockHash(block *model.Block) ([]byte, error) {
	transactions, err := bs.TransactionCoreService.GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.Transactions = transactions
	return commonUtils.GetBlockHash(block, bs.GetChainType())

}

// GetBlockByHeight return the last pushed block
func (bs *BlockService) GetBlockByHeight(height uint32) (*model.Block, error) {
	var (
		transactions []*model.Transaction
		block        *model.Block
		err          error
	)

	block, err = commonUtils.GetBlockByHeight(height, bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	transactions, err = bs.TransactionCoreService.GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.Transactions = transactions

	return block, nil
}

// GetGenesisBlock return the last pushed block
func (bs *BlockService) GetGenesisBlock() (*model.Block, error) {
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
func (bs *BlockService) GetBlocks() ([]*model.Block, error) {
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

// PopulateBlockData add transactions and published receipts to model.Block instance
func (bs *BlockService) PopulateBlockData(block *model.Block) error {
	txs, err := bs.TransactionCoreService.GetTransactionsByBlockID(block.ID)
	if err != nil {
		bs.Logger.Errorln(err)
		return blocker.NewBlocker(blocker.BlockErr, "error getting block transactions")
	}
	prs, err := bs.PublishedReceiptUtil.GetPublishedReceiptsByBlockHeight(block.Height)
	if err != nil {
		bs.Logger.Errorln(err)
		return blocker.NewBlocker(blocker.BlockErr, "error getting block published receipts")
	}
	block.Transactions = txs
	block.PublishedReceipts = prs
	return nil
}

// RemoveMempoolTransactions removes a list of transactions tx from mempool given their Ids
func (bs *BlockService) RemoveMempoolTransactions(transactions []*model.Transaction) error {
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

func (bs *BlockService) GetPayloadHashAndLength(block *model.Block) (payloadHash []byte, payloadLength uint32, err error) {
	var (
		digest = sha3.New256()
	)
	for _, tx := range block.GetTransactions() {
		if _, err := digest.Write(tx.GetTransactionHash()); err != nil {
			return nil, 0, err
		}
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, 0, err
		}
		payloadLength += txType.GetSize()
	}
	// filter only good receipt
	for _, br := range block.GetPublishedReceipts() {
		brBytes := bs.ReceiptUtil.GetSignedBatchReceiptBytes(br.BatchReceipt)
		_, err = digest.Write(brBytes)
		if err != nil {
			return nil, 0, err
		}
		payloadLength += uint32(len(brBytes))
	}
	payloadHash = digest.Sum([]byte{})
	return
}

// GenerateBlock generate block from transactions in mempool
func (bs *BlockService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinbase int64
		// only for mainchain
		sortedTransactions  []*model.Transaction
		publishedReceipts   []*model.PublishedReceipt
		err                 error
		digest              = sha3.New256()
		blockSmithPublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
	)
	newBlockHeight := previousBlock.Height + 1
	// calculate total coinbase to be added to the block
	totalCoinbase = bs.CoinbaseService.GetCoinbase()
	sortedTransactions, err = bs.MempoolService.SelectTransactionsFromMempool(timestamp)
	if err != nil {
		return nil, errors.New("MempoolReadError")
	}
	// select transactions from mempool to be added to the block
	for _, tx := range sortedTransactions {
		txType, errType := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if errType != nil {
			return nil, err
		}
		totalAmount += txType.GetAmount()
		totalFee += tx.Fee
	}
	// select published receipts to be added to the block
	publishedReceipts, err = bs.ReceiptService.SelectReceipts(
		timestamp, bs.ReceiptUtil.GetNumberOfMaxReceipts(
			len(bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))),
		previousBlock.Height,
	)
	if err != nil {
		return nil, err
	}

	// loop through transaction to build block hash
	if _, err = digest.Write(previousBlock.GetBlockSeed()); err != nil {
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
	block, err := bs.NewMainBlock(
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
		secretPhrase,
	)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GenerateGenesisBlock generate and return genesis block from a given template (see constant/genesis.go)
func (bs *BlockService) GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinBase int64
		blockTransactions                    []*model.Transaction
		payloadLength                        uint32
		digest                               = sha3.New256()
	)

	genesisTransactions, err := GetGenesisTransactions(bs.Chaintype, genesisEntries)
	if err != nil {
		return nil, err
	}
	for index, tx := range genesisTransactions {
		if _, err := digest.Write(tx.TransactionHash); err != nil {
			return nil, err
		}
		if tx.TransactionType == util.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
			totalAmount += tx.GetSendMoneyTransactionBody().Amount
		}
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}
		totalAmount += txType.GetAmount()
		totalFee += tx.Fee
		payloadLength += txType.GetSize()
		tx.TransactionIndex = uint32(index) + 1
		blockTransactions = append(blockTransactions, tx)
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
		nil,
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
func (bs *BlockService) AddGenesis() error {
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
func (bs *BlockService) CheckGenesis() bool {
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
func (bs *BlockService) ReceiveBlock(
	senderPublicKey []byte,
	lastBlock, block *model.Block,
	nodeSecretPhrase string,
	peer *model.Peer,
) (*model.BatchReceipt, error) {
	var err error
	// make sure block has previous block hash
	if block.GetPreviousBlockHash() == nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"last block hash does not exist",
		)
	}

	// check previous block hash of new block not same with current block hash and
	// check new block is not better than current block
	if !bytes.Equal(block.GetPreviousBlockHash(), lastBlock.GetBlockHash()) &&
		!(bytes.Equal(block.GetPreviousBlockHash(), lastBlock.GetPreviousBlockHash()) &&
			block.Timestamp < lastBlock.Timestamp) {
		return nil, status.Error(codes.InvalidArgument, "InvalidBlock")
	}

	// check new block is better than current block
	if bytes.Equal(block.GetPreviousBlockHash(), lastBlock.GetPreviousBlockHash()) &&
		block.Timestamp < lastBlock.Timestamp {
		lastBlock, err = commonUtils.GetBlockByHeight(lastBlock.Height-1, bs.QueryExecutor, bs.BlockQuery)
		if err != nil {
			return nil, status.Error(codes.Internal, "FailGetBlock")
		}
	}

	// pre validation block
	if err = bs.PreValidateBlock(block, lastBlock); err != nil {
		return nil, status.Error(codes.InvalidArgument, "InvalidBlock")
	}

	isQueued, err := bs.ProcessQueueBlock(block, peer)
	if err != nil {
		return nil, err
	}
	// process block when block don't have transaction
	if !isQueued {
		err = bs.ProcessCompletedBlock(block)
		if err != nil {
			return nil, err
		}
	}

	receiptKey, err := bs.ReceiptUtil.GetReceiptKey(
		block.GetBlockHash(), senderPublicKey,
	)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			err.Error(),
		)
	}
	// check if already broadcast receipt to this node
	_, err = bs.KVExecutor.Get(constant.KVdbTableBlockReminderKey + string(receiptKey))
	if err == nil {
		return nil, blocker.NewBlocker(blocker.BlockErr, "already send receipt for this block")
	}

	if err != badger.ErrKeyNotFound {
		return nil, blocker.NewBlocker(blocker.BlockErr, "failed get receipt key")
	}

	// generate receipt and return as response
	batchReceipt, err := bs.ReceiptService.GenerateBatchReceiptWithReminder(
		bs.Chaintype,
		block.GetBlockHash(),
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		constant.KVdbTableBlockReminderKey+string(receiptKey),
		constant.ReceiptDatumTypeBlock,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return batchReceipt, nil
}

// GetParticipationScore handle received block from another node
func (bs *BlockService) GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error) {
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
		blExt.BlocksmithAccountAddress, err = bs.BlocksmithService.GetBlocksmithAccountAddress(block)
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
	publishedReceipts, err = bs.PublishedReceiptUtil.GetPublishedReceiptsByBlockHeight(block.GetHeight())
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
		bs.ReceiptUtil.GetNumberOfMaxReceipts(len(nodeRegistryAtHeight)),
	)
	if err != nil {
		return nil, err
	}

	if includeReceipts {
		blExt.Block.PublishedReceipts = publishedReceipts
	}

	return blExt, nil
}

func (bs *BlockService) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
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

	_, err = bs.GetBlockByID(commonBlock.ID, false)
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
		tx, err := bs.TransactionUtil.ParseTransactionBytes(mempool.GetTransactionBytes(), true)
		if err != nil {
			return nil, err
		}
		txType, err = bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}

		err = bs.TransactionCoreService.UndoApplyUnconfirmedTransaction(txType)
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
	//
	// TODO: here we should also delete all snapshot files relative to the block manifests being rolled back during derived tables
	//  rollback. Something like this:
	//  - before rolling back derived queries, select all spine block manifest records from commonBlock.Height till last
	//  - delete all snapshots referenced by them
	//
	if mempoolsBackupBytes.Len() > 0 {
		kvdbMempoolsBackupKey := commonUtils.GetKvDbMempoolDBKey(bs.GetChainType())
		err = bs.KVExecutor.Insert(kvdbMempoolsBackupKey, mempoolsBackupBytes.Bytes(), int(constant.KVDBMempoolsBackupExpiry))
		if err != nil {
			return nil, err
		}
	}
	// remove peer memoization
	bs.NodeRegistrationService.ResetScrambledNodes()
	// clear block pool
	bs.BlockPoolService.ClearBlockPool()
	return poppedBlocks, nil
}

// WillSmith check if blocksmith need to calculate their smith time or need to smith or not
func (bs *BlockService) WillSmith(
	blocksmith *model.Blocksmith,
	blockchainProcessorLastBlockID int64,
) (int64, error) {
	var blocksmithScore int64
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return blockchainProcessorLastBlockID, blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}

	// caching: only calculate smith time once per new block
	if lastBlock.GetID() != blockchainProcessorLastBlockID {
		blockchainProcessorLastBlockID = lastBlock.GetID()
		bs.BlocksmithStrategy.SortBlocksmiths(lastBlock, true)
		// check if eligible to create block in this round
		blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(lastBlock)
		if blocksmithsMap[string(blocksmith.NodePublicKey)] == nil {
			return blockchainProcessorLastBlockID,
				blocker.NewBlocker(blocker.SmithingErr, "BlocksmithNotInBlocksmithList")
		}
		// calculate blocksmith score for the block type
		// try to get the node's participation score (ps) from node public key
		// if node is not registered, ps will be 0 and this node won't be able to smith
		// the default ps is 100000, smithing could be slower than when using account balances
		// since default balance was 1000 times higher than default ps
		blocksmithScore, err = bs.ParticipationScoreService.GetParticipationScore(blocksmith.NodePublicKey)
		if blocksmithScore <= 0 {
			bs.Logger.Info("Node has participation score <= 0. Either is not registered or has been expelled from node registry")
		}
		if err != nil || blocksmithScore < 0 {
			// no negative scores allowed
			blocksmithScore = 0
			bs.Logger.Errorf("Participation score calculation: %s", err)
		}
		err = bs.BlocksmithStrategy.CalculateSmith(
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
	// check for block pool duplicate
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(lastBlock)
	blocksmithIndex, ok := blocksmithsMap[string(blocksmith.NodePublicKey)]
	if !ok {
		return blockchainProcessorLastBlockID, err
	}
	blockPool := bs.BlockPoolService.GetBlock(*blocksmithIndex)
	if blockPool != nil {
		return blockchainProcessorLastBlockID, blocker.NewBlocker(
			blocker.BlockErr, "DuplicateBlockPool",
		)
	}
	return blockchainProcessorLastBlockID, nil
}

// ProcessCompletedBlock to process block that already having all needed transactions
func (bs *BlockService) ProcessCompletedBlock(block *model.Block) error {
	bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
	defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return err
	}
	//  check equality last block hash with previous block hash from received block
	if !bytes.Equal(lastBlock.GetBlockHash(), block.GetPreviousBlockHash()) {
		// check if incoming block is of higher quality
		// todo: moving this piece of code to another interface (block popper or process fork) the test will come later.
		if bytes.Equal(lastBlock.GetPreviousBlockHash(), block.PreviousBlockHash) &&
			block.Timestamp < lastBlock.Timestamp {
			previousBlock, err := commonUtils.GetBlockByHeight(lastBlock.Height-1, bs.QueryExecutor, bs.BlockQuery)
			if err != nil {
				return status.Error(codes.Internal,
					"fail to get last block",
				)
			}
			err = bs.ValidateBlock(block, previousBlock, time.Now().Unix())
			if err != nil {
				return status.Error(codes.InvalidArgument, "InvalidBlock")
			}
			lastBlocks, err := bs.PopOffToBlock(previousBlock)
			if err != nil {
				return err
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
		}
		return status.Error(codes.InvalidArgument,
			"previousBlockHashDoesNotMatchWithLastBlockHash",
		)
	}
	// Validate incoming block
	err = bs.ValidateBlock(block, lastBlock, time.Now().Unix())
	if err != nil {
		return status.Error(codes.InvalidArgument, "InvalidBlock")
	}
	err = bs.PushBlock(lastBlock, block, true, false)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

// ProcessQueueBlock process to queue block when waiting their transactions
func (bs *BlockService) ProcessQueueBlock(block *model.Block, peer *model.Peer) (needWaiting bool, err error) {
	// check block having transactions or not
	if len(block.TransactionIDs) == 0 {
		return false, nil
	}
	// check block already queued or not
	if bs.BlockIncompleteQueueService.GetBlockQueue(block.GetID()) != nil {
		return true, nil
	}
	var (
		txRequiredByBlock     = make(TransactionIDsMap)
		txRequiredByBlockArgs []interface{}
	)
	block.Transactions = make([]*model.Transaction, len(block.GetTransactionIDs()))
	for idx, txID := range block.TransactionIDs {
		txRequiredByBlock[txID] = idx
		// used as argument when quermockBlockDataying in mempool
		txRequiredByBlockArgs = append(txRequiredByBlockArgs, txID)
	}

	// find needed transactions in mempool
	var (
		caseQuery    = query.NewCaseQuery()
		mempoolQuery = query.NewMempoolQuery(bs.Chaintype)
		mempools     []*model.MempoolTransaction
	)
	// build query to select transaction in mempool transaction
	caseQuery.Select(mempoolQuery.TableName, mempoolQuery.Fields...)
	caseQuery.Where(caseQuery.In("id", txRequiredByBlockArgs...))
	selectQuery, args := caseQuery.Build()
	rows, err := bs.QueryExecutor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	mempools, err = mempoolQuery.BuildModel(mempools, rows)
	if err != nil {
		return false, err
	}
	for _, mempool := range mempools {
		tx, err := bs.TransactionUtil.ParseTransactionBytes(mempool.TransactionBytes, true)
		if err != nil {
			continue
		}
		block.Transactions[txRequiredByBlock[tx.GetID()]] = tx
		delete(txRequiredByBlock, tx.GetID())
	}
	// process when needed trasacntions are completed
	if len(txRequiredByBlock) == 0 {
		err := bs.ProcessCompletedBlock(block)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// saving temporary block
	bs.BlockIncompleteQueueService.AddBlockQueue(block)
	bs.BlockIncompleteQueueService.SetTransactionsRequired(block.GetID(), txRequiredByBlock)

	if peer == nil {
		bs.Logger.Errorf("Error peer is null, can not request block transactions from the Peer")
	}

	var txIds []int64
	for txID := range txRequiredByBlock {
		txIds = append(txIds, txID)
	}

	bs.BlockIncompleteQueueService.RequestBlockTransactions(txIds, block.GetID(), peer)
	return true, nil
}

// ReceivedValidatedBlockTransactionsListener will receive validated transactions to complete transactions of blocks queued
func (bs *BlockService) ReceivedValidatedBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsInterface interface{}, args ...interface{}) {
			transactions, ok := transactionsInterface.([]*model.Transaction)
			if !ok {
				bs.Logger.Fatalln("transactions casting failures in ReceivedValidatedBlockTransactionsListener")
			}
			for _, transaction := range transactions {
				var completedBlocks = bs.BlockIncompleteQueueService.AddTransaction(transaction)
				for _, block := range completedBlocks {
					err := bs.ProcessCompletedBlock(block)
					if err != nil {
						bs.Logger.Warn(blocker.BlockErr, err.Error())
					}
				}
			}
		},
	}
}

// ReceivedValidatedBlockTransactionsListener will send the transactions required by blocks
func (bs *BlockService) BlockTransactionsRequestedListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsIdsInterface interface{}, args ...interface{}) {
			bs.ChainWriteLock(constant.BlockchainSendingBlockTransactions)
			defer bs.ChainWriteUnlock(constant.BlockchainSendingBlockTransactions)

			var (
				transactions   []*model.Transaction
				transactionIds []int64
				peer           *model.Peer
				chainType      chaintype.ChainType
				blockID        int64
				ok             bool
			)

			// check number of arguments before casting the argument type
			if len(args) < 3 {
				bs.Logger.Fatalln("number of needed arguments too few in BlockTransactionsRequestedListener")
				return
			}
			chainType, ok = args[0].(*chaintype.MainChain)
			if !ok {
				bs.Logger.Fatalln("chaintype casting failures in BlockTransactionsRequestedListener")
			}

			// check chaintype
			if chainType != bs.Chaintype {
				bs.Logger.Warnf("chaintype is not macth, current chain is %s the incoming chain is %s",
					bs.Chaintype.GetName(), chainType.GetName())
				return
			}

			blockID, ok = args[1].(int64)
			if !ok {
				bs.Logger.Fatalln("blockID casting failures in BlockTransactionsRequestedListener")
			}

			peer, ok = args[2].(*model.Peer)
			if !ok {
				bs.Logger.Fatalln("peer casting failures in BlockTransactionsRequestedListener")
			}

			transactionIds, ok = transactionsIdsInterface.([]int64)
			if !ok {
				bs.Logger.Fatalln("transactionIds casting failures in BlockTransactionsRequestedListener")
			}

			var (
				remainingTxIDs []int64
				block          = bs.BlockPoolService.GetBlock(blockID)
			)
			// get transaction from block pool
			if block != nil {
				var (
					blockPoolTxs = block.GetTransactions()
					txMap        = make(map[int64]*model.Transaction)
				)
				for _, tx := range blockPoolTxs {
					txMap[tx.GetID()] = tx
				}

				for _, txID := range transactionIds {
					if txMap[txID] != nil {
						transactions = append(transactions, txMap[txID])
						continue
					}
					remainingTxIDs = append(remainingTxIDs, txID)
				}
			}

			// get remaining transactions from DB transaction if needed
			if len(transactions) < len(transactionIds) {
				if len(transactions) == 0 {
					remainingTxIDs = transactionIds
				}
				var remainingTxs, err = bs.TransactionCoreService.GetTransactionsByIds(remainingTxIDs)
				if err != nil {
					return
				}
				transactions = append(transactions, remainingTxs...)
			}
			bs.Observer.Notify(observer.SendBlockTransactions, transactions, bs.Chaintype, peer)
		},
	}
}
