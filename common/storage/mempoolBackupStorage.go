package storage

import (
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
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
	monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))

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
	monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))

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
	fmt.Println("get item")

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

// RemoveItem remove specific item by key
func (m *MempoolBackupStorage) RemoveItem(key interface{}) error {
	m.Lock()
	defer m.Unlock()

	id, ok := key.(int64)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	delete(m.mempools, id)
	monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, float64(m.size()))
	return nil
}

func (m *MempoolBackupStorage) size() int {
	var size int
	for _, v := range m.mempools {
		size += len(v)
	}
	return size
}

// GetSize get size of MempoolBackupStorage values
func (m *MempoolBackupStorage) GetSize() int64 {
	m.Lock()
	defer m.Unlock()

	return int64(m.size())
}

// ClearCache clear or remove all items from MempoolBackupStorage
func (m *MempoolBackupStorage) ClearCache() error {
	m.Lock()
	defer m.Unlock()

	m.mempools = make(map[int64][]byte)
	monitoring.SetCacheStorageMetrics(monitoring.TypeMempoolBackupCacheStorage, 0)
	return nil
}
