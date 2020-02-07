package strategy

import (
	"math"
	"math/big"
	"sort"
	"sync"

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

// CalculateSmith calculate seed, smithTime, and Deadline for mainchain
func (bss *BlocksmithStrategyMain) CalculateSmith(
	lastBlock *model.Block,
	blocksmithIndex int64,
	generator *model.Blocksmith,
	score int64,
) error {
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	generator.SmithTime = bss.GetSmithTime(blocksmithIndex, lastBlock)
	return nil
}

// GetSmithTime calculate smith time of a blocksmith for the new block by providing the blocksmith index and previous block
func (bss *BlocksmithStrategyMain) GetSmithTime(blocksmithIndex int64, previousBlock *model.Block) int64 {
	var (
		elapsedFromLastBlock int64
		skippedBlocksmiths   []*model.SkippedBlocksmith
	)
	ct := &chaintype.MainChain{}
	if blocksmithIndex < 1 {
		elapsedFromLastBlock = ct.GetSmithingPeriod()
	} else {
		elapsedFromLastBlock = blocksmithIndex*constant.SmithingBlocksmithTimeGap + ct.GetSmithingPeriod()
	}
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		skippedBlocksmithsQ := bss.SkippedBlocksmithQuery.GetSkippedBlocksmithsByBlockHeight(previousBlock.Height)
		skippedBlocksmithsRows, err := bss.QueryExecutor.ExecuteSelect(skippedBlocksmithsQ, false)
		if err != nil {
			bss.Logger.Error("GetSmithTimeError: ", err)
			return math.MaxInt64 // todo: how to return the error
		}
		defer skippedBlocksmithsRows.Close()
		skippedBlocksmiths, err = bss.SkippedBlocksmithQuery.BuildModel(skippedBlocksmiths, skippedBlocksmithsRows)
		if err != nil {
			bss.Logger.Error("GetSmithTimeError: ", err)
			return math.MaxInt64 // todo: how to return the error
		}
		if len(skippedBlocksmiths) > 0 {
			previousBlocksmithTime := previousBlock.Timestamp - constant.SmithingBlocksmithTimeGap
			estimatedPreviousBlockPersistTime := previousBlocksmithTime + constant.SmithingBlockCreationTime +
				constant.SmithingNetworkTolerance
			bss.LastEstimatedBlockPersistedTimestamp = estimatedPreviousBlockPersistTime
		} else {
			bss.LastEstimatedBlockPersistedTimestamp = previousBlock.GetTimestamp()
		}
		bss.LastEstimatedPersistedTimestampBlockID = previousBlock.ID
	}
	return bss.LastEstimatedBlockPersistedTimestamp + elapsedFromLastBlock
}
