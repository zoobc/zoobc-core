package strategy

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	Candidate struct {
		Blocksmith *model.Blocksmith
		StartTime  int64
		ExpiryTime int64
	}

	BlocksmithStrategyMain struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		LastSortedBlockID                      int64
		LastEstimatedBlockPersistedTimestamp   int64
		LastEstimatedPersistedTimestampBlockID int64
		SortedBlocksmithsLock                  sync.RWMutex
		SortedBlocksmithsMap                   map[string]*int64
		Chaintype                              chaintype.ChainType
		candidates                             []Candidate
		lastBlockHash                          []byte
	}
)

func NewBlocksmithStrategyMain(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	logger *log.Logger,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		QueryExecutor:          queryExecutor,
		NodeRegistrationQuery:  nodeRegistrationQuery,
		SkippedBlocksmithQuery: skippedBlocksmithQuery,
		Logger:                 logger,
		SortedBlocksmithsMap:   make(map[string]*int64),
		Chaintype:              &chaintype.MainChain{},
		candidates:             make([]Candidate, 0),
	}
}

func (bss *BlocksmithStrategyMain) isMe(lastCandidate Candidate, block *model.Block) bool {
	var (
		now = time.Now().Unix()
	)

	if now > lastCandidate.StartTime && bytes.Equal(lastCandidate.Blocksmith.NodePublicKey, block.BlocksmithPublicKey) {
		return true
	}
	return false
}

func (bss *BlocksmithStrategyMain) WillSmith(prevBlock *model.Block) (lastBlockID, blocksmithIndex int64, err error) {
	var (
		blockSmiths   []*model.Blocksmith
		lastCandidate Candidate
		candidate     Candidate
		now           = time.Now().Unix()
		// err           error
	)

	blockSmiths, err = bss.GetBlocksmiths(prevBlock)
	if err != nil {
		return 0, 0, errors.New("ErrorGetBlocksmiths")
	}

	if prevBlock.BlockHash != nil {
		bss.lastBlockHash = prevBlock.BlockHash
	}

	if !bytes.Equal(bss.lastBlockHash, prevBlock.BlockHash) {
		bss.candidates = []Candidate{}

		blockSeedBigInt := new(big.Int).SetBytes(prevBlock.BlockSeed)
		rand.Seed(blockSeedBigInt.Int64())
	}

	if len(bss.candidates) > 0 {
		lastCandidate = bss.candidates[len(bss.candidates)-1]
		isMe := bss.isMe(lastCandidate, prevBlock)
		if err != nil {
			return 0, 0, errors.New("ErrorIsMe")
		}

		if isMe && now < lastCandidate.ExpiryTime {
			return 0, 0, nil
		}
		if now < lastCandidate.StartTime+10 {
			return 0, 0, errors.New("Failed")
		}
	}
	idx := rand.Intn(len(blockSmiths))
	candidate = Candidate{
		Blocksmith: blockSmiths[idx],
		StartTime:  prevBlock.Timestamp + bss.Chaintype.GetSmithingPeriod() + int64(len(bss.candidates))*bss.Chaintype.GetBlocksmithTimeGap(),
		ExpiryTime: lastCandidate.StartTime + bss.Chaintype.GetBlocksmithNetworkTolerance() + bss.Chaintype.GetBlocksmithBlockCreationTime(),
	}

	bss.candidates = append(bss.candidates, candidate)
	lastBlockID = util.GetBlockIDFromHash(bss.lastBlockHash)
	return lastBlockID, int64(idx), nil
}

func (bss *BlocksmithStrategyMain) IsBlockValid(prevBlock, block *model.Block) error {
	var (
		err         error
		blockSmiths []*model.Blocksmith
	)
	blockSmiths, err = bss.GetBlocksmiths(prevBlock)
	if err != nil {
		return errors.New("ErrorGetBlocksmiths")
	}
	timeGap := block.Timestamp - prevBlock.Timestamp
	round := timeGap - bss.Chaintype.GetSmithingPeriod()/bss.Chaintype.GetBlocksmithTimeGap()

	blockSeedBigInt := new(big.Int).SetBytes(prevBlock.BlockSeed)
	rand.Seed(blockSeedBigInt.Int64())

	var idx int
	for i := 0; i < int(round); i++ {
		idx = rand.Intn(len(blockSmiths))
	}
	if bytes.Equal(blockSmiths[idx].NodePublicKey, block.BlocksmithPublicKey) {
		return nil
	}

	return errors.New("Failed")
}

