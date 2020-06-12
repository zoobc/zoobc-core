package service

import (
	"bytes"
	"database/sql"
	"math/big"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"golang.org/x/crypto/sha3"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		SelectNodesToBeExpelled() ([]*model.NodeRegistration, error)
		GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error)
		GetRegisteredNodes() ([]*model.NodeRegistration, error)
		GetRegisteredNodesWithNodeAddress() ([]*model.NodeRegistration, error)
		GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error)
		AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		GetNodeAdmittanceCycle() uint32
		BuildScrambledNodes(block *model.Block) error
		ResetScrambledNodes()
		GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32
		GetScrambleNodesByHeight(
			blockHeight uint32,
		) (*model.ScrambledNodes, error)
		AddParticipationScore(nodeID, scoreDelta int64, height uint32, dbTx bool) (newScore int64, err error)
		SetCurrentNodePublicKey(publicKey []byte)
		GetNodeAddressesInfoFromDb(nodeIDs []int64) ([]*model.NodeAddressInfo, error)
		GenerateNodeAddressInfo(
			nodeID int64,
			nodeAddress string,
			port uint32,
			nodeSecretPhrase string) (*model.NodeAddressInfo, error)
		UpdateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (updated bool, err error)
		ValidateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (found bool, err error)
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor                query.ExecutorInterface
		NodeAddressInfoQuery         query.NodeAddressInfoQueryInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmittanceCycle          uint32
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		ScrambledNodesLock           sync.RWMutex
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
		Signature                    crypto.SignatureInterface
		NodeRegistrationUtils        NodeRegistrationUtilsInterface
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	nodeAddressInfoQuery query.NodeAddressInfoQueryInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	blockQuery query.BlockQueryInterface,
	logger *log.Logger,
	blockchainStatusService BlockchainStatusServiceInterface,
	signature crypto.SignatureInterface,
	nodeRegistrationUtils NodeRegistrationUtilsInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:           queryExecutor,
		NodeAddressInfoQuery:    nodeAddressInfoQuery,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		ParticipationScoreQuery: participationScoreQuery,
		BlockQuery:              blockQuery,
		NodeAdmittanceCycle:     constant.NodeAdmittanceCycle,
		Logger:                  logger,
		ScrambledNodes:          map[uint32]*model.ScrambledNodes{},
		BlockchainStatusService: blockchainStatusService,
		Signature:               signature,
		NodeRegistrationUtils:   nodeRegistrationUtils,
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

func (nrs *NodeRegistrationService) GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistryAtHeightWithNodeAddress(height)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
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
	if (err != nil) || len(nodeRegistry) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
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
	if (err != nil) || len(nodeRegistry) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
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
	if (err != nil) || len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
	}

	return nodeRegistrations[0], nil
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

// GetNodeAdmittanceCycle get the offset, in number of blocks, when we accept and expel nodes from registry
func (nrs *NodeRegistrationService) GetNodeAdmittanceCycle() uint32 {
	if nrs.NodeAdmittanceCycle == 0 {
		return constant.NodeAdmittanceCycle
	}
	return nrs.NodeAdmittanceCycle
}

