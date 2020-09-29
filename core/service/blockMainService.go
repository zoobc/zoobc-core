package service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/mohae/deepcopy"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
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
			secretPhrase string,
		) (*model.Block, error)
		ReceivedValidatedBlockTransactionsListener() observer.Listener
		BlockTransactionsRequestedListener() observer.Listener
		ScanBlockPool() error
	}

	// TODO: rename to BlockMainService
	BlockService struct {
		sync.RWMutex
		Chaintype                   chaintype.ChainType
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
		NodeAddressInfoService      NodeAddressInfoServiceInterface
		BlocksmithService           BlocksmithServiceInterface
		FeeScaleService             fee.FeeScaleServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		FeeVoteRevealVoteQuery      query.FeeVoteRevealVoteQueryInterface
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
		PruneQuery                  []query.PruneQuery
		BlockStateStorage           storage.CacheStorageInterface
		BlockchainStatusService     BlockchainStatusServiceInterface
		ScrambleNodeService         ScrambleNodeServiceInterface
	}
)

func NewBlockMainService(
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	receiptService ReceiptServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	nodeAddressInfoService NodeAddressInfoServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	feeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface,
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
	feeScaleService fee.FeeScaleServiceInterface,
	pruneQuery []query.PruneQuery,
	blockStateStorage storage.CacheStorageInterface,
	blockchainStatusService BlockchainStatusServiceInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
) *BlockService {
	return &BlockService{
		Chaintype:                   ct,
		QueryExecutor:               queryExecutor,
		BlockQuery:                  blockQuery,
		MempoolQuery:                mempoolQuery,
		TransactionQuery:            transactionQuery,
		SkippedBlocksmithQuery:      skippedBlocksmithQuery,
		Signature:                   signature,
		MempoolService:              mempoolService,
		ReceiptService:              receiptService,
		NodeRegistrationService:     nodeRegistrationService,
		NodeAddressInfoService:      nodeAddressInfoService,
		ActionTypeSwitcher:          txTypeSwitcher,
		AccountBalanceQuery:         accountBalanceQuery,
		ParticipationScoreQuery:     participationScoreQuery,
		NodeRegistrationQuery:       nodeRegistrationQuery,
		FeeVoteRevealVoteQuery:      feeVoteRevealVoteQuery,
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
		FeeScaleService:             feeScaleService,
		PruneQuery:                  pruneQuery,
		BlockStateStorage:           blockStateStorage,
		BlockchainStatusService:     blockchainStatusService,
		ScrambleNodeService:         scrambleNodeService,
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
	bs.Unlock()
	monitoring.DecrementStatusLockCounter(bs.Chaintype, actionType)
	monitoring.SetBlockchainStatus(bs.Chaintype, constant.BlockchainStatusIdle)
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
	blockHash, err := commonUtils.GetBlockHash(block, bs.Chaintype)
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
	err := bs.BlocksmithStrategy.IsValidSmithTime(*blocksmithIndex, int64(len(blocksmithsMap)), previousLastBlock)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSmithTime")
	}
	return nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockService) ValidateBlock(block, previousLastBlock *model.Block) error {
	if err := bs.ValidatePayloadHash(block); err != nil {
		return err
	}

	// check if blocksmith can smith at the time
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousLastBlock)
	blocksmithIndex := blocksmithsMap[string(block.BlocksmithPublicKey)]
	if blocksmithIndex == nil {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidBlocksmith")
	}
	err := bs.BlocksmithStrategy.IsBlockTimestampValid(*blocksmithIndex, int64(len(blocksmithsMap)), previousLastBlock, block)
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
		mempoolMap      storage.MempoolMap
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
	err = bs.NodeRegistrationService.BeginCacheTransaction()
	if err != nil {
		bs.queryAndCacheRollbackProcess(fmt.Sprintf("NodeRegistryCacheBeginCacheTransaction - %s", err.Error()))
		return blocker.NewBlocker(blocker.BlockErr, err.Error())
	}
	/*
		Expiring Process: expiring the transactions that affected by current block height.
		Respecting Expiring escrow and multi signature transaction before push block process
	*/
	err = bs.TransactionCoreService.ExpiringEscrowTransactions(block.GetHeight(), block.GetTimestamp(), true)
	if err != nil {
		bs.queryAndCacheRollbackProcess(fmt.Sprintf("ExpiringEscrowTransactionsErr - %s", err.Error()))
		return blocker.NewBlocker(blocker.BlockErr, err.Error())
	}
	err = bs.TransactionCoreService.ExpiringPendingTransactions(block.GetHeight(), true)
	if err != nil {
		bs.queryAndCacheRollbackProcess(fmt.Sprintf("ExpiringPendingTransactionsErr - %s", err.Error()))
		return blocker.NewBlocker(blocker.BlockErr, err.Error())
	}

	/*
		Stopping liquid payment that already passes the time
	*/
	err = bs.TransactionCoreService.CompletePassedLiquidPayment(block)
	if err != nil {
		bs.queryAndCacheRollbackProcess(fmt.Sprintf("CompletePassedLiquidPaymentErr - %s", err.Error()))
		return blocker.NewBlocker(blocker.BlockErr, err.Error())
	}

	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err = bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		bs.queryAndCacheRollbackProcess("")
		return err
	}
	var transactionIDs = make([]int64, len(block.GetTransactions()))
	mempoolMap, err = bs.MempoolService.GetMempoolTransactions()
	if err != nil {
		bs.queryAndCacheRollbackProcess("")
		return err
	}
	// apply transactions and remove them from mempool
	for index, tx := range block.GetTransactions() {
		// assign block id and block height to tx
		tx.BlockID = block.ID
		tx.Height = block.Height
		tx.TransactionIndex = uint32(index) + 1
		transactionIDs[index] = tx.GetID()
		// validate tx here
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			bs.queryAndCacheRollbackProcess("")
			return err
		}
		// check if is in mempool : if yes, undo unconfirmed
		if _, ok := mempoolMap[tx.ID]; ok {
			err = bs.TransactionCoreService.UndoApplyUnconfirmedTransaction(txType)
			if err != nil {
				bs.queryAndCacheRollbackProcess("")
				return err
			}
		}

		if block.Height > 0 {
			err = bs.TransactionCoreService.ValidateTransaction(txType, true)
			if err != nil {
				bs.queryAndCacheRollbackProcess("")
				return err
			}
		}
		// validate tx body and apply/perform transaction-specific logic
		err = bs.TransactionCoreService.ApplyConfirmedTransaction(txType, block.GetTimestamp())
		if err == nil {
			transactionInsertQuery, transactionInsertValue := bs.TransactionQuery.InsertTransaction(tx)
			err := bs.QueryExecutor.ExecuteTransaction(transactionInsertQuery, transactionInsertValue...)
			if err != nil {
				bs.queryAndCacheRollbackProcess("")
				return err
			}
		} else {
			bs.queryAndCacheRollbackProcess("")
			return err
		}
	}

	linkedCount, err := bs.PublishedReceiptService.ProcessPublishedReceipts(block)
	if err != nil {
		bs.queryAndCacheRollbackProcess("")
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
			bs.queryAndCacheRollbackProcess("")
			return err
		}
		err = bs.updatePopScore(popScore, previousBlock, block)
		if err != nil {
			bs.queryAndCacheRollbackProcess("")
			return err
		}

		activeRegistries, scoreSum, err := bs.NodeRegistrationService.GetActiveRegistryNodeWithTotalParticipationScore()
		if err != nil {
			return blocker.NewBlocker(blocker.BlockErr, "NoActiveNodeRegistriesFound")
		}

		// selecting multiple account to be rewarded and split the total coinbase + totalFees evenly between them
		totalReward := block.TotalFee + block.TotalCoinBase

		lotteryAccounts, err := bs.CoinbaseService.CoinbaseLotteryWinners(
			activeRegistries,
			scoreSum,
			block.Timestamp,
			previousBlock,
		)
		if err != nil {
			bs.queryAndCacheRollbackProcess("")
			return err
		}
		if totalReward > 0 {
			if err := bs.BlocksmithService.RewardBlocksmithAccountAddresses(
				lotteryAccounts,
				totalReward,
				block.GetTimestamp(),
				block.Height,
			); err != nil {
				bs.queryAndCacheRollbackProcess("")
				return err
			}
		}
	}
	// nodeRegistryProcess precess to admit & expel node registry
	nodeAdmissionTimestamp, err := bs.nodeRegistryProcess(block)
	if err != nil {
		bs.queryAndCacheRollbackProcess("")
		return err
	}

	// building scrambled node registry
	if block.GetHeight() == bs.ScrambleNodeService.GetBlockHeightToBuildScrambleNodes(block.GetHeight()) {
		err = bs.ScrambleNodeService.BuildScrambledNodes(block)
		if err != nil {
			bs.queryAndCacheRollbackProcess("")
			return err
		}
	}
	// persist flag will only be turned off only when generate or receive block broadcasted by another peer
	if !persist { // block content are validated
		// get blocksmith index
		blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousBlock)
		blocksmithIndex = blocksmithsMap[string(block.BlocksmithPublicKey)]
		if blocksmithIndex == nil {
			bs.queryAndCacheRollbackProcess("")
			return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInSmithingList")
		}
		// handle if is first index
		if *blocksmithIndex > 0 {
			// check if current block is in pushable window
			err = bs.BlocksmithStrategy.CanPersistBlock(*blocksmithIndex, int64(len(blocksmithsMap)), previousBlock)
			if err != nil {
				// insert into block pool
				bs.BlockPoolService.InsertBlock(block, *blocksmithIndex)
				bs.queryAndCacheRollbackProcess("")
				if broadcast {
					// create copy of the block to avoid reference update on block pool
					b := deepcopy.Copy(block)
					blockToBroadcast, ok := b.(*model.Block)
					if !ok {
						return blocker.NewBlocker(blocker.AppErr, "FailCopyingBlock")
					}
					// add transactionIDs and remove transaction before broadcast
					blockToBroadcast.TransactionIDs = transactionIDs
					blockToBroadcast.Transactions = []*model.Transaction{}
					bs.Observer.Notify(observer.BroadcastBlock, blockToBroadcast, bs.Chaintype)
				}
				return nil
			}
			// if canPersistBlock return true ignore the passed `persist` flag
		}
		// block is in first place continue to persist block to database ignoring the `persist` flag
	}

	// if genesis
	if coreUtil.IsGenesis(previousBlock.GetID(), block) {
		// insert initial fee scale
		err := bs.FeeScaleService.InsertFeeScale(&model.FeeScale{
			FeeScale:    constant.OneZBC, // initial fee_scale 1
			BlockHeight: 0,
			Latest:      true,
		})
		if err != nil {
			bs.queryAndCacheRollbackProcess("initFeeScale:rollback-error")
			return err
		}
	}
	// adjust fee if end of fee-vote period
	_, adjust, err := bs.FeeScaleService.GetCurrentPhase(block.Timestamp, false)
	if err != nil {
		return err
	}
	if adjust {
		// fetch vote-reveals
		voteInfos, err := func() ([]*model.FeeVoteInfo, error) {
			var (
				result         []*model.FeeVoteInfo
				queryResult    []*model.FeeVoteRevealVote
				err            error
				latestFeeScale model.FeeScale
			)
			err = bs.FeeScaleService.GetLatestFeeScale(&latestFeeScale)
			if err != nil {
				return result, err
			}
			qry, args := bs.FeeVoteRevealVoteQuery.GetFeeVoteRevealsInPeriod(latestFeeScale.BlockHeight, block.Height)
			rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
			if err != nil {
				return result, err
			}
			defer rows.Close()
			queryResult, err = bs.FeeVoteRevealVoteQuery.BuildModel(queryResult, rows)
			if err != nil {
				return result, err
			}
			for _, vote := range queryResult {
				result = append(result, vote.VoteInfo)
			}
			return result, nil
		}()
		if err != nil {
			bs.queryAndCacheRollbackProcess("AdjustFeeRollbackErr")
			return err
		}
		// select vote
		vote := bs.FeeScaleService.SelectVote(voteInfos, fee.SendMoneyFeeConstant)
		// insert new fee-scale
		err = bs.FeeScaleService.InsertFeeScale(&model.FeeScale{
			FeeScale:    vote,
			BlockHeight: block.Height,
			Latest:      true,
		})
		if err != nil {
			bs.queryAndCacheRollbackProcess("AdjustFeeRollbackErr")
			return err
		}
	}

	// Delete prunable data
	if block.GetHeight() > (2 * constant.MinRollbackBlocks) {
		saveHeight := block.GetHeight() - (2 * constant.MinRollbackBlocks)
		for _, pQuery := range bs.PruneQuery {
			strQuery, args := pQuery.PruneData(saveHeight, constant.PruningChunkedSize)
			err = bs.QueryExecutor.ExecuteTransaction(strQuery, args...)
			if err != nil {
				bs.queryAndCacheRollbackProcess("PruneDataRollbackErr")
				return err
			}
		}
	}
	if !coreUtil.IsGenesis(previousBlock.GetID(), block) {
		if errRemoveMempool := bs.MempoolService.RemoveMempoolTransactions(block.GetTransactions()); errRemoveMempool != nil {
			bs.queryAndCacheRollbackProcess("RemoveMempoolTransactionsRollbackErr")
			// reset mempool cache
			initMempoolErr := bs.MempoolService.InitMempoolTransaction()
			if initMempoolErr != nil {
				bs.Logger.Errorf(initMempoolErr.Error())
			}
			return err
		}
	}

	// remove a list of remove node address info in cahce
	err = bs.NodeAddressInfoService.ExecuteWaitedNodeAddressInfoCache()
	if err != nil {
		return err
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	err = bs.NodeRegistrationService.CommitCacheTransaction()
	if err != nil {
		bs.Logger.Warnf("FailToCommitNodeRegistryCache-%v", err)
		_ = bs.NodeRegistrationService.InitializeCache()
	}
	// cache last block state
	// Note: Make sure every time calling query insert & rollback block, calling this SetItem too
	err = bs.UpdateLastBlockCache(block)
	if err != nil {
		return err
	}
	// cache next node admissiom timestamp
	err = bs.NodeRegistrationService.UpdateNextNodeAdmissionCache(nodeAdmissionTimestamp)
	if err != nil {
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

	bs.BlockchainStatusService.SetLastBlock(block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype, block)
	return nil
}

// queryAndCacheRollbackProcess process to rollback data database & cache after failed execute query
func (bs *BlockService) queryAndCacheRollbackProcess(rollbackErrLable string) {
	// clear list of candidate node address info to be remove in cache
	bs.NodeAddressInfoService.ClearWaitedNodeAddressInfoCache()
	err := bs.NodeRegistrationService.RollbackCacheTransaction()
	if err != nil {
		bs.Logger.Errorf("noderegistry:cacheRollbackErr - %s", err.Error())
	}
	if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
		bs.Logger.Errorf("%s:%s", rollbackErrLable, rollbackErr.Error())
	}
}

