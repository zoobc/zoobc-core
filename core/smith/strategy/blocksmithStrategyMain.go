package strategy

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"math"
	"math/big"
	"math/rand"
	"time"
)

type (
	Candidate struct {
		Blocksmith *model.Blocksmith
		StartTime  int64
		ExpiryTime int64
		Index      int64
	}

	BlocksmithStrategyMain struct {
		QueryExecutor                  query.ExecutorInterface
		NodeRegistrationQuery          query.NodeRegistrationQueryInterface
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		Logger                         *log.Logger
		Chaintype                      chaintype.ChainType
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
)

func NewBlocksmithStrategyMain(
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	logger *log.Logger,
	currentNodePublicKey []byte,
	activeNodeRegistryCacheStorage storage.CacheStorageInterface,
	rng *crypto.RandomNumberGenerator,
	chaintype chaintype.ChainType,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		QueryExecutor:                  queryExecutor,
		NodeRegistrationQuery:          nodeRegistrationQuery,
		SkippedBlocksmithQuery:         skippedBlocksmithQuery,
		Logger:                         logger,
		Chaintype:                      chaintype,
		candidates:                     make([]Candidate, 0),
		CurrentNodePublicKey:           currentNodePublicKey,
		ActiveNodeRegistryCacheStorage: activeNodeRegistryCacheStorage,
		rng:                            rng,
		me:                             Candidate{},
	}
}

func (bss *BlocksmithStrategyMain) WillSmith(prevBlock *model.Block) (int64, error) {
	var (
		lastCandidate   Candidate
		now             = time.Now().Unix()
		err             error
		blocksmithIndex = int64(-1)
	)
	if !bytes.Equal(bss.lastBlockHash, prevBlock.BlockHash) {
		bss.lastBlockHash = prevBlock.BlockHash
		bss.candidates = []Candidate{}
		bss.me = Candidate{}
		err = bss.rng.Reset(constant.BlocksmithSelectionSeedPrefix, prevBlock.BlockSeed)
		if err != nil {
			return blocksmithIndex, err
		}
	}
	if len(bss.candidates) > 0 {
		lastCandidate = bss.candidates[len(bss.candidates)-1]
		if now < lastCandidate.StartTime {
			return blocksmithIndex, errors.New("WillSmith:NowLessThanStartTime")
		}
	}

	if now >= lastCandidate.StartTime {
		if err := bss.AddCandidate(prevBlock); err != nil {
			return blocksmithIndex, err
		}
	}

	if bss.me.StartTime != 0 && now >= bss.me.StartTime && now < bss.me.ExpiryTime {
		return bss.me.Index, nil
	}
	return blocksmithIndex, errors.New("invalidExpiryTime")
}

func (bss *BlocksmithStrategyMain) convertRandomNumberToIndex(randNumber, activeNodeRegistryCount int64) int {
	rd := randNumber / activeNodeRegistryCount
	mult := rd * activeNodeRegistryCount
	rem := randNumber - mult
	return int(rem)
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
	round := bss.GetSmithingRound(prevBlock, &model.Block{Timestamp: now})
	currCandidateCount := len(bss.candidates)
	newCandidateCount := currCandidateCount
	for i := 0; i < round-currCandidateCount; i++ {
		var (
			idx        int
			randNumber int64
		)
		randNumber = bss.rng.Next()
		idx = bss.convertRandomNumberToIndex(randNumber, int64(activeNodeRegistryCount))
		blockSmith := model.Blocksmith{
			NodeID:        activeNodeRegistry[idx].Node.GetNodeID(),
			NodePublicKey: activeNodeRegistry[idx].Node.GetNodePublicKey(),
		}
		startTime := prevBlock.Timestamp + bss.Chaintype.GetSmithingPeriod() + int64(newCandidateCount)*bss.Chaintype.GetBlocksmithTimeGap()
		expiryTime := startTime + bss.Chaintype.GetBlocksmithNetworkTolerance() + bss.Chaintype.GetBlocksmithBlockCreationTime()
		candidate = Candidate{
			Blocksmith: &blockSmith,
			StartTime:  startTime,
			ExpiryTime: expiryTime,
			Index:      int64(newCandidateCount),
		}
		if bytes.Equal(candidate.Blocksmith.NodePublicKey, bss.CurrentNodePublicKey) {
			// set self as candidate if found same node public key
			bss.me = candidate
		}
		bss.candidates = append(bss.candidates, candidate)
		newCandidateCount++
	}
	return nil
}

func (bss *BlocksmithStrategyMain) CalculateCumulativeDifficulty(prevBlock, block *model.Block) string {
	round := bss.GetSmithingRound(prevBlock, block)
	prevCummulativeDiff, _ := new(big.Int).SetString(prevBlock.GetCumulativeDifficulty(), 10)
	currentCumulativeDifficulty := new(big.Int).SetInt64(constant.CumulativeDifficultyDivisor / int64(round))
	newCummulativeDifficulty := new(big.Int).Add(prevCummulativeDiff, currentCumulativeDifficulty)
	return newCummulativeDifficulty.String()
}

