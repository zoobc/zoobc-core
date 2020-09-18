package storage

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeAddressInfoStorageInterface interface {
		AddAwaitedRemoveItem(key NodeAddressInfoStorageKey) error
		ClearAwaitedRemoveItems() error
	}
	// NodeAddressInfoStorage represent list of node address info
	NodeAddressInfoStorage struct {
		sync.RWMutex
		nodeAddressInfoMapByID          map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort map[string]map[int64]bool
		nodeAddressInfoMapByStatus      map[model.NodeAddressStatus]map[int64]map[string]bool
		awaitedRemoveList               map[int64]map[string]bool
	}
	// NodeAddressInfoStorage represent a key for NodeAddressInfoStorage
	NodeAddressInfoStorageKey struct {
		NodeID      int64
		AddressPort string
		Statuses    []model.NodeAddressStatus
	}
)

func NewNodeAddressInfoStorage() *NodeAddressInfoStorage {
	return &NodeAddressInfoStorage{
		nodeAddressInfoMapByID:          make(map[int64]map[string]model.NodeAddressInfo),
		nodeAddressInfoMapByAddressPort: make(map[string]map[int64]bool),
		nodeAddressInfoMapByStatus:      make(map[model.NodeAddressStatus]map[int64]map[string]bool),
		awaitedRemoveList:               make(map[int64]map[string]bool),
	}
}

func (nas *NodeAddressInfoStorage) SetItem(key, item interface{}) error {
	var nodeAddressInfo, ok = item.(model.NodeAddressInfo)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKey")
	}
	fullAddress := nodeAddressInfo.Address + ":" + strconv.Itoa(int(nodeAddressInfo.Port))
	nas.Lock()
	defer nas.Unlock()
	if nas.nodeAddressInfoMapByID[nodeAddressInfo.NodeID] == nil {
		nas.nodeAddressInfoMapByID[nodeAddressInfo.NodeID] = make(map[string]model.NodeAddressInfo)
	}
	nas.nodeAddressInfoMapByID[nodeAddressInfo.NodeID][fullAddress] = nodeAddressInfo

	if nas.nodeAddressInfoMapByAddressPort[fullAddress] == nil {
		nas.nodeAddressInfoMapByAddressPort[fullAddress] = make(map[int64]bool)
	}
	nas.nodeAddressInfoMapByAddressPort[fullAddress][nodeAddressInfo.NodeID] = true

	if nas.nodeAddressInfoMapByStatus[nodeAddressInfo.Status] == nil {
		nas.nodeAddressInfoMapByStatus[nodeAddressInfo.Status] = make(map[int64]map[string]bool)
	}
	if nas.nodeAddressInfoMapByStatus[nodeAddressInfo.Status][nodeAddressInfo.NodeID] == nil {
		nas.nodeAddressInfoMapByStatus[nodeAddressInfo.Status][nodeAddressInfo.NodeID] = make(map[string]bool)
	}
	nas.nodeAddressInfoMapByStatus[nodeAddressInfo.Status][nodeAddressInfo.NodeID][fullAddress] = true
	return nil
}

func (nas *NodeAddressInfoStorage) GetItem(key, item interface{}) error {
	nas.RLock()
	defer nas.RUnlock()
	storageKey, ok := key.(NodeAddressInfoStorageKey)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:NodeAddressInfoStorageKey")
	}
	nodeAddresses, ok := item.(*[]*model.NodeAddressInfo)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:*[]*model.NodeAddressInfo")
	}
	// staus node address info is always required when getting node address info
	if len(storageKey.Statuses) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "StatusNodeAddressInfoRequired")
	}
	switch {
	case storageKey.NodeID != 0:
		for _, status := range storageKey.Statuses {
			for fullAddressPort := range nas.nodeAddressInfoMapByStatus[status][storageKey.NodeID] {
				*nodeAddresses = nas.append(*nodeAddresses, nas.nodeAddressInfoMapByID[storageKey.NodeID][fullAddressPort])
			}
		}
	case storageKey.AddressPort != "":
		for _, status := range storageKey.Statuses {
			for nodeID := range nas.nodeAddressInfoMapByAddressPort[storageKey.AddressPort] {
				if nas.nodeAddressInfoMapByStatus[status][nodeID][storageKey.AddressPort] {
					*nodeAddresses = nas.append(*nodeAddresses, nas.nodeAddressInfoMapByID[nodeID][storageKey.AddressPort])
				}
			}
		}
	default:
		for _, status := range storageKey.Statuses {
			for nodeID, addressPortPotitions := range nas.nodeAddressInfoMapByStatus[status] {
				for addressPort := range addressPortPotitions {
					*nodeAddresses = nas.append(*nodeAddresses, nas.nodeAddressInfoMapByID[nodeID][addressPort])
				}
			}
		}

	}
	return nil
}