// GetBlocksmiths select the blocksmiths for a given block and calculate the SmithOrder (for smithing) and NodeOrder (for block rewards)
func (bss *BlocksmithStrategyMain) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	var (
		activeBlocksmiths, blocksmiths []*model.Blocksmith
	)
	// get all registered nodes with participation score > 0
	rows, err := bss.QueryExecutor.ExecuteSelect(bss.NodeRegistrationQuery.GetActiveNodeRegistrationsByHeight(
		block.Height), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activeBlocksmiths, err = bss.NodeRegistrationQuery.BuildBlocksmith(activeBlocksmiths, rows)
	if err != nil {
		return nil, err
	}
	monitoring.SetNodeScore(activeBlocksmiths)
	monitoring.SetActiveRegisteredNodesCount(len(activeBlocksmiths))
	// add smithorder and nodeorder to be used to select blocksmith and coinbase rewards
	for _, blocksmith := range activeBlocksmiths {
		blocksmith.BlockSeed, err = coreUtil.GetBlockSeed(blocksmith.NodeID, block)
		if err != nil {
			return nil, err
		}
		blocksmith.NodeOrder = coreUtil.CalculateNodeOrder(blocksmith.Score, blocksmith.BlockSeed, blocksmith.NodeID)
		blocksmiths = append(blocksmiths, blocksmith)
	}
	return blocksmiths, nil
}

func (bss *BlocksmithStrategyMain) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	copy(result, bss.SortedBlocksmiths)
	return result
}

// GetSortedBlocksmithsMap get the sorted blocksmiths in map
func (bss *BlocksmithStrategyMain) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	var (
		result = make(map[string]*int64)
	)
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	if block.ID != bss.LastSortedBlockID || block.ID == constant.MainchainGenesisBlockID {
		bss.SortBlocksmiths(block, false)
	}
	for k, v := range bss.SortedBlocksmithsMap {
		result[k] = v
	}
	return result
}

