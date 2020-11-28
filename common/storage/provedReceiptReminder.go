package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"sync"
)

type (
	ProvedReceiptReminderStorage struct {
		limit int
		sync.RWMutex
		// reminders map[receipt_key]
		reminders []ProvedReceiptReminderObject
	}

	ProvedReceiptReminderObject struct {
		ReferenceBlockHeight uint32
		ReferenceBlockHash   []byte
		MerkleRoot           []byte
	}
)

func NewProvedReceiptReminderStorage(limit int) *ProvedReceiptReminderStorage {
	return &ProvedReceiptReminderStorage{
		limit:     limit,
		reminders: make([]ProvedReceiptReminderObject, 0),
	}
}

func (rs *ProvedReceiptReminderStorage) Pop() error {
	return nil
}

func (rs *ProvedReceiptReminderStorage) PopTo(index uint32) error {
	return nil
}

// Push add new item into storage
func (rs *ProvedReceiptReminderStorage) Push(item interface{}) error {
	var (
		provedReceipt ProvedReceiptReminderObject
		ok            bool
	)

	if provedReceipt, ok = item.(ProvedReceiptReminderObject); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ProvedReceiptReminder:ValueMustBe:PublishedReceipt")
	}
	rs.Lock()
	defer rs.Unlock()
	if len(rs.reminders) >= rs.limit {
		if len(rs.reminders) != 0 {
			rs.reminders = rs.reminders[1:] // remove first (oldest) cache to make room for new batches
		}
	}
	rs.reminders = append(rs.reminders, provedReceipt)
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetTop(item interface{}) error {
	rs.RLock()
	defer rs.RUnlock()
	topIndex := len(rs.reminders)
	if topIndex == 0 {
		return blocker.NewBlocker(blocker.CacheEmpty, "ProvedReceiptReminderStorage:Empty")
	}
	reminderCopy, ok := item.(*ProvedReceiptReminderObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNot:storage.ProvedReceiptReminderObject")
	}
	*reminderCopy = rs.reminders[topIndex-1]
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetAtIndex(index uint32, item interface{}) error {
	rs.RLock()
	defer rs.RUnlock()
	if int(index) >= len(rs.reminders) {
		return blocker.NewBlocker(
			blocker.NotFound,
			fmt.Sprintf(
				"ProvedReceiptReminderStorage:GetAtIndex:IndexOutOfRange:have= %d - requested= %d",
				len(rs.reminders),
				index,
			),
		)
	}
	var (
		provedReceipt *ProvedReceiptReminderObject
		ok            bool
	)

	if provedReceipt, ok = item.(*ProvedReceiptReminderObject); !ok {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"ProvedReceiptReminder:GetAtIndex:ValueMustBe:*storage.ProvedReceiptReminderObject",
		)
	}
	*provedReceipt = rs.reminders[index]
	return nil
}

func (rs *ProvedReceiptReminderStorage) GetAll(items interface{}) error {
	rs.RLock()
	defer rs.RUnlock()

	if k, ok := items.(*[]ProvedReceiptReminderObject); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr,
		"ProvedReceiptReminder:GetAll:ItemsMustBe(*[]ProvedReceiptReminderObject)",
	)
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

func (rs *ProvedReceiptReminderStorage) Clear() error {
	rs.Lock()
	defer rs.Unlock()

	rs.reminders = make([]ProvedReceiptReminderObject, 0)
	return nil
}
