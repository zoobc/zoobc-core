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
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
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
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType lastChange")
}

func (n *NodeShardCacheStorage) GetAllItems(item interface{}) error {
	return nil
}

func (n *NodeShardCacheStorage) GetTotalItems() int {
	n.RLock()
	defer n.RUnlock()
	var totalItems int
	for _, IDs := range n.shardMap.NodeShards {
		totalItems += len(IDs)
	}
	return totalItems
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

func (n *NodeShardCacheStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (n *NodeShardCacheStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (n *NodeShardCacheStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
