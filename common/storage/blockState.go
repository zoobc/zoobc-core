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
	"encoding/binary"
	"encoding/json"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlockStateStorage represent last state of block
	BlockStateStorage struct {
		sync.RWMutex
		// save last block in bytes to make esier to convert when requesting the block
		lastBlockBytes []byte
	}
)

// NewBlockStateStorage returns BlockStateStorage instance
func NewBlockStateStorage() *BlockStateStorage {
	return &BlockStateStorage{}
}

// SetItem setter of BlockStateStorage
func (bs *BlockStateStorage) SetItem(lastUpdate, block interface{}) error {
	bs.Lock()
	defer bs.Unlock()
	var (
		newBlock, ok = (block).(model.Block)
	)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}
	newBlockBytes, err := json.Marshal(newBlock)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "Failed marshal block")
	}
	bs.lastBlockBytes = newBlockBytes
	return nil
}

func (bs *BlockStateStorage) SetItems(_ interface{}) error {
	return nil
}

// GetItem getter of BlockStateStorage
func (bs *BlockStateStorage) GetItem(lastUpdate, block interface{}) error {
	bs.RLock()
	defer bs.RUnlock()

	var (
		ok        bool
		err       error
		blockCopy *model.Block
	)

	if bs.lastBlockBytes == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptyCache")
	}

	blockCopy, ok = block.(*model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	err = json.Unmarshal(bs.lastBlockBytes, blockCopy)
	if err != nil {
		return blocker.NewBlocker(blocker.BlockErr, "Failed unmarshal block bytes")
	}
	return nil
}

func (bs *BlockStateStorage) GetAllItems(item interface{}) error {
	return nil
}

func (bs *BlockStateStorage) GetTotalItems() int {
	// this storage only have 1 item
	return 1
}

func (bs *BlockStateStorage) RemoveItem(key interface{}) error {
	return nil
}

// GetSize return the size of BlockStateStorage
func (bs *BlockStateStorage) GetSize() int64 {
	return int64(binary.Size(bs.lastBlockBytes))
}

// ClearCache cleaner of BlockStateStorage
func (bs *BlockStateStorage) ClearCache() error {
	bs.lastBlockBytes = nil
	return nil
}
