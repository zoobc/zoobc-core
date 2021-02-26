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
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	// MempoolBackupStorage cache storage for backup transactions that want to rollback
	MempoolBackupStorage struct {
		sync.RWMutex
		// mempools map[ID]mempool_byte
		mempools map[int64][]byte
	}
)

// NewMempoolBackupStorage create new instance of MempoolBackupStorage
func NewMempoolBackupStorage() *MempoolBackupStorage {
	return &MempoolBackupStorage{
		mempools: make(map[int64][]byte),
	}
}

// SetItem add new item on mempoolBackup
func (m *MempoolBackupStorage) SetItem(key, item interface{}) error {
	m.Lock()
	defer m.Unlock()

	var (
		id          int64
		mempoolByte []byte
		ok          bool
	)

	if id, ok = key.(int64); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if mempoolByte, ok = item.([]byte); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}

	m.mempools[id] = mempoolByte
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))
	}
	return nil
}

// SetItems replace and set bulk items
func (m *MempoolBackupStorage) SetItems(items interface{}) error {
	m.Lock()
	defer m.Unlock()

	var (
		nItems map[int64][]byte
		ok     bool
	)
	nItems, ok = items.(map[int64][]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType items")
	}
	m.mempools = nItems
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))
	}
	return nil
}

// GetItem get an item from MempoolBackupStorage by key and refill reference item
func (m *MempoolBackupStorage) GetItem(key, item interface{}) error {
	m.Lock()
	defer m.Unlock()

	var (
		id          int64
		mempoolByte *[]byte
		ok          bool
	)

	if id, ok = key.(int64); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if mempoolByte, ok = item.(*[]byte); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}

	*mempoolByte = m.mempools[id]

	return nil
}

// GetAllItems get all from MempoolBackupStorage and refill reference item
func (m *MempoolBackupStorage) GetAllItems(item interface{}) error {

	m.Lock()
	defer m.Unlock()

	mempoolsBackup, ok := item.(*map[int64][]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	*mempoolsBackup = m.mempools

	return nil
}

func (m *MempoolBackupStorage) GetTotalItems() int {
	m.RLock()
	var totalItems = len(m.mempools)
	m.RUnlock()
	return totalItems
}

// RemoveItem remove specific item by key
func (m *MempoolBackupStorage) RemoveItem(key interface{}) error {
	m.Lock()
	defer m.Unlock()

	id, ok := key.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	delete(m.mempools, id)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))
	}
	return nil
}

func (m *MempoolBackupStorage) size() int64 {
	var size int
	for _, v := range m.mempools {
		s := len(v)
		size += s
	}
	return int64(size)
}

// GetSize get size of MempoolBackupStorage values
func (m *MempoolBackupStorage) GetSize() int64 {
	m.RLock()
	defer m.RUnlock()

	return m.size()
}

// ClearCache clear or remove all items from MempoolBackupStorage
func (m *MempoolBackupStorage) ClearCache() error {
	m.Lock()
	defer m.Unlock()

	m.mempools = make(map[int64][]byte)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, 0)
	}
	return nil
}

func (m *MempoolBackupStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (m *MempoolBackupStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (m *MempoolBackupStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