func (nrs *NodeRegistrationService) BuildScrambledNodesAtHeight(blockHeight uint32) error {
	var (
		nearestBlock    model.Block
		nodeRegistries  []*model.NodeRegistration
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
		err             error
	)
	nearestHeight := nrs.GetBlockHeightToBuildScrambleNodes(blockHeight)
	nearestBlockRow, _ := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(nearestHeight), false)
	err = nrs.BlockQuery.Scan(&nearestBlock, nearestBlockRow)
	if err != nil {
		return err
	}
	nodeRegistries, err = nrs.sortNodeRegistries(&nearestBlock)
	if err != nil {
		return err
	}

	// Restructure & validating node address
	for key, node := range nodeRegistries {
		// STEF node.GetNodeAddress() must change into getting ip address from peer table by nodeID
		// note that we already have the address in node struct: see GetNodeRegistryAtHeightWithNodeAddress
		fullAddress := nrs.NodeRegistrationQuery.ExtractNodeAddress(node.GetNodeAddress())
		// Checking port of address,
		nodeInfo := p2pUtil.GetNodeInfo(fullAddress)
		fullAddresss := p2pUtil.GetFullAddressPeer(&model.Peer{
			Info: nodeInfo,
		})
		peer := &model.Peer{
			Info: &model.Node{
				ID:            node.GetNodeID(),
				Address:       nodeInfo.GetAddress(),
				Port:          nodeInfo.GetPort(),
				SharedAddress: nodeInfo.GetAddress(),
			},
		}
		index := key
		newIndexNodes[fullAddresss] = &index
		newAddressNodes = append(newAddressNodes, peer)
	}

	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	// memoize the scrambled nodes
	nrs.ScrambledNodes[nearestBlock.Height] = &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  nearestBlock.Height,
	}
	return nil
}

