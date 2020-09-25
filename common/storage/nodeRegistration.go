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
	}
	switch castedIdx := idx.(type) {
	case nil:
		return blocker.NewBlocker(blocker.ValidationErr, "KeyCannotBeNil")
	case int:
		itemIndex = castedIdx
	case int64:
		itemIndex, ok = n.nodeIDIndexes[castedIdx]
		if !ok {
			return blocker.NewBlocker(blocker.ValidationErr, "NotFound")
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
	}
	*nodeRegistries = make([]NodeRegistry, len(n.nodeRegistries))
	for i := 0; i < len(n.nodeRegistries); i++ {
		(*nodeRegistries)[i] = n.copy(n.nodeRegistries[i])
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
		idxInt := n.transactionalNodeIDIndexes[castedIdx]
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
		idToRemove  int64
	)
	n.transactionalLock.Lock()
	defer n.transactionalLock.Unlock()
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
