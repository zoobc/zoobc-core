package service

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sort"
)

type (
	ScrambleNodeServiceInterface interface {
		GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32
		GetScrambleNodesByHeight(
			blockHeight uint32,
		) (*model.ScrambledNodes, error)
		BuildScrambledNodes(block *model.Block) error
		BuildScrambledNodesAtHeight(blockHeight uint32) error
		ResetScrambledNodes()
	}

	ScrambleNodeService struct {
		NodeRegistrationService NodeRegistrationServiceInterface
		NodeAddressInfoService  NodeAddressInfoServiceInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
	}
)

func NewScrambleNodeService() *ScrambleNodeService {
	return &ScrambleNodeService{}
}

func (sns *ScrambleNodeService) InitializeScrambleCache() error {
	return nil
}

// BuildScrambleNodes, build sorted scramble nodes based on node registry
func (sns *ScrambleNodeService) BuildScrambledNodes(block *model.Block) error {
	return sns.sortNodeRegistries(block)
}

func (*ScrambleNodeService) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight - (lastBlockHeight % constant.PriorityStrategyBuildScrambleNodesGap)
}

// ResetScrambledNodes todo: update this to `PopOffScrambleToHeight`
func (sns *ScrambleNodeService) ResetScrambledNodes() {
	// nrs.ScrambledNodes = map[uint32]*model.ScrambledNodes{}
}

func (sns *ScrambleNodeService) BuildScrambledNodesAtHeight(blockHeight uint32) error {
	var (
		nearestBlock model.Block
		err          error
	)
	nearestHeight := sns.GetBlockHeightToBuildScrambleNodes(blockHeight)
	// todo: get block at height should be via service instead of executor
	nearestBlockRow, _ := sns.QueryExecutor.ExecuteSelectRow(sns.BlockQuery.GetBlockByHeight(nearestHeight), false)
	err = sns.BlockQuery.Scan(&nearestBlock, nearestBlockRow)
	if err != nil {
		return err
	}
	return sns.sortNodeRegistries(&nearestBlock)
}

func (sns *ScrambleNodeService) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	// var (
	// 	newAddressNodes []*model.Peer
	// 	newIndexNodes   = make(map[string]*int)
	// 	// err             error
	// )
	// nearestHeight := sns.GetBlockHeightToBuildScrambleNodes(blockHeight)
	// sns.ScrambledNodesLock.RLock()
	// scrambleNodeExist := nrs.ScrambledNodes[nearestHeight]
	// nrs.ScrambledNodesLock.RUnlock()
	// if scrambleNodeExist == nil || blockHeight < constant.ScrambleNodesSafeHeight {
	// 	err = nrs.BuildScrambledNodesAtHeight(nearestHeight)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	// nrs.ScrambledNodesLock.Lock()
	// defer nrs.ScrambledNodesLock.Unlock()
	// scrambledNodes := nrs.ScrambledNodes[nearestHeight]
	// newAddressNodes = append(newAddressNodes, scrambledNodes.AddressNodes...)
	// // in the window, deep copy the nodes
	// for key, indexNode := range scrambledNodes.IndexNodes {
	// 	tempVal := *indexNode
	// 	newIndexNodes[key] = &tempVal
	// }
	// return &model.ScrambledNodes{
	// 	AddressNodes: newAddressNodes,
	// 	IndexNodes:   newIndexNodes,
	// 	BlockHeight:  scrambledNodes.BlockHeight,
	// }, nil
	return nil, nil
}

// sortNodeRegistries this function is responsible of selecting and sorting registered nodes so that nodes/peers in scrambledNodes map changes
// order at a given interval
// note: this algorithm is deterministic for the whole network so that,
// at any point in time every node can calculate this map autonomously, given its node registry is updated
func (sns *ScrambleNodeService) sortNodeRegistries(
	block *model.Block,
) error {
	var (
		nodeRegistries  []*model.NodeRegistration
		newAddressNodes []*model.Peer
		newIndexNodes   = make(map[string]*int)
		err             error
	)

	nodeRegistries, err = sns.NodeRegistrationService.GetNodeRegistryAtHeight(block.GetHeight())
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
		nai, err := sns.NodeAddressInfoService.GetAddressInfoByNodeID(node.GetNodeID(), model.NodeAddressStatus_NodeAddressPending)
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
	// memoize result to cache layer
	// sns.ScrambledNodes[block.Height] = &model.ScrambledNodes{
	// 	AddressNodes: newAddressNodes,
	// 	IndexNodes:   newIndexNodes,
	// 	BlockHeight:  block.Height,
	// }

	return nil
}
