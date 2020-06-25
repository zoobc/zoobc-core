package service

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
)

type (
	// NodeRegistrationUtilsInterface represents interface for NodeRegistrationUtils
	NodeRegistrationUtilsInterface interface {
		GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte
		GetRegisteredNodesWithConsolidatedAddresses(height uint32) ([]*model.NodeRegistration, error)
	}

	// NodeRegistrationUtils nodeRegistration helper service methods
	NodeRegistrationUtils struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
	}
)

func NewNodeRegistrationUtils(
	executor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	logger *log.Logger,
) *NodeRegistrationUtils {
	return &NodeRegistrationUtils{
		QueryExecutor:         executor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		Logger:                logger,
	}
}

// GetUnsignedNodeAddressInfoBytes get NodeAddressInfo message bytes
func (nru *NodeRegistrationUtils) GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(nodeAddressMessage.GetNodeID())))
	addressLengthBytes := util.ConvertUint32ToBytes(uint32(len([]byte(
		nodeAddressMessage.Address,
	))))
	buffer.Write(addressLengthBytes)
	buffer.Write([]byte(nodeAddressMessage.Address))

	buffer.Write(util.ConvertUint32ToBytes(nodeAddressMessage.Port))
	buffer.Write(util.ConvertUint32ToBytes(nodeAddressMessage.BlockHeight))
	buffer.Write(nodeAddressMessage.BlockHash)
	return buffer.Bytes()
}

// GetRegisteredNodesWithConsolidatedAddresses returns registered nodes that have relative node address info records,
// selecting pending addresses, when available, over confirmed ones
func (nru *NodeRegistrationUtils) GetRegisteredNodesWithConsolidatedAddresses(height uint32) ([]*model.NodeRegistration, error) {
	// get all registry with addresses, grouped by nodeID and ordered by status
	rows, err := nru.QueryExecutor.ExecuteSelect(
		nru.NodeRegistrationQuery.GetNodeRegistryAtHeightWithNodeAddress(height),
		false,
	)
	if err != nil {
		nru.Logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	nodeRegistries, err := nru.NodeRegistrationQuery.BuildModelWithAddressInfo([]*model.NodeRegistration{}, rows)
	if err != nil {
		nru.Logger.Error(err.Error())
		return nil, err
	}

	mapRegistries := make(map[int64]*model.NodeRegistration)
	// consolidate the registry into a list of unique node Ids, preferring pending addresses rather than confirmed when present
	for _, nr := range nodeRegistries {
		if prevNr, ok := mapRegistries[nr.GetNodeID()]; ok &&
			prevNr.GetNodeAddressInfo().GetStatus() == model.NodeAddressStatus_NodeAddressPending {
			continue
		}
		mapRegistries[nr.GetNodeID()] = nr
	}
	// rebuild the registry array
	var res []*model.NodeRegistration
	for _, nr := range mapRegistries {
		res = append(res, nr)
	}
	return res, nil
}