func (nas *NodeAddressInfoStorage) GetAllItems(item interface{}) error {
	nas.RLock()
	defer nas.RUnlock()
	nodeAddresses, ok := item.(*[]*model.NodeAddressInfo)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeItemExpected:*[]*model.NodeAddressInfo")
	}
	for _, nodeAddressInfos := range nas.nodeAddressInfoMapByID {
		for _, nodeAddressInfo := range nodeAddressInfos {
			*nodeAddresses = nas.append(*nodeAddresses, nodeAddressInfo)
		}
	}
	return nil
}

func (nas *NodeAddressInfoStorage) RemoveItem(key interface{}) error {
	nas.Lock()
	defer nas.Unlock()
	if key != nil {
		storageKey, ok := key.(NodeAddressInfoStorageKey)
		if !ok {
			return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeKeyExpected:NodeAddressInfoStorageKey")
		}
		// to remove node AddressInfo is require status and node ID
		if len(storageKey.Statuses) == 0 {
			return blocker.NewBlocker(blocker.ValidationErr, "StatusNodeAddressInfoRequired")
		}
		if storageKey.NodeID == 0 {
			return blocker.NewBlocker(blocker.ValidationErr, "NodeIDNodeAddressInfoRequired")
		}
		for _, status := range storageKey.Statuses {
			for fullAddressPort := range nas.nodeAddressInfoMapByStatus[status][storageKey.NodeID] {
				delete(nas.nodeAddressInfoMapByID[storageKey.NodeID], fullAddressPort)
				delete(nas.nodeAddressInfoMapByAddressPort[fullAddressPort], storageKey.NodeID)
				delete(nas.nodeAddressInfoMapByStatus[status][storageKey.NodeID], fullAddressPort)
			}
		}
	}

	// Remove all waiting node address info on remove list
	for nodeID, nodePotitionsByAddressPort := range nas.awaitedRemoveList {
		for fullAddress := range nodePotitionsByAddressPort {
			status := nas.nodeAddressInfoMapByID[nodeID][fullAddress].Status
			delete(nas.nodeAddressInfoMapByStatus[status][nodeID], fullAddress)
			delete(nas.nodeAddressInfoMapByAddressPort[fullAddress], nodeID)
			delete(nas.nodeAddressInfoMapByID[nodeID], fullAddress)
		}
	}
	_ = nas.ClearAwaitedRemoveItems()
	return nil
}

func (nas *NodeAddressInfoStorage) GetSize() int64 {
	nas.Lock()
	defer nas.Unlock()
	var (
		nasBytes bytes.Buffer
		enc      = gob.NewEncoder(&nasBytes)
	)
	_ = enc.Encode(nas.nodeAddressInfoMapByID)
	_ = enc.Encode(nas.nodeAddressInfoMapByAddressPort)
	_ = enc.Encode(nas.nodeAddressInfoMapByStatus)
	_ = enc.Encode(nas.awaitedRemoveList)
	return int64(len(nasBytes.Bytes()))
}

func (nas *NodeAddressInfoStorage) ClearCache() error {
	nas.Lock()
	defer nas.Unlock()
	nas.nodeAddressInfoMapByID = make(map[int64]map[string]model.NodeAddressInfo)
	nas.nodeAddressInfoMapByAddressPort = make(map[string]map[int64]bool)
	nas.nodeAddressInfoMapByStatus = make(map[model.NodeAddressStatus]map[int64]map[string]bool)
	nas.awaitedRemoveList = make(map[int64]map[string]bool)
	return nil
}

func (nas *NodeAddressInfoStorage) AddAwaitedRemoveItem(storageKey NodeAddressInfoStorageKey) error {
	nas.Lock()
	defer nas.Unlock()
	// to remove node AddressInfo is require status and node ID
	if len(storageKey.Statuses) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "StatusNodeAddressInfoRequired")
	}
	if storageKey.NodeID == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeIDNodeAddressInfoRequired")
	}

	for _, status := range storageKey.Statuses {
		for fullAddressPort := range nas.nodeAddressInfoMapByStatus[status][storageKey.NodeID] {
			if nas.awaitedRemoveList[storageKey.NodeID] == nil {
				nas.awaitedRemoveList[storageKey.NodeID] = make(map[string]bool)
			}
			nas.awaitedRemoveList[storageKey.NodeID][fullAddressPort] = true
		}
	}
	return nil
}

func (nas *NodeAddressInfoStorage) ClearAwaitedRemoveItems() error {
	nas.awaitedRemoveList = make(map[int64]map[string]bool)
	return nil
}

func (nas *NodeAddressInfoStorage) append(
	nodeAddresses []*model.NodeAddressInfo,
	nodeAddress model.NodeAddressInfo,
) []*model.NodeAddressInfo {
	var copyNodeAddress = nodeAddress
	copy(copyNodeAddress.BlockHash, nodeAddress.BlockHash)
	copy(copyNodeAddress.Signature, nodeAddress.Signature)
	return append(nodeAddresses, &copyNodeAddress)
}
