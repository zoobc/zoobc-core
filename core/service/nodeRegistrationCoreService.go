package service

import (
	"math/big"
	"sort"
	"sync"

	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"golang.org/x/crypto/sha3"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		SelectNodesToBeExpelled() ([]*model.NodeRegistration, error)
		GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error)
		GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error)
		AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		GetNodeAdmittanceCycle() uint32
		BuildScrambledNodes(block *model.Block) error
		ResetMemoizedScrambledNodes()
		ResetScrambledNodes()
		GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32
		GetLatestScrambledNodes() *model.ScrambledNodes
		GetScrambleNodesByHeight(
			blockHeight uint32,
		) (*model.ScrambledNodes, error)
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmittanceCycle          uint32
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		ScrambledNodesLock           sync.RWMutex
		MemoizedLatestScrambledNodes *model.ScrambledNodes
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	blockQuery query.BlockQueryInterface,
	logger *log.Logger,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:           queryExecutor,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		ParticipationScoreQuery: participationScoreQuery,
		BlockQuery:              blockQuery,
		NodeAdmittanceCycle:     constant.NodeAdmittanceCycle,
		Logger:                  logger,
		ScrambledNodes:          map[uint32]*model.ScrambledNodes{},
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
	qry := nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(height)
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
		addParticipationScoreQry := nrs.ParticipationScoreQuery.AddParticipationScore(
			nodeRegistration.NodeID,
			constant.DefaultParticipationScore,
			height)
		queries = append(queries, addParticipationScoreQry...)
		if err := nrs.QueryExecutor.ExecuteTransactions(queries); err != nil {
			return err
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
	nearestBlockRow := nrs.QueryExecutor.ExecuteSelectRow(nrs.BlockQuery.GetBlockByHeight(nearestHeight))
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
		nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.GetHeight()),
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
	nrs.ResetMemoizedScrambledNodes()
	return nil
}

func (nrs *NodeRegistrationService) ResetMemoizedScrambledNodes() {
	nrs.MemoizedLatestScrambledNodes = nil
}

func (nrs *NodeRegistrationService) ResetScrambledNodes() {
	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	nrs.ScrambledNodes = map[uint32]*model.ScrambledNodes{}
}

func (nrs *NodeRegistrationService) GetLatestScrambledNodes() *model.ScrambledNodes {
	if len(nrs.ScrambledNodes) < 1 {
		return &model.ScrambledNodes{
			AddressNodes: []*model.Peer{},
		}
	}
	var (
		newIndexNodes   = make(map[string]*int)
		newAddressNodes []*model.Peer
		lastBlock       *model.Block
		err             error
	)
	lastBlock, err = util.GetLastBlock(nrs.QueryExecutor, nrs.BlockQuery)
	if err != nil {
		nrs.Logger.Error(err)
		return &model.ScrambledNodes{
			AddressNodes: []*model.Peer{},
		}
	}
	nearestBlockHeight := nrs.GetBlockHeightToBuildScrambleNodes(lastBlock.Height)
	if nrs.ScrambledNodes[nearestBlockHeight] == nil {
		err = nrs.BuildScrambledNodesAtHeight(nearestBlockHeight)
		if err != nil {
			return nil
		}
	}

	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()
	if nrs.MemoizedLatestScrambledNodes != nil {
		if nrs.MemoizedLatestScrambledNodes.BlockHeight == nrs.ScrambledNodes[nearestBlockHeight].BlockHeight {
			return nrs.MemoizedLatestScrambledNodes
		}
	}

	newAddressNodes = append(newAddressNodes, nrs.ScrambledNodes[nearestBlockHeight].AddressNodes...)

	for key, indexNode := range nrs.ScrambledNodes[nearestBlockHeight].IndexNodes {
		tempVal := *indexNode
		newIndexNodes[key] = &tempVal
	}

	nrs.MemoizedLatestScrambledNodes = &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  nrs.ScrambledNodes[nearestBlockHeight].BlockHeight,
	}

	return nrs.MemoizedLatestScrambledNodes
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
	if nrs.ScrambledNodes[nearestHeight] == nil {
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
