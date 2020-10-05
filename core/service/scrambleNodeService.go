package service

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"golang.org/x/crypto/sha3"
)

type (
	ScrambleNodeServiceInterface interface {
		InitializeScrambleCache(lastBlockHeight uint32) error
		GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32
		GetScrambleNodesByHeight(
			blockHeight uint32,
		) (*model.ScrambledNodes, error)
		BuildScrambledNodes(block *model.Block) error
		BuildScrambledNodesAtHeight(blockHeight uint32) (*model.ScrambledNodes, error)
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

func (sns *ScrambleNodeService) InitializeScrambleCache(lastBlockHeight uint32) error {
	var (
		topScramble model.ScrambledNodes
		err         error
	)
	err = sns.cacheStorage.GetTop(&topScramble)
	if err != nil {
		blockerErr, ok := err.(blocker.Blocker)
		if ok {
			if blockerErr.Type != blocker.CacheEmpty {
				return err
			}
		} else {
			return err
		}
	}

	// clear memory
	err = sns.cacheStorage.Clear()
	if err != nil {
		return err
	}
	var revertedScrambleBlocks = make([]model.Block, 0)
	firstHeight := sns.GetBlockHeightToBuildScrambleNodes(lastBlockHeight)
	getBlockAtHeight := func(height uint32) (model.Block, error) {
		var block model.Block
		nearestBlockRow, _ := sns.QueryExecutor.ExecuteSelectRow(sns.BlockQuery.GetBlockByHeight(height), false)
		err := sns.BlockQuery.Scan(&block, nearestBlockRow)
		return block, err
	}
	firstBlock, err := getBlockAtHeight(firstHeight)
	if err != nil {
		return err
	}
	revertedScrambleBlocks = append(revertedScrambleBlocks, firstBlock)
	if firstHeight != 0 {
		startHeight := firstHeight
		for !(startHeight == 0 || len(revertedScrambleBlocks) > int(constant.MaxScrambleCacheRound)) {
			startHeight -= constant.PriorityStrategyBuildScrambleNodesGap
			scrambleBlock, err := getBlockAtHeight(startHeight)
			if err != nil {
				if err == sql.ErrNoRows {
					break
				}
				return err
			}
			revertedScrambleBlocks = append(revertedScrambleBlocks, scrambleBlock)
		}
	}
	for i := len(revertedScrambleBlocks) - 1; i >= 0; i-- {
		err := sns.BuildScrambledNodes(&revertedScrambleBlocks[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildScrambledNodes build sorted scramble nodes based on node registry
func (sns *ScrambleNodeService) BuildScrambledNodes(block *model.Block) error {
	scrambleNodes, err := sns.ScrambleNodeRegistries(block)
	if err != nil {
		return err
	}
	err = sns.cacheStorage.Push(*scrambleNodes)
	if err != nil {
		return err
	}
	return nil
}

func (*ScrambleNodeService) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight - (lastBlockHeight % constant.PriorityStrategyBuildScrambleNodesGap)
}

// PopOffScrambleToHeight delete cache of scrambles to given height's nearest scramble
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
	err = sns.cacheStorage.PopTo(index)
	return err
}

// BuildScrambledNodesAtHeight build scramble node at custom height, used to build older scramble node
// this function will not store the scramble result in cache
func (sns *ScrambleNodeService) BuildScrambledNodesAtHeight(blockHeight uint32) (*model.ScrambledNodes, error) {
	var (
		nearestBlock model.Block
		err          error
	)
	nearestHeight := sns.GetBlockHeightToBuildScrambleNodes(blockHeight)
	// todo: get block at height should be via service instead of executor
	nearestBlockRow, _ := sns.QueryExecutor.ExecuteSelectRow(sns.BlockQuery.GetBlockByHeight(nearestHeight), false)
	err = sns.BlockQuery.Scan(&nearestBlock, nearestBlockRow)
	if err != nil {
		return nil, err
	}
	scrambleNodes, err := sns.ScrambleNodeRegistries(&nearestBlock)
	return scrambleNodes, err
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
	if blockHeight == firstCachedScramble.BlockHeight {
		return &firstCachedScramble, nil
	}
	if blockHeight < firstCachedScramble.BlockHeight {
		// looking for an older scramble that's not cached, look into database
		scrambleNodes, err := sns.BuildScrambledNodesAtHeight(blockHeight)
		return scrambleNodes, err
	}
	nearestScrambleHeight := sns.GetBlockHeightToBuildScrambleNodes(blockHeight)
	index = (nearestScrambleHeight - firstCachedScramble.BlockHeight) / constant.PriorityStrategyBuildScrambleNodesGap
	err = sns.cacheStorage.GetAtIndex(index, &result)
	return &result, err
}

// ScrambleNodeRegistries this function is responsible of selecting and sorting registered nodes so that nodes/peers in scrambledNodes map changes
// order at a given interval
// note: this algorithm is deterministic for the whole network so that,
// at any point in time every node can calculate this map autonomously, given its node registry is updated
func (sns *ScrambleNodeService) ScrambleNodeRegistries(block *model.Block) (*model.ScrambledNodes, error) {
	var (
		nodeRegistries       []*model.NodeRegistration
		newAddressNodes      []*model.Peer
		newIndexNodes        = make(map[string]*int)
		nodePublicKeyToIDMap = make(map[string]int64)
		err                  error
	)

	nodeRegistries, err = sns.NodeRegistrationService.GetNodeRegistryAtHeight(block.GetHeight())
	if err != nil {
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
	// Restructure & validating node address
	for key, node := range nodeRegistries {
		nai, err := sns.NodeAddressInfoService.GetAddressInfoByNodeIDWithPreferredStatus(
			node.GetNodeID(),
			model.NodeAddressStatus_NodeAddressPending,
		)
		if err != nil {
			return nil, err
		}
		peer := &model.Peer{
			Info: &model.Node{
				ID:        node.GetNodeID(),
				PublicKey: node.GetNodePublicKey(),
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
		nodePublicKeyToIDMap[hex.EncodeToString(node.GetNodePublicKey())] = node.NodeID
		newAddressNodes = append(newAddressNodes, peer)
	}

	return &model.ScrambledNodes{
		AddressNodes:         newAddressNodes,
		IndexNodes:           newIndexNodes,
		NodePublicKeyToIDMap: nodePublicKeyToIDMap,
		BlockHeight:          block.Height,
	}, nil

}
