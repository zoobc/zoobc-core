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
package service

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// NodeAddressInfoServiceInterface represents interface for NodeAddressInfoService
	NodeAddressInfoServiceInterface interface {
		GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte
		GenerateNodeAddressInfo(
			nodeID int64,
			nodeAddress string,
			port uint32,
			nodeSecretPhrase string) (*model.NodeAddressInfo, error)
		ValidateNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) (found bool, err error)
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
		UpdateOrInsertAddressInfo(
			nodeAddressInfo *model.NodeAddressInfo,
			updatedStatus model.NodeAddressStatus,
		) (updated bool, err error)
		ConfirmNodeAddressInfo(pendingNodeAddressInfo *model.NodeAddressInfo) error
		DeletePendingNodeAddressInfo(nodeID int64) error
		DeleteNodeAddressInfoByNodeIDInDBTx(nodeID int64) error
		CountNodesAddressByStatus() (map[model.NodeAddressStatus]int, error)
		CountRegistredNodeAddressWithAddressInfo() (int, error)
		ClearUpdateNodeAddressInfoCache() error
		BeginCacheTransaction() error
		RollbackCacheTransaction() error
		CommitCacheTransaction() error
	}

	// NodeAddressInfoService nodeRegistration helper service methods
	NodeAddressInfoService struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		MainBlocksStorage       storage.CacheStackStorageInterface
		Logger                  *log.Logger
	}
)

