package service

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
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
		AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		GetNextNodeAdmissionTimestamp(blockHeight uint32) (int64, error)
		InsertNextNodeAdmissionTimestamp(lastAdmissionTimestamp int64, blockHeight uint32, dbTx bool) error
		BuildScrambledNodes(block *model.Block) error
		ResetScrambledNodes()
		GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32
		GetScrambleNodesByHeight(
			blockHeight uint32,
		) (*model.ScrambledNodes, error)
		AddParticipationScore(nodeID, scoreDelta int64, height uint32, dbTx bool) (newScore int64, err error)
		SetCurrentNodePublicKey(publicKey []byte)
		GetNodeAddressesInfoFromDb(
			nodeIDs []int64,
			addressStatuses []model.NodeAddressStatus,
		) ([]*model.NodeAddressInfo, error)
		GetNodeAddressInfoFromDbByAddressPort(
			address string,
			port uint32,
			nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error)
		GenerateNodeAddressInfo(
			nodeID int64,
			nodeAddress string,
			port uint32,
			nodeSecretPhrase string) (*model.NodeAddressInfo, error)
		UpdateNodeAddressInfo(
			nodeAddressInfo *model.NodeAddressInfo,
			updatedStatus model.NodeAddressStatus,
		) (updated bool, err error)
		DeletePendingNodeAddressInfo(nodeID int64) error
		ValidateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (found bool, err error)
		ConfirmPendingNodeAddress(pendingNodeAddressInfo *model.NodeAddressInfo) error
		CountNodesAddressByStatus() (map[model.NodeAddressStatus]int, error)
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor                query.ExecutorInterface
		NodeAddressInfoQuery         query.NodeAddressInfoQueryInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmissionTimestampQuery  query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmission            *model.NodeAdmissionTimestamp
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
	nodeAddressInfoQuery query.NodeAddressInfoQueryInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	blockQuery query.BlockQueryInterface,
	nodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface,
	logger *log.Logger,
	blockchainStatusService BlockchainStatusServiceInterface,
	signature crypto.SignatureInterface,
	nodeAddressInfoService NodeAddressInfoServiceInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:               queryExecutor,
		NodeAddressInfoQuery:        nodeAddressInfoQuery,
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

// CountNodesAddressByStatus return a map with a count of nodes addresses in db for every node address status
func (nrs *NodeRegistrationService) CountNodesAddressByStatus() (map[model.NodeAddressStatus]int, error) {
	qry := nrs.NodeAddressInfoQuery.GetNodeAddressInfo()
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeAddressesInfo, err := nrs.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		return nil, err
	}

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
	return nil, nil
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
	}
	return nil
}

// GetNextNodeAdmissionTimestamp get the next node admission timestamp
func (nrs *NodeRegistrationService) GetNextNodeAdmissionTimestamp(blockHeight uint32) (int64, error) {
	if nrs.NextNodeAdmission == nil || blockHeight <= nrs.NextNodeAdmission.BlockHeight {
		var (
			err               error
			row               *sql.Row
			nextNodeAdmission model.NodeAdmissionTimestamp
		)
		row, err = nrs.QueryExecutor.ExecuteSelectRow(
			nrs.NodeAdmissionTimestampQuery.GetNextNodeAdmision(),
			false,
		)
		if err != nil {
			return 0, err
		}
		err = nrs.NodeAdmissionTimestampQuery.Scan(&nextNodeAdmission, row)
		if err != nil {
			return 0, err
		}
		nrs.NextNodeAdmission = &nextNodeAdmission
	}
	return nrs.NextNodeAdmission.Timestamp, nil
}

// InsertNextNodeAdmissionTimestamp set new next node admission timestamp
func (nrs *NodeRegistrationService) InsertNextNodeAdmissionTimestamp(
	lastAdmissionTimestamp int64,
	blockHeight uint32,
	dbTx bool,
) error {
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
		return err
	}
	activeBlocksmiths, err = nrs.NodeRegistrationQuery.BuildBlocksmith(activeBlocksmiths, rows)
	if err != nil {
		return err
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
		return err
	}

	nrs.NextNodeAdmission = nextNodeAdmission
	return nil
}

func (nrs *NodeRegistrationService) BuildScrambledNodesAtHeight(blockHeight uint32) error {
	var (
		nearestBlock model.Block
		err          error
	)
	nearestHeight := nrs.GetBlockHeightToBuildScrambleNodes(blockHeight)
	nearestBlockRow, _ := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(nearestHeight), false)
	err = nrs.BlockQuery.Scan(&nearestBlock, nearestBlockRow)
	if err != nil {
		return err
	}
	return nrs.sortNodeRegistries(&nearestBlock)
}

