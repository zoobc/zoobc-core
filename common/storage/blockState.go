package storage

import (
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlockStateStorage represent last state of block
	BlockStateStorage struct {
		sync.RWMutex
		lastBlock model.Block
	}
)

// NewBlockStateStorage returns BlockStateStorage instance
func NewBlockStateStorage() *BlockStateStorage {
	return &BlockStateStorage{
		lastBlock: model.Block{},
	}
}

// SetItem setter of BlockStateStorage
func (bs *BlockStateStorage) SetItem(lastUpdate, block interface{}) error {
	bs.Lock()
	defer bs.Unlock()
	var (
		ok       bool
		newBlock model.Block
	)

	// copy block
	newBlock, ok = (deepcopy.Copy(block)).(model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item or FailCopyingBlock")
	}
	bs.lastBlock = newBlock
	return nil
}

// GetItem getter of BlockStateStorage
func (bs *BlockStateStorage) GetItem(lastUpdate, block interface{}) error {
	bs.RLock()
	defer bs.RUnlock()

	var (
		ok        bool
		blockCopy *model.Block
	)

	if bs.lastBlock.BlockHash == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptyCache")
	}

	blockCopy, ok = block.(*model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	// copy cache block value into reference variable requester
	*blockCopy, ok = (deepcopy.Copy(bs.lastBlock)).(model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	return nil
}

func (bs *BlockStateStorage) GetAllItems(item interface{}) error {
	return nil
}

func (bs *BlockStateStorage) RemoveItem(key interface{}) error {
	return nil
}

// GetSize return the size of BlockStateStorage
func (bs *BlockStateStorage) GetSize() int64 {
	return int64(bs.lastBlock.XXX_Size())
}

// ClearCache cleaner of BlockStateStorage
func (bs *BlockStateStorage) ClearCache() error {
	bs.lastBlock = model.Block{}
	return nil
}
