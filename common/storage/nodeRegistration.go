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
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	NodeRegistryCacheStorage struct {
		isInTransaction   bool
		transactionalLock sync.RWMutex
		sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	// NodeRegistry store in-memory representation of node registry, excluding its NodeAddressInfo which is cache on
	// different storage struct
	NodeRegistry struct {
		Node               model.NodeRegistration
		ParticipationScore int64
	}
)

// NewNodeRegistryCacheStorage returns NodeRegistryCacheStorage instance
func NewNodeRegistryCacheStorage(
	metricLabel monitoring.CacheStorageType,
	sortFunc func([]NodeRegistry),
) *NodeRegistryCacheStorage {
	return &NodeRegistryCacheStorage{
		isInTransaction: false,
		nodeRegistries:  make([]NodeRegistry, 0),
		nodeIDIndexes:   make(map[int64]int),
		metricLabel:     metricLabel,
		sortItems:       sortFunc,
	}
}

// SetItem don't require index in node registry cache implementation, since it's a sorted array
func (n *NodeRegistryCacheStorage) SetItem(idx, item interface{}) error {
	nodeRegistry, ok := item.(NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:Storage.NodeRegistry")
	}
	n.Lock()
	defer n.Unlock()

	var tempPreviousCopy NodeRegistry
	switch castedIdx := idx.(type) {
	case nil:
		n.nodeRegistries = append(n.nodeRegistries, n.copy(nodeRegistry))
		n.sortItems(n.nodeRegistries)
		n.nodeIDIndexes = make(map[int64]int)
		for i, registry := range n.nodeRegistries {
			n.nodeIDIndexes[registry.Node.GetNodeID()] = i
		}
	case int:
		// update by index
		tempPreviousCopy = n.copy(n.nodeRegistries[castedIdx])
		n.nodeRegistries[castedIdx] = n.copy(nodeRegistry)
	case int64:
		// update by nodeID
		tempPreviousCopy = n.nodeRegistries[castedIdx]
		n.nodeRegistries[castedIdx] = n.copy(nodeRegistry)
	}
	// if this is pending node registry storage, and
	if n.metricLabel == monitoring.TypePendingNodeRegistryStorage {
		// locked balance has been updated
		if tempPreviousCopy.Node.GetLockedBalance() != nodeRegistry.Node.GetLockedBalance() {
			n.sortItems(n.nodeRegistries)
			n.nodeIDIndexes = make(map[int64]int)
			for i, registry := range n.nodeRegistries {
				n.nodeIDIndexes[registry.Node.GetNodeID()] = i
			}
		}
	}

	if monitoring.IsMonitoringActive() {
		go monitoring.SetCacheStorageMetrics(n.metricLabel, float64(n.size()))
	}
	return nil
}

func (n *NodeRegistryCacheStorage) SetItems(items interface{}) error {
	registries, ok := items.([]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemsMustBe:[]Storage.NodeRegistry")
	}
	n.Lock()
	defer n.Unlock()
	n.nodeRegistries = make([]NodeRegistry, len(registries))
	for i := 0; i < len(registries); i++ {
		n.nodeRegistries[i] = n.copy(registries[i])
	}
	// sort the updated registries
	n.sortItems(n.nodeRegistries)
	n.nodeIDIndexes = make(map[int64]int)
	// map registries sorted index for faster access
	for i := 0; i < len(n.nodeRegistries); i++ {
		n.nodeIDIndexes[n.nodeRegistries[i].Node.GetNodeID()] = i
	}
	if monitoring.IsMonitoringActive() {
		go monitoring.SetCacheStorageMetrics(n.metricLabel, float64(n.size()))
	}
	return nil
}

