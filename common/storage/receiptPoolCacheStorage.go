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
		receipts map[string][]model.Receipt
	}
)

func NewReceiptPoolCacheStorage() *ReceiptPoolCacheStorage {
	return &ReceiptPoolCacheStorage{
		receipts: make(map[string][]model.Receipt),
	}
}

// SetItem set new value to ReceiptPoolCacheStorage
//      - key: nil
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) SetItem(key, item interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		ok    bool
		nItem model.Receipt
		nKey  string
	)
	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptPoolCacheStorage:InvalidKeyType:Expect-string")
	}
	if nItem, ok = item.(model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptPoolCacheStorage:InvalidItemType:Expect-model.Receipt")
	}

	brs.receipts[nKey] = append(brs.receipts[nKey], nItem)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// SetItems not used in receipt pool storage
func (brs *ReceiptPoolCacheStorage) SetItems(items interface{}) error {
	return nil
}

// GetItem getting single item of ReceiptPoolCacheStorage refill the reference item
//      - key: receiptKey which is a string
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) GetItem(key, item interface{}) error {
	var (
		ok    bool
		nKey  string
		nItem *[]model.Receipt
	)
	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptPoolCacheStorage:InvalidKeyType:Expect-string")
	}
	if nItem, ok = item.(*[]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptPoolCacheStorage:InvalidItemType:Expect-*model.Receipt")
	}
	*nItem = brs.receipts[nKey]
	return nil
}

// GetAllItems not used in receipt pool storage
func (brs *ReceiptPoolCacheStorage) GetAllItems(items interface{}) error {
	return nil
}

func (brs *ReceiptPoolCacheStorage) GetTotalItems() int {
	brs.Lock()
	var totalItems = len(brs.receipts)
	brs.Unlock()
	return totalItems
}

func (brs *ReceiptPoolCacheStorage) RemoveItem(_ interface{}) error {
	return nil
}

func (brs *ReceiptPoolCacheStorage) size() int64 {
	var size int64
	for _, cache := range brs.receipts {
		for _, receipt := range cache {
			var s int
			s += len(receipt.GetSenderPublicKey())
			s += len(receipt.GetRecipientPublicKey())
			s += 4 // this is cache.GetDatumType()
			s += len(receipt.GetDatumHash())
			s += 4 // this is cache.GetReferenceBlockHeight()
			s += len(receipt.GetRMRLinked())
			s += len(receipt.GetReferenceBlockHash())
			s += len(receipt.GetRecipientSignature())
			size += int64(s)
		}
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

	brs.receipts = make(map[string][]model.Receipt)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, 0)
	}
	return nil
}