func NewNodeAddressInfoService(
	executor query.ExecutorInterface,
	nodeAddressInfoQuery query.NodeAddressInfoQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	blockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	nodeAddressesInfoStorage, mainBlockStateStorage, activeNodeRegistryCache storage.CacheStorageInterface,
	mainBlocksStorage storage.CacheStackStorageInterface,
	logger *log.Logger,
) *NodeAddressInfoService {
	return &NodeAddressInfoService{
		QueryExecutor:           executor,
		NodeAddressInfoQuery:    nodeAddressInfoQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		BlockQuery:              blockQuery,
		Signature:               signature,
		NodeAddressInfoStorage:  nodeAddressesInfoStorage,
		MainBlockStateStorage:   mainBlockStateStorage,
		ActiveNodeRegistryCache: activeNodeRegistryCache,
		MainBlocksStorage:       mainBlocksStorage,
		Logger:                  logger,
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

// GenerateNodeAddressInfo generate a nodeAddressInfo signed message
func (nru *NodeAddressInfoService) GenerateNodeAddressInfo(
	nodeID int64,
	nodeAddress string,
	port uint32,
	nodeSecretPhrase string) (*model.NodeAddressInfo, error) {
	var (
		safeBlockHeight uint32
		safeBlock       *storage.BlockCacheObject
		lastBlock       model.Block
		err             = nru.MainBlockStateStorage.GetItem(nil, &lastBlock)
	)
	if err != nil {
		return nil, err
	}
	// get a rollback-safe block for node address info message, to make sure evey peer can validate it
	// note: a disadvantage of this is, once a node address is written to db, it cannot be updated in the first 720 blocks
	if lastBlock.GetHeight() < constant.MinRollbackBlocks {
		safeBlockHeight = 0
	} else {
		safeBlockHeight = lastBlock.GetHeight() - constant.MinRollbackBlocks
	}
	safeBlock, err = util.GetBlockByHeightUseBlocksCache(
		safeBlockHeight,
		nru.QueryExecutor,
		nru.BlockQuery,
		nru.MainBlocksStorage,
	)
	if err != nil {
		return nil, err
	}
	nodeAddressInfo := &model.NodeAddressInfo{
		NodeID:      nodeID,
		Address:     nodeAddress,
		Port:        port,
		BlockHeight: safeBlock.Height,
		BlockHash:   safeBlock.BlockHash,
	}
	nodeAddressInfoBytes := nru.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	nodeAddressInfo.Signature = nru.Signature.SignByNode(nodeAddressInfoBytes, nodeSecretPhrase)
	return nodeAddressInfo, nil
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

// CountRegistredNodeAddressWithAddressInfo return the number of node registry that already have node address info
func (nru *NodeAddressInfoService) CountRegistredNodeAddressWithAddressInfo() (int, error) {
	var (
		counter       int
		countQuery    = query.GetTotalRecordOfSelect(nru.NodeRegistrationQuery.GetActiveNodeRegistrationsWithNodeAddress())
		rowCount, err = nru.QueryExecutor.ExecuteSelectRow(countQuery, false)
	)
	if err != nil {
		return 0, err
	}
	err = rowCount.Scan(
		&counter,
	)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func (nru *NodeAddressInfoService) InsertAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	qry, args := nru.NodeAddressInfoQuery.InsertNodeAddressInfo(nodeAddressInfo)
	_, err := nru.QueryExecutor.ExecuteStatement(qry, args...)
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
	qry, args := nru.NodeAddressInfoQuery.UpdateNodeAddressInfo(nodeAddressInfo)
	_, err := nru.QueryExecutor.ExecuteStatement(qry, args)
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

// ConfirmPendingNodeAddress confirm a pending address by inserting or replacing the previously confirmed one and deleting the pending address
func (nru *NodeAddressInfoService) ConfirmNodeAddressInfo(pendingNodeAddressInfo *model.NodeAddressInfo) error {
	pendingNodeAddressInfo.Status = model.NodeAddressStatus_NodeAddressConfirmed
	var (
		isDbTransactionHighPriority = false
		queries                     = nru.NodeAddressInfoQuery.ConfirmNodeAddressInfo(pendingNodeAddressInfo)
		err                         = nru.QueryExecutor.BeginTx(isDbTransactionHighPriority, monitoring.ConfirmNodeAddressInfoOwnerProcess)
	)
	if err != nil {
		return err
	}
	err = nru.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := nru.QueryExecutor.RollbackTx(isDbTransactionHighPriority)
		if rollbackErr != nil {
			log.Errorln(rollbackErr.Error())
		}
		return err
	}
	err = nru.QueryExecutor.CommitTx(isDbTransactionHighPriority)
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

	if monitoring.IsMonitoringActive() {
		if countResult, err := nru.CountRegistredNodeAddressWithAddressInfo(); err == nil {
			monitoring.SetNodeAddressInfoCount(countResult)
		}
		if cna, err := nru.CountNodesAddressByStatus(); err == nil {
			for status, counter := range cna {
				monitoring.SetNodeAddressStatusCount(counter, status)
			}
		}
	}
	return nil
}

// DeletePendingNodeAddressInfo to delete pending node addrress based on node ID
func (nru *NodeAddressInfoService) DeletePendingNodeAddressInfo(nodeID int64) error {
	var (
		nodeAddressInfoStatuses = []model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending}
		qry, args               = nru.NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID(
			nodeID,
			nodeAddressInfoStatuses,
		)
		isDbTransactionHighPriority = false
		// start db transaction here
		err = nru.QueryExecutor.BeginTx(isDbTransactionHighPriority, monitoring.DeletePendingNodeAddressInfoOwnerProcess)
	)
	if err != nil {
		return err
	}
	err = nru.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		if rollbackErr := nru.QueryExecutor.RollbackTx(isDbTransactionHighPriority); rollbackErr != nil {
			nru.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	err = nru.QueryExecutor.CommitTx(isDbTransactionHighPriority)
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
	txNodeAddressInfoCache, ok := nru.NodeAddressInfoStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastNodeAddressInfoStorageAsTransactionalCacheInterface")
	}
	// add into list of awaited remove node address info
	return txNodeAddressInfoCache.TxRemoveItem(
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

// BeginCacheTransaction to begin transactional process of NodeAddressInfoStorage
func (nru *NodeAddressInfoService) BeginCacheTransaction() error {
	txNodeAddressInfoCache, ok := nru.NodeAddressInfoStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastNodeAddressInfoStorageAsTransactionalCacheInterface")
	}
	// node address info cache implementation cannot return error on rollback
	return txNodeAddressInfoCache.Begin()
}

// RollbackCacheTransaction to rollback all transactional precess from NodeAddressInfoStorage
func (nru *NodeAddressInfoService) RollbackCacheTransaction() error {
	txNodeAddressInfoCache, ok := nru.NodeAddressInfoStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastNodeAddressInfoStorageAsTransactionalCacheInterface")
	}
	return txNodeAddressInfoCache.Rollback()
}

// CommitCacheTransaction to commiut all transactional process from NodeAddressInfoStorage
func (nru *NodeAddressInfoService) CommitCacheTransaction() error {
	txNodeAddressInfoCache, ok := nru.NodeAddressInfoStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastNodeAddressInfoStorageAsTransactionalCacheInterface")
	}
	return txNodeAddressInfoCache.Commit()
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

// InsertOrUpdateAddressInfo updates or adds (in case new) a node address info record to db
func (nru *NodeAddressInfoService) UpdateOrInsertAddressInfo(
	nodeAddressInfo *model.NodeAddressInfo,
	updatedStatus model.NodeAddressStatus,
) (updated bool, err error) {
	var (
		addressAlreadyUpdated bool
		nodeAddressesInfo     []*model.NodeAddressInfo
	)
	// validate first
	addressAlreadyUpdated, err = nru.ValidateNodeAddressInfo(nodeAddressInfo)
	if err != nil || addressAlreadyUpdated {
		return addressAlreadyUpdated, err
	}

	nodeAddressInfo.Status = updatedStatus
	// if a node with same id and status already exist, update
	if nodeAddressesInfo, err = nru.GetAddressInfoByNodeID(
		nodeAddressInfo.NodeID,
		[]model.NodeAddressStatus{nodeAddressInfo.Status},
	); err != nil {
		return false, err
	}
	if len(nodeAddressesInfo) > 0 {
		// check if new address info is more recent than previous
		if nodeAddressInfo.GetBlockHeight() <= nodeAddressesInfo[0].GetBlockHeight() {
			return false, nil
		}
		err = nru.UpdateAddrressInfo(nodeAddressInfo)
		if err != nil {
			nru.Logger.Errorf("UpdateAddressInfoFail - %v - nodeAddressInfo:%d-%s:%d\n",
				err, nodeAddressInfo.NodeID, nodeAddressInfo.Address, nodeAddressInfo.Port)
			return false, err
		}
		return true, nil
	}
	err = nru.InsertAddressInfo(nodeAddressInfo)
	if err != nil {
		return false, err
	}
	if monitoring.IsMonitoringActive() {
		if countResult, err := nru.CountRegistredNodeAddressWithAddressInfo(); err == nil {
			monitoring.SetNodeAddressInfoCount(countResult)
		}
		if cna, err := nru.CountNodesAddressByStatus(); err == nil {
			for status, counter := range cna {
				monitoring.SetNodeAddressStatusCount(counter, status)
			}
		}
	}
	return true, nil
}

// ValidateNodeAddressInfo validate message data against:
// - main blocks: block height and hash
// - node registry: nodeID and message signature (use node public key in registry to validate the signature)
// Validation also fails if there is already a nodeAddressInfo record in db with same nodeID, address, port
func (nru *NodeAddressInfoService) ValidateNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) (found bool, err error) {
	var (
		nodeRegistry      storage.NodeRegistry
		nodeAddressesInfo []*model.NodeAddressInfo
	)
	err = nru.ActiveNodeRegistryCache.GetItem(nodeAddressInfo.GetNodeID(), &nodeRegistry)
	if err != nil {
		return false, err
	}

	if nodeAddressesInfo, err = nru.GetAddressInfoByNodeID(
		nodeAddressInfo.GetNodeID(),
		[]model.NodeAddressStatus{
			model.NodeAddressStatus_NodeAddressConfirmed,
			model.NodeAddressStatus_NodeAddressPending},
	); err != nil {
		return
	}

	for _, nai := range nodeAddressesInfo {
		if nodeAddressInfo.GetAddress() == nai.GetAddress() &&
			nodeAddressInfo.GetPort() == nai.GetPort() {
			// in case address for this node exists
			found = true
			return
		}
		if nai.GetStatus() == model.NodeAddressStatus_NodeAddressPending && nai.BlockHeight >= nodeAddressInfo.BlockHeight {
			found = true
			err = blocker.NewBlocker(blocker.ValidationErr, "OutdatedNodeAddressInfo")
			return
		}
	}
	var block *storage.BlockCacheObject
	block, err = util.GetBlockByHeightUseBlocksCache(
		nodeAddressInfo.GetBlockHeight(),
		nru.QueryExecutor,
		nru.BlockQuery,
		nru.MainBlocksStorage,
	)
	if err != nil {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHeight: "+err.Error())
		return
	}
	// validate block hash
	if !bytes.Equal(nodeAddressInfo.GetBlockHash(), block.BlockHash) {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHash")
		return
	}

	// validate the message signature
	unsignedBytes := nru.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	if !nru.Signature.VerifyNodeSignature(
		unsignedBytes,
		nodeAddressInfo.GetSignature(),
		nodeRegistry.Node.GetNodePublicKey(),
	) {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
		return
	}

	return false, nil
}