func (n *NodeRegistryCacheStorage) GetItem(idx, item interface{}) error {
	var (
		itemIndex int
	)
	nodeRegistry, ok := item.(*NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:*Storage.NodeRegistry")
	}
	if !n.isInTransaction {
		n.RLock()
		defer n.RUnlock()
	} else {
		n.transactionalLock.RLock()
		defer n.transactionalLock.RUnlock()
	}
	switch castedIdx := idx.(type) {
	case nil:
		return blocker.NewBlocker(blocker.ValidationErr, "KeyCannotBeNil")
	case int:
		itemIndex = castedIdx
	case int64:
		itemIndex, ok = n.nodeIDIndexes[castedIdx]
		if !ok {
			return blocker.NewBlocker(blocker.NotFound, "NodeRegistryNotFound")
		}
	default:
		return blocker.NewBlocker(blocker.ValidationErr, "UnknownType")
	}

	*nodeRegistry = n.copy(n.nodeRegistries[itemIndex])
	return nil
}

func (n *NodeRegistryCacheStorage) GetAllItems(item interface{}) error {
	nodeRegistries, ok := item.(*[]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:*Storage.NodeRegistry")
	}
	if !n.isInTransaction {
		n.RLock()
		defer n.RUnlock()
	} else {
		n.transactionalLock.RLock()
		defer n.transactionalLock.RUnlock()
	}
	*nodeRegistries = make([]NodeRegistry, len(n.nodeRegistries))
	for i := 0; i < len(n.nodeRegistries); i++ {
		(*nodeRegistries)[i] = n.copy(n.nodeRegistries[i])
	}
	return nil
}

func (n *NodeRegistryCacheStorage) GetTotalItems() int {
	n.RLock()
	var totalItems = len(n.transactionalNodeRegistries)
	n.RUnlock()
	return totalItems
}

func (n *NodeRegistryCacheStorage) RemoveItem(idx interface{}) error {
	n.Lock()
	defer n.Unlock()
	var (
		idxToRemove int
		idToRemove  int64
	)
	switch castedIdx := idx.(type) {
	case nil:
		return blocker.NewBlocker(blocker.ValidationErr, "TxRemoveItem:IdxCannotBeNil")
	case int64:
		idxToRemove = n.transactionalNodeIDIndexes[castedIdx]
		idToRemove = castedIdx
	case int:
		idToRemove = n.transactionalNodeRegistries[castedIdx].Node.GetNodeID()
		idxToRemove = castedIdx
	default:
		return blocker.NewBlocker(blocker.ValidationErr, "UnknownType")
	}
	if idxToRemove >= len(n.transactionalNodeRegistries) {
		return blocker.NewBlocker(blocker.NotFound, "RemoveItem:IndexOutOfRange")
	}
	tempLeft := n.nodeRegistries[:idxToRemove]
	tempRight := n.nodeRegistries[idxToRemove+1:]
	n.nodeRegistries = append(tempLeft, tempRight...)
	delete(n.nodeIDIndexes, idToRemove)
	if monitoring.IsMonitoringActive() {
		go monitoring.SetCacheStorageMetrics(n.metricLabel, float64(n.size()))
	}
	return nil
}

func (n *NodeRegistryCacheStorage) size() int64 {
	var (
		nBytes bytes.Buffer
		enc    = gob.NewEncoder(&nBytes)
	)
	_ = enc.Encode(n.nodeRegistries)
	_ = enc.Encode(n.nodeIDIndexes)
	return int64(nBytes.Len())
}

func (n *NodeRegistryCacheStorage) GetSize() int64 {
	n.RLock()
	defer n.RUnlock()
	return n.size()
}

func (n *NodeRegistryCacheStorage) ClearCache() error {
	n.Lock()
	defer n.Unlock()
	n.nodeRegistries = make([]NodeRegistry, 0)
	n.nodeIDIndexes = make(map[int64]int)
	if monitoring.IsMonitoringActive() {
		go monitoring.SetCacheStorageMetrics(n.metricLabel, float64(0))
	}
	return nil
}

// Transactional implementation