// SortBlocksmiths sort the list of active node of current block height, index start from 0.
func (bss *BlocksmithStrategyMain) SortBlocksmiths(block *model.Block, withLock bool) {
	if block.ID == bss.LastSortedBlockID && block.ID != constant.MainchainGenesisBlockID {
		return
	}
	// fetch valid blocksmiths
	var blocksmiths []*model.Blocksmith
	nextBlocksmiths, err := bss.GetBlocksmiths(block)
	if err != nil {
		bss.Logger.Errorf("SortBlocksmith (Main):GetBlocksmiths fail: %s", err)
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
	// clean up bss.SortedBlocksmithsMap
	bss.SortedBlocksmithsMap = make(map[string]*int64)
	// copying the sorted list to map[string(publicKey)]index
	for index, blocksmith := range blocksmiths {
		blocksmithIndex := int64(index)
		bss.SortedBlocksmithsMap[string(blocksmith.NodePublicKey)] = &blocksmithIndex
	}
	// set last sorted block id
	bss.LastSortedBlockID = block.ID
	bss.SortedBlocksmiths = blocksmiths

	monitoring.SetNextSmith(blocksmiths, bss.SortedBlocksmithsMap)
}

// CalculateScore calculate the blocksmith score
func (bss *BlocksmithStrategyMain) CalculateScore(generator *model.Blocksmith, score int64) error {
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	return nil
}

func (bss *BlocksmithStrategyMain) EstimateLastBlockPersistedTime(
	previousBlock *model.Block,
	ct chaintype.ChainType,
) error {
	var (
		skippedBlocksmiths []*model.SkippedBlocksmith
	)
	skippedBlocksmiths, err := func() ([]*model.SkippedBlocksmith, error) {
		skippedBlocksmithsQ := bss.SkippedBlocksmithQuery.GetSkippedBlocksmithsByBlockHeight(previousBlock.Height)
		skippedBlocksmithsRows, err := bss.QueryExecutor.ExecuteSelect(skippedBlocksmithsQ, false)
		if err != nil {
			return nil, err
		}
		defer skippedBlocksmithsRows.Close()
		if _, err := bss.SkippedBlocksmithQuery.BuildModel(skippedBlocksmiths, skippedBlocksmithsRows); err != nil {
			return nil, err
		}
		return skippedBlocksmiths, nil
	}()
	if err != nil {
		return err
	}
	if len(skippedBlocksmiths) > 0 {
		previousBlocksmithTime := previousBlock.Timestamp - ct.GetBlocksmithTimeGap()
		estimatedPreviousBlockPersistTime := previousBlocksmithTime + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
		bss.LastEstimatedBlockPersistedTimestamp = estimatedPreviousBlockPersistTime
	} else {
		bss.LastEstimatedBlockPersistedTimestamp = previousBlock.GetTimestamp()
	}
	bss.LastEstimatedPersistedTimestampBlockID = previousBlock.ID
	return nil
}

// CanPersistBlock check if currentTime is a time to persist the provided block.
// This function uses current node time, which make it unsafe to validate past block.
// numberOfBlocksmiths must be > 0
func (bss *BlocksmithStrategyMain) CanPersistBlock(
	blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
	var (
		err                                                                             error
		ct                                                                              = &chaintype.MainChain{}
		currentTime                                                                     = time.Now().Unix()
		remainder, prevRoundBegin, prevRoundExpired, prevRound2Begin, prevRound2Expired int64
	)
	// always return true for the first block | keeping in mind genesis block's timestamps is far behind, let fork processor
	// handle to get highest cum-diff block
	if previousBlock.GetHeight() == 0 {
		return nil
	}
	// calculate estimated starting time
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		err = bss.EstimateLastBlockPersistedTime(previousBlock, ct)
		if err != nil {
			return err
		}
	}
	// check if is valid time
	// calculate total time before every blocksmiths are skipped
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()
	timeSinceLastBlock := currentTime - bss.LastEstimatedBlockPersistedTimestamp
	if timeSinceLastBlock < ct.GetSmithingPeriod() {
		return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
	}
	modTimeSinceLastBlock := timeSinceLastBlock - ct.GetSmithingPeriod()
	timeRound := math.Floor(float64(modTimeSinceLastBlock) / float64(timeForOneRound))
	remainder = modTimeSinceLastBlock % timeForOneRound
	nearestRoundBeginning := currentTime - remainder
	if timeRound > 0 { // if more than one round has passed, calculate previous round start-expiry time for overlap
		prevRoundStart := nearestRoundBeginning - timeForOneRound
		prevRoundBegin = prevRoundStart + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRoundExpired = prevRoundBegin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
	if timeRound > 1 { // handle small network, go one more round
		prevRound2Start := nearestRoundBeginning - 2*timeForOneRound
		prevRound2Begin = prevRound2Start + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRound2Expired = prevRound2Begin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
	// calculate current round begin and expiry time
	allowedBeginTime := blocksmithIndex*ct.GetBlocksmithTimeGap() + nearestRoundBeginning
	expiredTime := allowedBeginTime + ct.GetBlocksmithBlockCreationTime() +
		ct.GetBlocksmithNetworkTolerance()
	// check if current time is in {(expire-timeGap) < x < (expire)} in either previous round or current round
	if (currentTime > (expiredTime-ct.GetBlocksmithTimeGap()) && currentTime <= expiredTime) ||
		(currentTime > (prevRoundExpired-ct.GetBlocksmithTimeGap()) && currentTime <= prevRoundExpired) ||
		(currentTime > (prevRound2Expired-ct.GetBlocksmithTimeGap()) && currentTime <= prevRound2Expired) {
		return nil
	}
	return blocker.NewBlocker(blocker.BlockErr, "CannotPersistBlock")
}

func (bss *BlocksmithStrategyMain) IsValidSmithTime(
	blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
	var (
		err                                         error
		currentTime                                 = time.Now().Unix()
		ct                                          = &chaintype.MainChain{}
		remainder, prevRoundBegin, prevRoundExpired int64
	)
	// calculate estimated starting time
	if bss.LastEstimatedPersistedTimestampBlockID != previousBlock.ID {
		err = bss.EstimateLastBlockPersistedTime(previousBlock, ct)
		if err != nil {
			return err
		}
	}
	// calculate total time before every blocksmiths are skipped
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()

	timeSinceLastBlock := currentTime - bss.LastEstimatedBlockPersistedTimestamp
	if timeSinceLastBlock < ct.GetSmithingPeriod() {
		return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
	}
	modTimeSinceLastBlock := timeSinceLastBlock - ct.GetSmithingPeriod()
	timeRound := math.Floor(float64(modTimeSinceLastBlock) / float64(timeForOneRound))
	remainder = modTimeSinceLastBlock % timeForOneRound
	// find the time of beginning of the list
	nearestRoundBeginning := currentTime - remainder
	if timeRound > 0 { // if more than one round has passed, calculate previous round start-expiry time for overlap
		prevRoundStart := nearestRoundBeginning - timeForOneRound
		prevRoundBegin = prevRoundStart + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRoundExpired = prevRoundBegin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
	// calculate current round begin and expiry time
	allowedBeginTime := blocksmithIndex*ct.GetBlocksmithTimeGap() + nearestRoundBeginning
	expiredTime := allowedBeginTime + ct.GetBlocksmithBlockCreationTime() +
		ct.GetBlocksmithNetworkTolerance()
	// if currentTime overlap with either currentRound window or previous round window, it's considered valid time
	if (currentTime >= allowedBeginTime && currentTime <= expiredTime) ||
		(currentTime >= prevRoundBegin && currentTime <= prevRoundExpired) {
		return nil
	}
	return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
}
