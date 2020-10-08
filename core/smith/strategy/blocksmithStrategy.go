package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithStrategyInterface interface {
		WillSmith(prevBlock *model.Block) (int64, int64, error)
		IsBlockValid(prevBlock, block *model.Block) error
		isMe(lastCandidate Candidate, block *model.Block) bool

		// old
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		SortBlocksmiths(block *model.Block, withLock bool)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		GetSortedBlocksmithsMap(block *model.Block) map[string]*int64
		CalculateScore(generator *model.Blocksmith, score int64) error
		IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock, currentBlock *model.Block) error
		CanPersistBlock(blocksmithIndex, numberOfBlocksmiths int64, previousBlock *model.Block) error
		IsValidSmithTime(
			blocksmithIndex, numberOfBlocksmiths int64,
			previousBlock *model.Block,

		) error
	}
)
