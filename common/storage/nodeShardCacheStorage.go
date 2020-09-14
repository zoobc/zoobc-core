package storage

import (
	"bytes"
	"sync"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/blocker"
)

type (
	NodeShardCacheStorage struct {
		sync.RWMutex
		// representation of sorted chunk_hashes hashed
		lastChange [32]byte
		shardMap   ShardMap
	}
	ShardMap struct {
		NodeShards  map[int64][]uint64
		ShardChunks map[uint64][][]byte
	}
)

func NewNodeShardCacheStorage() *NodeShardCacheStorage {
	return &NodeShardCacheStorage{
		shardMap: ShardMap{
			NodeShards:  make(map[int64][]uint64),
			ShardChunks: make(map[uint64][][]byte),
		},
	}
}

// SetItem setter of NodeShardCacheStorage
func (n *NodeShardCacheStorage) SetItem(lastChange, item interface{}) error {
	n.Lock()
	defer n.Unlock()

	if last, ok := lastChange.([32]byte); ok {
		n.lastChange = last
	} else {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType lastChange")
	}
	if shardMap, ok := item.(ShardMap); ok {
		n.shardMap.NodeShards = shardMap.NodeShards
		n.shardMap.ShardChunks = shardMap.ShardChunks
	} else {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	return nil
}

func (n *NodeShardCacheStorage) SetItems(_ interface{}) error {
	return nil
}

// GetItem getter of NodShardCacheStorage
func (n *NodeShardCacheStorage) GetItem(lastChange, item interface{}) error {
	n.RLock()
	defer n.RUnlock()

	var (
		shardMapCopy *ShardMap
	)
	if last, ok := lastChange.([32]byte); ok {
		if bytes.Equal(last[:], n.lastChange[:]) {
			shardMapCopy, ok = item.(*ShardMap)
			if !ok {
				return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
			}
			for i, uint64s := range n.shardMap.NodeShards {
				copy(shardMapCopy.NodeShards[i], uint64s)
			}
			for u, i := range n.shardMap.ShardChunks {
				copy(shardMapCopy.ShardChunks[u], i)
			}
		}
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType lastChange")
}

func (n *NodeShardCacheStorage) GetAllItems(item interface{}) error {
	return nil
}

func (n *NodeShardCacheStorage) RemoveItem(key interface{}) error {
	return nil
}

func (n *NodeShardCacheStorage) GetSize() int64 {
	var result int64
	for _, uint64s := range n.shardMap.NodeShards {
		result += 8
		result += int64(len(uint64s)) * 8
	}
	for _, i := range n.shardMap.ShardChunks {
		result += 8
		result += int64(len(i) * sha3.New256().Size())
	}
	return result
}

func (n *NodeShardCacheStorage) ClearCache() error {
	n.shardMap = ShardMap{
		NodeShards:  make(map[int64][]uint64),
		ShardChunks: make(map[uint64][][]byte),
	}
	return nil
}
