package storage

import (
	"bytes"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"golang.org/x/crypto/sha3"
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
		if monitoring.IsMonitoringActive() {
			monitoring.SetCacheStorageMetrics(monitoring.TypeNodeShardCacheStorage, float64(n.size()))
		}
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
		// return blocker.NewBlocker(blocker.NotFound, "item not found")
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

func (n *NodeShardCacheStorage) size() int64 {
	var size int
	for _, uint64s := range n.shardMap.NodeShards {
		var s int
		s += 8
		s += len(uint64s) * 8
		size += s
	}
	for _, i := range n.shardMap.ShardChunks {
		var s int
		s += 8
		s += len(i) * sha3.New256().Size()
	}
	return int64(size)
}

func (n *NodeShardCacheStorage) GetSize() int64 {
	n.RLock()
	defer n.RUnlock()

	return n.size()
}

func (n *NodeShardCacheStorage) ClearCache() error {
	n.shardMap = ShardMap{
		NodeShards:  make(map[int64][]uint64),
		ShardChunks: make(map[uint64][][]byte),
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeNodeShardCacheStorage, 0)
	}

	return nil
}
