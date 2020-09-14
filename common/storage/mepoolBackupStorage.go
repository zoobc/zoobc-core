package storage

import (
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
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
	return nil
}

// SetItems replace and set bulk items
func (m *MempoolBackupStorage) SetItems(items interface{}) error {
	var (
		nItems map[int64][]byte
		ok     bool
	)
	fmt.Println("Set items mempools")
	nItems, ok = items.(map[int64][]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType items")
	}
	m.mempools = nItems
	fmt.Println("Set items mempools success")
	return nil
}

// GetItem get an item from MempoolBackupStorage by key and refill reference item
func (m *MempoolBackupStorage) GetItem(key, item interface{}) error {
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
	fmt.Println("Get item mempools success")

	return nil
}

// GetAllItems get all from MempoolBackupStorage and refill reference item
func (m *MempoolBackupStorage) GetAllItems(item interface{}) error {
	fmt.Println("Get all items")

	mempoolsBackup, ok := item.(*map[int64][]byte)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	*mempoolsBackup = m.mempools
	fmt.Println("Get all items success")

	return nil
}

// RemoveItem remove specific item by key
func (m *MempoolBackupStorage) RemoveItem(key interface{}) error {
	id, ok := key.(int64)
	fmt.Println("Remove item")

	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	delete(m.mempools, id)
	fmt.Println("Remove item success")
	return nil
}

// GetSize get size of MempoolBackupStorage values
func (m *MempoolBackupStorage) GetSize() int64 {
	var size int
	for _, v := range m.mempools {
		size += len(v)
	}
	return int64(size)
}

// ClearCache clear or remove all items from MempoolBackupStorage
func (m *MempoolBackupStorage) ClearCache() error {
	m.mempools = make(map[int64][]byte)
	fmt.Println("clearCache items mempools success")
	return nil
}
