package storage

import (
	"bytes"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
)

type (
	NodeShardCacheStorage struct {
		sync.RWMutex
		// representation of sorted NodeIDs hashed
		lastChange [32]byte
		nodeShards map[int64][]uint64
	}
)

func NewNodeShardCacheStorage() *NodeShardCacheStorage {
	return &NodeShardCacheStorage{
		nodeShards: make(map[int64][]uint64),
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
	if nodeShards, ok := item.(map[int64][]uint64); ok {
		n.nodeShards = nodeShards
	} else {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")

	}
	return nil
}

// GetItem getter of NodShardCacheStorage
func (n *NodeShardCacheStorage) GetItem(lastChange, item interface{}) error {
	n.RLock()
	defer n.RUnlock()

	var (
		mapCopy map[int64][]uint64
	)
	if last, ok := lastChange.([32]byte); ok {
		if bytes.Equal(last[:], n.lastChange[:]) {
			mapCopy, ok = item.(map[int64][]uint64)
			if !ok {
				return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
			}
			for i, uint64s := range n.nodeShards {
				mapCopy[i] = uint64s
			}
		}
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType lastChange")
}

func (n *NodeShardCacheStorage) GetSize() int64 {
	var result int64
	for _, uint64s := range n.nodeShards {
		result += 8
		result += int64(len(uint64s)) * 8
	}
	return result
}

func (n *NodeShardCacheStorage) ClearCache() error {
	n.nodeShards = make(map[int64][]uint64)
	return nil
}
