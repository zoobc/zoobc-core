package strategy

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	BlocksmithStrategySpine struct {
		QueryExecutor                  query.ExecutorInterface
		SpinePublicKeyQuery            query.SpinePublicKeyQueryInterface
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                         *log.Logger
		SortedBlocksmiths              []*model.Blocksmith
		LastSortedBlockID              int64
		SortedBlocksmithsLock          sync.RWMutex
		SortedBlocksmithsMap           map[string]*int64
		SpineBlockQuery                query.BlockQueryInterface
		CurrentNodePublicKey           []byte
		Chaintype                      chaintype.ChainType
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
)

func NewBlocksmithStrategySpine(
	queryExecutor query.ExecutorInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	logger *log.Logger,
	spineBlockQuery query.BlockQueryInterface,
	currentNodePublicKey []byte,
	activeNodeRegistryCacheStorage storage.CacheStorageInterface,
	rng *crypto.RandomNumberGenerator,
) *BlocksmithStrategySpine {
	return &BlocksmithStrategySpine{
		QueryExecutor:                  queryExecutor,
		SpinePublicKeyQuery:            spinePublicKeyQuery,
		Logger:                         logger,
		SortedBlocksmithsMap:           make(map[string]*int64),
		SpineBlockQuery:                spineBlockQuery,
		CurrentNodePublicKey:           currentNodePublicKey,
		Chaintype:                      &chaintype.SpineChain{},
		candidates:                     make([]Candidate, 0),
		me:                             Candidate{},
		ActiveNodeRegistryCacheStorage: activeNodeRegistryCacheStorage,
		rng:                            rng,
	}
}

func (bss *BlocksmithStrategySpine) WillSmith(prevBlock *model.Block) (int64, error) {
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

func (bss *BlocksmithStrategySpine) convertRandomNumberToIndex(randNumber, activeNodeRegistryCount int64) int {
	rd := randNumber / activeNodeRegistryCount
	mult := rd * activeNodeRegistryCount
	rem := randNumber - mult
	return int(rem)
}

func (bss *BlocksmithStrategySpine) AddCandidate(prevBlock *model.Block) error {
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

func (bss *BlocksmithStrategySpine) CalculateCumulativeDifficulty(prevBlock, block *model.Block) string {
	round := bss.GetSmithingRound(prevBlock, block)
	prevCummulativeDiff, _ := new(big.Int).SetString(prevBlock.GetCumulativeDifficulty(), 10)
	currentCumulativeDifficulty := new(big.Int).SetInt64(constant.CumulativeDifficultyDivisor / int64(round))
	newCummulativeDifficulty := new(big.Int).Add(prevCummulativeDiff, currentCumulativeDifficulty)
	return newCummulativeDifficulty.String()
}

func (bss *BlocksmithStrategySpine) IsBlockValid(prevBlock, block *model.Block) error {
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

func (bss *BlocksmithStrategySpine) CanPersistBlock(previousBlock, block *model.Block, timestamp int64) error {
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

// GetSkippedBlocksmiths return the list of skipped blocksmiths
// previousBlock must be latest last block since we'll be fetching registered nodes from cache.
func (bss *BlocksmithStrategySpine) GetBlocksBlocksmiths(previousBlock, block *model.Block) ([]*model.Blocksmith, error) {
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

func (bss *BlocksmithStrategySpine) GetSmithingIndex(
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

func (bss *BlocksmithStrategySpine) GetSmithingRound(previousBlock, block *model.Block) int {
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

func (bss *BlocksmithStrategySpine) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	bss.SortedBlocksmithsLock.RLock()
	defer bss.SortedBlocksmithsLock.RUnlock()
	var result = make([]*model.Blocksmith, len(bss.SortedBlocksmiths))
	copy(result, bss.SortedBlocksmiths)
	return result
}

// CalculateScore calculate the blocksmith score
func (bss *BlocksmithStrategySpine) CalculateScore(generator *model.Blocksmith, score int64) error {
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	return nil
}

func (bss *BlocksmithStrategySpine) IsValidSmithTime(blocksmithIndex, numberOfBlocksmiths int64, previousBlock *model.Block) error {
	var (
		currentTime                      = time.Now().Unix()
		ct                               = &chaintype.SpineChain{}
		prevRoundBegin, prevRoundExpired int64
	)
	// avoid division by zero in case there are no blocksmiths in the network (edge case)
	if numberOfBlocksmiths < 1 {
		return blocker.NewBlocker(blocker.SmithingPending, "NoBlockSmiths")
	}
	// calculate total time before every blocksmiths are skipped
	timeForOneRound := numberOfBlocksmiths * ct.GetBlocksmithTimeGap()
	timeSinceLastBlock := currentTime - previousBlock.GetTimestamp()

	if timeSinceLastBlock < ct.GetSmithingPeriod() {
		return blocker.NewBlocker(blocker.SmithingPending, "SmithingPending")
	}
	modTimeSinceLastBlock := timeSinceLastBlock - ct.GetSmithingPeriod()
	timeRound := math.Floor(float64(modTimeSinceLastBlock) / float64(timeForOneRound))
	if timeForOneRound <= 0 || numberOfBlocksmiths <= 0 {

		return blocker.NewBlocker(blocker.SmithingPending, "NUmberOfBlockSmithsLessThanWhatIsNeeded")
	}
	remainder := modTimeSinceLastBlock % timeForOneRound
	nearestRoundBeginning := currentTime - remainder
	if timeRound > 0 { // if more than one round has passed, calculate previous round start-expiry time for overlap
		prevRoundStart := nearestRoundBeginning - timeForOneRound
		prevRoundBegin = prevRoundStart + blocksmithIndex*ct.GetBlocksmithTimeGap()
		prevRoundExpired = prevRoundBegin + ct.GetBlocksmithBlockCreationTime() +
			ct.GetBlocksmithNetworkTolerance()
	}
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
