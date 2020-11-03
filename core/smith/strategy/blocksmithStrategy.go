package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlocksmithStrategyInterface interface for finding the valid blocksmith (block creator)
	BlocksmithStrategyInterface interface {
		// WillSmith is used by node to check if it is its own time to create block yet
		WillSmith(prevBlock *model.Block) (int64, error)
		// IsBlockValid validate if provided `block` is valid given the previousBlock
		IsBlockValid(prevBlock, block *model.Block) error
		// CalculateCumulativeDifficulty calculate quality of block by giving the previous block as
		// base calculation
		CalculateCumulativeDifficulty(prevBlock, block *model.Block) string
		// GetBlocksBlocksmiths return the candidates of blocksmith of provided block, by calculating the time-gap
		// between previousBlock and block
		GetBlocksBlocksmiths(previousBlock, block *model.Block) ([]*model.Blocksmith, error)
		// GetSmithingRound get the number of time we should be running a random number generate given previousBlock
		// and block
		GetSmithingRound(previousBlock, block *model.Block) int
		// CanPersistBlock check if block can be persisted or not (from block-pool to database)
		CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error
	}
)
