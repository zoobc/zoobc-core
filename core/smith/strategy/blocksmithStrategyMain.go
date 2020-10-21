package strategy

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/storage"

	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	Candidate struct {
		Blocksmith *model.Blocksmith
		StartTime  int64
		ExpiryTime int64
	}

	Me struct {
		ExpiryTime      int64
		LastBlockID     int64
		BlocksmithIndex int64
	}

	BlocksmithStrategyMain struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		ActiveNodeRegistryCacheStorage         storage.CacheStorageInterface
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
		me                                     Me
		lastBlockHash                          []byte
		lastTimeAddCandidate                   int64
		CurrentNodePublicKey                   []byte
	}
)

func NewBlocksmithStrategyMain(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	logger *log.Logger,
	currentNodePublicKey []byte,
	activeNodeRegistryCacheStorage storage.CacheStorageInterface,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		QueryExecutor:                  queryExecutor,
		NodeRegistrationQuery:          nodeRegistrationQuery,
		SkippedBlocksmithQuery:         skippedBlocksmithQuery,
		Logger:                         logger,
		SortedBlocksmithsMap:           make(map[string]*int64),
		Chaintype:                      &chaintype.MainChain{},
		candidates:                     make([]Candidate, 0),
		me:                             Me{},
		CurrentNodePublicKey:           currentNodePublicKey,
		ActiveNodeRegistryCacheStorage: activeNodeRegistryCacheStorage,
	}
}

func (bss *BlocksmithStrategyMain) isMe(lastCandidate Candidate) bool {
	var (
		now = time.Now().Unix()
	)

	if now > lastCandidate.StartTime && bytes.Equal(lastCandidate.Blocksmith.NodePublicKey, bss.CurrentNodePublicKey) {
		return true
	}
	return false
}

func (bss *BlocksmithStrategyMain) WillSmith(prevBlock *model.Block) (lastBlockID, blocksmithIndex int64, err error) {
	var (
		lastCandidate Candidate
		now           = time.Now().Unix()
	)

	fmt.Printf("\n==============================\nWillSmith... at %v\n", time.Now())

	if prevBlock.BlockHash != nil {
		bss.lastBlockHash = prevBlock.BlockHash
	}
	if !bytes.Equal(bss.lastBlockHash, prevBlock.BlockHash) {
		bss.candidates = []Candidate{}
		bss.me = Me{}

		blockSeedBigInt := new(big.Int).SetBytes(prevBlock.BlockSeed)
		rand.Seed(blockSeedBigInt.Int64())
	}

	// check Me first
	if now < bss.me.ExpiryTime {
		return bss.me.LastBlockID, bss.me.BlocksmithIndex, nil
	}

	candidatesCount := len(bss.candidates)
	fmt.Printf("\tcandidatesCount: %v\n", candidatesCount)

	if candidatesCount > 0 {
		lastCandidate = bss.candidates[candidatesCount-1]
		isMe := bss.isMe(lastCandidate)

		if !isMe {
			return 0, 0, errors.New("isMeFalse")
		}

		if now < lastCandidate.StartTime {
			return 0, 0, errors.New("invalidStartTime")
		}

		if err := bss.AddCandidate(prevBlock); err != nil {
			return 0, 0, errors.New("failedAddCandidate")
		}

	} else {
		if err := bss.AddCandidate(prevBlock); err != nil {
			return 0, 0, errors.New("failedAddCandidate")
		} else {
			lastCandidate = bss.candidates[0]
		}
	}
	if now < lastCandidate.ExpiryTime {
		fmt.Printf("\tsmithing is true  at %v\n", time.Now())
		bss.me = Me{
			ExpiryTime:      lastCandidate.ExpiryTime,
			BlocksmithIndex: int64(candidatesCount) - 1,
			LastBlockID:     prevBlock.ID,
		}

		return prevBlock.ID, int64(candidatesCount) - 1, nil
	}

	return 0, 0, errors.New("invalidExpiryTime")
}

