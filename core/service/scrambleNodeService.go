package service

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
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
		PopOffScrambleToHeight(height uint32) error
	}

	ScrambleNodeService struct {
		NodeRegistrationService NodeRegistrationServiceInterface
		NodeAddressInfoService  NodeAddressInfoServiceInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		cacheStorage            storage.CacheStackStorageInterface
	}
)

func NewScrambleNodeService(
	nodeRegistrationService NodeRegistrationServiceInterface,
	nodeAddressInfoService NodeAddressInfoServiceInterface,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	scrambleNodeStackStorage storage.CacheStackStorageInterface,
) *ScrambleNodeService {
	return &ScrambleNodeService{
		NodeRegistrationService: nodeRegistrationService,
		NodeAddressInfoService:  nodeAddressInfoService,
		QueryExecutor:           queryExecutor,
		BlockQuery:              blockQuery,
		cacheStorage:            scrambleNodeStackStorage,
	}
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
func (sns *ScrambleNodeService) PopOffScrambleToHeight(height uint32) error {
	var (
		firstCachedScramble model.ScrambledNodes
		index               uint32
		err                 error
	)
	err = sns.cacheStorage.GetAtIndex(0, &firstCachedScramble)
	if err != nil {
		return err
	}
	nearestScrambleHeight := sns.GetBlockHeightToBuildScrambleNodes(height)
	index = (nearestScrambleHeight - firstCachedScramble.BlockHeight) / constant.PriorityStrategyBuildScrambleNodesGap
	err = sns.cacheStorage.PopTo(index - 1)
	return err
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
	var (
		index               uint32
		err                 error
		firstCachedScramble model.ScrambledNodes
		result              model.ScrambledNodes
	)
	err = sns.cacheStorage.GetAtIndex(0, &firstCachedScramble)
	if err != nil {
		return nil, err
	}
	nearestScrambleHeight := sns.GetBlockHeightToBuildScrambleNodes(blockHeight)
	index = (nearestScrambleHeight - firstCachedScramble.BlockHeight) / constant.PriorityStrategyBuildScrambleNodesGap
	err = sns.cacheStorage.GetAtIndex(index, &result)
	return &result, err
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
	newScramble := &model.ScrambledNodes{
		AddressNodes: newAddressNodes,
		IndexNodes:   newIndexNodes,
		BlockHeight:  block.Height,
	}
	err = sns.cacheStorage.Push(newScramble)
	return err
}
