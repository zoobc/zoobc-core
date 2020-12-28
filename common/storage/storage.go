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

type (
	CacheStorageInterface interface {
		// SetItem take any item and store to its specific storage implementation
		SetItem(key, item interface{}) error
		// SetItems take all items that stored and refill item reference
		SetItems(item interface{}) error
		// GetItem take variable and assign implementation stored item to it
		GetItem(key, item interface{}) error
		// GetAllItems fetch all cached items
		GetAllItems(item interface{}) error
		// GetTotalItems fetch the number of total cached items
		GetTotalItems() int
		// RemoveItem remove item by providing the key(s)
		RemoveItem(key interface{}) error
		// GetSize return the size of storage in number of `byte`
		GetSize() int64
		// ClearCache empty the storage item
		ClearCache() error
	}

	CacheStackStorageInterface interface {
		// Pop delete the latest item on the stack
		Pop() error
		// Push item into the stack, if exceed size first item is deleted and shifted
		Push(interface{}) error
		// PopTo takes index (uint32) and delete item to the given index (start from 0)
		PopTo(uint32) error
		// GetAll return all item in the stack to given `interface` arguments
		GetAll(interface{}) error
		// GetAtIndex return item at given index
		GetAtIndex(uint32, interface{}) error
		// GetTop return top item on the stack
		GetTop(interface{}) error
		// Clear clean up the whole stack and reinitialize with new array
		Clear() error
	}

	TransactionalCache interface {
		// Begin prepare state of cache for transactional writes, must called at start of tx writes
		Begin() error
		// Commit finalize transactional writes to the struct
		Commit() error
		// Rollback release locks and return state of struct to original before tx modifications are made
		Rollback() error
		// TxSetItem set individual item
		TxSetItem(id, item interface{}) error
		// TxSetItems replace items in bulk
		TxSetItems(items interface{}) error
		// TxRemoveItem remove item with given ID
		TxRemoveItem(id interface{}) error
	}
)
