package service

import (
	"bytes"
	"database/sql"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		SelectNodesToBeExpelled() ([]*model.NodeRegistration, error)
		GetActiveRegisteredNodes() ([]*model.NodeRegistration, error)
		GetActiveRegistry() ([]storage.NodeRegistry, float64, error)
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
		// cache controllers
		InitializeCache() error
		BeginCacheTransaction() error
		RollbackCacheTransaction() error
		CommitCacheTransaction() error
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor                   query.ExecutorInterface
		AccountBalanceQuery             query.AccountBalanceQueryInterface
		NodeRegistrationQuery           query.NodeRegistrationQueryInterface
		ParticipationScoreQuery         query.ParticipationScoreQueryInterface
		NodeAdmissionTimestampQuery     query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage        storage.CacheStorageInterface
		ActiveNodeRegistryCacheStorage  storage.CacheStorageInterface
		PendingNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                          *log.Logger
		BlockchainStatusService         BlockchainStatusServiceInterface
		CurrentNodePublicKey            []byte
		NodeAddressInfoService          NodeAddressInfoServiceInterface
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	nodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface,
	logger *log.Logger,
	blockchainStatusService BlockchainStatusServiceInterface,
	nodeAddressInfoService NodeAddressInfoServiceInterface,
	nextNodeAdmissionStorage, activeNodeRegistryCacheStorage,
	pendingNodeRegistryCache storage.CacheStorageInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:                   queryExecutor,
		AccountBalanceQuery:             accountBalanceQuery,
		NodeRegistrationQuery:           nodeRegistrationQuery,
		ParticipationScoreQuery:         participationScoreQuery,
		Logger:                          logger,
		BlockchainStatusService:         blockchainStatusService,
		NodeAddressInfoService:          nodeAddressInfoService,
		NodeAdmissionTimestampQuery:     nodeAdmissionTimestampQuery,
		NextNodeAdmissionStorage:        nextNodeAdmissionStorage,
		ActiveNodeRegistryCacheStorage:  activeNodeRegistryCacheStorage,
		PendingNodeRegistryCacheStorage: pendingNodeRegistryCache,
	}
}

// InitializeCache prefill the node registry cache with latest state from database
func (nrs *NodeRegistrationService) InitializeCache() error {
	var (
		pendingQry, activeQry                             string
		cachePendingNodeRegistry, cacheActiveNodeRegistry []storage.NodeRegistry
	)
	err := nrs.PendingNodeRegistryCacheStorage.ClearCache()
	if err != nil {
		return err
	}
	err = nrs.ActiveNodeRegistryCacheStorage.ClearCache()
	if err != nil {
		return err
	}
	// pending
	pendingQry = nrs.NodeRegistrationQuery.GetAllNodeRegistryByStatus(model.NodeRegistrationState_NodeQueued) // limit = 0 get all records
	pendingNodeRegistryRows, err := nrs.QueryExecutor.ExecuteSelect(pendingQry, false)
	if err != nil {
		return err
	}
	defer pendingNodeRegistryRows.Close()
	cachePendingNodeRegistry, err = nrs.NodeRegistrationQuery.BuildModelWithParticipationScore(
		cachePendingNodeRegistry, pendingNodeRegistryRows)
	if err != nil {
		return err
	}
	// active
	activeQry = nrs.NodeRegistrationQuery.GetAllNodeRegistryByStatus(model.NodeRegistrationState_NodeRegistered) // limit = 0 get all records
	activeNodeRegistrationRows, err := nrs.QueryExecutor.ExecuteSelect(activeQry, false)
	if err != nil {
		return err
	}
	defer activeNodeRegistrationRows.Close()
	cacheActiveNodeRegistry, err = nrs.NodeRegistrationQuery.BuildModelWithParticipationScore(
		cacheActiveNodeRegistry, activeNodeRegistrationRows,
	)
	if err != nil {
		return err
	}
	err = nrs.PendingNodeRegistryCacheStorage.SetItems(cachePendingNodeRegistry)
	if err != nil {
		return err
	}
	err = nrs.ActiveNodeRegistryCacheStorage.SetItems(cacheActiveNodeRegistry)
	return err
}

