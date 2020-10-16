package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	ReceiptPoolCacheStorage struct {
		sync.RWMutex
		receipts []model.Receipt
	}
)

func NewReceiptPoolCacheStorage() *ReceiptPoolCacheStorage {
	return &ReceiptPoolCacheStorage{
		receipts: make([]model.Receipt, 0),
	}
}

// SetItem set new value to ReceiptPoolCacheStorage
//      - key: nil
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) SetItem(_, item interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		ok    bool
		nItem model.Receipt
	)

	if nItem, ok = item.(model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}

	brs.receipts = append(brs.receipts, nItem)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// SetItems store and replace the old items.
//      - items: []model.BatchReceipt
func (brs *ReceiptPoolCacheStorage) SetItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItems []model.Receipt
		ok     bool
	)
	nItems, ok = items.([]model.Receipt)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	brs.receipts = nItems
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// GetItem getting single item of ReceiptPoolCacheStorage refill the reference item
//      - key: receiptKey which is a string
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) GetItem(_, _ interface{}) error {
	return nil
}

// GetAllItems get all items of ReceiptPoolCacheStorage
//      - items: *map[string]BatchReceipt
func (brs *ReceiptPoolCacheStorage) GetAllItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItem *[]model.Receipt
		ok    bool
	)
	if nItem, ok = items.(*[]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	*nItem = brs.receipts
	return nil
}

func (brs *ReceiptPoolCacheStorage) RemoveItem(_ interface{}) error {
	return nil
}

func (brs *ReceiptPoolCacheStorage) size() int64 {
	var size int64
	for _, cache := range brs.receipts {
		var s int
		s += len(cache.GetSenderPublicKey())
		s += len(cache.GetRecipientPublicKey())
		s += 4 // this is cache.GetDatumType()
		s += len(cache.GetDatumHash())
		s += 4 // this is cache.GetReferenceBlockHeight()
		s += len(cache.GetRMRLinked())
		s += len(cache.GetReferenceBlockHash())
		s += len(cache.GetRecipientSignature())
		size += int64(s)
	}
	return size
}
func (brs *ReceiptPoolCacheStorage) GetSize() int64 {
	brs.RLock()
	defer brs.RUnlock()

	return brs.size()
}

func (brs *ReceiptPoolCacheStorage) ClearCache() error {
	brs.Lock()
	defer brs.Unlock()

	brs.receipts = make([]model.Receipt, 0)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, 0)
	}
	return nil
}