// Begin prepare data to begin doing transactional change to the cache, this implementation
// will never return error
func (n *NodeRegistryCacheStorage) Begin() error {
	n.Lock()
	n.transactionalLock.Lock()
	defer n.transactionalLock.Unlock()
	n.isInTransaction = true
	n.transactionalNodeIDIndexes = make(map[int64]int)
	n.transactionalNodeRegistries = make([]NodeRegistry, len(n.nodeRegistries))
	for i := 0; i < len(n.nodeRegistries); i++ {
		n.transactionalNodeRegistries[i] = n.copy(n.nodeRegistries[i])
		n.transactionalNodeIDIndexes[n.nodeRegistries[i].Node.NodeID] = i
	}
	return nil
}

// Commit of node registry cache replace all value in n.nodeRegistries
// this implementation will never return error
func (n *NodeRegistryCacheStorage) Commit() error {
	n.transactionalLock.Lock()
	defer func() {
		n.isInTransaction = false
		n.transactionalNodeIDIndexes = make(map[int64]int)
		n.transactionalNodeRegistries = make([]NodeRegistry, 0)
		if monitoring.IsMonitoringActive() {
			go monitoring.SetCacheStorageMetrics(n.metricLabel, float64(n.size()))
		}
		n.Unlock()
		n.transactionalLock.Unlock()
	}()
	// re-initialize actual value
	n.nodeRegistries = make([]NodeRegistry, len(n.transactionalNodeRegistries))
	n.nodeIDIndexes = make(map[int64]int)
	for i := 0; i < len(n.transactionalNodeRegistries); i++ {
		n.nodeRegistries[i] = n.transactionalNodeRegistries[i]
		n.nodeIDIndexes[n.transactionalNodeRegistries[i].Node.GetNodeID()] = i
	}
	return nil
}

// Rollback return the state of cache to before any changes made, either to transactional data
// or actual committed data. This implementation will never return error.
func (n *NodeRegistryCacheStorage) Rollback() error {
	n.transactionalLock.Lock()
	defer func() {
		n.isInTransaction = false
		n.Unlock()
		n.transactionalLock.Unlock()
	}()
	n.transactionalNodeIDIndexes = make(map[int64]int)
	n.transactionalNodeRegistries = make([]NodeRegistry, 0)
	return nil
}

func (n *NodeRegistryCacheStorage) TxSetItem(idx, item interface{}) error {
	nodeRegistry, ok := item.(NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:Storage.NodeRegistry")
	}
	// if id is not nil, mean we set item based on its nodeID (int64) or index (int)
	n.transactionalLock.Lock()
	defer n.transactionalLock.Unlock()
	var tempPreviousCopy NodeRegistry
	switch castedIdx := idx.(type) {
	case nil:
		n.transactionalNodeRegistries = append(n.transactionalNodeRegistries, n.copy(nodeRegistry))
		n.sortItems(n.transactionalNodeRegistries)
		for i, registry := range n.transactionalNodeRegistries {
			n.transactionalNodeIDIndexes[registry.Node.GetNodeID()] = i
		}
	case int:
		// update by index
		tempPreviousCopy = n.copy(n.transactionalNodeRegistries[castedIdx])
		n.transactionalNodeRegistries[castedIdx] = n.copy(nodeRegistry)
	case int64:
		// update by nodeID
		idxInt, ok := n.transactionalNodeIDIndexes[castedIdx]
		if !ok {
			return blocker.NewBlocker(blocker.NotFound, "NodeRegistryCacheStorage:TxSetItem-NotFound")
		}
		tempPreviousCopy = n.transactionalNodeRegistries[idxInt]
		n.transactionalNodeRegistries[idxInt] = n.copy(nodeRegistry)
	}

	// if this is pending node registry storage, and
	if n.metricLabel == monitoring.TypePendingNodeRegistryStorage {
		// locked balance has been updated
		if tempPreviousCopy.Node.GetLockedBalance() != nodeRegistry.Node.GetLockedBalance() {
			n.sortItems(n.transactionalNodeRegistries)
			n.transactionalNodeIDIndexes = make(map[int64]int)
			for i, registry := range n.transactionalNodeRegistries {
				n.transactionalNodeIDIndexes[registry.Node.GetNodeID()] = i
			}
		}
	}
	return nil
}