// ScanBlockPool scan the whole block pool to check if there are any block that's legal to be pushed yet
func (bs *BlockService) ScanBlockPool() error {
	bs.ChainWriteLock(constant.BlockchainStatusReceivingBlockScanBlockPool)
	defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlockScanBlockPool)
	var (
		previousBlock model.Block
		err           error
	)
	err = bs.BlockStateStorage.GetItem(bs.Chaintype.GetTypeInt(), &previousBlock)
	if err != nil {
		return err
	}
	blocks := bs.BlockPoolService.GetBlocks()
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmiths(&previousBlock)
	for index, block := range blocks {
		err = bs.BlocksmithStrategy.CanPersistBlock(index, int64(len(blocksmithsMap)), &previousBlock)
		if err != nil {
			continue
		}

		err = bs.ValidateBlock(block, &previousBlock)
		if err != nil {
			bs.Logger.Warnf("ScanBlockPool:blockValidationFail: %v\n", blocker.NewBlocker(blocker.ValidateMainBlockErr, err.Error(), block, previousBlock))
			return blocker.NewBlocker(
				blocker.BlockErr, "ScanBlockPool:ValidateBlockFail",
			)
		}
		err = bs.PushBlock(&previousBlock, block, true, true)

		if err != nil {
			bs.Logger.Warnf("ScanBlockPool:PushBlockFail: %v\n", blocker.NewBlocker(blocker.PushMainBlockErr, err.Error(), block, previousBlock))
			return blocker.NewBlocker(
				blocker.BlockErr, "ScanBlockPool:PushBlockFail",
			)
		}
		break
	}
	return nil
}

