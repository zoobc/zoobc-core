package storage

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"sync"
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
	} else {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
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
	return nil
}

func (m *MempoolCacheStorage) GetSize() int64 {
	var size int64
	for _, memObj := range m.mempoolMap {
		size += 8 * 3 // key + feePerByte + arrivalTimestamp + blockHeight
		size += int64(memObj.TransactionByteSize)
	}
	return size
}

func (m *MempoolCacheStorage) ClearCache() error {
	m.mempoolMap = make(MempoolMap)
	return nil
}
