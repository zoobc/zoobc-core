package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlocksmithServiceInterface interface {
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
		SortBlocksmiths(block *model.Block)
		GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith
		GetSortedBlocksmithsMap(block *model.Block) map[string]*int64
	}
)
