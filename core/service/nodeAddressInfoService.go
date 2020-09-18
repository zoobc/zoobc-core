package service

import (
	"bytes"
	"fmt"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
)

type (
	// NodeAddressInfoServiceInterface represents interface for NodeAddressInfoService
	NodeAddressInfoServiceInterface interface {
		GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte
		GetAddressInfoTableWithConsolidatedAddresses(preferredStatus model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		GetAddressInfoByNodeIDWithPreferredStatus(nodeID int64, preferredStatus model.NodeAddressStatus) (*model.NodeAddressInfo, error)
		GetAddressInfoByNodeID(nodeID int64, addressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		GetAddressInfoByNodeIDs(nodeIDs []int64, addressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		GetAddressInfoByAddressPort(
			address string,
			port uint32,
			nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		GetAddressInfoByStatus(nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		InsertAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error
		UpdateAddrressInfo(nodeAddressInfo *model.NodeAddressInfo) error
		ConfirmNodeAddressInfo(pendingNodeAddressInfo *model.NodeAddressInfo) error
		DeletePendingNodeAddressInfo(nodeID int64) error
		DeleteNodeAddressInfoByNodeIDInDBTx(nodeID int64) error
		CountNodesAddressByStatus() (map[model.NodeAddressStatus]int, error)
		ClearUpdateNodeAddressInfoCache() error
		ExecuteWaitedNodeAddressInfoCache() error
		ClearWaitedNodeAddressInfoCache()
	}

	// NodeAddressInfoService nodeRegistration helper service methods
	NodeAddressInfoService struct {
		QueryExecutor          query.ExecutorInterface
		NodeAddressInfoQuery   query.NodeAddressInfoQueryInterface
		NodeAddressInfoStorage *storage.NodeAddressInfoStorage
		Logger                 *log.Logger
	}
)

func NewNodeAddressInfoService(
	executor query.ExecutorInterface,
	nodeAddressInfoQuery query.NodeAddressInfoQueryInterface,
	nodeAddressesInfoStorage *storage.NodeAddressInfoStorage,
	logger *log.Logger,
) *NodeAddressInfoService {
	return &NodeAddressInfoService{
		QueryExecutor:          executor,
		NodeAddressInfoQuery:   nodeAddressInfoQuery,
		NodeAddressInfoStorage: nodeAddressesInfoStorage,
		Logger:                 logger,
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

// GetAddressInfoTableWithConsolidatedAddresses returns registered nodes that have relative node address info records,
// selecting addresses with 'preferredStatus', when available, over the other ones
func (nru *NodeAddressInfoService) GetAddressInfoTableWithConsolidatedAddresses(
	preferredStatus model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	// get all address info table, grouped by nodeID and ordered by status
	var (
		nodeAddressesInfo []*model.NodeAddressInfo
		mapAddresses      = make(map[int64]*model.NodeAddressInfo)
		err               = nru.NodeAddressInfoStorage.GetAllItems(&nodeAddressesInfo)
	)
	if err != nil {
		return nil, err
	}
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

// GetAddressInfoByNodeIDWithPreferredStatus returns a single address info from relative node id,
// preferring 'preferredStatus' address status over the others
func (nru *NodeAddressInfoService) GetAddressInfoByNodeIDWithPreferredStatus(
	nodeID int64,
	preferredStatus model.NodeAddressStatus,
) (*model.NodeAddressInfo, error) {
	// get a node address info given a node id
	var (
		err               error
		nodeAddressesInfo []*model.NodeAddressInfo
		keyGetItem        = storage.NodeAddressInfoStorageKey{
			NodeID: nodeID,
			Statuses: []model.NodeAddressStatus{
				model.NodeAddressStatus_NodeAddressPending,
				model.NodeAddressStatus_NodeAddressConfirmed,
			},
		}
	)
	err = nru.NodeAddressInfoStorage.GetItem(keyGetItem, &nodeAddressesInfo)
	if err != nil {
		return nil, err
	}
	// select node address based on status preference
	if len(nodeAddressesInfo) == 0 {
		return nil, nil
	}
	mapAddresses := make(map[int64]*model.NodeAddressInfo)
	for _, nai := range nodeAddressesInfo {
		if prevNr, ok := mapAddresses[nai.GetNodeID()]; ok &&
			prevNr.GetStatus() == preferredStatus {
			continue
		}
		mapAddresses[nai.GetNodeID()] = nai
	}
	return mapAddresses[nodeID], nil
}

// GetAddressInfoByNodeID return a list of node address info that have provied nodeID
func (nru *NodeAddressInfoService) GetAddressInfoByNodeID(
	nodeID int64,
	addressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	var (
		nodeAddressesInfo []*model.NodeAddressInfo
		keyGetItem        = storage.NodeAddressInfoStorageKey{
			Statuses: addressStatuses,
			NodeID:   nodeID,
		}
		err = nru.NodeAddressInfoStorage.GetItem(keyGetItem, &nodeAddressesInfo)
	)
	if err != nil {
		return nil, err
	}
	return nodeAddressesInfo, nil

}

// GetAddressInfoByNodeIDs return a list of node address info that have provied nodeIDs
func (nru *NodeAddressInfoService) GetAddressInfoByNodeIDs(
	nodeIDs []int64,
	addressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	var (
		nodeAddressInfos []*model.NodeAddressInfo
		keyGetItem       = storage.NodeAddressInfoStorageKey{
			Statuses: addressStatuses,
		}
	)
	for _, nodeID := range nodeIDs {
		keyGetItem.NodeID = nodeID
		err := nru.NodeAddressInfoStorage.GetItem(keyGetItem, &nodeAddressInfos)
		if err != nil {
			return nil, err
		}
	}
	return nodeAddressInfos, nil
}

// GetAddressInfoByStatus return a list of Node Address Info that have provided statuses
func (nru *NodeAddressInfoService) GetAddressInfoByStatus(
	nodeAddressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	var (
		nodeAddresses []*model.NodeAddressInfo
		keyGetItem    = storage.NodeAddressInfoStorageKey{
			Statuses: nodeAddressStatuses,
		}
		err = nru.NodeAddressInfoStorage.GetItem(keyGetItem, &nodeAddresses)
	)
	if err != nil {
		return nil, err
	}
	return nodeAddresses, nil
}

// GetAddressInfoByAddressPort returns a node address info given and address and port pairs
func (nru *NodeAddressInfoService) GetAddressInfoByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	var (
		nodeAddressesInfo []*model.NodeAddressInfo
		err               = nru.NodeAddressInfoStorage.GetItem(
			storage.NodeAddressInfoStorageKey{
				AddressPort: fmt.Sprintf("%s:%d", address, port),
				Statuses:    nodeAddressStatuses,
			},
			&nodeAddressesInfo,
		)
	)
	if err != nil {
		return nil, err
	}
	return nodeAddressesInfo, nil
}

// CountNodesAddressByStatus return a map with a count of nodes addresses in db for every node address status
func (nru *NodeAddressInfoService) CountNodesAddressByStatus() (map[model.NodeAddressStatus]int, error) {
	var (
		nodeAddressesInfo []*model.NodeAddressInfo
		err               = nru.NodeAddressInfoStorage.GetAllItems(&nodeAddressesInfo)
	)
	if err != nil {
		return nil, err
	}
	addressStatusCounter := make(map[model.NodeAddressStatus]int)
	for _, nai := range nodeAddressesInfo {
		addressStatus := nai.GetStatus()
		// init map key to avoid npe
		if _, ok := addressStatusCounter[addressStatus]; !ok {
			addressStatusCounter[addressStatus] = 0
		}
		addressStatusCounter[addressStatus]++
	}
	for status, counter := range addressStatusCounter {
		monitoring.SetNodeAddressStatusCount(counter, status)
	}
	return addressStatusCounter, nil
}

func (nru *NodeAddressInfoService) InsertAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	var err = nru.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	qry, args := nru.NodeAddressInfoQuery.InsertNodeAddressInfo(nodeAddressInfo)
	err = nru.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		errRollback := nru.QueryExecutor.RollbackTx()
		nru.Logger.Error(errRollback)
		return err
	}
	err = nru.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	// Add into node address info storage cache
	err = nru.NodeAddressInfoStorage.SetItem(nil, *nodeAddressInfo)
	if err != nil {
		return err
	}
	return nil
}

func (nru *NodeAddressInfoService) UpdateAddrressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	var err = nru.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	qryArgs := nru.NodeAddressInfoQuery.UpdateNodeAddressInfo(nodeAddressInfo)
	err = nru.QueryExecutor.ExecuteTransactions(qryArgs)
	if err != nil {
		errRollback := nru.QueryExecutor.RollbackTx()
		nru.Logger.Error(errRollback)
		return err
	}
	err = nru.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	// followed update query, will directly replace the old  node address info based on node ID
	err = nru.NodeAddressInfoStorage.SetItem(nil, *nodeAddressInfo)
	if err != nil {
		return err
	}
	return nil
}

func (nru *NodeAddressInfoService) ConfirmNodeAddressInfo(pendingNodeAddressInfo *model.NodeAddressInfo) error {
	pendingNodeAddressInfo.Status = model.NodeAddressStatus_NodeAddressConfirmed
	var (
		queries = nru.NodeAddressInfoQuery.ConfirmNodeAddressInfo(pendingNodeAddressInfo)
		err     = nru.QueryExecutor.BeginTx()
	)
	if err != nil {
		return err
	}
	err = nru.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := nru.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			log.Errorln(rollbackErr.Error())
		}
		return err
	}
	err = nru.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	// first remove all node address info based on provided node ID
	err = nru.NodeAddressInfoStorage.RemoveItem(storage.NodeAddressInfoStorageKey{
		NodeID: pendingNodeAddressInfo.NodeID,
		Statuses: []model.NodeAddressStatus{
			model.NodeAddressStatus_NodeAddressConfirmed,
			model.NodeAddressStatus_NodeAddressPending,
		},
	})
	if err != nil {
		return err
	}
	// then add new address info
	err = nru.NodeAddressInfoStorage.SetItem(nil, *pendingNodeAddressInfo)
	if err != nil {
		return err
	}
	return nil
}

func (nru *NodeAddressInfoService) DeletePendingNodeAddressInfo(nodeID int64) error {
	var (
		nodeAddressInfoStatuses = []model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending}
		qry, args               = nru.NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID(
			nodeID,
			nodeAddressInfoStatuses,
		)
		// start db transaction here
		err = nru.QueryExecutor.BeginTx()
	)
	if err != nil {
		return err
	}
	err = nru.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		if rollbackErr := nru.QueryExecutor.RollbackTx(); rollbackErr != nil {
			nru.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	err = nru.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	err = nru.NodeAddressInfoStorage.RemoveItem(
		storage.NodeAddressInfoStorageKey{
			NodeID:   nodeID,
			Statuses: nodeAddressInfoStatuses,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodeAddressInfoByNodeIDInDBTx will remove node adddress info in PushBlock process
func (nru *NodeAddressInfoService) DeleteNodeAddressInfoByNodeIDInDBTx(nodeID int64) error {
	var (
		removeNodeAddressInfoQ, removeNodeAddressInfoArgs = nru.NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID(
			nodeID,
			[]model.NodeAddressStatus{
				model.NodeAddressStatus_NodeAddressPending,
				model.NodeAddressStatus_NodeAddressConfirmed,
				model.NodeAddressStatus_Unset,
			},
		)
		err = nru.QueryExecutor.ExecuteTransaction(removeNodeAddressInfoQ, removeNodeAddressInfoArgs...)
	)
	if err != nil {
		return err
	}
	// add into list of awaited remove node address info
	return nru.NodeAddressInfoStorage.AddAwaitedRemoveItem(
		storage.NodeAddressInfoStorageKey{
			NodeID: nodeID,
			Statuses: []model.NodeAddressStatus{
				model.NodeAddressStatus_NodeAddressPending,
				model.NodeAddressStatus_NodeAddressConfirmed,
				model.NodeAddressStatus_Unset,
			},
		},
	)
}

// RemoveWaitedNodeAddressInfoCache will remove all node address info on
func (nru *NodeAddressInfoService) ExecuteWaitedNodeAddressInfoCache() error {
	return nru.NodeAddressInfoStorage.RemoveItem(nil)
}

func (nru *NodeAddressInfoService) ClearWaitedNodeAddressInfoCache() {
	_ = nru.NodeAddressInfoStorage.ClearAwaitedRemoveItems()
}

// ClearUpdateNodeAddressInfoCache to clear cache node address info and pull again from DB
func (nru *NodeAddressInfoService) ClearUpdateNodeAddressInfoCache() error {
	var rows, err = nru.QueryExecutor.ExecuteSelect(
		nru.NodeAddressInfoQuery.GetNodeAddressInfo(),
		false,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	nodeAddressesInfos, err := nru.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		return err
	}
	err = nru.NodeAddressInfoStorage.ClearCache()
	if err != nil {
		return err
	}
	for _, nodeAddressesInfo := range nodeAddressesInfos {
		err = nru.NodeAddressInfoStorage.SetItem(nil, *nodeAddressesInfo)
		if err != nil {
			if errStorage := nru.NodeAddressInfoStorage.ClearCache(); errStorage != nil {
				nru.Logger.Error(errStorage.Error())
			}
			return err
		}
	}
	return nil
}
