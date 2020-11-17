package storage

import (
	"bytes"
	"encoding/gob"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"sync"
)

type (
	ProvedReceiptReminderStorage struct {
		sync.RWMutex
		// reminders map[receipt_key]
		reminders map[uint32]model.PublishedReceipt
	}
)

func NewProvedReceiptReminderStorage() *ProvedReceiptReminderStorage {
	return &ProvedReceiptReminderStorage{
		reminders: make(map[uint32]model.PublishedReceipt),
	}
}

// SetItem add new item into storage
func (rs *ProvedReceiptReminderStorage) SetItem(key, item interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	var (
		height           uint32
		publishedReceipt model.PublishedReceipt
		ok               bool
	)
	if height, ok = key.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:KeyMustBeUINT32")
	}

	if publishedReceipt, ok = item.(model.PublishedReceipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ValueMustBe:Model.PublishedReceipt")
	}

	if len(rs.reminders) >= constant.PriorityStrategyMaxPriorityPeers*int(constant.MinRollbackBlocks) {
		rs.reminders = make(map[uint32]model.PublishedReceipt)
	}
	rs.reminders[height] = publishedReceipt
	return nil
}

func (rs *ProvedReceiptReminderStorage) SetItems(_ interface{}) error {
	return nil
}
func (rs *ProvedReceiptReminderStorage) GetItem(key, item interface{}) error {
	rs.RLock()
	defer rs.RUnlock()

	var (
		height           uint32
		publishedReceipt *model.PublishedReceipt
		ok               bool
	)
	if height, ok = key.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:KeyMustBeUINT32")
	}

	if publishedReceipt, ok = item.(*model.PublishedReceipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ValueMustBe:Model.PublishedReceipt")
	}

	*publishedReceipt = rs.reminders[height]
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetAllItems(items interface{}) error {
	rs.RLock()
	defer rs.RUnlock()

	if k, ok := items.(*map[uint32]model.PublishedReceipt); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ItemsMustBe(*map[uint32]model.PublishedReceipt)")
}

func (rs *ProvedReceiptReminderStorage) GetTotalItems() int {
	rs.Lock()
	var totalItems = len(rs.reminders)
	rs.Unlock()
	return totalItems
}

func (rs *ProvedReceiptReminderStorage) RemoveItem(key interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	if k, ok := key.(uint32); ok {
		delete(rs.reminders, k)
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:KeyMustBeUINT32")
}

func (rs *ProvedReceiptReminderStorage) GetSize() int64 {
	rs.RLock()
	defer rs.RUnlock()
	var (
		rsBytes bytes.Buffer
		enc     = gob.NewEncoder(&rsBytes)
	)
	_ = enc.Encode(rs.reminders)
	return int64(rsBytes.Len())
}

func (rs *ProvedReceiptReminderStorage) ClearCache() error {
	rs.Lock()
	defer rs.Unlock()

	rs.reminders = make(map[uint32]model.PublishedReceipt)
	return nil
}
