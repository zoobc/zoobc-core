package strategy

import (
	"math"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlocksmithStrategyMain struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		LastSortedBlockID                      int64
		LastEstimatedBlockPersistedTimestamp   int64
		LastEstimatedPersistedTimestampBlockID int64
		SortedBlocksmithsLock                  sync.RWMutex
		SortedBlocksmithsMap                   map[string]*int64
	}
)

func NewBlocksmithStrategyMain(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	logger *log.Logger,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		QueryExecutor:          queryExecutor,
		NodeRegistrationQuery:  nodeRegistrationQuery,
		SkippedBlocksmithQuery: skippedBlocksmithQuery,
		Logger:                 logger,
		SortedBlocksmithsMap:   make(map[string]*int64),
	}
}

// GetBlocksmiths select the blocksmiths for a given block and calculate the SmithOrder (for smithing) and NodeOrder (for block rewards)
func (bss *BlocksmithStrategyMain) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	var (
		activeBlocksmiths, blocksmiths []*model.Blocksmith
	)
	// get all registered nodes with participation score > 0
	rows, err := bss.QueryExecutor.ExecuteSelect(bss.NodeRegistrationQuery.GetActiveNodeRegistrationsByHeight(
		block.Height), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activeBlocksmiths, err = bss.NodeRegistrationQuery.BuildBlocksmith(activeBlocksmiths, rows)
	if err != nil {
		return nil, err
	}
	monitoring.SetNodeScore(activeBlocksmiths)
	monitoring.SetActiveRegisteredNodesCount(len(activeBlocksmiths))
	// add smithorder and nodeorder to be used to select blocksmith and coinbase rewards
	for _, blocksmith := range activeBlocksmiths {
		blocksmith.BlockSeed, err = coreUtil.GetBlockSeed(blocksmith.NodeID, block)
		if err != nil {
			return nil, err
		}
		blocksmith.NodeOrder = coreUtil.CalculateNodeOrder(blocksmith.Score, blocksmith.BlockSeed, blocksmith.NodeID)
		blocksmiths = append(blocksmiths, blocksmith)
	}
	return blocksmiths, nil
}

func (bss *BlocksmithStrategyMain) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	copy(result, bss.SortedBlocksmiths)
	return result
}

// GetSortedBlocksmithsMap get the sorted blocksmiths in map
func (bss *BlocksmithStrategyMain) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	var (
		result = make(map[string]*int64)
	)
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	for k, v := range bss.SortedBlocksmithsMap {
		result[k] = v
	}
	return result
}

// SortBlocksmiths sort the list of active node of current block height, index start from 0.
func (bss *BlocksmithStrategyMain) SortBlocksmiths(block *model.Block, withLock bool) {
	if block.ID == bss.LastSortedBlockID && block.ID != constant.MainchainGenesisBlockID {
		return
	}
	// fetch valid blocksmiths
	var blocksmiths []*model.Blocksmith
	nextBlocksmiths, err := bss.GetBlocksmiths(block)
	if err != nil {
		bss.Logger.Errorf("SortBlocksmith (Main):GetBlocksmiths fail: %s", err)
		return
	}
	// copy the nextBlocksmiths pointers array into an array of blocksmiths
	blocksmiths = append(blocksmiths, nextBlocksmiths...)
	// sort blocksmiths by SmithOrder
	sort.SliceStable(blocksmiths, func(i, j int) bool {
		if blocksmiths[i].BlockSeed == blocksmiths[j].BlockSeed {
			return blocksmiths[i].NodeID < blocksmiths[j].NodeID
		}
		// ascending sort
		return blocksmiths[i].BlockSeed < blocksmiths[j].BlockSeed
	})

	if withLock {
		bss.SortedBlocksmithsLock.Lock()
		defer bss.SortedBlocksmithsLock.Unlock()
	}
	// copying the sorted list to map[string(publicKey)]index
	for index, blocksmith := range blocksmiths {
		blocksmithIndex := int64(index)
		bss.SortedBlocksmithsMap[string(blocksmith.NodePublicKey)] = &blocksmithIndex
	}
	// set last sorted block id
	bss.LastSortedBlockID = block.ID
	bss.SortedBlocksmiths = blocksmiths
}

