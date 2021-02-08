// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
		metricLabel     monitoring.CacheStorageType
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
		blocksMapID     map[int64]*int
	}
	// BlockCacheObject represent selected field from model.Block want to cache
	BlockCacheObject struct {
		ID        int64
		Height    uint32
		Timestamp int64
		BlockHash []byte
	}
)

func NewBlocksStorage(metricLabel monitoring.CacheStorageType) *BlocksStorage {
	return &BlocksStorage{
		metricLabel: metricLabel,
		itemLimit:   int(constant.MaxBlocksCacheStorage),
		blocks:      make([]BlockCacheObject, 0, constant.MinRollbackBlocks),
		blocksMapID: make(map[int64]*int, constant.MinRollbackBlocks),
	}
}

func (b *BlocksStorage) Pop() error {
	if len(b.blocks) > 0 {
		b.Lock()
		defer b.Unlock()

		lastBlocksIndex := len(b.blocks) - 1
		delete(b.blocksMapID, b.blocks[lastBlocksIndex].ID)
		b.blocks = b.blocks[:lastBlocksIndex]
		return nil
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(b.metricLabel, float64(b.size()))
	}
	// no more to pop
	return blocker.NewBlocker(blocker.ValidationErr, "StackEmpty")
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
			// remove first (oldest) cache to make room for new block
			delete(b.blocksMapID, b.blocks[0].ID)
			b.blocks = b.blocks[1:]

		}
	}
	b.blocks = append(b.blocks, b.copy(blockCacheObjectCopy))
	b.lastBlockHeight = blockCacheObjectCopy.Height
	newIndexBlock := len(b.blocks) - 1
	b.blocksMapID[blockCacheObjectCopy.ID] = &newIndexBlock
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(b.metricLabel, float64(b.size()))
	}
	return nil
}

// PopTo pop the cache blocks from the provided height to the last height
func (b *BlocksStorage) PopTo(height uint32) error {
	b.Lock()
	defer b.Unlock()
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
	// delete on blocksMapID
	for i := lastIndex; i > heightIndex; i-- {
		delete(b.blocksMapID, b.blocks[i].ID)
	}
	b.blocks = b.blocks[:heightIndex+1]
	b.lastBlockHeight = height
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(b.metricLabel, float64(b.size()))
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
	var blockCacheObjCopy, ok = item.(*BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObject")
	}
	b.RLock()
	defer b.RUnlock()
	if height > b.lastBlockHeight {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	var (
		lastIndex   = len(b.blocks) - 1
		heightIndex = lastIndex - int(b.lastBlockHeight-height)
	)
	if heightIndex < 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	*blockCacheObjCopy = b.copy(b.blocks[heightIndex])
	return nil
}

func (b *BlocksStorage) GetTop(item interface{}) error {
	b.RLock()
	defer b.RUnlock()
	topIndex := len(b.blocks) - 1
	if topIndex < 0 {
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
	b.RLock()
	defer b.RUnlock()
	b.blocks = make([]BlockCacheObject, 0, b.itemLimit)
	b.lastBlockHeight = 0
	b.blocksMapID = make(map[int64]*int, b.itemLimit)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(b.metricLabel, 0)
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
	// copy array type to remove reference
	var blockHash = make([]byte, len(blockCacheObject.BlockHash))
	copy(blockHash, blockCacheObject.BlockHash)

	blockCacheObjectCopy = BlockCacheObject{
		ID:        blockCacheObject.ID,
		Height:    blockCacheObject.Height,
		Timestamp: blockCacheObject.Timestamp,
		BlockHash: blockHash,
	}
	return blockCacheObjectCopy
}

// CacheStorageInterface implementation

// SetItem not implementaed, set intem already implement in push CacheStackStorageInterface
func (b *BlocksStorage) SetItem(key, item interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotImplemented")
}

// SetItem not implementaed, set intem already implement in push CacheStackStorageInterface
func (b *BlocksStorage) SetItems(item interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotImplemented")
}

// GetItem take variable and assign implementation stored item to it
func (b *BlocksStorage) GetItem(key, item interface{}) error {
	b.RLock()
	defer b.RUnlock()
	blockID, ok := key.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotInt64")
	}
	blockCacheObjCopy, ok := item.(*BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotBlockCacheObject")
	}
	index := b.blocksMapID[blockID]
	if index == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemNotFound")
	}
	*blockCacheObjCopy = b.copy(b.blocks[*index])
	if len(blockCacheObjCopy.BlockHash) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemNotFound")
	}
	return nil
}

// GetAllItems fetch all cached items
func (b *BlocksStorage) GetAllItems(item interface{}) error {
	return b.GetAll(item)
}

// GetTotalItems fetch the number of total cached items
func (b *BlocksStorage) GetTotalItems() int {
	return len(b.blocks)
}

// RemoveItem not implementaed, set intem already implement in Pop CacheStackStorageInterface
func (b *BlocksStorage) RemoveItem(key interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotImplemented")
}

// GetSize return the size of storage in number of `byte`
func (b *BlocksStorage) GetSize() int64 {
	return int64(b.size())
}

// ClearCache empty the storage item
func (b *BlocksStorage) ClearCache() error {
	return b.Clear()
}