func (bss *BlocksmithStrategyMain) AddCandidate(prevBlock *model.Block) error {
	var (
		activeNodeRegistry []storage.NodeRegistry
		candidate          Candidate
		now                = time.Now().Unix()
		err                error
	)

	// get node registry
	err = bss.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return err
	}

	activeNodeRegistryCount := len(activeNodeRegistry)
	round := float64(1)
	gap := float64(now - prevBlock.Timestamp)
	if gap > 15 {
		round += math.Floor(gap / 10)
	}
	for i := float64(0); i < round; i++ {
		idx := rand.Intn(activeNodeRegistryCount)

		blockSmith := model.Blocksmith{
			NodeID:        activeNodeRegistry[idx].Node.GetNodeID(),
			NodePublicKey: activeNodeRegistry[idx].Node.GetNodePublicKey(),
		}
		startTime := prevBlock.Timestamp + bss.Chaintype.GetSmithingPeriod() + int64(len(bss.candidates))*bss.Chaintype.GetBlocksmithTimeGap()
		expiryTime := startTime + bss.Chaintype.GetBlocksmithNetworkTolerance() + bss.Chaintype.GetBlocksmithBlockCreationTime()
		candidate = Candidate{
			Blocksmith: &blockSmith,
			StartTime:  startTime,
			ExpiryTime: expiryTime,
		}
		bss.candidates = append(bss.candidates, candidate)
	}

	return nil
}

func (bss *BlocksmithStrategyMain) CalculateCumulativeDifficulty(prevBlock, block *model.Block) string {
	round := bss.GetSmithingRound(prevBlock, block)
	currentCumulativeDifficulty := constant.CumulativeDifficultyDivisor / int64(round)
	return strconv.FormatInt(currentCumulativeDifficulty, 16)
}

func (bss *BlocksmithStrategyMain) IsBlockValid(prevBlock, block *model.Block) error {
	var (
		activeNodeRegistry []storage.NodeRegistry
		err                error
		blockSmiths        []*model.Blocksmith
	)

	// get node registry
	err = bss.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return err
	}

	for _, anr := range activeNodeRegistry {
		blockSmiths = append(blockSmiths, &model.Blocksmith{
			NodeID:        anr.Node.GetNodeID(),
			NodePublicKey: anr.Node.GetNodePublicKey(),
		})
	}

	round := bss.GetSmithingRound(prevBlock, block)

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

func (bss *BlocksmithStrategyMain) CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error {
	round := bss.GetSmithingRound(previousBlock, block)
	if round <= 1 {
		return nil
	}
	blocksmithBaseTime := bss.Chaintype.GetSmithingPeriod() + bss.Chaintype.GetBlocksmithBlockCreationTime() + bss.Chaintype.GetBlocksmithNetworkTolerance()
	previousExpiryTimestamp := previousBlock.GetTimestamp() + blocksmithBaseTime + int64(round-1)*bss.Chaintype.GetBlocksmithTimeGap()
	currentExpiryTimestamp := previousExpiryTimestamp + bss.Chaintype.GetBlocksmithTimeGap()
	if timestamp > previousExpiryTimestamp && timestamp < previousExpiryTimestamp+currentExpiryTimestamp {
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "%s-PendingPersist", bss.Chaintype.GetName())
}

// GetSkippedBlocksmiths return the list of skipped blocksmiths
// previousBlock must be latest last block since we'll be fetching registered nodes from cache.
func (bss *BlocksmithStrategyMain) GetBlocksBlocksmiths(previousBlock, block *model.Block) ([]*model.Blocksmith, error) {
	var (
		activeNodeRegistry []storage.NodeRegistry
		result             = make([]*model.Blocksmith, 0)
		err                error
	)
	// get node registry
	err = bss.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return nil, err
	}
	// get round
	round := bss.GetSmithingRound(previousBlock, block)
	blockSeedBigInt := new(big.Int).SetBytes(previousBlock.BlockSeed)
	rand.Seed(blockSeedBigInt.Int64())
	for i := 0; i < round; i++ {
		skippedNodeIdx := rand.Intn(len(activeNodeRegistry))
		result = append(result, &model.Blocksmith{
			NodeID:        activeNodeRegistry[skippedNodeIdx].Node.GetNodeID(),
			NodePublicKey: activeNodeRegistry[skippedNodeIdx].Node.GetNodePublicKey(),
		})
	}
	return result, nil
}

func (bss *BlocksmithStrategyMain) GetSmithingRound(previousBlock, block *model.Block) int {
	var (
		round = 1 // round start from 1
	)

	timeGap := block.GetTimestamp() - previousBlock.GetTimestamp()
	firstBlocksmithTime := bss.Chaintype.GetSmithingPeriod() + bss.Chaintype.GetBlocksmithBlockCreationTime() + bss.Chaintype.GetBlocksmithNetworkTolerance()
	if timeGap < firstBlocksmithTime {
		return round // first blocksmith
	}
	afterFirstBlocksmith := math.Ceil(float64(timeGap-firstBlocksmithTime) / float64(bss.Chaintype.GetBlocksmithTimeGap()))
	round += int(afterFirstBlocksmith)
	return round
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
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	copy(result, bss.SortedBlocksmiths)
	return result
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