func (n *NodeRegistryCacheStorage) TxSetItems(items interface{}) error {
	registries, ok := items.([]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemsMustBe:[]Storage.NodeRegistry")
	}
	n.transactionalLock.Lock()
	defer n.transactionalLock.Unlock()
	n.transactionalNodeRegistries = make([]NodeRegistry, len(registries))
	n.transactionalNodeIDIndexes = make(map[int64]int)
	n.transactionalNodeRegistries = registries
	for i := 0; i < len(registries); i++ {
		n.transactionalNodeRegistries[i] = n.copy(registries[i])
	}
	// resort the node registries in transaction
	n.sortItems(n.transactionalNodeRegistries)
	// re-assign node-order in map for fast access
	for i := 0; i < len(n.transactionalNodeRegistries); i++ {
		n.transactionalNodeIDIndexes[registries[i].Node.GetNodeID()] = i
	}
	return nil
}

// TxRemoveItem remove an item from transactional state given the index
func (n *NodeRegistryCacheStorage) TxRemoveItem(idx interface{}) error {
	var (
		idxToRemove int
	)
	n.transactionalLock.Lock()
	defer n.transactionalLock.Unlock()
	switch castedIdx := idx.(type) {
	case nil:
		return blocker.NewBlocker(blocker.ValidationErr, "TxRemoveItem:IdxCannotBeNil")
	case int64:
		idxToRemove = n.transactionalNodeIDIndexes[castedIdx]
	case int:
		idxToRemove = castedIdx

	default:
		return blocker.NewBlocker(blocker.ValidationErr, "TxRemoveItem:UnknownType")
	}
	if idxToRemove >= len(n.transactionalNodeRegistries) {
		return blocker.NewBlocker(blocker.NotFound, "TxRemoveItem:IndexOutOfRange")
	}
	tempLeft := n.transactionalNodeRegistries[:idxToRemove]
	tempRight := n.transactionalNodeRegistries[idxToRemove+1:]
	n.transactionalNodeRegistries = append(tempLeft, tempRight...)
	n.transactionalNodeIDIndexes = make(map[int64]int)
	for i := 0; i < len(n.transactionalNodeRegistries); i++ {
		n.transactionalNodeIDIndexes[n.transactionalNodeRegistries[i].Node.GetNodeID()] = i
	}
	return nil
}

// copy manually copy the object to avoid referencing by the user of cache object
// this implementation also avoid the heavier alternative like `deepcopy` or `json.Marshal`
func (n *NodeRegistryCacheStorage) copy(src NodeRegistry) NodeRegistry {
	result := NodeRegistry{
		Node: model.NodeRegistration{
			NodeID:             src.Node.GetNodeID(),
			NodePublicKey:      make([]byte, len(src.Node.GetNodePublicKey())),
			AccountAddress:     src.Node.AccountAddress,
			RegistrationHeight: src.Node.RegistrationHeight,
			LockedBalance:      src.Node.LockedBalance,
			RegistrationStatus: src.Node.RegistrationStatus,
			Latest:             src.Node.Latest,
			Height:             src.Node.GetHeight(),
			NodeAddressInfo:    nil,
		},
		ParticipationScore: src.ParticipationScore,
	}
	copy(result.Node.NodePublicKey, src.Node.GetNodePublicKey())
	return result
}

func (n *NodeRegistryCacheStorage) GetItems(keys, items interface{}) error {
	return nil
}

func (n *NodeRegistryCacheStorage) RemoveItems(keys interface{}) error {
	return nil
}

func (n *NodeRegistryCacheStorage) CacheRegularCleaningListener() observer.Listener {
	return observer.Listener{}
}