func (nrs *NodeRegistrationService) sortNodeRegistries(
	block *model.Block,
) ([]*model.NodeRegistration, error) {
	var nodeRegistries []*model.NodeRegistration
	// get node registry list
	rows, err := nrs.QueryExecutor.ExecuteSelect(
		nrs.NodeRegistrationQuery.GetNodeRegistryAtHeightWithNodeAddress(block.GetHeight()),
		false,
	)
	if err != nil {
		nrs.Logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	nodeRegistries, err = nrs.NodeRegistrationQuery.BuildModel(nodeRegistries, rows)
	if err != nil {
		nrs.Logger.Error(err.Error())
		return nil, err
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
	return nodeRegistries, nil
}

// BuildScrambleNodes,  buil sorted scramble nodes based on node registry
func (nrs *NodeRegistrationService) BuildScrambledNodes(block *model.Block) error {
	var (
		nodeRegistries  []*model.NodeRegistration
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
		err             error
	)
	nodeRegistries, err = nrs.sortNodeRegistries(block)
	if err != nil {
		return err
	}
	// Restructure & validating node address
	for key, node := range nodeRegistries {
		// STEF node.GetNodeAddress() must change into getting ip address from peer table by nodeID
		// note that we already have the address in node struct: see GetNodeRegistryAtHeightWithNodeAddress
		fullAddress := nrs.NodeRegistrationQuery.ExtractNodeAddress(node.GetNodeAddress())
		// Checking port of address,
		nodeInfo := p2pUtil.GetNodeInfo(fullAddress)
		fullAddresss := p2pUtil.GetFullAddressPeer(&model.Peer{
			Info: nodeInfo,
		})
		peer := &model.Peer{
			Info: &model.Node{
				ID:            node.GetNodeID(),
				Address:       nodeInfo.GetAddress(),
				Port:          nodeInfo.GetPort(),
				SharedAddress: nodeInfo.GetAddress(),
			},
		}
		index := key
		newIndexNodes[fullAddresss] = &index
		newAddressNodes = append(newAddressNodes, peer)
	}

	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	// memoize the scrambled nodes
	nrs.ScrambledNodes[block.Height] = &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  block.Height,
	}
	return nil
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
	if scrambleNodeExist == nil {
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

// GetNodeAddressesInfoFromDb returns a list of node address info messages given a list of nodeIDs
func (nrs *NodeRegistrationService) GetNodeAddressesInfoFromDb(nodeIDs []int64) ([]*model.NodeAddressInfo, error) {
	qry := nrs.NodeAddressInfoQuery.GetNodeAddressInfoByNodeIDs(nodeIDs)
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

// UpdateNodeAddressInfo updates or adds (in case new) a node address info record to db
// NOTE: nodeAddressInfo is supposed to have been already validated
func (nrs *NodeRegistrationService) UpdateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (updated bool, err error) {
	// check if already exist and if new one is more recent
	nodeAddressesInfo, err := nrs.GetNodeAddressesInfoFromDb([]int64{nodeAddressMessage.NodeID})
	if err != nil {
		return false, err
	}
	if len(nodeAddressesInfo) > 0 {
		prevNodeAddressInfo := nodeAddressesInfo[0]
		if prevNodeAddressInfo.GetAddress() == nodeAddressMessage.GetAddress() &&
			prevNodeAddressInfo.GetPort() == nodeAddressMessage.GetPort() {
			nrs.Logger.Debugf("node address info for node %d already up to date", nodeAddressMessage.NodeID)
			return false, err
		}
		err = nrs.QueryExecutor.BeginTx()
		if err != nil {
			return false, err
		}
		qryArgs := nrs.NodeAddressInfoQuery.UpdateNodeAddressInfo(nodeAddressMessage)
		err := nrs.QueryExecutor.ExecuteTransactions(qryArgs)
		if err != nil {
			_ = nrs.QueryExecutor.RollbackTx()
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
	qry, args := nrs.NodeAddressInfoQuery.InsertNodeAddressInfo(nodeAddressMessage)
	err = nrs.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		_ = nrs.QueryExecutor.RollbackTx()
		return false, err
	}
	err = nrs.QueryExecutor.CommitTx()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ValidateNodeAddressInfo validate message data against:
// - main blocks: block height and hash
// - node registry: nodeID and message signature (use node public key in registry to validate the signature)
func (nrs *NodeRegistrationService) ValidateNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) (bool, error) {
	var (
		block            model.Block
		nodeRegistration model.NodeRegistration
	)

	// validate nodeID
	qry, args := nrs.NodeRegistrationQuery.GetNodeRegistrationByID(nodeAddressInfo.GetNodeID())
	row, _ := nrs.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	err := nrs.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, blocker.NewBlocker(blocker.ValidationErr, "NodeIDNotFound")
		}
		return true, err
	}

	// validate the message signature
	unsignedBytes := nrs.NodeRegistrationUtils.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	if !nrs.Signature.VerifyNodeSignature(
		unsignedBytes,
		nodeAddressInfo.GetSignature(),
		nodeRegistration.GetNodePublicKey(),
	) {
		return true, blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}

	// validate block height
	blockRow, _ := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(nodeAddressInfo.GetBlockHeight()), false)
	err = nrs.BlockQuery.Scan(&block, blockRow)
	if err != nil || nodeAddressInfo.GetBlockHeight() <= block.GetHeight() {
		if err == sql.ErrNoRows {
			return true, blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHeight")
		}
		return true, err
	}
	// validate block hash
	if !bytes.Equal(nodeAddressInfo.GetBlockHash(), block.GetBlockHash()) {
		return true, blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHash")
	}

	// in case node address info for nodeID exists, validate that its height is lower than the one in the message
	qry = nrs.NodeAddressInfoQuery.GetNodeAddressInfoByNodeIDs([]int64{nodeAddressInfo.GetNodeID()})
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	prevAddressInfo, err := nrs.NodeAddressInfoQuery.BuildModel([]*model.NodeAddressInfo{}, rows)
	if err != nil {
		return true, err
	}
	if len(prevAddressInfo) > 0 {
		if prevAddressInfo[0].BlockHeight >= nodeAddressInfo.BlockHeight {
			return true, blocker.NewBlocker(blocker.ValidationErr, "OutdatedNodeAddressInfo")
		}
	}

	return true, nil
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
	nodeAddressInfoBytes := nrs.NodeRegistrationUtils.GetUnsignedNodeAddressInfoBytes(nodeAddressInfo)
	nodeAddressInfo.Signature = nrs.Signature.SignByNode(nodeAddressInfoBytes, nodeSecretPhrase)
	return nodeAddressInfo, nil
}
