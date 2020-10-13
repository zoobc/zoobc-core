package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
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
