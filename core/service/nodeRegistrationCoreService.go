package service

import (
	"bytes"
	"database/sql"
	"math/big"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		SelectNodesToBeExpelled() ([]*model.NodeRegistration, error)
		GetRegisteredNodes() ([]*model.NodeRegistration, error)
		GetRegisteredNodesWithNodeAddress() ([]*model.NodeRegistration, error)
		GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error)
		GetNodeRegistrationByNodeID(nodeID int64) (*model.NodeRegistration, error)
		GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error)
		AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		GetNextNodeAdmissionTimestamp() (*model.NodeAdmissionTimestamp, error)
		InsertNextNodeAdmissionTimestamp(
			lastAdmissionTimestamp int64, blockHeight uint32, dbTx bool,
		) (*model.NodeAdmissionTimestamp, error)
		UpdateNextNodeAdmissionCache(newNextNodeAdmission *model.NodeAdmissionTimestamp) error
		AddParticipationScore(nodeID, scoreDelta int64, height uint32, dbTx bool) (newScore int64, err error)
		SetCurrentNodePublicKey(publicKey []byte)
		GenerateNodeAddressInfo(
			nodeID int64,
			nodeAddress string,
			port uint32,
			nodeSecretPhrase string) (*model.NodeAddressInfo, error)
		UpdateNodeAddressInfo(
			nodeAddressInfo *model.NodeAddressInfo,
			updatedStatus model.NodeAddressStatus,
		) (updated bool, err error)
		ValidateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (found bool, err error)
		ConfirmPendingNodeAddress(pendingNodeAddressInfo *model.NodeAddressInfo) error
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmissionTimestampQuery  query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage     storage.CacheStorageInterface
		MainBlockStateStorage        storage.CacheStorageInterface
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		ScrambledNodesLock           sync.RWMutex
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
		Signature                    crypto.SignatureInterface
		NodeAddressInfoService       NodeAddressInfoServiceInterface
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	blockQuery query.BlockQueryInterface,
	nodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface,
	logger *log.Logger,
	blockchainStatusService BlockchainStatusServiceInterface,
	signature crypto.SignatureInterface,
	nodeAddressInfoService NodeAddressInfoServiceInterface,
	nextNodeAdmissionStorage, mainBlockStateStorage storage.CacheStorageInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:               queryExecutor,
		AccountBalanceQuery:         accountBalanceQuery,
		NodeRegistrationQuery:       nodeRegistrationQuery,
		ParticipationScoreQuery:     participationScoreQuery,
		BlockQuery:                  blockQuery,
		Logger:                      logger,
		ScrambledNodes:              map[uint32]*model.ScrambledNodes{},
		BlockchainStatusService:     blockchainStatusService,
		Signature:                   signature,
		NodeAddressInfoService:      nodeAddressInfoService,
		NodeAdmissionTimestampQuery: nodeAdmissionTimestampQuery,
		NextNodeAdmissionStorage:    nextNodeAdmissionStorage,
		MainBlockStateStorage:       mainBlockStateStorage,
	}
}

// SelectNodesToBeAdmitted Select n (=limit) queued nodes with the highest locked balance
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(limit, model.NodeRegistrationState_NodeQueued)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return nodeRegistrations, nil
}

// SelectNodesToBeExpelled Select n (=limit) registered nodes with participation score = 0
func (nrs *NodeRegistrationService) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsWithZeroScore(model.NodeRegistrationState_NodeRegistered)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return nodeRegistrations, nil
}

func (nrs *NodeRegistrationService) GetRegisteredNodes() ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetActiveNodeRegistrations()
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistry, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}

	return nodeRegistry, nil
}

func (nrs *NodeRegistrationService) GetRegisteredNodesWithNodeAddress() ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetActiveNodeRegistrationsWithNodeAddress()
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistry, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}

	return nodeRegistry, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error) {
	rows, err := nrs.QueryExecutor.ExecuteSelect(nrs.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, nodePublicKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}

	if len(nodeRegistrations) > 0 {
		return nodeRegistrations[0], nil
	}
	return nil, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistrationByNodeID(nodeID int64) (*model.NodeRegistration, error) {
	qry, args := nrs.NodeRegistrationQuery.GetNodeRegistrationByID(nodeID)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}

	if len(nodeRegistrations) > 0 {
		return nodeRegistrations[0], nil
	}
	return nil, blocker.NewBlocker(blocker.DBErr, "noNodeRegistrationFound")
}

