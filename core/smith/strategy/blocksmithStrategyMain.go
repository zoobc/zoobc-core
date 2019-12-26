package strategy

import (
	"math/big"
	"sort"
	"sync"

	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlocksmithStrategyMain struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
	}
)

func NewBlocksmithStrategyMain(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	logger *log.Logger,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		QueryExecutor:         queryExecutor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		Logger:                logger,
		SortedBlocksmithsMap:  make(map[string]*int64),
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
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block)
	}
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	copy(result, bss.SortedBlocksmiths)
	return result
}

// GetSortedBlocksmithsMap get the sorted blocksmiths in map
func (bss *BlocksmithStrategyMain) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	var (
		result = make(map[string]*int64)
	)
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block)
	}
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	for k, v := range bss.SortedBlocksmithsMap {
		result[k] = v
	}
	return result
}

func (bss *BlocksmithStrategyMain) SortBlocksmiths(block *model.Block) {
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
		bi, bj := blocksmiths[i], blocksmiths[j]
		res := bi.BlockSeed - bj.BlockSeed
		if res == 0 {
			res = bi.NodeID - bj.NodeID
		}
		// ascending sort
		return res < 0
	})
	bss.SortedBlocksmithsLock.Lock()
	defer bss.SortedBlocksmithsLock.Unlock()
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

// GetSmithTime calculate smith time of a blocksmith
func (bss *BlocksmithStrategyMain) GetSmithTime(blocksmithIndex int64, block *model.Block) int64 {
	elapsedFromLastBlock := (blocksmithIndex + 1) * constant.SmithingStartTimeMain
	return block.GetTimestamp() + elapsedFromLastBlock
}
