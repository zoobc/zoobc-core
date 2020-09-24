package storage

import (
	"bytes"
	"encoding/gob"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"sync"
)

type (
	NodeRegistryCacheStorage struct {
		isInTransaction bool
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
	n.nodeRegistries = make([]NodeRegistry, 0)
	for _, registry := range registries {
		n.nodeRegistries = append(n.nodeRegistries, n.copy(registry))
	}
	n.sortItems(n.nodeRegistries)
	n.nodeIDIndexes = make(map[int64]int)
	for i, registry := range n.nodeRegistries {
		n.nodeIDIndexes[registry.Node.GetNodeID()] = i
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
	}
	switch castedIdx := idx.(type) {
	case nil:
		return blocker.NewBlocker(blocker.ValidationErr, "KeyCannotBeNil")
	case int:
		itemIndex = castedIdx
	case int64:
		itemIndex = n.nodeIDIndexes[castedIdx]
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
	}
	for _, nr := range n.nodeRegistries {
		*nodeRegistries = append(*nodeRegistries, n.copy(nr))
	}
	return nil
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
	n.isInTransaction = true
	n.transactionalNodeIDIndexes = make(map[int64]int)
	n.transactionalNodeRegistries = make([]NodeRegistry, 0)
	for i, registry := range n.nodeRegistries {
		n.transactionalNodeRegistries = append(n.transactionalNodeRegistries, n.copy(registry))
		n.transactionalNodeIDIndexes[registry.Node.GetNodeID()] = i
	}
	return nil
}

// Commit of node registry cache replace all value in n.nodeRegistries
// this implementation will never return error
func (n *NodeRegistryCacheStorage) Commit() error {
	defer func() {
		n.isInTransaction = false
		n.transactionalNodeIDIndexes = make(map[int64]int)
		n.transactionalNodeRegistries = make([]NodeRegistry, 0)
		n.Unlock()
	}()
	for i, txRegistry := range n.transactionalNodeRegistries {
		n.nodeRegistries = append(n.nodeRegistries, n.copy(txRegistry))
		n.nodeIDIndexes[txRegistry.Node.GetNodeID()] = i
	}
	return nil
}

// Rollback return the state of cache to before any changes made, either to transactional data
// or actual committed data. This implementation will never return error.
func (n *NodeRegistryCacheStorage) Rollback() error {
	defer func() {
		n.isInTransaction = false
		n.Unlock()
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
		idxInt := n.transactionalNodeIDIndexes[castedIdx]
		tempPreviousCopy = n.transactionalNodeRegistries[idxInt]
		n.transactionalNodeRegistries[idxInt] = n.copy(nodeRegistry)
	}

	// if this is pending node registry storage, and
	if n.metricLabel == monitoring.TypePendingNodeRegistryStorage {
		// locked balance has been updated
		if tempPreviousCopy.Node.GetLockedBalance() != nodeRegistry.Node.GetLockedBalance() {
			n.sortItems(n.transactionalNodeRegistries)
		}
	}
	return nil
}

func (n *NodeRegistryCacheStorage) TxSetItems(items interface{}) error {
	registries, ok := items.([]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemsMustBe:[]Storage.NodeRegistry")
	}
	n.transactionalNodeRegistries = make([]NodeRegistry, 0)
	n.transactionalNodeRegistries = registries
	for i, registry := range registries {
		n.transactionalNodeRegistries = append(n.transactionalNodeRegistries, n.copy(registry))
		n.transactionalNodeIDIndexes[registry.Node.GetNodeID()] = i
	}
	n.sortItems(n.transactionalNodeRegistries)
	return nil
}

// TxRemoveItem remove an item from transactional state given the index
func (n *NodeRegistryCacheStorage) TxRemoveItem(idx interface{}) error {
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
	tempLeft := n.transactionalNodeRegistries[:idxToRemove]
	tempRight := n.transactionalNodeRegistries[idxToRemove+1:]
	n.transactionalNodeRegistries = append(tempLeft, tempRight...)
	delete(n.transactionalNodeIDIndexes, idToRemove)
	return nil
}

// copy manually copy the object to avoid referencing by the user of cache object
// this implementation also avoid the heavier alternative like `deepcopy` or `json.Marshal`
func (n *NodeRegistryCacheStorage) copy(src NodeRegistry) NodeRegistry {
	result := NodeRegistry{
		Node: model.NodeRegistration{
			NodeID:             src.Node.GetNodeID(),
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
