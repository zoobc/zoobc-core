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
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	ScrambleCacheStackStorage struct {
		itemLimit int
		sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
)

func NewScrambleCacheStackStorage() *ScrambleCacheStackStorage {
	// store 2 times rollback height worth of scramble nodes
	return &ScrambleCacheStackStorage{
		itemLimit:      int(constant.MaxScrambleCacheRound),
		scrambledNodes: make([]model.ScrambledNodes, 0, constant.MaxScrambleCacheRound),
	}
}

func (s *ScrambleCacheStackStorage) Pop() error {
	if len(s.scrambledNodes) > 0 {
		s.Lock()
		defer s.Unlock()
		s.scrambledNodes = s.scrambledNodes[:len(s.scrambledNodes)-1]
		return nil
	}
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	// no more to pop
	return blocker.NewBlocker(blocker.ValidationErr, "StackEmpty")
}

func (s *ScrambleCacheStackStorage) Push(item interface{}) error {
	scrambleCopy, ok := item.(model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	s.Lock()
	defer s.Unlock()
	if len(s.scrambledNodes) >= s.itemLimit {
		if len(s.scrambledNodes) != 0 {
			s.scrambledNodes = s.scrambledNodes[1:] // remove first (oldest) cache to make room for new scramble
		}
	}
	s.scrambledNodes = append(s.scrambledNodes, scrambleCopy)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	return nil
}

// PopTo pop the scramble stack to index-th element (last element = index-th element)
func (s *ScrambleCacheStackStorage) PopTo(index uint32) error {
	if int(index)+1 > len(s.scrambledNodes) {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	s.Lock()
	defer s.Unlock()
	s.scrambledNodes = s.scrambledNodes[:index+1]
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, float64(s.size()))
	}
	return nil
}

func (s *ScrambleCacheStackStorage) GetAll(items interface{}) error {
	scrambledNodesCopy, ok := items.(*[]model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNodeList")
	}
	s.RLock()
	defer s.RUnlock()
	*scrambledNodesCopy = make([]model.ScrambledNodes, len(s.scrambledNodes))
	for i, scrambleNode := range s.scrambledNodes {
		var tempScramble = s.copy(scrambleNode)
		(*scrambledNodesCopy)[i] = tempScramble
	}
	return nil
}

func (s *ScrambleCacheStackStorage) GetAtIndex(index uint32, item interface{}) error {
	if int(index) >= len(s.scrambledNodes) {
		return blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("IndexOutOfRange-Has %d scramble round - requested index %d",
				len(s.scrambledNodes),
				index,
			))
	}
	scrambleNodeCopy, ok := item.(*model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	s.RLock()
	defer s.RUnlock()
	*scrambleNodeCopy = s.copy(s.scrambledNodes[int(index)])
	return nil
}

func (s *ScrambleCacheStackStorage) GetTop(item interface{}) error {
	s.RLock()
	defer s.RUnlock()
	topIndex := len(s.scrambledNodes)
	if topIndex == 0 {
		return blocker.NewBlocker(blocker.CacheEmpty, "EmptyScramble")
	}
	scrambleNodeCopy, ok := item.(*model.ScrambledNodes)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemIsNotScrambleNode")
	}
	*scrambleNodeCopy = s.copy(s.scrambledNodes[topIndex-1])
	return nil
}

func (s *ScrambleCacheStackStorage) Clear() error {
	s.scrambledNodes = make([]model.ScrambledNodes, 0, s.itemLimit)
	if monitoring.IsMonitoringActive() {
		monitoring.SetCacheStorageMetrics(monitoring.TypeScrambleNodeCacheStorage, 0)
	}
	return nil
}

func (s *ScrambleCacheStackStorage) size() int {
	var size int
	var (
		scrambleBytes bytes.Buffer
		enc           = gob.NewEncoder(&scrambleBytes)
	)
	_ = enc.Encode(s.scrambledNodes)
	size = scrambleBytes.Len()
	return size
}

func (s *ScrambleCacheStackStorage) copy(src model.ScrambledNodes) model.ScrambledNodes {
	var (
		result                  model.ScrambledNodes
		newIndexNodes           = make(map[string]*int, len(src.AddressNodes))
		newNodePublicKeyToIDMap = make(map[string]int64, len(src.AddressNodes))
		newPeers                = make([]*model.Peer, len(src.AddressNodes))
	)
	for i, node := range src.AddressNodes {
		idx := i
		newNodePublicKeyToIDMap[hex.EncodeToString(node.GetInfo().GetPublicKey())] = node.GetInfo().GetID()
		scrambleDNodeMapKey := fmt.Sprintf("%d", node.GetInfo().GetID())
		newIndexNodes[scrambleDNodeMapKey] = &idx
		tempPeer := model.Peer{
			Info: &model.Node{
				ID:            node.GetInfo().GetID(),
				PublicKey:     node.GetInfo().GetPublicKey(),
				SharedAddress: node.GetInfo().GetSharedAddress(),
				Address:       node.GetInfo().GetAddress(),
				Port:          node.GetInfo().GetPort(),
				AddressStatus: node.GetInfo().GetAddressStatus(),
				Version:       node.GetInfo().GetVersion(),
				CodeName:      node.GetInfo().GetCodeName(),
			},
			LastInboundRequest:  node.GetLastInboundRequest(),
			BlacklistingCause:   node.GetBlacklistingCause(),
			BlacklistingTime:    node.GetBlacklistingTime(),
			ResolvingTime:       node.GetResolvingTime(),
			ConnectionAttempted: node.GetConnectionAttempted(),
			UnresolvingTime:     node.GetUnresolvingTime(),
		}
		newPeers[i] = &tempPeer
	}
	result = model.ScrambledNodes{
		IndexNodes:           newIndexNodes,
		NodePublicKeyToIDMap: newNodePublicKeyToIDMap,
		AddressNodes:         newPeers,
		BlockHeight:          src.BlockHeight,
	}
	return result
}

func (s *ScrambleCacheStackStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (s *ScrambleCacheStackStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (s *ScrambleCacheStackStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