// sortNodeRegistries this function is responsible of selecting and sorting registered nodes so that nodes/peers in scrambledNodes map changes
// order at a given interval
// note: this algorithm is deterministic for the whole network so that,
// at any point in time every node can calculate this map autonomously, given its node registry is updated
func (nrs *NodeRegistrationService) sortNodeRegistries(
	block *model.Block,
) error {
	var (
		nodeRegistries  []*model.NodeRegistration
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
		err             error
	)

	// get node registry list
	rows, err := nrs.QueryExecutor.ExecuteSelect(
		nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.GetHeight()),
		false,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	nodeRegistries, err = nrs.NodeRegistrationQuery.BuildModel(nodeRegistries, rows)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	// sort node registry
	sort.SliceStable(nodeRegistries, func(i, j int) bool {
		ni, nj := nodeRegistries[i], nodeRegistries[j]

		// Get Hash of joined  with block seed & node ID
		// TODO : Enhance, to precomputing the hash/bigInt before sorting
		// 		  to avoid repeated hash computation while sorting
		hashI := sha3.Sum256(append(block.GetBlockSeed(), byte(ni.GetNodeID())))
		hashJ := sha3.Sum256(append(block.GetBlockSeed(), byte(nj.GetNodeID())))
		resI := new(big.Int).SetBytes(hashI[:])
		resJ := new(big.Int).SetBytes(hashJ[:])

		res := resI.Cmp(resJ)
		// Ascending sort
		return res < 0
	})
	// Restructure & validating node address
	for key, node := range nodeRegistries {
		nai, err := nrs.NodeAddressInfoService.GetAddressInfoByNodeID(node.GetNodeID(), model.NodeAddressStatus_NodeAddressPending)
		if err != nil {
			return err
		}
		peer := &model.Peer{
			Info: &model.Node{
				ID: node.GetNodeID(),
			},
		}
		// p2p: add peer to index and address nodes only if node has address
		scrambleDNodeMapKey := fmt.Sprintf("%d", node.GetNodeID())
		if nai != nil {
			peer.Info.Address = nai.GetAddress()
			peer.Info.Port = nai.GetPort()
			peer.Info.SharedAddress = nai.GetAddress()
			peer.Info.AddressStatus = nai.GetStatus()
		}
		index := key
		newIndexNodes[scrambleDNodeMapKey] = &index
		newAddressNodes = append(newAddressNodes, peer)
	}

	// build the scrambled node map
	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	// memoize the scrambled nodes
	nrs.ScrambledNodes[block.Height] = &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  block.Height,
	}
	// STEF temporary monitoring parameter
	// computing the hash of scrambled nodes and extracting 1st 8 bytes into an int64 (little endian)
	var digest = sha3.New256()
	for _, sn := range nrs.ScrambledNodes[block.Height].AddressNodes {
		if _, err := digest.Write(commonUtils.ConvertUint64ToBytes(uint64(sn.GetInfo().GetID()))); err != nil {
			break
		}
	}
	scrambledHash := binary.LittleEndian.Uint64(digest.Sum([]byte{}))
	monitoring.SetScrambledNodes(int64(scrambledHash))
	// var h = new(codec.CborHandle)
	// var b = make([]byte, 0)
	// enc := codec.NewEncoderBytes(&b, h)
	// if err = enc.Encode(nrs.ScrambledNodes[block.Height].AddressNodes); err == nil {
	// 	hash := sha3.Sum256(b)
	// 	scrambledHash := binary.LittleEndian.Uint64(hash[:])
	// 	monitoring.SetScrambledNodes(int64(scrambledHash))
	// }

	return nil
}

// BuildScrambleNodes,  build sorted scramble nodes based on node registry
func (nrs *NodeRegistrationService) BuildScrambledNodes(block *model.Block) error {
	return nrs.sortNodeRegistries(block)
}

func (nrs *NodeRegistrationService) ResetScrambledNodes() {
	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	nrs.ScrambledNodes = map[uint32]*model.ScrambledNodes{}
}

func (nrs *NodeRegistrationService) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	var (
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
		err             error
	)
	nearestHeight := nrs.GetBlockHeightToBuildScrambleNodes(blockHeight)
	nrs.ScrambledNodesLock.RLock()
	scrambleNodeExist := nrs.ScrambledNodes[nearestHeight]
	nrs.ScrambledNodesLock.RUnlock()
	if scrambleNodeExist == nil || blockHeight < constant.ScrambleNodesSafeHeight {
		err = nrs.BuildScrambledNodesAtHeight(nearestHeight)
		if err != nil {
			return nil, err
		}
	}
	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	scrambledNodes := nrs.ScrambledNodes[nearestHeight]
	newAddressNodes = append(newAddressNodes, scrambledNodes.AddressNodes...)
	// in the window, deep copy the nodes
	for key, indexNode := range scrambledNodes.IndexNodes {
		tempVal := *indexNode
		newIndexNodes[key] = &tempVal
	}
	return &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  scrambledNodes.BlockHeight,
	}, nil
}

