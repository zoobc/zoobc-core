package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
)

type (
	ProvedReceiptReminderStorage struct {
		sync.RWMutex
		// reminders map[receipt_key]
		reminders map[string]chaintype.ChainType
	}
)

func NewProvedReceiptReminderStorage() *ProvedReceiptReminderStorage {
	return &ProvedReceiptReminderStorage{
		reminders: make(map[string]chaintype.ChainType),
	}
}

// SetItem add new item into storage
func (rs *ProvedReceiptReminderStorage) SetItem(key, item interface{}) error {
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

	if len(rs.reminders) >= constant.PriorityStrategyMaxPriorityPeers*int(constant.MinRollbackBlocks) {
		rs.reminders = make(map[string]chaintype.ChainType)
	}
	rs.reminders[reminder] = nItem
	return nil
}

func (rs *ProvedReceiptReminderStorage) SetItems(_ interface{}) error {
	return nil
}
func (rs *ProvedReceiptReminderStorage) GetItem(key, item interface{}) error {
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

func (rs *ProvedReceiptReminderStorage) GetAllItems(key interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	if k, ok := key.(*map[string]chaintype.ChainType); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}
func (rs *ProvedReceiptReminderStorage) RemoveItem(key interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	if k, ok := key.(string); ok {
		delete(rs.reminders, k)
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}
func (rs *ProvedReceiptReminderStorage) GetSize() int64 {
	rs.Lock()
	defer rs.Unlock()

	var size int
	for k, v := range rs.reminders {
		size += len(k) + int(v.GetTypeInt())
	}
	return int64(size)
}

func (rs *ProvedReceiptReminderStorage) ClearCache() error {
	rs.Lock()
	defer rs.Unlock()

	rs.reminders = make(map[string]chaintype.ChainType)
	return nil
}