// nodeRegistryProcess all process related with node registry at the end of push block
func (bs *BlockService) nodeRegistryProcess(
	block *model.Block,
) (*model.NodeAdmissionTimestamp, error) {
	var (
		err               error
		nextNodeAdmission *model.NodeAdmissionTimestamp
	)
	// admit nodes from registry at genesis and regular intervals
	// expel nodes from node registry as soon as they reach zero participation score
	err = bs.expelNodes(block)
	if err != nil {
		return nil, err
	}
	nextNodeAdmission, err = bs.NodeRegistrationService.GetNextNodeAdmissionTimestamp()
	if err != nil {
		return nil, err
	}
	if block.Timestamp >= nextNodeAdmission.Timestamp && block.Height != 0 {
		// insert new next node admission timestamp
		nextNodeAdmission, err = bs.NodeRegistrationService.InsertNextNodeAdmissionTimestamp(
			nextNodeAdmission.Timestamp,
			block.Height,
			true,
		)
		if err != nil {
			return nil, err
		}
		err = bs.admitNodes(block)
		if err != nil {
			return nil, err
		}
	}
	return nextNodeAdmission, nil
}

// adminNodes select and admit nodes from node registry
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

// expelNodes select and expel nodes from node registry
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

	return err
}

// GetBlockByID return a block by its ID
// withAttachedData if true returns extra attached data for the block (transactions)
func (bs *BlockService) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
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
		var transactions, err = bs.TransactionCoreService.GetTransactionsByBlockID(block.ID)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		block.Transactions = transactions
	}
	return &block, nil
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