func (nrs *NodeRegistrationService) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight - (lastBlockHeight % constant.PriorityStrategyBuildScrambleNodesGap)
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
	var qry string
	if len(nodeIDs) > 0 {
		qry = nrs.NodeAddressInfoQuery.GetNodeAddressInfoByNodeIDs(nodeIDs, addressStatuses)
	} else {
		qry = nrs.NodeAddressInfoQuery.GetNodeAddressInfoByStatus(addressStatuses)
	}
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeAddressesInfo, err := nrs.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		return nil, err
	}

	return nodeAddressesInfo, nil
}

// GetNodeAddressInfoFromDbByAddressPort returns a node address info given and address and port pairs
func (nrs *NodeRegistrationService) GetNodeAddressInfoFromDbByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	qry, args := nrs.NodeAddressInfoQuery.GetNodeAddressInfoByAddressPort(address, port, nodeAddressStatuses)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeAddressesInfo, err := nrs.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return nodeAddressesInfo, nil
}

// UpdateNodeAddressInfo updates or adds (in case new) a node address info record to db
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
	if nodeAddressesInfo, err = nrs.GetNodeAddressesInfoFromDb(
		[]int64{nodeAddressInfo.NodeID},
		[]model.NodeAddressStatus{nodeAddressInfo.Status},
	); err != nil {
		return false, err
	}
	if len(nodeAddressesInfo) > 0 {
		// check if new address info is more recent than previous
		if nodeAddressInfo.GetBlockHeight() < nodeAddressesInfo[0].GetBlockHeight() {
			return false, nil
		}
		err = nrs.QueryExecutor.BeginTx()
		if err != nil {
			return false, err
		}
		qryArgs := nrs.NodeAddressInfoQuery.UpdateNodeAddressInfo(nodeAddressInfo)
		err = nrs.QueryExecutor.ExecuteTransactions(qryArgs)
		if err != nil {
			_ = nrs.QueryExecutor.RollbackTx()
			nrs.Logger.Error(err)
			return false, err
		}
		err = nrs.QueryExecutor.CommitTx()
		if err != nil {
			return false, err
		}
		return true, nil
	}

	err = nrs.QueryExecutor.BeginTx()
	if err != nil {
		return false, err
	}
	qry, args := nrs.NodeAddressInfoQuery.InsertNodeAddressInfo(nodeAddressInfo)
	err = nrs.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		_ = nrs.QueryExecutor.RollbackTx()
		nrs.Logger.Error(err)
		return false, err
	}
	err = nrs.QueryExecutor.CommitTx()
	if err != nil {
		return false, err
	}
	if monitoring.IsMonitoringActive() {
		if registeredNodesWithAddress, err := nrs.GetRegisteredNodesWithNodeAddress(); err == nil {
			monitoring.SetNodeAddressInfoCount(len(registeredNodesWithAddress))
		}
		if cna, err := nrs.CountNodesAddressByStatus(); err == nil {
			for status, counter := range cna {
				monitoring.SetNodeAddressStatusCount(counter, status)
			}
		}
	}
	return true, nil
}

func (nrs *NodeRegistrationService) DeletePendingNodeAddressInfo(nodeID int64) error {
	qry, args := nrs.NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID(
		nodeID,
		[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending})
	// start db transaction here
	err := nrs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	err = nrs.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		if rollbackErr := nrs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			nrs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	return nrs.QueryExecutor.CommitTx()
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

	if nodeAddressesInfo, err = nrs.GetNodeAddressesInfoFromDb([]int64{nodeAddressInfo.GetNodeID()},
		[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressConfirmed, model.NodeAddressStatus_NodeAddressPending}); err != nil {
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
		safeBlockHeight uint32
		safeBlock       model.Block
	)
	lastBlock, err := commonUtils.GetLastBlock(nrs.QueryExecutor, nrs.BlockQuery)
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
func (nrs *NodeRegistrationService) ConfirmPendingNodeAddress(pendingNodeAddressInfo *model.NodeAddressInfo) error {
	queries := nrs.NodeAddressInfoQuery.ConfirmNodeAddressInfo(pendingNodeAddressInfo)
	executor := nrs.QueryExecutor
	err := executor.BeginTx()
	if err != nil {
		return err
	}
	err = executor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := executor.RollbackTx()
		if rollbackErr != nil {
			log.Errorln(rollbackErr.Error())
		}
		return err
	}
	err = executor.CommitTx()
	if err != nil {
		return err
	}
	if monitoring.IsMonitoringActive() {
		if registeredNodesWithAddress, err := nrs.GetRegisteredNodesWithNodeAddress(); err == nil {
			monitoring.SetNodeAddressInfoCount(len(registeredNodesWithAddress))
		}
		if cna, err := nrs.CountNodesAddressByStatus(); err == nil {
			for status, counter := range cna {
				monitoring.SetNodeAddressStatusCount(counter, status)
			}
		}
	}

	return nil
}