// AdmitNodes update given node registrations' registrationStatus field to NodeRegistrationState_NodeRegistered (=0)
// and set default participation score to it
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	// prepare all node registrations to be updated (set registrationStatus to NodeRegistrationState_NodeRegistered and new height)
	// and default participation scores to be added
	for _, nodeRegistration := range nodeRegistrations {
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeRegistered)
		nodeRegistration.Height = height
		// update the node registry (set registrationStatus to zero)
		queries := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		// add default participation score to the node
		updateParticipationScoreQuery := nrs.ParticipationScoreQuery.UpdateParticipationScore(
			nodeRegistration.NodeID,
			constant.DefaultParticipationScore,
			height)
		queries = append(queries, updateParticipationScoreQuery...)
		if err := nrs.QueryExecutor.ExecuteTransactions(queries); err != nil {
			return err
		}
		if bytes.Equal(nrs.CurrentNodePublicKey, nodeRegistration.NodePublicKey) {
			nrs.BlockchainStatusService.SetIsBlocksmith(true)
		}
	}
	return nil
}

// ExpelNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting registrationStatus field to 3 (deleted) and locked balance to zero
func (nrs *NodeRegistrationService) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	for _, nodeRegistration := range nodeRegistrations {
		// update the node registry (set registrationStatus to 1 and lockedbalance to 0)
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)
		nodeRegistration.LockedBalance = 0
		nodeRegistration.Height = height
		nodeQueries := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		// return lockedbalance to the node's owner account
		updateAccountBalanceQ := nrs.AccountBalanceQuery.AddAccountBalance(
			nodeRegistration.LockedBalance,
			map[string]interface{}{
				"account_address": nodeRegistration.AccountAddress,
				"block_height":    height,
			},
		)
		queries := append(updateAccountBalanceQ, nodeQueries...)
		if err := nrs.QueryExecutor.ExecuteTransactions(queries); err != nil {
			return err
		}
		// remove the node_address_info
		err := nrs.NodeAddressInfoService.DeleteNodeAddressInfoByNodeIDInDBTx(nodeRegistration.NodeID)
		if err != nil {
			return err
		}

	}
	return nil
}

// GetNextNodeAdmissionTimestamp get the next node admission timestamp
func (nrs *NodeRegistrationService) GetNextNodeAdmissionTimestamp() (*model.NodeAdmissionTimestamp, error) {
	var (
		nextNodeAdmission model.NodeAdmissionTimestamp
		err               = nrs.NextNodeAdmissionStorage.GetItem(nil, &nextNodeAdmission)
	)
	if err != nil {
		return nil, err
	}
	return &nextNodeAdmission, nil
}

