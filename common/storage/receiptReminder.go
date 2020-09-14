package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
)

type (
	ReceiptReminderStorage struct {
		sync.RWMutex
		// reminders map[receipt_key]
		reminders map[string]chaintype.ChainType
	}
)

func NewReceiptReminderStorage() *ReceiptReminderStorage {
	return &ReceiptReminderStorage{
		reminders: make(map[string]chaintype.ChainType),
	}
}

// SetItem add new item into storage
func (rs *ReceiptReminderStorage) SetItem(key, item interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	var (
		reminder string
		nItem    chaintype.ChainType
		ok       bool
	)
	if reminder, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}

	if nItem, ok = item.(chaintype.ChainType); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")

	}

	if rs.GetSize() >= constant.PriorityStrategyMaxPriorityPeers*int64(constant.MinRollbackBlocks) {
		if err := rs.ClearCache(); err != nil {
			return err
		}
	}
	rs.reminders[reminder] = nItem
	return nil
}
func (rs *ReceiptReminderStorage) SetItems(_ interface{}) error {
	return nil
}
func (rs *ReceiptReminderStorage) GetItem(key, item interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	var (
		reminder string
		nItem    *chaintype.ChainType
		ok       bool
	)

	if reminder, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if nItem, ok = item.(*chaintype.ChainType); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	*nItem = rs.reminders[reminder]
	return nil
}

func (rs *ReceiptReminderStorage) GetAllItems(key interface{}) error {
	if k, ok := key.(*map[string]chaintype.ChainType); ok {
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
	return int64(len(rs.reminders))
}

func (rs *ReceiptReminderStorage) ClearCache() error {
	rs.reminders = make(map[string]chaintype.ChainType)
	return nil
}
