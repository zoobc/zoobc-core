package service

import (
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlocksmithServiceInterface interface {
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		SortBlocksmiths(block *model.Block)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		GetSortedBlocksmithsMap(block *model.Block) map[string]*int64
	}
	BlocksmithService struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		LastSortedBlockHeight    uint32
		SortedBlocksmithsMapLock sync.RWMutex
		SortedBlocksmithsMap     map[string]*int64
	}
)

func NewBlocksmithService(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	logger *log.Logger,
) *BlocksmithService {
	return &BlocksmithService{
		QueryExecutor:         queryExecutor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		Logger:                logger,
		SortedBlocksmithsMap:  make(map[string]*int64),
	}
}

// GetBlocksmiths select the blocksmiths for a given block and calculate the SmithOrder (for smithing) and NodeOrder (for block rewards)
func (bss *BlocksmithService) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
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

func (bss *BlocksmithService) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	if block.Height != bss.LastSortedBlockHeight || block.Height == 0 {
		bss.SortBlocksmiths(block)
	}
	return bss.SortedBlocksmiths
}

func (bss *BlocksmithService) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	if block.Height != bss.LastSortedBlockHeight || block.Height == 0 {
		bss.SortBlocksmiths(block)
	}
	bss.SortedBlocksmithsMapLock.RLock()
	defer bss.SortedBlocksmithsMapLock.RUnlock()
	return bss.SortedBlocksmithsMap
}

func (bss *BlocksmithService) SortBlocksmiths(block *model.Block) {
	if block.Height == bss.LastSortedBlockHeight && block.Height != 0 {
		return
	}
	// fetch valid blocksmiths
	var blocksmiths []*model.Blocksmith
	nextBlocksmiths, err := bss.GetBlocksmiths(block)
	if err != nil {
		bss.Logger.Errorf("SortBlocksmith: %s", err)
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
	bss.copyBlocksmithsToMap(blocksmiths)
	bss.SortedBlocksmiths = blocksmiths
}

func (bss *BlocksmithService) copyBlocksmithsToMap(blocksmiths []*model.Blocksmith) {
	bss.SortedBlocksmithsMapLock.Lock()
	defer bss.SortedBlocksmithsMapLock.Unlock()
	for index, blocksmith := range blocksmiths {
		blocksmithIndex := int64(index)
		bss.SortedBlocksmithsMap[string(blocksmith.NodePublicKey)] = &blocksmithIndex
	}
}
