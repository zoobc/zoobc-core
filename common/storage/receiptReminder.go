package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
)

type (
	ReceiptReminderStorage struct {
		sync.RWMutex
		// reminders map[nodeKeys]datumHash
		reminders map[string][]byte
	}
)

func NewReceiptReminderStorage() *ReceiptReminderStorage {
	return &ReceiptReminderStorage{
		reminders: make(map[string][]byte),
	}
}

// SetItem add new item into storage
func (rs *ReceiptReminderStorage) SetItem(key, item interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	var (
		reminder string
		nItem    []byte
		ok       bool
	)
	if reminder, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}

	if nItem, ok = item.([]byte); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")

	}
	rs.reminders[reminder] = append(rs.reminders[reminder], nItem...)
	return nil
}

func (rs *ReceiptReminderStorage) GetItem(key, item interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	var (
		reminder string
		nItem    *[]byte
		ok       bool
	)

	if reminder, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if nItem, ok = item.(*[]byte); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	*nItem = rs.reminders[reminder]
	return nil
}

func (rs *ReceiptReminderStorage) GetAllItems(key interface{}) error {
	if k, ok := key.(*map[string][]byte); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}
func (rs *ReceiptReminderStorage) RemoveItem(key interface{}) error {
	if k, ok := key.(string); ok {
		delete(rs.reminders, k)
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}
func (rs *ReceiptReminderStorage) GetSize() int64 {
	var size int
	for _, v := range rs.reminders {
		size += len(v)
	}
	return int64(size)
}

func (rs *ReceiptReminderStorage) ClearCache() error {
	rs.reminders = nil
	return nil
}
