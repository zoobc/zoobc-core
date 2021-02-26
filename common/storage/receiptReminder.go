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
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/observer"
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

	if len(rs.reminders) >= constant.PriorityStrategyMaxPriorityPeers*int(constant.MinRollbackBlocks) {
		rs.reminders = make(map[string]chaintype.ChainType)
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
	rs.Lock()
	defer rs.Unlock()

	if k, ok := key.(*map[string]chaintype.ChainType); ok {
		*k = rs.reminders
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}

func (rs *ReceiptReminderStorage) GetTotalItems() int {
	rs.Lock()
	var totalItems = len(rs.reminders)
	rs.Unlock()
	return totalItems
}

func (rs *ReceiptReminderStorage) RemoveItem(key interface{}) error {
	rs.Lock()
	defer rs.Unlock()

	if k, ok := key.(string); ok {
		delete(rs.reminders, k)
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
}
func (rs *ReceiptReminderStorage) GetSize() int64 {
	rs.Lock()
	defer rs.Unlock()

	var size int
	for k, v := range rs.reminders {
		size += len(k) + int(v.GetTypeInt())
	}
	return int64(size)
}

func (rs *ReceiptReminderStorage) ClearCache() error {
	rs.Lock()
	defer rs.Unlock()

	rs.reminders = make(map[string]chaintype.ChainType)
	return nil
}

func (rs *ReceiptReminderStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (rs *ReceiptReminderStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (rs *ReceiptReminderStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
