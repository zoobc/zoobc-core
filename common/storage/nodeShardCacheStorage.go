package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
)

type (
	NodeShardCacheStorage struct {
		sync.RWMutex
		nodeShards map[int64][]uint64
	}
)

func NewNodeShardCacheStorage() *NodeShardCacheStorage {
	return &NodeShardCacheStorage{
		nodeShards: make(map[int64][]uint64),
	}
}

func (n *NodeShardCacheStorage) SetItem(item interface{}) error {
	var (
		ok bool
	)
	n.Lock()
	defer n.Unlock()
	n.nodeShards, ok = item.(map[int64][]uint64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType")
	}
	return nil
}

func (n *NodeShardCacheStorage) GetItem(item interface{}) error {
	n.RLock()
	var (
		ok      bool
		mapCopy map[int64][]uint64
	)
	mapCopy, ok = item.(map[int64][]uint64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType")
	}
	for i, uint64s := range n.nodeShards {
		mapCopy[i] = uint64s
	}
	n.RUnlock()
	return nil
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