// GetLastBlock return the last pushed block from block state storage
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	var (
		lastBlock model.Block
		err       = bs.BlockStateStorage.GetItem(nil, &lastBlock)
	)
	if err != nil {
		return nil, err
	}
	return &lastBlock, nil
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

// UpdateLastBlockCache to update the state of last block cache
func (bs *BlockService) UpdateLastBlockCache(block *model.Block) error {
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

// GenerateBlock generate block from transactions in mempool, pass empty flag to generate an empty block
func (bs *BlockService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
	empty bool,
) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinbase int64
		// only for mainchain
		sortedTransactions  []*model.Transaction
		publishedReceipts   []*model.PublishedReceipt
		err                 error
		digest              = sha3.New256()
		blockSmithPublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
		newBlockHeight      = previousBlock.Height + 1
	)

	// calculate total coinbase to be added to the block
	totalCoinbase = bs.CoinbaseService.GetCoinbase(timestamp, previousBlock.Timestamp)
	if !empty {
		sortedTransactions, err = bs.MempoolService.SelectTransactionsFromMempool(timestamp, newBlockHeight)
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
	previousBlockHash, err := commonUtils.GetBlockHash(previousBlock, bs.Chaintype)
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

	sort.SliceStable(genesisTransactions, func(i, j int) bool {
		return genesisTransactions[i].GetID() < genesisTransactions[j].GetID()
	})

	for index, tx := range genesisTransactions {
		if _, err := digest.Write(tx.TransactionHash); err != nil {
			return nil, err
		}
		if tx.TransactionType == commonUtils.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
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
		return err
	}
	return nil
}

