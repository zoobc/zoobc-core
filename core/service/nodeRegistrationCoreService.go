package service

import (
	"math/big"
	"sort"
	"sync"

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
		GetScrambledNodes() *ScrambledNodes
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAdmittanceCycle     uint32
		Logger                  *log.Logger
		ScrambledNodes          *ScrambledNodes
		ScrambledNodesLock      sync.RWMutex
		MomoizedScrambledNodes  *ScrambledNodes
	}

	ScrambledNodes struct {
		IndexNodes   map[string]*int // if we use normal int, we won't be able to detect null values
		AddressNodes []*model.Peer
		BlockHeight  uint32
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	logger *log.Logger,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:           queryExecutor,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		ParticipationScoreQuery: participationScoreQuery,
		NodeAdmittanceCycle:     constant.NodeAdmittanceCycle,
		Logger:                  logger,
	}
}

// SelectNodesToBeAdmitted Select n (=limit) queued nodes with the highest locked balance
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(limit, uint32(model.NodeRegistrationState_NodeQueued))
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

// AdmitNodes update given node registrations' registrationStatus field to 0 (= node registered) and set default participation score to it
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	err := nrs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	// prepare all node registrations to be updated (set registrationStatus to 0 and new height) and default participation scores to be added
	for _, nodeRegistration := range nodeRegistrations {
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeRegistered)
		nodeRegistration.Height = height
		// update the node registry (set registrationStatus to zero)
		queries := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		ps := &model.ParticipationScore{
			NodeID: nodeRegistration.NodeID,
			Score:  constant.MaxParticipationScore / 10,
			Latest: true,
			Height: height,
		}
		// add default participation score to the node
		insertParticipationScoreQ, insertParticipationScoreArg := nrs.ParticipationScoreQuery.InsertParticipationScore(ps)
		queries = append(queries,
			append([]interface{}{insertParticipationScoreQ}, insertParticipationScoreArg...),
		)
		err = nrs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			if rollbackErr := nrs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				nrs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}

	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}

// ExpelNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting registrationStatus field to 3 (deleted) and locked balance to zero
func (nrs *NodeRegistrationService) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	err := nrs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	for _, nodeRegistration := range nodeRegistrations {
		// update the node registry (set registrationStatus to 1 and lockedbalance to 0)
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)
		nodeRegistration.LockedBalance = 0
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
		err := nrs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			if rollbackErr := nrs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				nrs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
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

// BuildScrambleNodes,  buil sorted scramble nodes based on node registry
func (nrs *NodeRegistrationService) BuildScrambledNodes(block *model.Block) error {
	var (
		nodeRegistries  []*model.NodeRegistration
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
	)
	// get node registry list
	rows, err := nrs.QueryExecutor.ExecuteSelect(
		nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.GetHeight()),
		false,
	)
	if err != nil {
		nrs.Logger.Error(err.Error())
		return err
	}
	defer rows.Close()
	nodeRegistries, err = nrs.NodeRegistrationQuery.BuildModel(nodeRegistries, rows)
	if err != nil {
		nrs.Logger.Error(err.Error())
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
		fullAddress := nrs.NodeRegistrationQuery.ExtractNodeAddress(node.GetNodeAddress())
		// Checking port of address,
		nodeInfo := p2pUtil.GetNodeInfo(fullAddress)
		fullAddresss := p2pUtil.GetFullAddressPeer(&model.Peer{
			Info: nodeInfo,
		})
		peer := &model.Peer{
			Info: &model.Node{
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
	nrs.ScrambledNodes = &ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  block.Height,
	}
	nrs.ResetMemoizedScrambledNodes()
	return nil
}

func (nrs *NodeRegistrationService) ResetMemoizedScrambledNodes() {
	nrs.MomoizedScrambledNodes = nil
}

func (nrs *NodeRegistrationService) GetScrambledNodes() *ScrambledNodes {
	if nrs.ScrambledNodes == nil {
		return &ScrambledNodes{
			AddressNodes: []*model.Peer{},
		}
	}

	var (
		newIndexNodes   = make(map[string]*int)
		newAddressNodes []*model.Peer
	)

	nrs.ScrambledNodesLock.Lock()
	defer nrs.ScrambledNodesLock.Unlock()

	if nrs.MomoizedScrambledNodes != nil && nrs.MomoizedScrambledNodes.BlockHeight == nrs.ScrambledNodes.BlockHeight {
		return nrs.MomoizedScrambledNodes
	}

	for _, addressNode := range nrs.ScrambledNodes.AddressNodes {
		newAddressNodes = append(newAddressNodes, addressNode)
	}

	for key, indexNode := range nrs.ScrambledNodes.IndexNodes {
		tempVal := *indexNode
		newIndexNodes[key] = &tempVal
	}

	nrs.MomoizedScrambledNodes = &ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  nrs.ScrambledNodes.BlockHeight,
	}

	return nrs.MomoizedScrambledNodes
}
