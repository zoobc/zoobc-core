package storage

import (
	"bytes"
	"encoding/gob"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"sync"
)

type (
	// ReceiptBatchStorage implements stack storage limiting to 40 blocks worth of receipt batch, each batch will be
	// representing a set of receipts collected and finalized at certain height
	ReceiptBatchStorage struct {
		sync.RWMutex
		itemLimit int
		batches   []ReceiptBatchObject
	}

	// ReceiptBatchObject represent receipt batch finalized in a block
	// ReceiptBatch field's content order must be consensus where the first item will always be `previousBlock` receipts
	// even if there is no receipt present [same rule apply for other transaction data included]
	ReceiptBatchObject struct {
		BlockHeight  uint32
		BlockHash    []byte
		MerkleRoot   []byte
		ReceiptBatch [][]model.Receipt
	}
)

func NewReceiptBatchStackStorage() *ReceiptBatchStorage {
	// store 40 block worth of
	return &ReceiptBatchStorage{
		itemLimit: int(constant.MaxReceiptBatchCacheRound),
		batches:   make([]ReceiptBatchObject, 0, constant.MaxReceiptBatchCacheRound),
	}
}

func (r *ReceiptBatchStorage) Pop() error {
	if len(r.batches) > 0 {
		r.Lock()
		defer r.Unlock()
		r.batches = r.batches[:len(r.batches)-1]
		return nil
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeReceiptBatchStorage, float64(r.size()))
	}
	// no more to pop
	return blocker.NewBlocker(blocker.CacheEmpty, "ReceiptBatchStorage:Empty")
}

func (r *ReceiptBatchStorage) Push(item interface{}) error {
	batchCopy, ok := item.(ReceiptBatchObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNot:storage.ReceiptBatchObject")
	}
	r.Lock()
	defer r.Unlock()
	if len(r.batches) >= r.itemLimit {
		if len(r.batches) != 0 {
			r.batches = r.batches[1:] // remove first (oldest) cache to make room for new batches
		}
	}
	r.batches = append(r.batches, batchCopy)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeReceiptBatchStorage, float64(r.size()))
	}
	return nil
}

func (r *ReceiptBatchStorage) PopTo(index uint32) error {
	if int(index)+1 > len(r.batches) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	r.Lock()
	defer r.Unlock()
	r.batches = r.batches[:index+1]
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeReceiptBatchStorage, float64(r.size()))
	}
	return nil
}

func (r *ReceiptBatchStorage) GetAll(items interface{}) error {
	batchesCopy, ok := items.(*[]ReceiptBatchObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNot:storage.ReceiptBatchObject")
	}
	r.RLock()
	defer r.RUnlock()
	*batchesCopy = r.batches
	return nil
}

func (r *ReceiptBatchStorage) GetAtIndex(index uint32, item interface{}) error {
	if int(index) >= len(r.batches) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	batchCopy, ok := item.(*ReceiptBatchObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNot:storage.ReceiptBatchObject")
	}
	r.RLock()
	defer r.RUnlock()
	*batchCopy = r.batches[int(index)]
	return nil
}

func (r *ReceiptBatchStorage) GetTop(item interface{}) error {
	r.RLock()
	defer r.RUnlock()
	topIndex := len(r.batches)
	if topIndex == 0 {
		return blocker.NewBlocker(blocker.CacheEmpty, "ReceiptBatchStorage:Empty")
	}
	batchCopy, ok := item.(*ReceiptBatchObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNot:storage.ReceiptBatchObject")
	}
	*batchCopy = r.batches[topIndex-1]
	return nil
}

func (r *ReceiptBatchStorage) Clear() error {
	r.batches = make([]ReceiptBatchObject, 0, r.itemLimit)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeReceiptBatchStorage, 0)
	}
	return nil
}

func (r *ReceiptBatchStorage) size() int {
	var size int
	var (
		batchesBytes bytes.Buffer
		enc          = gob.NewEncoder(&batchesBytes)
	)
	_ = enc.Encode(r.batches)
	size = batchesBytes.Len()
	return size
}