// SelectNodesToBeAdmitted Select n (=limit) queued nodes with the highest locked balance
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	var (
		pendingNodeRegistries  = make([]storage.NodeRegistry, 0)
		selectedNodeRegistries = make([]*model.NodeRegistration, 0)
		err                    error
	)
	// get all pending registry (sorted by locked balance highest to lowest already)
	err = nrs.PendingNodeRegistryCacheStorage.GetAllItems(&pendingNodeRegistries)
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(limit) && i < len(pendingNodeRegistries); i++ {
		selectedNodeRegistries = append(selectedNodeRegistries, &pendingNodeRegistries[i].Node)
	}
	return selectedNodeRegistries, nil
}

// SelectNodesToBeExpelled Select n (=limit) registered nodes with participation score = 0
func (nrs *NodeRegistrationService) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	var (
		activeNodeRegistry    = make([]storage.NodeRegistry, 0)
		zeroScoreNodeRegistry = make([]*model.NodeRegistration, 0)
		err                   error
	)
	err = nrs.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return nil, err
	}
	for _, registry := range activeNodeRegistry {
		if registry.ParticipationScore <= 0 {
			zeroScoreNodeRegistry = append(zeroScoreNodeRegistry, &registry.Node)
		}
	}
	return zeroScoreNodeRegistry, nil
}

func (nrs *NodeRegistrationService) GetActiveRegisteredNodes() ([]*model.NodeRegistration, error) {
	var (
		activeNodeRegistry []storage.NodeRegistry
		nodeRegistries     []*model.NodeRegistration
		err                error
	)
	err = nrs.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return nil, err
	}
	for _, registry := range activeNodeRegistry {
		nodeRegistries = append(nodeRegistries, &registry.Node)
	}
	return nodeRegistries, nil
}

func (nrs *NodeRegistrationService) GetActiveRegistry() ([]storage.NodeRegistry, float64, error) {
	var (
		activeNodeRegistry []storage.NodeRegistry
		err                error
	)
	err = nrs.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return nil, 0, err
	}

	scoreSum := float64(0)
	for _, registry := range activeNodeRegistry {
		scoreSum += float64(registry.ParticipationScore / constant.OneZBC)
	}
	return activeNodeRegistry, scoreSum, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error) {
	var (
		err          error
		row          *sql.Row
		nodeRegistry model.NodeRegistration
	)
	row, err = nrs.QueryExecutor.ExecuteSelectRow(nrs.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, nodePublicKey)
	if err != nil {
		return nil, err
	}

	err = nrs.NodeRegistrationQuery.Scan(&nodeRegistry, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, "noNodeRegistrationFound")
		}
		return nil, err
	}
	return &nodeRegistry, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistrationByNodeID(nodeID int64) (*model.NodeRegistration, error) {
	var (
		qry          string
		args         []interface{}
		err          error
		row          *sql.Row
		nodeRegistry model.NodeRegistration
	)
	qry, args = nrs.NodeRegistrationQuery.GetNodeRegistrationByID(nodeID)
	row, err = nrs.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	if err != nil {
		return nil, err
	}

	err = nrs.NodeRegistrationQuery.Scan(&nodeRegistry, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, "noNodeRegistrationFound")
		}
		return nil, err
	}

	return &nodeRegistry, err
}

// AdmitNodes update given node registrations' registrationStatus field to NodeRegistrationState_NodeRegistered (=0)
// and set default participation score to it
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	var (
		activeNodeRegistries = make([]storage.NodeRegistry, 0)
		pendingIDsToRemove   []int64
		err                  error
	)
	err = nrs.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistries)
	if err != nil {
		return err
	}
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
		// handle cache, remove from pending cache & add active cache
		pendingIDsToRemove = append(pendingIDsToRemove, nodeRegistration.NodeID)
		activeNodeRegistries = append(activeNodeRegistries, storage.NodeRegistry{
			Node:               *nodeRegistration,
			ParticipationScore: constant.DefaultParticipationScore,
		})

	}
	txActiveCache, ok := nrs.ActiveNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	txPendingCache, ok := nrs.PendingNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	// remove pending
	for _, id := range pendingIDsToRemove {
		// look up from updated pending node registry (temp) cache
		err = txPendingCache.TxRemoveItem(id)
		if err != nil {
			return err
		}
	}
	// update transactional cache state
	err = txActiveCache.TxSetItems(activeNodeRegistries)
	return err
}

// ExpelNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting registrationStatus field to 3 (deleted) and locked balance to zero
func (nrs *NodeRegistrationService) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	var (
		activeNodeIDsToRemove []int64
	)
	for _, nodeRegistration := range nodeRegistrations {
		// update the node registry (set registrationStatus to 1 and locked balance to 0)
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
		activeNodeIDsToRemove = append(activeNodeIDsToRemove, nodeRegistration.GetNodeID())
	}
	txActiveCache, ok := nrs.ActiveNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	for _, id := range activeNodeIDsToRemove {
		err := txActiveCache.TxRemoveItem(id)
		if err != nil {
			return err
		}
	}
	// no need to re-sort as the slicing of the cache will keep the order in place
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
		nodeRegistry storage.NodeRegistry
	)

	err = nrs.ActiveNodeRegistryCacheStorage.GetItem(nodeID, &nodeRegistry)
	if err != nil {
		return 0, blocker.NewBlocker(blocker.AppErr, "FailGetNodeRegistryFromCache")
	}
	// don't update the score if already max allowed
	if nodeRegistry.ParticipationScore >= constant.MaxParticipationScore && scoreDelta > 0 {
		nrs.Logger.Debugf("Node id %d: score is already the maximum allowed and won't be increased", nodeID)
		return constant.MaxParticipationScore, nil
	}
	if nodeRegistry.ParticipationScore <= 0 && scoreDelta < 0 {
		nrs.Logger.Debugf("Node id %d: score is already 0. new score won't be decreased", nodeID)
		return 0, nil
	}
	// check if updating the score will overflow the max score and if so, set the new score to max allowed
	// note: we use big integers to make sure we manage the very unlikely case where the addition overflows max int64
	scoreDeltaBig := big.NewInt(scoreDelta)
	prevScoreBig := big.NewInt(nodeRegistry.ParticipationScore)
	maxScoreBig := big.NewInt(constant.MaxParticipationScore)
	newScoreBig := new(big.Int).Add(prevScoreBig, scoreDeltaBig)
	if newScoreBig.Cmp(maxScoreBig) > 0 {
		newScore = constant.MaxParticipationScore
	} else if newScoreBig.Cmp(big.NewInt(0)) < 0 {
		newScore = 0
	} else {
		newScore = nodeRegistry.ParticipationScore + scoreDelta
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

func (nrs *NodeRegistrationService) BeginCacheTransaction() error {
	txActiveCache, ok := nrs.ActiveNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	txPendingCache, ok := nrs.PendingNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastPendingNodeRegistryAsTransactionalCacheInterface")
	}
	// node registration cache implementation cannot return error on rollback
	_ = txActiveCache.Begin()
	_ = txPendingCache.Begin()
	return nil
}

func (nrs *NodeRegistrationService) RollbackCacheTransaction() error {
	txActiveCache, ok := nrs.ActiveNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	txPendingCache, ok := nrs.PendingNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastPendingNodeRegistryAsTransactionalCacheInterface")
	}
	// node registration cache implementation cannot return error on rollback
	_ = txActiveCache.Rollback()
	_ = txPendingCache.Rollback()
	return nil
}

func (nrs *NodeRegistrationService) CommitCacheTransaction() error {
	txActiveCache, ok := nrs.ActiveNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastActiveNodeRegistryAsTransactionalCacheInterface")
	}
	txPendingCache, ok := nrs.PendingNodeRegistryCacheStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastPendingNodeRegistryAsTransactionalCacheInterface")
	}
	// node registration cache implementation cannot return error on commit
	_ = txActiveCache.Commit()
	_ = txPendingCache.Commit()
	return nil
}
