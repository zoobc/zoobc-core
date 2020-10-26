package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithStrategyInterface interface {
		WillSmith(prevBlock *model.Block) (int64, error)
		IsBlockValid(prevBlock, block *model.Block) error
		CalculateCumulativeDifficulty(prevBlock, block *model.Block) string
		GetBlocksBlocksmiths(previousBlock, block *model.Block) ([]*model.Blocksmith, error)

		// old
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		CalculateScore(generator *model.Blocksmith, score int64) error
		GetSmithingRound(previousBlock, block *model.Block) int
		CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error
	}
)
