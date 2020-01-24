package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithStrategyInterface interface {
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		SortBlocksmiths(block *model.Block)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		GetSortedBlocksmithsMap(block *model.Block) map[string]*int64
		CalculateSmith(lastBlock *model.Block, blocksmithIndex int64, generator *model.Blocksmith, score int64) error
		GetSmithTime(blocksmithIndex int64, block *model.Block) int64
	}
)