// CalculateScore calculate the blocksmith score
func (bss *BlocksmithStrategyMain) CalculateScore(generator *model.Blocksmith, score int64) error {
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	return nil
}

func (bss *BlocksmithStrategyMain) EstimateLastBlockPersistedTime(
	previousBlock *model.Block,
	ct chaintype.ChainType,
) error {
	var (
		skippedBlocksmiths []*model.SkippedBlocksmith
	)
	skippedBlocksmiths, err := func() ([]*model.SkippedBlocksmith, error) {
		skippedBlocksmithsQ := bss.SkippedBlocksmithQuery.GetSkippedBlocksmithsByBlockHeight(previousBlock.Height)
		skippedBlocksmithsRows, err := bss.QueryExecutor.ExecuteSelect(skippedBlocksmithsQ, false)
		if err != nil {
			return nil, err
		}
		defer skippedBlocksmithsRows.Close()
		if _, err := bss.SkippedBlocksmithQuery.BuildModel(skippedBlocksmiths, skippedBlocksmithsRows); err != nil {
			return nil, err
		}
		return skippedBlocksmiths, nil
	}()
	if err != nil {
		return err
	}
	if len(skippedBlocksmiths) > 0 {
		previousBlocksmithTime := previousBlock.Timestamp - ct.GetBlocksmithTimeGap()
		estimatedPreviousBlockPersistTime := previousBlocksmithTime + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
		bss.LastEstimatedBlockPersistedTimestamp = estimatedPreviousBlockPersistTime
	} else {
		bss.LastEstimatedBlockPersistedTimestamp = previousBlock.GetTimestamp()
	}
	bss.LastEstimatedPersistedTimestampBlockID = previousBlock.ID
	return nil
}

// IsBlockTimestampValid check if the block provided (currentBlock) has valid timestamp based on the previous block
// of the current node. This function is save to be called on download process, it does not make use of node current time.
func (bss *BlocksmithStrategyMain) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	var (
		err error
		ct  = &chaintype.MainChain{}
	)
	// calculate estimated starting time
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		err = bss.EstimateLastBlockPersistedTime(previousBlock, ct)
		if err != nil {
			return err
		}
	}
	// check if is valid time
	timeGapCurrentLastBlock := currentBlock.GetTimestamp() - bss.LastEstimatedBlockPersistedTimestamp
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()
	// exception: first blocksmith check
	if blocksmithIndex == 0 && timeGapCurrentLastBlock >= ct.GetSmithingPeriod() {
		if timeGapCurrentLastBlock <= ct.GetSmithingPeriod()+ct.GetBlocksmithBlockCreationTime()+ct.GetBlocksmithNetworkTolerance() {
			return nil
		}
	}
	remainder := (timeGapCurrentLastBlock - ct.GetSmithingPeriod()) % timeForOneRound
	if remainder >= blocksmithIndex*ct.GetBlocksmithTimeGap() {
		if remainder > ct.GetBlocksmithTimeGap()+ct.GetBlocksmithBlockCreationTime()+ct.GetBlocksmithNetworkTolerance() {
			return blocker.NewBlocker(blocker.BlockErr, "BlocksmithExpired")
		}
		return nil
	}
	return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
}

