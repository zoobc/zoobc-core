package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithStrategyInterface interface {
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		SortBlocksmiths(block *model.Block, withLock bool)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		GetSortedBlocksmithsMap(block *model.Block) map[string]*int64
		CalculateScore(generator *model.Blocksmith, score int64) error
		IsBlockTimestampValid(blocksmithIndex int64, previousBlock, currentBlock *model.Block) error
		CanPersistBlock(
			blocksmithIndex int64,
			previousBlock *model.Block,
		) error
		IsValidSmithTime(
			blocksmithIndex int64,
			previousBlock *model.Block,
		) error
	}
)
