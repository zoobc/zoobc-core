package storage

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	// Blockstorage will cache last 720 blocks
	BlocksStorage struct {
		sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	// BlockCacheObject represent selected field from model.Block want to cache
	BlockCacheObject struct {
		ID        int64
		Height    uint32
		BlockHash []byte
	}
)

func NewBlocksStorage() *BlocksStorage {
	return &BlocksStorage{
		itemLimit: int(constant.MaxBlocksCacheStorage),
		blocks:    make([]BlockCacheObject, 0, constant.MinRollbackBlocks),
	}
}

func (b *BlocksStorage) Pop() error {
	return nil
}

// Push add new item into list & remove the oldest one if needed
func (b *BlocksStorage) Push(item interface{}) error {
	blockCacheObjectCopy, ok := item.(BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObject")
	}
	b.Lock()
	defer b.Unlock()
	if len(b.blocks) >= b.itemLimit {
		if len(b.blocks) != 0 {
			b.blocks = b.blocks[1:] // remove first (oldest) cache to make room for new block
		}
	}
	b.blocks = append(b.blocks, b.copy(blockCacheObjectCopy))
	b.lastBlockHeight = blockCacheObjectCopy.Height
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBlocksCacheStorage, float64(b.size()))
	}
	return nil
}

// PopTo pop the cache blocks from the provided height to the last height
func (b *BlocksStorage) PopTo(height uint32) error {
	if height > b.lastBlockHeight {
		return blocker.NewBlocker(blocker.ValidationErr, "HeightOutOfRange")
	}
	var rangePop = int(b.lastBlockHeight - height)
	if rangePop > len(b.blocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "NumberPopOutOfRange")
	}
	var (
		lastIndex   = len(b.blocks) - 1
		heightIndex = lastIndex - rangePop
	)
	b.Lock()
	defer b.Unlock()
	b.blocks = b.blocks[:heightIndex+1]
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBlocksCacheStorage, float64(b.size()))
	}
	return nil
}

func (b *BlocksStorage) GetAll(items interface{}) error {
	blocksCopy, ok := items.(*[]BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObjectList")
	}
	b.Lock()
	defer b.Unlock()
	*blocksCopy = make([]BlockCacheObject, len(b.blocks))
	for i := range b.blocks {
		(*blocksCopy)[i] = b.copy(b.blocks[i])
	}
	return nil
}

// GetAtIndex get block cache object based on given index
func (b *BlocksStorage) GetAtIndex(height uint32, item interface{}) error {
	if height > b.lastBlockHeight {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	blockCacheObjCopy, ok := item.(*BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObject")
	}
	b.RLock()
	defer b.RUnlock()
	var (
		lastIndex   = len(b.blocks) - 1
		heightIndex = lastIndex - int(b.lastBlockHeight-height)
	)
	*blockCacheObjCopy = b.copy(b.blocks[heightIndex])
	return nil
}

func (b *BlocksStorage) GetTop(item interface{}) error {
	b.RLock()
	defer b.RUnlock()
	topIndex := len(b.blocks) - 1
	if topIndex == -1 {
		return blocker.NewBlocker(blocker.CacheEmpty, "EmptyBlocksCache")
	}
	blockCacheObjCopy, ok := item.(*BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObject")
	}
	*blockCacheObjCopy = b.copy(b.blocks[topIndex])
	return nil
}

func (b *BlocksStorage) Clear() error {
	b.blocks = make([]BlockCacheObject, 0, b.itemLimit)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBlocksCacheStorage, 0)
	}
	return nil
}

func (b *BlocksStorage) size() int {
	var (
		blocksBytes bytes.Buffer
		enc         = gob.NewEncoder(&blocksBytes)
	)
	_ = enc.Encode(b.blocks)
	_ = enc.Encode(b.itemLimit)
	_ = enc.Encode(b.lastBlockHeight)
	return blocksBytes.Len()
}

func (b *BlocksStorage) copy(blockCacheObject BlockCacheObject) (blockCacheObjectCopy BlockCacheObject) {
	blockCacheObjectCopy = blockCacheObject
	// copy array type to remove reference
	copy(blockCacheObjectCopy.BlockHash, blockCacheObject.BlockHash)
	return blockCacheObjectCopy
}
