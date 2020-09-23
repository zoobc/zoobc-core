package storage

import (
	"bytes"
	"encoding/gob"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"sync"
)

type (
	NodeRegistryCacheStorage struct {
		sync.RWMutex
		nodeRegistries []NodeRegistry
	}
	// NodeRegistry store in-memory representation of node registry, excluding its NodeAddressInfo which is cache on
	// different storage struct
	NodeRegistry struct {
		Node               model.NodeRegistration
		ParticipationScore int64
	}
)

// NewNodeRegistryCacheStorage returns NodeRegistryCacheStorage instance
func NewNodeRegistryCacheStorage() *NodeRegistryCacheStorage {
	return &NodeRegistryCacheStorage{}
}

func (n *NodeRegistryCacheStorage) SetItem(index, item interface{}) error {
	indexInt, ok := (index).(int)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexMustBeInteger")
	}
	nodeRegistry, ok := item.(NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:Storage.NodeRegistry")
	}
	n.Lock()
	defer n.Unlock()
	if indexInt > len(n.nodeRegistries)-1 {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	n.nodeRegistries[indexInt] = n.copy(nodeRegistry)
	return nil
}

func (n *NodeRegistryCacheStorage) SetItems(items interface{}) error {
	registries, ok := items.([]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemsMustBe:[]Storage.NodeRegistry")
	}
	n.Lock()
	defer n.Unlock()
	for _, nr := range registries {
		n.nodeRegistries = append(n.nodeRegistries, n.copy(nr))
	}
	return nil
}

func (n *NodeRegistryCacheStorage) GetItem(index, item interface{}) error {
	indexInt, ok := (index).(int)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexMustBeInteger")
	}
	nodeRegistry, ok := item.(*NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:*Storage.NodeRegistry")
	}
	n.RLock()
	defer n.RUnlock()
	if indexInt > len(n.nodeRegistries)-1 {
		return blocker.NewBlocker(blocker.ValidationErr, "IndexOutOfRange")
	}
	*nodeRegistry = n.copy(n.nodeRegistries[indexInt])
	return nil
}

func (n *NodeRegistryCacheStorage) GetAllItems(item interface{}) error {
	nodeRegistries, ok := item.(*[]NodeRegistry)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ItemTypeMustBe:*Storage.NodeRegistry")
	}
	n.RLock()
	defer n.RUnlock()
	for _, nr := range n.nodeRegistries {
		*nodeRegistries = append(*nodeRegistries, n.copy(nr))
	}
	return nil
}

func (n *NodeRegistryCacheStorage) RemoveItem(index interface{}) error {
	return nil
}

func (n *NodeRegistryCacheStorage) GetSize() int64 {
	n.RLock()
	defer n.RUnlock()
	var (
		nBytes bytes.Buffer
		enc    = gob.NewEncoder(&nBytes)
	)
	_ = enc.Encode(n.nodeRegistries)
	return int64(nBytes.Len())
}

func (n *NodeRegistryCacheStorage) ClearCache() error {
	n.Lock()
	defer n.Unlock()
	n.nodeRegistries = make([]NodeRegistry, 0)
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