func (bss *BlocksmithStrategyMain) IsBlockValid(prevBlock, block *model.Block) error {
	var (
		activeNodeRegistry []storage.NodeRegistry
		err                error
	)
	// get node registry
	err = bss.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return err
	}

	round := bss.GetSmithingRound(prevBlock, block)
	rng := crypto.NewRandomNumberGenerator()
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, prevBlock.BlockSeed)
	if err != nil {
		return err
	}
	var (
		validRandomNumbers []int64
		idx                int
	)
	// check for n-previous round also if round > 1
	gap := bss.Chaintype.GetBlocksmithNetworkTolerance() + bss.Chaintype.GetBlocksmithBlockCreationTime()
	validNumberOfRounds := 1 + gap/bss.Chaintype.GetBlocksmithTimeGap()
	for i := 0; i < round; i++ {
		randomNumber := rng.Next()
		if int64(i) > (int64(round) - validNumberOfRounds) {
			validRandomNumbers = append(validRandomNumbers, randomNumber)
		}
	}
	for i := 0; i < len(validRandomNumbers); i++ {
		idx = bss.convertRandomNumberToIndex(validRandomNumbers[i], int64(len(activeNodeRegistry)))
		if bytes.Equal(activeNodeRegistry[idx].Node.NodePublicKey, block.BlocksmithPublicKey) {
			return nil
		}
	}
	return errors.New("IsBlockValid:Failed-InvalidSmithingTime")
}

func (bss *BlocksmithStrategyMain) CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error {
	var (
		activeNodeRegistry []storage.NodeRegistry
		err                error
	)
	// get node registry
	err = bss.ActiveNodeRegistryCacheStorage.GetAllItems(&activeNodeRegistry)
	if err != nil {
		return err
	}

	blocksmithIndex, _ := bss.GetSmithingIndex(previousBlock, block, activeNodeRegistry)
	if blocksmithIndex <= 1 {
		return nil
	}
	previousExpiryTimestamp := previousBlock.GetTimestamp() + bss.Chaintype.GetSmithingPeriod() +
		bss.Chaintype.GetBlocksmithBlockCreationTime() + bss.Chaintype.GetBlocksmithNetworkTolerance() +
		int64(blocksmithIndex-1)*bss.Chaintype.GetBlocksmithTimeGap()
	currentExpiryTimestamp := previousExpiryTimestamp + bss.Chaintype.GetBlocksmithTimeGap()
	if timestamp > previousExpiryTimestamp && timestamp < currentExpiryTimestamp {
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "%s-PendingPersist", bss.Chaintype.GetName())
}

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

func (bss *BlocksmithStrategyMain) GetSmithingIndex(
	previousBlock, block *model.Block, activeRegistries []storage.NodeRegistry,
) (int, error) {
	var (
		round = 1 // round start from 1
		err   error
	)
	timeGap := block.GetTimestamp() - previousBlock.GetTimestamp()
	if timeGap < bss.Chaintype.GetSmithingPeriod()+bss.Chaintype.GetBlocksmithTimeGap() {
		return 0, nil // first blocksmith
	}

	afterFirstBlocksmith := math.Floor(float64(timeGap-bss.Chaintype.GetSmithingPeriod()) / float64(bss.Chaintype.GetBlocksmithTimeGap()))
	round += int(afterFirstBlocksmith)
	rng := crypto.NewRandomNumberGenerator()
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, previousBlock.BlockSeed)
	if err != nil {
		return 0, err
	}

	for i := 0; i < round; i++ {
		randomNumber := rng.Next()
		idx := bss.convertRandomNumberToIndex(randomNumber, int64(len(activeRegistries)))
		if bytes.Equal(activeRegistries[idx].Node.GetNodePublicKey(), block.GetBlocksmithPublicKey()) {
			return i, nil
		}
	}
	return 0, blocker.NewBlocker(blocker.ValidationErr, "GetSmithingIndex:BlocksmithNotFound")
}

func (bss *BlocksmithStrategyMain) GetSmithingRound(previousBlock, block *model.Block) int {
	var (
		round = 1 // round start from 1
	)
	timeGap := block.GetTimestamp() - previousBlock.GetTimestamp()
	if timeGap < bss.Chaintype.GetSmithingPeriod()+bss.Chaintype.GetBlocksmithTimeGap() {
		return round // first blocksmith
	}
	afterFirstBlocksmith := math.Floor(float64(timeGap-bss.Chaintype.GetSmithingPeriod()) / float64(bss.Chaintype.GetBlocksmithTimeGap()))
	round += int(afterFirstBlocksmith)
	return round
}
