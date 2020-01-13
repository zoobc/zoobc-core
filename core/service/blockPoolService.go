package service

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlockPoolServiceInterface interface the block pool to smithing process
	BlockPoolServiceInterface interface {
		GetBlocks() map[int64]*model.Block
		GetBlock(index int64) *model.Block
		InsertBlock(block *model.Block, index int64)
		ClearBlockPool()
	}
	BlockPoolService struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
)

func NewBlockPoolService() *BlockPoolService {
	return &BlockPoolService{
		BlockQueue: make(map[int64]*model.Block),
	}
}

func (bps *BlockPoolService) GetBlocks() map[int64]*model.Block {
	var result = make(map[int64]*model.Block)
	bps.BlockQueueLock.RLock()
	defer bps.BlockQueueLock.RUnlock()
	for k, v := range bps.BlockQueue {
		result[k] = v
	}
	return result
}
func (bps *BlockPoolService) GetBlock(index int64) *model.Block {
	bps.BlockQueueLock.RLock()
	defer bps.BlockQueueLock.RUnlock()
	block := bps.BlockQueue[index]
	return block
}

func (bps *BlockPoolService) InsertBlock(block *model.Block, index int64) {
	bps.BlockQueueLock.Lock()
	defer bps.BlockQueueLock.Unlock()
	bps.BlockQueue[index] = block
}

func (bps *BlockPoolService) ClearBlockPool() {
	bps.BlockQueueLock.Lock()
	defer bps.BlockQueueLock.Unlock()
	for k := range bps.BlockQueue {
		delete(bps.BlockQueue, k)
	}
}
