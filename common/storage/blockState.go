package storage

import (
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlockStateStorage represent last state of block
	BlockStateStorage struct {
		sync.RWMutex
		// save last block in bytes to make esier to convert when requesting the block
		lastBlockBytes []byte
	}
)

// NewBlockStateStorage returns BlockStateStorage instance
func NewBlockStateStorage() *BlockStateStorage {
	return &BlockStateStorage{}
}

// SetItem setter of BlockStateStorage
func (bs *BlockStateStorage) SetItem(lastUpdate, block interface{}) error {
	bs.Lock()
	defer bs.Unlock()
	var (
		newBlock, ok = (block).(model.Block)
	)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	newBlockBytes, err := json.Marshal(newBlock)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "Failed marshal block")
	}
	bs.lastBlockBytes = newBlockBytes
	return nil
}

func (bs *BlockStateStorage) SetItems(_ interface{}) error {
	return nil
}

// GetItem getter of BlockStateStorage
func (bs *BlockStateStorage) GetItem(lastUpdate, block interface{}) error {
	bs.RLock()
	defer bs.RUnlock()

	var (
		ok        bool
		err       error
		blockCopy *model.Block
	)

	if bs.lastBlockBytes == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptyCache")
	}

	blockCopy, ok = block.(*model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	err = json.Unmarshal(bs.lastBlockBytes, blockCopy)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "Failed unmarshal block bytes")
	}
	return nil
}

func (bs *BlockStateStorage) GetAllItems(item interface{}) error {
	return nil
}

func (bs *BlockStateStorage) GetTotalItems() int {
	// this storage only have 1 item
	return 1
}

func (bs *BlockStateStorage) RemoveItem(key interface{}) error {
	return nil
}

// GetSize return the size of BlockStateStorage
func (bs *BlockStateStorage) GetSize() int64 {
	return int64(binary.Size(bs.lastBlockBytes))
}

// ClearCache cleaner of BlockStateStorage
func (bs *BlockStateStorage) ClearCache() error {
	bs.lastBlockBytes = nil
	return nil
}
