// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
