package storage

import (
	"bytes"
	"encoding/gob"
	"github.com/zoobc/zoobc-core/common/blocker"
	"math"
	"sync"
)

type (
	ProvedReceiptReminderStorage struct {
		limit int
		sync.RWMutex
		// reminders map[receipt_key]
		reminders map[uint32]ProvedReceiptReminderObject
	}

	ProvedReceiptReminderObject struct {
		MerkleRoot []byte
	}
)

func NewProvedReceiptReminderStorage(limit int) *ProvedReceiptReminderStorage {
	return &ProvedReceiptReminderStorage{
		limit:     limit,
		reminders: make(map[uint32]ProvedReceiptReminderObject),
	}
}

// SetItem add new item into storage
func (rs *ProvedReceiptReminderStorage) SetItem(key, item interface{}) error {
	var (
		height        uint32
		provedReceipt ProvedReceiptReminderObject
		ok            bool
	)
	if height, ok = key.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:KeyMustBeUINT32")
	}

	if provedReceipt, ok = item.(ProvedReceiptReminderObject); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ValueMustBe:PublishedReceipt")
	}
	rs.Lock()
	defer rs.Unlock()
	if len(rs.reminders) >= rs.limit {
		if len(rs.reminders) != 0 {
			var minHeight uint32 = math.MaxUint32
			for height := range rs.reminders {
				if height < minHeight {
					minHeight = height
				}
			}
			delete(rs.reminders, minHeight)
		}
	}
	rs.reminders[height] = provedReceipt
	return nil
}

// SetItems is not needed in proved receipt reminder
func (rs *ProvedReceiptReminderStorage) SetItems(_ interface{}) error {
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetItem(key, item interface{}) error {
	rs.RLock()
	defer rs.RUnlock()

	var (
		height        uint32
		provedReceipt *ProvedReceiptReminderObject
		ok            bool
	)
	if height, ok = key.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:KeyMustBeUINT32")
	}

	if provedReceipt, ok = item.(*ProvedReceiptReminderObject); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ValueMustBe:*storage.ProvedReceiptReminderObject")
	}
	*provedReceipt = rs.reminders[height]
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetAllItems(items interface{}) error {
	rs.RLock()
	defer rs.RUnlock()

	if k, ok := items.(*map[uint32]ProvedReceiptReminderObject); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ItemsMustBe(*map[uint32]ProvedReceiptReminderObject)")
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

	rs.reminders = make(map[uint32]ProvedReceiptReminderObject)
	return nil
}
