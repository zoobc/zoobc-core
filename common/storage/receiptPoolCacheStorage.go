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
//      - item: model.Receipt
func (brs *ReceiptPoolCacheStorage) SetItem(key, item interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		ok    bool
		nItem model.Receipt
		nKey  string
	)

	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid key type")
	}

	if nItem, ok = item.(model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid receipt item")
	}

	brs.receipts[nKey] = append(brs.receipts[nKey], nItem)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// SetItems store and replace the old items.
//      - items: []model.Receipt
func (brs *ReceiptPoolCacheStorage) SetItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItems map[string][]model.Receipt
		ok     bool
	)
	nItems, ok = items.(map[string][]model.Receipt)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid receipt item")
	}
	brs.receipts = nItems
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// GetItem getting single item of ReceiptPoolCacheStorage refill the reference item
//      - key: receiptKey which is a string
//      - item: *[]model.Receipt
func (brs *ReceiptPoolCacheStorage) GetItem(key, items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItems *[]model.Receipt
		nKey   string
		ok     bool
	)

	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid key type")
	}

	if nItems, ok = items.(*[]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid receipt item")
	}

	*nItems = brs.receipts[nKey]
	return nil
}

// GetAllItems get all items of ReceiptPoolCacheStorage
//      - items: *map[string][]model.Receipt
func (brs *ReceiptPoolCacheStorage) GetAllItems(items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		nItem *map[string][]model.Receipt
		ok    bool
	)
	if nItem, ok = items.(*map[string][]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid receipt item")
	}
	*nItem = brs.receipts
	return nil
}

func (brs *ReceiptPoolCacheStorage) RemoveItem(key interface{}) error {

	brs.Lock()
	defer brs.Unlock()

	var (
		nKey string
		ok   bool
	)

	nKey, ok = key.(string)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid key type")
	}

	delete(brs.receipts, nKey)
	return nil
}

func (brs *ReceiptPoolCacheStorage) size() int64 {
	var size int64
	for _, receipts := range brs.receipts {
		for _, receipt := range receipts {
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
