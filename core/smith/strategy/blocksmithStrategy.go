package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithStrategyInterface interface {
		WillSmith(prevBlock *model.Block) (int64, int64, error)
		IsBlockValid(prevBlock, block *model.Block) error
		isMe(lastCandidate Candidate, block *model.Block) bool
		CalculateCumulativeDifficulty(prevBlock, block *model.Block) string

		// old
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		CalculateScore(generator *model.Blocksmith, score int64) error
		GetSmithingRound(previousBlock, block *model.Block) int
		CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error
	}
)
