package strategy

import (
	"bytes"
	"encoding/binary"
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
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
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
		CurrentNodePublicKey  []byte
		Chaintype             chaintype.ChainType
		candidates            []Candidate
		lastBlockHash         []byte
	}
)

func NewBlocksmithStrategySpine(
	queryExecutor query.ExecutorInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	logger *log.Logger,
	spineBlockQuery query.BlockQueryInterface,
	currentNodePublicKey []byte,
) *BlocksmithStrategySpine {
	return &BlocksmithStrategySpine{
		QueryExecutor:        queryExecutor,
		SpinePublicKeyQuery:  spinePublicKeyQuery,
		Logger:               logger,
		SortedBlocksmithsMap: make(map[string]*int64),
		SpineBlockQuery:      spineBlockQuery,
		CurrentNodePublicKey: currentNodePublicKey,
		candidates:           make([]Candidate, 0),
	}
}

func (bss *BlocksmithStrategySpine) IsBlockValid(prevBlock, block *model.Block) error {
	var (
		err         error
		blockSmiths []*model.Blocksmith
	)
	blockSmiths, err = bss.GetBlocksmiths(prevBlock)
	if err != nil {
		return errors.New("ErrorGetBlocksmiths")
	}
	timeGap := block.Timestamp - prevBlock.Timestamp
	round := (timeGap - bss.Chaintype.GetSmithingPeriod()) / bss.Chaintype.GetBlocksmithTimeGap()

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

func (bss *BlocksmithStrategySpine) isMe(lastCandidate Candidate, block *model.Block) bool {
	var (
		now = time.Now().Unix()
	)
	if now > lastCandidate.StartTime && bytes.Equal(lastCandidate.Blocksmith.NodePublicKey, bss.CurrentNodePublicKey) {
		return true
	}
	return false
}

func (bss *BlocksmithStrategySpine) WillSmith(prevBlock *model.Block) (lastBlockID, blocksmithIndex int64, err error) {
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

func (bss *BlocksmithStrategySpine) CalculateCumulativeDifficulty(prevBlock, block *model.Block) string {
	return "0"
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
	for k, v := range bss.SortedBlocksmithsMap {
		result[k] = v
	}
	return result
}

// CalculateScore calculate the blocksmith score of spinechain
func (bss *BlocksmithStrategySpine) CalculateScore(generator *model.Blocksmith, score int64) error {
	// FIXME: ask @barton probably the way we compute spine blocksmith has to be reviewed, since we don't have ps and receipts,
	//		  attached to spine blocks
	generator.Score = big.NewInt(score / int64(constant.ScalarReceiptScore))
	return nil
}

func (*BlocksmithStrategySpine) CanPersistBlock(
	blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
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