// InsertNextNodeAdmissionTimestamp set new next node admission timestamp
func (nrs *NodeRegistrationService) InsertNextNodeAdmissionTimestamp(
	lastAdmissionTimestamp int64,
	blockHeight uint32,
	dbTx bool,
) (*model.NodeAdmissionTimestamp, error) {
	var (
		rows              *sql.Rows
		err               error
		delayAdmission    int64
		nextNodeAdmission *model.NodeAdmissionTimestamp
		activeBlocksmiths []*model.Blocksmith
		insertQueries     [][]interface{}
	)

	// get all registered nodes
	rows, err = nrs.QueryExecutor.ExecuteSelect(
		nrs.NodeRegistrationQuery.GetActiveNodeRegistrationsByHeight(blockHeight),
		dbTx,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activeBlocksmiths, err = nrs.NodeRegistrationQuery.BuildBlocksmith(activeBlocksmiths, rows)
	if err != nil {
		return nil, err
	}
	// calculate next delay node admission timestamp
	delayAdmission = constant.NodeAdmissionBaseDelay / int64(len(activeBlocksmiths))
	delayAdmission = commonUtils.MinInt64(
		commonUtils.MaxInt64(delayAdmission, constant.NodeAdmissionMinDelay),
		constant.NodeAdmissionMaxDelay,
	)
	nextNodeAdmission = &model.NodeAdmissionTimestamp{
		Timestamp:   lastAdmissionTimestamp + delayAdmission,
		BlockHeight: blockHeight,
		Latest:      true,
	}
	insertQueries = nrs.NodeAdmissionTimestampQuery.InsertNextNodeAdmission(nextNodeAdmission)
	err = nrs.QueryExecutor.ExecuteTransactions(insertQueries)
	if err != nil {
		return nil, err
	}
	return nextNodeAdmission, nil
}

func (nrs *NodeRegistrationService) UpdateNextNodeAdmissionCache(newNextNodeAdmission *model.NodeAdmissionTimestamp) error {
	var (
		err               error
		row               *sql.Row
		nextNodeAdmission model.NodeAdmissionTimestamp
	)
	if newNextNodeAdmission != nil {
		err = nrs.NextNodeAdmissionStorage.SetItem(nil, *newNextNodeAdmission)
		if err != nil {
			return err
		}
		return nil
	}
	// get next node admission from DB
	row, err = nrs.QueryExecutor.ExecuteSelectRow(
		nrs.NodeAdmissionTimestampQuery.GetNextNodeAdmision(),
		false,
	)
	if err != nil {
		return err
	}
	err = nrs.NodeAdmissionTimestampQuery.Scan(&nextNodeAdmission, row)
	if err != nil {
		return err
	}
	// update next node admission timestamp storage
	err = nrs.NextNodeAdmissionStorage.SetItem(nil, nextNodeAdmission)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeRegistryAtHeight get active node registry list at the given height
func (nrs *NodeRegistrationService) GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error) {
	var (
		result []*model.NodeRegistration
		err    error
	)

	rows, err := nrs.QueryExecutor.ExecuteSelect(
		nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(height),
		false,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	return nrs.NodeRegistrationQuery.BuildModel(result, rows)
}

// AddParticipationScore updates a node's participation score by increment/deincrement a previous score by a given number
func (nrs *NodeRegistrationService) AddParticipationScore(nodeID, scoreDelta int64, height uint32, dbTx bool) (newScore int64, err error) {
	var (
		ps model.ParticipationScore
	)
	qry, args := nrs.ParticipationScoreQuery.GetParticipationScoreByNodeID(nodeID)
	row, err := nrs.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, blocker.NewBlocker(blocker.DBErr, "ParticipationScoreNotFound")
	}
	err = nrs.ParticipationScoreQuery.Scan(&ps, row)
	if err != nil {
		return 0, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// don't update the score if already max allowed
	if ps.Score >= constant.MaxParticipationScore && scoreDelta > 0 {
		nrs.Logger.Debugf("Node id %d: score is already the maximum allowed and won't be increased", nodeID)
		return constant.MaxParticipationScore, nil
	}
	if ps.Score <= 0 && scoreDelta < 0 {
		nrs.Logger.Debugf("Node id %d: score is already 0. new score won't be decreased", nodeID)
		return 0, nil
	}
	// check if updating the score will overflow the max score and if so, set the new score to max allowed
	// note: we use big integers to make sure we manage the very unlikely case where the addition overflows max int64
	scoreDeltaBig := big.NewInt(scoreDelta)
	prevScoreBig := big.NewInt(ps.Score)
	maxScoreBig := big.NewInt(constant.MaxParticipationScore)
	newScoreBig := new(big.Int).Add(prevScoreBig, scoreDeltaBig)
	if newScoreBig.Cmp(maxScoreBig) > 0 {
		newScore = constant.MaxParticipationScore
	} else if newScoreBig.Cmp(big.NewInt(0)) < 0 {
		newScore = 0
	} else {
		newScore = ps.Score + scoreDelta
	}

	// finally update the participation score
	updateParticipationScoreQuery := nrs.ParticipationScoreQuery.UpdateParticipationScore(nodeID, newScore, height)
	err = nrs.QueryExecutor.ExecuteTransactions(updateParticipationScoreQuery)
	return newScore, err
}

// SetCurrentNodePublicKey set the public key of running node, this information will be used to check if current node is
// being admitted and can start unlock smithing process
func (nrs *NodeRegistrationService) SetCurrentNodePublicKey(publicKey []byte) {
	nrs.CurrentNodePublicKey = publicKey
}

// GetNodeAddressesInfoFromDb returns a list of node address info messages given a list of nodeIDs and address statuses
func (nrs *NodeRegistrationService) GetNodeAddressesInfoFromDb(
	nodeIDs []int64,
	addressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	var nodeAddressesInfo []*model.NodeAddressInfo
	var err error
	if len(nodeIDs) > 0 {
		nodeAddressesInfo, err = nrs.NodeAddressInfoService.GetAddressInfoByNodeIDs(nodeIDs, addressStatuses)
		if err != nil {
			return nil, err
		}
	} else {
		nodeAddressesInfo, err = nrs.NodeAddressInfoService.GetAddressInfoByStatus(addressStatuses)
		if err != nil {
			return nil, err
		}
	}
	return nodeAddressesInfo, nil
}

// UpdateNodeAddressInfo updates or adds (in case new) a node address info record to db
// TODO @sukrawidhyawan: will completely move this function into node address info service
// after node address info cache stable
func (nrs *NodeRegistrationService) UpdateNodeAddressInfo(
	nodeAddressInfo *model.NodeAddressInfo,
	updatedStatus model.NodeAddressStatus,
) (updated bool, err error) {
	var (
		addressAlreadyUpdated bool
		nodeAddressesInfo     []*model.NodeAddressInfo
	)
	// validate first
	addressAlreadyUpdated, err = nrs.ValidateNodeAddressInfo(nodeAddressInfo)
	if err != nil || addressAlreadyUpdated {
		return false, err
	}

	nodeAddressInfo.Status = updatedStatus
	// if a node with same id and status already exist, update
	if nodeAddressesInfo, err = nrs.NodeAddressInfoService.GetAddressInfoByNodeID(
		nodeAddressInfo.NodeID,
		[]model.NodeAddressStatus{nodeAddressInfo.Status},
	); err != nil {
		return false, err
	}
	if len(nodeAddressesInfo) > 0 {
		// check if new address info is more recent than previous
		if nodeAddressInfo.GetBlockHeight() < nodeAddressesInfo[0].GetBlockHeight() {
			return false, nil
		}
		err = nrs.NodeAddressInfoService.UpdateAddrressInfo(nodeAddressInfo)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	err = nrs.NodeAddressInfoService.InsertAddressInfo(nodeAddressInfo)
	if err != nil {
		return false, err
	}
	if monitoring.IsMonitoringActive() {
		if registeredNodesWithAddress, err := nrs.GetRegisteredNodesWithNodeAddress(); err == nil {
			monitoring.SetNodeAddressInfoCount(len(registeredNodesWithAddress))
		}
		if cna, err := nrs.NodeAddressInfoService.CountNodesAddressByStatus(); err == nil {
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
func (nrs *NodeRegistrationService) ValidateNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) (found bool, err error) {
	var (
		block             model.Block
		nodeRegistration  model.NodeRegistration
		nodeAddressesInfo []*model.NodeAddressInfo
	)

	// validate nodeID
	qry, args := nrs.NodeRegistrationQuery.GetNodeRegistrationByID(nodeAddressInfo.GetNodeID())
	row, _ := nrs.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	err = nrs.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err == sql.ErrNoRows {
			err = blocker.NewBlocker(blocker.ValidationErr, "NodeIDNotFound")
			return
		}
		return
	}

	// validate the message signature
	unsignedBytes := nrs.NodeAddressInfoService.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	if !nrs.Signature.VerifyNodeSignature(
		unsignedBytes,
		nodeAddressInfo.GetSignature(),
		nodeRegistration.GetNodePublicKey(),
	) {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
		return
	}

	// validate block height
	blockRow, _ := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(nodeAddressInfo.GetBlockHeight()), false)
	err = nrs.BlockQuery.Scan(&block, blockRow)
	if err != nil {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHeight")
		return
	}
	// validate block hash
	if !bytes.Equal(nodeAddressInfo.GetBlockHash(), block.GetBlockHash()) {
		err = blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHash")
		return
	}

	if nodeAddressesInfo, err = nrs.NodeAddressInfoService.GetAddressInfoByNodeID(
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
	return false, nil
}

// GenerateNodeAddressInfo generate a nodeAddressInfo signed message
func (nrs *NodeRegistrationService) GenerateNodeAddressInfo(
	nodeID int64,
	nodeAddress string,
	port uint32,
	nodeSecretPhrase string) (*model.NodeAddressInfo, error) {
	var (
		safeBlockHeight      uint32
		safeBlock, lastBlock model.Block
		err                  = nrs.MainBlockStateStorage.GetItem(nil, &lastBlock)
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
	rows, err := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(safeBlockHeight), false)
	if err != nil {
		return nil, err
	}
	err = nrs.BlockQuery.Scan(&safeBlock, rows)
	if err != nil {
		return nil, err
	}

	nodeAddressInfo := &model.NodeAddressInfo{
		NodeID:      nodeID,
		Address:     nodeAddress,
		Port:        port,
		BlockHeight: safeBlock.GetHeight(),
		BlockHash:   safeBlock.GetBlockHash(),
	}
	nodeAddressInfoBytes := nrs.NodeAddressInfoService.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	nodeAddressInfo.Signature = nrs.Signature.SignByNode(nodeAddressInfoBytes, nodeSecretPhrase)
	return nodeAddressInfo, nil
}

// ConfirmPendingNodeAddress confirm a pending address by inserting or replacing the previously confirmed one and deleting the pending address
// TODO @sukrawidhyawan: will completely move this function into node address info service
// after node address info cache stable
func (nrs *NodeRegistrationService) ConfirmPendingNodeAddress(pendingNodeAddressInfo *model.NodeAddressInfo) error {
	var err = nrs.NodeAddressInfoService.ConfirmNodeAddressInfo(pendingNodeAddressInfo)
	if err != nil {
		return err
	}
	if monitoring.IsMonitoringActive() {
		if registeredNodesWithAddress, err := nrs.GetRegisteredNodesWithNodeAddress(); err == nil {
			monitoring.SetNodeAddressInfoCount(len(registeredNodesWithAddress))
		}
		if cna, err := nrs.NodeAddressInfoService.CountNodesAddressByStatus(); err == nil {
			for status, counter := range cna {
				monitoring.SetNodeAddressStatusCount(counter, status)
			}
		}
	}

	return nil
}
