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
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	// MempoolCacheStorage cache layer for mempool transaction
	MempoolCacheStorage struct {
		sync.RWMutex
		mempoolMap MempoolMap
	}
	MempoolCacheObject struct {
		Tx                  model.Transaction
		ArrivalTimestamp    int64
		FeePerByte          int64
		TransactionByteSize uint32
		BlockHeight         uint32
	}
	MempoolMap map[int64]MempoolCacheObject
)

func NewMempoolStorage() *MempoolCacheStorage {
	return &MempoolCacheStorage{
		mempoolMap: make(MempoolMap),
	}
}

func (m *MempoolCacheStorage) SetItem(key, item interface{}) error {
	m.Lock()
	defer m.Unlock()

	if mempoolMap, ok := item.(MempoolCacheObject); ok {
		keyInt64, ok := key.(int64)
		if !ok {
			return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
		}
		m.mempoolMap[keyInt64] = mempoolMap
		if monitoring.IsMonitoringActive() {
			monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolCacheStorage, float64(m.size()))
			monitoring.SetMempoolTransactionCount(len(m.mempoolMap))
		}
	} else {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	return nil
}

func (m *MempoolCacheStorage) SetItems(_ interface{}) error {
	return nil
}
func (m *MempoolCacheStorage) GetItem(key, item interface{}) error {
	m.RLock()
	defer m.RUnlock()

	if keyInt64, ok := key.(int64); ok {
		txCopy, ok := item.(*MempoolCacheObject)
		if !ok {
			return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
		}
		*txCopy = m.mempoolMap[keyInt64]
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType Key")
}

func (m *MempoolCacheStorage) GetAllItems(item interface{}) error {
	m.RLock()
	defer m.RUnlock()
	if item == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemCannotBeNil")
	}
	itemCopy, ok := item.(MempoolMap)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeItem")
	}
	for k, tx := range m.mempoolMap {
		itemCopy[k] = tx
	}
	return nil
}

func (m *MempoolCacheStorage) GetTotalItems() int {
	m.RLock()
	var totalItems = len(m.mempoolMap)
	m.RUnlock()
	return totalItems
}

func (m *MempoolCacheStorage) RemoveItem(keys interface{}) error {
	m.Lock()
	defer m.Unlock()
	ids, ok := keys.([]int64)
	if !ok {
		id, ok := keys.(int64)
		if !ok {
			return blocker.NewBlocker(blocker.ValidationErr, "WrongType Key")
		}
		delete(m.mempoolMap, id)
		return nil
	}
	for _, id := range ids {
		delete(m.mempoolMap, id)
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolCacheStorage, float64(m.size()))
		monitoring.SetMempoolTransactionCount(len(m.mempoolMap))
	}
	return nil
}

func (m *MempoolCacheStorage) size() int {
	var size int
	for _, memObj := range m.mempoolMap {
		size += 8 * 3
		size += int(memObj.TransactionByteSize)
	}
	return size
}

func (m *MempoolCacheStorage) GetSize() int64 {
	m.RLock()
	defer m.RUnlock()

	return int64(m.size())
}

func (m *MempoolCacheStorage) ClearCache() error {
	m.mempoolMap = make(MempoolMap)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolCacheStorage, 0)
	}
	return nil
}

func (m *MempoolCacheStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (m *MempoolCacheStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (m *MempoolCacheStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