// CheckGenesis check if genesis has been added
func (bs *BlockService) CheckGenesis() (bool, error) {
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
	// or if broadcast block is our current last block
	if !bytes.Equal(block.GetPreviousBlockHash(), lastBlock.GetBlockHash()) &&
		!bytes.Equal(block.GetPreviousBlockHash(), lastBlock.GetPreviousBlockHash()) {
		return nil, status.Error(codes.InvalidArgument, "InvalidBlock")
	}

	// check if received the exact same block as current node's last block
	if bytes.Equal(block.GetBlockHash(), lastBlock.GetBlockHash()) {
		return nil, status.Error(codes.InvalidArgument, "DuplicateBlock")
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
		return nil, status.Error(codes.InvalidArgument, "BlockFailPrevalidation")
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

	// check if already broadcast receipt to this node
	duplicated, duplicatedErr := bs.ReceiptService.IsDuplicated(senderPublicKey, block.GetBlockHash())
	if duplicatedErr != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			duplicatedErr.Error(),
		)
	}
	if duplicated {
		return nil, blocker.NewBlocker(blocker.BlockErr, "already send receipt for this block")
	}

	// generate receipt and return as response
	batchReceipt, err := bs.ReceiptService.GenerateBatchReceiptWithReminder(
		bs.Chaintype, block.GetBlockHash(),
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		constant.ReceiptDatumTypeBlock,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return batchReceipt, nil
}

