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
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	ReceiptPoolCacheStorage struct {
		sync.RWMutex
		receipts map[string][]model.Receipt // receipts grouped by their datum hash
	}
)

func NewReceiptPoolCacheStorage() *ReceiptPoolCacheStorage {
	return &ReceiptPoolCacheStorage{
		receipts: make(map[string][]model.Receipt, 0),
	}
}

// SetItem set new value to ReceiptPoolCacheStorage
//      - key: nil
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) SetItem(key, item interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		ok          bool
		nItem       model.Receipt
		receiptHash string
	)

	if receiptHash, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}

	if nItem, ok = item.(model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt item")
	}

	if _, ok = brs.receipts[receiptHash]; !ok {
		brs.receipts[receiptHash] = make([]model.Receipt, 0)
	}

	brs.receipts[receiptHash] = append(brs.receipts[receiptHash], nItem)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, float64(brs.size()))
	}
	return nil
}

// SetItems store and replace the old items. (not implemented)
func (brs *ReceiptPoolCacheStorage) SetItems(items interface{}) error {
	return nil
}

// GetItem getting single item of ReceiptPoolCacheStorage refill the reference item
//      - key: receiptKey which is a string
//      - item: BatchReceiptCache
func (brs *ReceiptPoolCacheStorage) GetItem(_, _ interface{}) error {
	return nil
}

// GetItems getting multiple items of ReceiptPoolCacheStorage refill the reference item
//      - keys: receiptKey which is an array of string
//      - items: BatchReceiptCache (map of array of receipts)
//				 map is by default passed as reference
func (brs *ReceiptPoolCacheStorage) GetItems(keys, items interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		keysParsed []string
		result     map[string][]model.Receipt
		ok         bool
	)
	if keysParsed, ok = keys.([]string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt keys")
	}

	if result, ok = items.(map[string][]model.Receipt); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt items")
	}

	for _, key := range keysParsed {
		matchedData, ok := brs.receipts[key]
		if ok {
			result[key] = matchedData
		}
	}

	return nil
}

// GetAllItems get all items of ReceiptPoolCacheStorage
//      - items: *map[string]BatchReceipt
func (brs *ReceiptPoolCacheStorage) GetAllItems(items interface{}) error {
	return nil
}

func (brs *ReceiptPoolCacheStorage) GetTotalItems() int {
	brs.Lock()
	var total int
	for _, receipts := range brs.receipts {
		total += len(receipts)
	}
	brs.Unlock()
	return total
}

func (brs *ReceiptPoolCacheStorage) RemoveItem(_ interface{}) error {
	return nil
}

func (brs *ReceiptPoolCacheStorage) RemoveItems(keys interface{}) error {
	brs.Lock()
	defer brs.Unlock()

	var (
		keysParsed []string
		ok         bool
	)
	if keysParsed, ok = keys.([]string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "invalid batch receipt keys")
	}

	for _, key := range keysParsed {
		delete(brs.receipts, key)
	}

	return nil
}

func (brs *ReceiptPoolCacheStorage) size() int64 {
	var size int64
	for _, cacheArr := range brs.receipts {
		for _, cache := range cacheArr {
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

	brs.receipts = make(map[string][]model.Receipt, 0)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeBatchReceiptCacheStorage, 0)
	}
	return nil
}

func (brs ReceiptPoolCacheStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args ...interface{}) {
			var (
				b  *model.Block
				ok bool
			)
			b, ok = block.(*model.Block)
			if !ok {
				// brs.Logger.Fatalln("Block casting failures in SendBlockListener")
				return
			}

			brs.Lock()
			defer brs.Unlock()

			for key, receiptGroup := range brs.receipts {
				newReceiptList := []model.Receipt{}
				for _, receipt := range receiptGroup {
					if receipt.ReferenceBlockHeight >= b.GetHeight()-constant.ReceiptPoolMaxLife-constant.MinRollbackBlocks {
						newReceiptList = append(newReceiptList, receipt)
					}
				}

				if len(newReceiptList) == 0 {
					delete(brs.receipts, key)
				} else if len(newReceiptList) != len(receiptGroup) {
					brs.receipts[key] = newReceiptList
				}
			}
		},
	}
}