// CanPersistBlock check if currentTime is a time to persist the provided block.
// This function uses current node time, which make it unsafe to validate past block.
// numberOfBlocksmiths must be > 0
func (bss *BlocksmithStrategyMain) CanPersistBlock(
	blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
	var (
		err                                         error
		ct                                          = &chaintype.MainChain{}
		currentTime                                 = time.Now().Unix()
		remainder, prevRoundBegin, prevRoundExpired int64
	)
	// always return true for the first block | keeping in mind genesis block's timestamps is far behind, let fork processor
	// handle to get highest cum-diff block
	if previousBlock.GetHeight() == 0 {
		return nil
	}
	// calculate estimated starting time
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		err = bss.EstimateLastBlockPersistedTime(previousBlock, ct)
		if err != nil {
			return err
		}
	}
	// check if is valid time
	// calculate total time before every blocksmiths are skipped
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()
	timeSinceLastBlock := currentTime - bss.LastEstimatedBlockPersistedTimestamp
	if timeSinceLastBlock < ct.GetSmithingPeriod() {
		return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
	}
	modTimeSinceLastBlock := timeSinceLastBlock - ct.GetSmithingPeriod()
	timeRound := math.Floor(float64(modTimeSinceLastBlock) / float64(timeForOneRound))
	remainder = modTimeSinceLastBlock % timeForOneRound
	nearestRoundBeginning := currentTime - remainder
	if timeRound > 0 { // if more than one round has passed, calculate previous round start-expiry time for overlap
		prevRoundStart := nearestRoundBeginning - timeForOneRound
		prevRoundBegin = prevRoundStart + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRoundExpired = prevRoundBegin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
	// calculate current round begin and expiry time
	allowedBeginTime := blocksmithIndex*ct.GetBlocksmithTimeGap() + nearestRoundBeginning
	expiredTime := allowedBeginTime + ct.GetBlocksmithBlockCreationTime() +
		ct.GetBlocksmithNetworkTolerance()
	// check if current time is in {(expire-timeGap) < x < (expire)} in either previous round or current round
	if (currentTime > (expiredTime-ct.GetBlocksmithTimeGap()) && currentTime <= expiredTime) ||
		(currentTime > (prevRoundExpired-ct.GetBlocksmithTimeGap()) && currentTime <= prevRoundExpired) {
		return nil
	}
	return blocker.NewBlocker(blocker.BlockErr, "CannotPersistBlock")
}

func (bss *BlocksmithStrategyMain) IsValidSmithTime(
	blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
	var (
		err                                         error
		currentTime                                 = time.Now().Unix()
		ct                                          = &chaintype.MainChain{}
		remainder, prevRoundBegin, prevRoundExpired int64
	)
	// calculate estimated starting time
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		err = bss.EstimateLastBlockPersistedTime(previousBlock, ct)
		if err != nil {
			return err
		}
	}
	// calculate total time before every blocksmiths are skipped
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()

	timeSinceLastBlock := currentTime - bss.LastEstimatedBlockPersistedTimestamp
	if timeSinceLastBlock < ct.GetSmithingPeriod() {
		return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
	}
	modTimeSinceLastBlock := timeSinceLastBlock - ct.GetSmithingPeriod()
	timeRound := math.Floor(float64(modTimeSinceLastBlock) / float64(timeForOneRound))
	remainder = modTimeSinceLastBlock % timeForOneRound
	// find the time of beginning of the list
	nearestRoundBeginning := currentTime - remainder
	if timeRound > 0 { // if more than one round has passed, calculate previous round start-expiry time for overlap
		prevRoundStart := nearestRoundBeginning - timeForOneRound
		prevRoundBegin = prevRoundStart + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRoundExpired = prevRoundBegin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
	// calculate current round begin and expiry time
	allowedBeginTime := blocksmithIndex*ct.GetBlocksmithTimeGap() + nearestRoundBeginning
	expiredTime := allowedBeginTime + ct.GetBlocksmithBlockCreationTime() +
		ct.GetBlocksmithNetworkTolerance()
	// if currentTime overlap with either currentRound window or previous round window, it's considered valid time
	if (currentTime >= allowedBeginTime && currentTime <= expiredTime) ||
		(currentTime >= prevRoundBegin && currentTime <= prevRoundExpired) {
		return nil
	}
	return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
}
