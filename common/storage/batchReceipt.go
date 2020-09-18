package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BatchReceiptCacheStorage struct {
		sync.RWMutex
		receipts []model.BatchReceipt
	}
)

func NewBatchReceiptCacheStorage() *BatchReceiptCacheStorage {
	return &BatchReceiptCacheStorage{
		receipts: make([]model.BatchReceipt, 0),
	}
}

// SetItem set new value to BatchReceiptCacheStorage
//      - key: nil
//      - item: BatchReceiptCache
func (brs *BatchReceiptCacheStorage) SetItem(_, item interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		ok    bool
		nItem model.BatchReceipt
	)

	if nItem, ok = item.(model.BatchReceipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid item")
	}

	brs.receipts = append(brs.receipts, nItem)
	return nil
}

// SetItems store and replace the old items.
//      - items: []model.BatchReceipt
func (brs *BatchReceiptCacheStorage) SetItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItems []model.BatchReceipt
		ok     bool
	)
	nItems, ok = items.([]model.BatchReceipt)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid item")
	}
	brs.receipts = nItems
	return nil
}

// GetItem getting single item of BatchReceiptCacheStorage refill the reference item
//      - key: receiptKey which is a string
//      - item: BatchReceiptCache
func (brs *BatchReceiptCacheStorage) GetItem(key, item interface{}) error {

	return nil
}

// GetAllItems get all items of BatchReceiptCacheStorage
//      - items: *map[string]BatchReceipt
func (brs *BatchReceiptCacheStorage) GetAllItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItem *[]model.BatchReceipt
		ok    bool
	)
	if nItem, ok = items.(*[]model.BatchReceipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid item")
	}
	*nItem = brs.receipts
	return nil
}

func (brs *BatchReceiptCacheStorage) RemoveItem(_ interface{}) error {
	return nil
}

func (brs *BatchReceiptCacheStorage) GetSize() int64 {
	brs.Lock()
	defer brs.Unlock()

	var size int
	for _, cache := range brs.receipts {
		size += cache.XXX_Size()
	}
	return int64(size)
}

func (brs *BatchReceiptCacheStorage) ClearCache() error {
	brs.Lock()
	defer brs.Unlock()

	brs.receipts = make([]model.BatchReceipt, 0)
	return nil
}
