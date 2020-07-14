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
		blocks map[int32]model.Block
	}
)

var blockStateStorageInstance *BlockStateStorage

// NewBlockStateStorage returns BlockStateStorage instance
func NewBlockStateStorage(chainTypeInt int32, block model.Block) *BlockStateStorage {
	if blockStateStorageInstance == nil {
		blockStateStorageInstance = &BlockStateStorage{
			blocks: map[int32]model.Block{
				chainTypeInt: block,
			},
		}
		return blockStateStorageInstance
	}
	blockStateStorageInstance.blocks[chainTypeInt] = block
	return blockStateStorageInstance
}

// SetItem setter of BlockStateStorage
func (bs *BlockStateStorage) SetItem(chaintypeInt, block interface{}) error {
	bs.Lock()
	defer bs.Unlock()
	var (
		ok           bool
		chainTypeInt int32
		newBlock     model.Block
	)
	// todo? : make sure integer of chaintype is existing chaintype
	chainTypeInt, ok = chaintypeInt.(int32)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType  lastChange, expected int32")
	}

	newBlock, ok = (deepcopy.Copy(block)).(model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item or FailCopyingBlock")
	}

	bs.blocks[chainTypeInt] = newBlock
	return nil
}

// GetItem getter of BlockStateStorage
func (bs *BlockStateStorage) GetItem(chaintypeInt, block interface{}) error {
	bs.RLock()
	defer bs.RUnlock()

	var (
		ok           bool
		chainTypeInt int32
		blockCopy    *model.Block
	)
	chainTypeInt, ok = chaintypeInt.(int32)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType lastChange, expected int32")
	}

	blockCopy, ok = block.(*model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	// NOTE: when BlockHash is nil it means empty block
	if bs.blocks[chainTypeInt].BlockHash == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "Chaintype not found")
	}
	// copy chache block value into reference variable requester
	*blockCopy, ok = (deepcopy.Copy(bs.blocks[chainTypeInt])).(model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	return nil
}

// GetSize return the size of BlockStateStorage
func (bs *BlockStateStorage) GetSize() int64 {
	var size int64
	for _, block := range bs.blocks {
		size += int64(block.XXX_Size())
	}
	return size
}

// ClearCache cleaner of BlockStateStorage
func (bs *BlockStateStorage) ClearCache() error {
	bs.blocks = make(map[int32]model.Block)
	return nil
}
