package service

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
)

type (
	// NodeAddressInfoServiceInterface represents interface for NodeAddressInfoService
	NodeAddressInfoServiceInterface interface {
		GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte
		GetRegisteredNodesWithConsolidatedAddresses(
			height uint32,
			preferredStatus model.NodeAddressStatus) ([]*model.NodeRegistration, error)
		GetAddressInfoTableWithConsolidatedAddresses(preferredStatus model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
	}

	// NodeAddressInfoService nodeRegistration helper service methods
	NodeAddressInfoService struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		NodeAddressInfoQuery  query.NodeAddressInfoQueryInterface
		Logger                *log.Logger
	}
)

func NewNodeAddressInfoService(
	executor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	nodeAddressInfoQuery query.NodeAddressInfoQueryInterface,
	logger *log.Logger,
) *NodeAddressInfoService {
	return &NodeAddressInfoService{
		QueryExecutor:         executor,
		NodeRegistrationQuery: nodeRegistrationQuery,
		NodeAddressInfoQuery:  nodeAddressInfoQuery,
		Logger:                logger,
	}
}

// GetUnsignedNodeAddressInfoBytes get NodeAddressInfo message bytes
func (nru *NodeAddressInfoService) GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte {
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
// selecting addresses with 'preferredStatus', when available, over the other ones
func (nru *NodeAddressInfoService) GetRegisteredNodesWithConsolidatedAddresses(
	height uint32,
	preferredStatus model.NodeAddressStatus) ([]*model.NodeRegistration, error) {
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
			prevNr.GetNodeAddressInfo().GetStatus() == preferredStatus {
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

// GetAddressInfoTableWithConsolidatedAddresses returns registered nodes that have relative node address info records,
// selecting addresses with 'preferredStatus', when available, over the other ones
func (nru *NodeAddressInfoService) GetAddressInfoTableWithConsolidatedAddresses(
	preferredStatus model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	// get all address info table, grouped by nodeID and ordered by status
	rows, err := nru.QueryExecutor.ExecuteSelect(
		nru.NodeAddressInfoQuery.GetNodeAddressInfo(),
		false,
	)
	if err != nil {
		nru.Logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	nodeAddressesInfo, err := nru.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		nru.Logger.Error(err.Error())
		return nil, err
	}

	mapAddresses := make(map[int64]*model.NodeAddressInfo)
	// consolidate the registry into a list of unique node Ids, preferring pending addresses rather than confirmed when present
	for _, nai := range nodeAddressesInfo {
		if prevNr, ok := mapAddresses[nai.GetNodeID()]; ok &&
			prevNr.GetStatus() == preferredStatus {
			continue
		}
		mapAddresses[nai.GetNodeID()] = nai
	}
	// rebuild the addressInfo array
	var res []*model.NodeAddressInfo
	for _, nai := range mapAddresses {
		res = append(res, nai)
	}
	return res, nil
}
