package storage

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	ReceiptPool struct {
		sync.Mutex
		receipts map[string][]model.Receipt
	}
)

func NewReceiptPool() *ReceiptPool {
	return &ReceiptPool{
		receipts: make(map[string][]model.Receipt),
	}
}

func (r *ReceiptPool) SetItem(key, item interface{}) error {
	r.Lock()
	defer r.Unlock()

	var (
		ok    bool
		nItem model.Receipt
		nKey  string
	)
	if nItem, ok = item.(model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	r.receipts[nKey] = append(r.receipts[nKey], nItem)
	return nil
}

func (r *ReceiptPool) SetItems(items interface{}) error {
	r.Lock()
	defer r.Unlock()

	var (
		nItems map[string][]model.Receipt
		ok     bool
	)

	nItems, ok = items.(map[string][]model.Receipt)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	r.receipts = nItems
	return nil
}

func (r *ReceiptPool) GetItem(key, items interface{}) error {
	r.Lock()
	defer r.Unlock()

	var (
		nKey   string
		nItems *[]model.Receipt
		ok     bool
	)

	nKey, ok = key.(string)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}

	nItems, ok = items.(*[]model.Receipt)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}

	*nItems = r.receipts[nKey]
	return nil
}

func (r *ReceiptPool) GetAllItems(items interface{}) error {
	r.Lock()
	defer r.Unlock()

	var (
		nItems *map[string][]model.Receipt
		ok     bool
	)

	if nItems, ok = items.(*map[string][]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}
	*nItems = r.receipts
	return nil
}

func (r *ReceiptPool) RemoveItem(key interface{}) error {
	r.Lock()
	defer r.Unlock()

	var (
		nKey string
		ok   bool
	)
	if nKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}

	delete(r.receipts, nKey)
	return nil
}

func (r *ReceiptPool) GetSize() int64 {
	panic("implement me")
}

func (r *ReceiptPool) ClearCache() error {
	r.Lock()
	defer r.Unlock()

	r.receipts = make(map[string][]model.Receipt)
	return nil
}

func (r *ReceiptPool) size() int64 {
	panic("implement me")
}
