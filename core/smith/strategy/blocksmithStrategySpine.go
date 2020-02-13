package strategy

import (
	"encoding/binary"
	"math/big"
	"sort"
	"sync"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlocksmithStrategySpine struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
		SpineBlockQuery       query.BlockQueryInterface
	}
)

func NewBlocksmithStrategySpine(
	queryExecutor query.ExecutorInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	logger *log.Logger,
	spineBlockQuery query.BlockQueryInterface,
) *BlocksmithStrategySpine {
	return &BlocksmithStrategySpine{
		QueryExecutor:        queryExecutor,
		SpinePublicKeyQuery:  spinePublicKeyQuery,
		Logger:               logger,
		SortedBlocksmithsMap: make(map[string]*int64),
		SpineBlockQuery:      spineBlockQuery,
	}
}

// GetBlocksmiths select the blocksmiths for a given block and calculate the SmithOrder (for smithing) and NodeOrder (for block rewards)
func (bss *BlocksmithStrategySpine) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	var (
		validBlocksmiths, blocksmiths []*model.Blocksmith
	)
	// get all registered nodes with participation score > 0
	rows, err := bss.QueryExecutor.ExecuteSelect(bss.SpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(0, block.Height), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	validBlocksmiths, err = bss.SpinePublicKeyQuery.BuildBlocksmith(validBlocksmiths, rows)
	if err != nil {
		return nil, err
	}
	// TODO: do we want to add monitoring param for spine blocksmiths count?
	// monitoring.SetActiveRegisteredNodesCount(len(validBlocksmiths))
	// add smithorder to be used to select blocksmith
	for _, blocksmith := range validBlocksmiths {
		// FIXME: ask @barton double check with him that generating a pseudo random id to compute the blockSeed is ok
		pseudoNodeID := int64(binary.LittleEndian.Uint64(blocksmith.NodePublicKey))
		blocksmith.BlockSeed, err = coreUtil.GetBlockSeed(pseudoNodeID, block)
		if err != nil {
			return nil, err
		}
		// FIXME: ask @barton how to compute or assign a score to spine blocksmiths, since we don't have any participation score?
		//		  at the moment we always assign a default score to all blocksmiths
		blocksmith.Score = big.NewInt(constant.DefaultParticipationScore)
		blocksmith.NodeOrder = coreUtil.CalculateNodeOrder(blocksmith.Score, blocksmith.BlockSeed, pseudoNodeID)
		blocksmiths = append(blocksmiths, blocksmith)
	}

	return blocksmiths, nil
}

func (bss *BlocksmithStrategySpine) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.SpinechainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	copy(result, bss.SortedBlocksmiths)
	return result
}

// GetSortedBlocksmithsMap get the sorted blocksmiths in map
func (bss *BlocksmithStrategySpine) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	var (
		result = make(map[string]*int64)
	)
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.SpinechainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	for k, v := range bss.SortedBlocksmithsMap {
		result[k] = v
	}
	return result
}

func (bss *BlocksmithStrategySpine) SortBlocksmiths(block *model.Block, withLock bool) {
	if block.ID == bss.LastSortedBlockID && block.ID != constant.SpinechainGenesisBlockID {
		return
	}

	var (
		prevHeight = block.Height
		prevBlock  model.Block
		err        error
	)

	// always calculate sorted blocksmiths from previous block, otherwise when downloading the spine blocks it could happen
	// that the node is unable to validate a block if it is smithed by a newly registered node that has his public key included in the same
	// block the node is trying to validate (in that scenario the node's public key isn't in the db yet because the block hasn't been
	// pushed yet)
	if block.Height > 0 {
		prevHeight = block.Height - 1
	}
	blockAtHeightQ := bss.SpineBlockQuery.GetBlockByHeight(prevHeight)
	blockAtHeightRow, _ := bss.QueryExecutor.ExecuteSelectRow(blockAtHeightQ, false)
	err = bss.SpineBlockQuery.Scan(&prevBlock, blockAtHeightRow)
	if err != nil {
		bss.Logger.Errorf("SortBlocksmith (Spine):GetBlockByHeight fail: %s", err)
		return
	}

	// fetch valid blocksmiths
	var blocksmiths []*model.Blocksmith
	nextBlocksmiths, err := bss.GetBlocksmiths(&prevBlock)
	if err != nil {
		bss.Logger.Errorf("SortBlocksmith (Spine):GetBlocksmiths fail: %s", err)
		return
	}
	// copy the nextBlocksmiths pointers array into an array of blocksmiths
	blocksmiths = append(blocksmiths, nextBlocksmiths...)
	// sort blocksmiths by SmithOrder
	sort.SliceStable(blocksmiths, func(i, j int) bool {
		if blocksmiths[i].BlockSeed == blocksmiths[j].BlockSeed {
			return blocksmiths[i].NodeID < blocksmiths[j].NodeID
		}
		// ascending sort
		return blocksmiths[i].BlockSeed < blocksmiths[j].BlockSeed
	})

	if withLock {
		bss.SortedBlocksmithsLock.Lock()
		defer bss.SortedBlocksmithsLock.Unlock()
	}
	// copying the sorted list to map[string(publicKey)]index
	for index, blocksmith := range blocksmiths {
		blocksmithIndex := int64(index)
		bss.SortedBlocksmithsMap[string(blocksmith.NodePublicKey)] = &blocksmithIndex
	}
	// set last sorted block id
	bss.LastSortedBlockID = block.ID
	bss.SortedBlocksmiths = blocksmiths
}

// CalculateSmith calculate seed, smithTime, and Deadline
func (bss *BlocksmithStrategySpine) CalculateSmith(
	lastBlock *model.Block,
	blocksmithIndex int64,
	generator *model.Blocksmith,
	score int64,
) error {
	// FIXME: ask @barton probably the way we compute spine blocksmith has to be reviewed, since we don't have ps and receipts,
	//		  attached to spine blocks
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	generator.SmithTime = bss.GetSmithTime(blocksmithIndex, lastBlock)
	return nil
}

// GetSmithTime calculate smith time of a blocksmith
func (bss *BlocksmithStrategySpine) GetSmithTime(blocksmithIndex int64, block *model.Block) int64 {
	ct := &chaintype.SpineChain{}
	elapsedFromLastBlock := (blocksmithIndex + 1) * ct.GetSmithingPeriod()
	return block.GetTimestamp() + elapsedFromLastBlock
}
