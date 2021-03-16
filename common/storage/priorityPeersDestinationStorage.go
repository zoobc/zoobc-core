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
)

type (
	// PriorityPeersDestinationCacheHybridStorage memoizes everytime a node public key is a priority of any node
	PriorityPeersDestinationCacheHybridStorage struct {
		minCount   int
		safeHeight uint32
		sync.RWMutex
		priorityPeersDestinations map[string][]uint32
	}
)

func NewPriorityPeersDestinationCacheHybridStorage() *PriorityPeersDestinationCacheHybridStorage {
	// store 2 times rollback height worth of scramble nodes
	return &PriorityPeersDestinationCacheHybridStorage{
		minCount:                  int(constant.PriorityStrategyMaxPriorityPeers),
		safeHeight:                2 * (constant.MinRollbackBlocks + constant.BatchReceiptLookBackHeight),
		priorityPeersDestinations: make(map[string][]uint32),
	}
}

// SetItem take any item and store to its specific storage implementation
func (ppdc *PriorityPeersDestinationCacheHybridStorage) SetItem(key, item interface{}) error {
	// storing item (blockHeight) at which key (public key) is a priority of any node
	// and removing irrelevant blockHeights
	// should be performed every push block
	var (
		pubKey      string
		blockHeight uint32
		ok          bool
	)

	ppdc.Lock()
	defer ppdc.Unlock()

	if pubKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if blockHeight, ok = item.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}

	if _, ok := ppdc.priorityPeersDestinations[pubKey]; !ok {
		ppdc.priorityPeersDestinations[pubKey] = make([]uint32, 0)
	}

	// storing the item
	ppdc.priorityPeersDestinations[pubKey] = append(ppdc.priorityPeersDestinations[pubKey], blockHeight)

	// removing irrelevant items but keeping safe heights
	if blockHeight > ppdc.safeHeight && len(ppdc.priorityPeersDestinations[pubKey]) > ppdc.minCount {
		usedBlockHeight := blockHeight - ppdc.safeHeight
		firstCycleCap := blockHeight - (ppdc.safeHeight / 2)
		heightsToEvaluate := ppdc.priorityPeersDestinations[pubKey]
		safeItemsCount := 0
		lastIndex := len(heightsToEvaluate) - 1

		for lastIndex >= 0 {
			if heightsToEvaluate[lastIndex] < usedBlockHeight {
				break
			} else if heightsToEvaluate[lastIndex] <= firstCycleCap {
				safeItemsCount++
			}
			if safeItemsCount > ppdc.minCount {
				break
			}
			lastIndex--
		}

		if lastIndex > -1 && safeItemsCount > ppdc.minCount {
			ppdc.priorityPeersDestinations[pubKey] = heightsToEvaluate[lastIndex+1:]
		}
	}

	return nil
}

// GetItem take variable and assign implementation stored item to it
func (ppdc *PriorityPeersDestinationCacheHybridStorage) GetItem(key, item interface{}) error {
	// get block heights at which key (public key) is a priority peers of any node
	var (
		pubKey       string
		blockHeights *[]uint32
		ok           bool
	)

	ppdc.Lock()
	defer ppdc.Unlock()

	if pubKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}
	if blockHeights, ok = item.(*[]uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item")
	}

	if _, ok := ppdc.priorityPeersDestinations[pubKey]; !ok {
		return nil
	}
	for _, blockHeight := range ppdc.priorityPeersDestinations[pubKey] {
		*blockHeights = append(*blockHeights, blockHeight)
	}

	return nil
}

// RemoveItems remove item by providing the keys
func (ppdc *PriorityPeersDestinationCacheHybridStorage) RemoveItems(key interface{}) error {
	// remove items which has specific `key`
	// should be performed everytime a node is kicked out (triggered in spine public key algorithm)
	var (
		pubKey string
		ok     bool
	)

	ppdc.Lock()
	defer ppdc.Unlock()

	if pubKey, ok = key.(string); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}

	delete(ppdc.priorityPeersDestinations, pubKey)

	return nil
}

// RemoveItem remove item by providing the key
func (ppdc *PriorityPeersDestinationCacheHybridStorage) RemoveItem(key interface{}) error {
	// remove blockHeights upto specific key (height)
	var (
		popOffToHeight uint32
		ok             bool
	)

	ppdc.Lock()
	defer ppdc.Unlock()

	if popOffToHeight, ok = key.(uint32); !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType key")
	}

	for key, blockHeights := range ppdc.priorityPeersDestinations {
		indexCount := 0
		for indexCount < len(blockHeights) {
			if blockHeights[indexCount] > popOffToHeight {
				break
			}
			indexCount++
		}
		if indexCount < len(blockHeights) {
			ppdc.priorityPeersDestinations[key] = blockHeights[:indexCount]
		}
	}

	return nil
}

// GetSize return the size of storage in number of `byte`
func (ppdc *PriorityPeersDestinationCacheHybridStorage) GetSize() int64 {
	var accumulatedSize int64
	for _, heights := range ppdc.priorityPeersDestinations {
		accumulatedSize += int64(len(heights) * 4) // 4 bytes for each uint32
	}
	return accumulatedSize
}

// ClearCache empty the storage item
func (ppdc *PriorityPeersDestinationCacheHybridStorage) ClearCache() error {
	ppdc.Lock()
	defer ppdc.Unlock()

	ppdc.priorityPeersDestinations = make(map[string][]uint32)
	return nil
}