func (bs *BlockService) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	var (
		publishedReceipts []*model.PublishedReceipt
		err               error
	)
	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return nil, err
	}
	minRollbackHeight := commonUtils.GetMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork process
		bs.Logger.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return nil, nil
	}

	_, err = bs.GetBlockByID(commonBlock.ID, false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var poppedBlocks []*model.Block
	block := lastBlock

	// TODO:
	// Need to refactor this codes with better solution in the future
	// https://github.com/zoobc/zoobc-core/pull/514#discussion_r355297318
	publishedReceipts, err = bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
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
		publishedReceipts, err = bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		block.PublishedReceipts = publishedReceipts
	}

	// Backup existing transactions from mempool before rollback
	// note: rollback process do inside Backup Mempools func
	err = bs.MempoolService.BackupMempools(commonBlock)
	if err != nil {
		return nil, err
	}

	// cache last block state
	// Note: Make sure every time calling query insert & rollback block, calling this SetItem too
	err = bs.UpdateLastBlockCache(nil)
	if err != nil {
		return nil, err
	}
	// update cache next node admissiom timestamp after rollback
	err = bs.NodeRegistrationService.UpdateNextNodeAdmissionCache(nil)
	if err != nil {
		return nil, err
	}
	// TODO: here we should also delete all snapshot files relative to the block manifests being rolled back during derived tables
	//  rollback. Something like this:
	//  - before rolling back derived queries, select all spine block manifest records from commonBlock.Height till last
	//  - delete all snapshots referenced by them
	//

	// remove peer memoization
	err = bs.ScrambleNodeService.PopOffScrambleToHeight(commonBlock.Height)
	if err != nil {
		return nil, err
	}
	// clear block pool
	bs.BlockPoolService.ClearBlockPool()
	// re-initialize node-registry cache
	err = bs.NodeRegistrationService.InitializeCache()
	if err != nil {
		return nil, err
	}
	// Need to sort ascending since was descended in above by Height
	sort.Slice(poppedBlocks, func(i, j int) bool {
		return poppedBlocks[i].GetHeight() < poppedBlocks[j].GetHeight()
	})

	return poppedBlocks, nil
}

// WillSmith check if blocksmith need to calculate their smith time or need to smith or not
func (bs *BlockService) WillSmith(
	blocksmith *model.Blocksmith,
	blockchainProcessorLastBlockID int64,
) (lastBlockID, blocksmithIndex int64, err error) {
	var blocksmithScore int64
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return blockchainProcessorLastBlockID, blocksmithIndex, blocker.NewBlocker(
			blocker.SmithingErr, "genesis block has not been applied")
	}

	// caching: only calculate smith time once per new block
	if lastBlock.GetID() != blockchainProcessorLastBlockID {
		blockchainProcessorLastBlockID = lastBlock.GetID()
		bs.BlocksmithStrategy.SortBlocksmiths(lastBlock, true)
		// check if eligible to create block in this round
		blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(lastBlock)
		blocksmithIdx := blocksmithsMap[string(blocksmith.NodePublicKey)]
		if blocksmithIdx == nil {
			return blockchainProcessorLastBlockID, blocksmithIndex,
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
			return 0, 0, blocker.NewBlocker(blocker.ZeroParticipationScoreErr, "participation score = 0")
		}
		err = bs.BlocksmithStrategy.CalculateScore(blocksmith, blocksmithScore)
		if err != nil {
			return blockchainProcessorLastBlockID, blocksmithIndex, err
		}
		monitoring.SetBlockchainSmithIndex(bs.GetChainType(), *blocksmithIdx)
	}
	// check for block pool duplicate
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(lastBlock)
	blocksmithIdxPtr, ok := blocksmithsMap[string(blocksmith.NodePublicKey)]
	if !ok {
		return blockchainProcessorLastBlockID, blocksmithIndex, blocker.NewBlocker(
			blocker.BlockErr, "BlocksmithNotInSmithingList",
		)
	}
	blocksmithIndex = *blocksmithIdxPtr
	blockPool := bs.BlockPoolService.GetBlock(blocksmithIndex)
	if blockPool != nil {
		return blockchainProcessorLastBlockID, blocksmithIndex, blocker.NewBlocker(
			blocker.BlockErr, "DuplicateBlockPool",
		)
	}
	// check if it's legal to create block for current blocksmith now
	err = bs.BlocksmithStrategy.IsValidSmithTime(blocksmithIndex, int64(len(blocksmithsMap)), lastBlock)
	if err == nil {
		return blockchainProcessorLastBlockID, blocksmithIndex, nil
	}
	return blockchainProcessorLastBlockID, blocksmithIndex, blocker.NewBlocker(
		blocker.SmithingErr, "NotTimeToSmithYet",
	)
}

