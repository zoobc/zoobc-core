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
		GetRegisteredNodes() ([]*model.NodeRegistration, error)
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
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor               query.ExecutorInterface
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage    storage.CacheStorageInterface
		Logger                      *log.Logger
		BlockchainStatusService     BlockchainStatusServiceInterface
		CurrentNodePublicKey        []byte
		NodeAddressInfoService      NodeAddressInfoServiceInterface
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
	nextNodeAdmissionStorage storage.CacheStorageInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:               queryExecutor,
		AccountBalanceQuery:         accountBalanceQuery,
		NodeRegistrationQuery:       nodeRegistrationQuery,
		ParticipationScoreQuery:     participationScoreQuery,
		Logger:                      logger,
		BlockchainStatusService:     blockchainStatusService,
		NodeAddressInfoService:      nodeAddressInfoService,
		NodeAdmissionTimestampQuery: nodeAdmissionTimestampQuery,
		NextNodeAdmissionStorage:    nextNodeAdmissionStorage,
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