// ProcessCompletedBlock to process block that already having all needed transactions
func (bs *BlockService) ProcessCompletedBlock(block *model.Block) error {
	bs.ChainWriteLock(constant.BlockchainStatusReceivingBlockProcessCompletedBlock)
	defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlockProcessCompletedBlock)
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
			err = bs.ValidateBlock(block, previousBlock)
			if err != nil {
				bs.Logger.Warnf("ProcessCompletedBlock:blockValidationFail: %v\n",
					blocker.NewBlocker(blocker.ValidateMainBlockErr, err.Error(), block, previousBlock))
				return status.Error(codes.InvalidArgument, "InvalidBlock")
			}
			lastBlocks, err := bs.PopOffToBlock(previousBlock)
			if err != nil {
				return err
			}

			err = bs.PushBlock(previousBlock, block, true, true)
			if err != nil {
				bs.Logger.Warn("Push ProcessCompletedBlock:fail ",
					blocker.NewBlocker(blocker.PushMainBlockErr, err.Error(), block, previousBlock))
				errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false, true)
				if errPushBlock != nil {
					bs.Logger.Errorf("ProcessCompletedBlock pushing back popped off block fail: %v",
						blocker.NewBlocker(blocker.PushMainBlockErr, err.Error(), block, previousBlock))
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
	err = bs.ValidateBlock(block, lastBlock)
	if err != nil {
		bs.Logger.Warnf("ProcessCompletedBlock2:blockValidationFail: %v\n", blocker.NewBlocker(blocker.ValidateMainBlockErr, err.Error(), block, lastBlock))
		return status.Error(codes.InvalidArgument, "InvalidBlock")
	}
	err = bs.PushBlock(lastBlock, block, true, false)
	if err != nil {
		bs.Logger.Errorf("ProcessCompletedBlock2 push Block fail: %v", blocker.NewBlocker(blocker.PushMainBlockErr, err.Error(), block, lastBlock))
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
		txRequiredByBlock = make(TransactionIDsMap)
	)
	block.Transactions = make([]*model.Transaction, len(block.GetTransactionIDs()))
	for idx, txID := range block.TransactionIDs {
		txRequiredByBlock[txID] = idx
	}

	// find needed transactions in mempool
	mempoolCacheObjects, err := bs.MempoolService.GetMempoolTransactions()
	if err != nil {
		return false, err
	}

	for txID, txIdx := range txRequiredByBlock {
		if memObj, ok := mempoolCacheObjects[txID]; ok {
			block.Transactions[txIdx] = &memObj.Tx
			delete(txRequiredByBlock, memObj.Tx.GetID())
		}
	}
	// process when needed transactions are completed
	if len(txRequiredByBlock) == 0 {
		err := bs.ProcessCompletedBlock(block)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	// check if block has any txIDs that're already in `transactions` table, if yes, the block is rejected for
	// including applied txs
	var txIds []int64
	for txID := range txRequiredByBlock {
		txIds = append(txIds, txID)
	}
	duplicateTxs, err := bs.TransactionCoreService.GetTransactionsByIds(txIds)
	if err != nil {
		return false, err
	}
	if len(duplicateTxs) > 0 {
		return false, blocker.NewBlocker(blocker.ValidationErr, "BlockContainAppliedTransactions")
	}
	// saving temporary block
	bs.BlockIncompleteQueueService.AddBlockQueue(block)
	bs.BlockIncompleteQueueService.SetTransactionsRequired(block.GetID(), txRequiredByBlock)

	if peer == nil {
		bs.Logger.Errorf("Error peer is null, can not request block transactions from the Peer")
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

// BlockTransactionsRequestedListener will send the transactions required by blocks
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
