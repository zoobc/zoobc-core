// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package strategy

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"math"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	// Candidate represent single blocksmith that may create the next block
	Candidate struct {
		Blocksmith *model.Blocksmith
		StartTime  int64
		ExpiryTime int64
		Index      int64
	}

	BlocksmithStrategyMain struct {
		Chaintype                      chaintype.ChainType
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		BlockQuery                     query.BlockQueryInterface
		QueryExecutor                  query.ExecutorInterface
		BlockCacheStorage              storage.CacheStackStorageInterface
		Logger                         *log.Logger
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
)

func NewBlocksmithStrategyMain(
	logger *log.Logger,
	currentNodePublicKey []byte,
	activeNodeRegistryCacheStorage storage.CacheStorageInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	blockQuery query.BlockQueryInterface,
	blockCacheStorage storage.CacheStackStorageInterface,
	queryExecutor query.ExecutorInterface,
	rng *crypto.RandomNumberGenerator,
	chaintype chaintype.ChainType,
) *BlocksmithStrategyMain {
	return &BlocksmithStrategyMain{
		Logger:                         logger,
		Chaintype:                      chaintype,
		CurrentNodePublicKey:           currentNodePublicKey,
		ActiveNodeRegistryCacheStorage: activeNodeRegistryCacheStorage,
		QueryExecutor:                  queryExecutor,
		SkippedBlocksmithQuery:         skippedBlocksmithQuery,
		BlockQuery:                     blockQuery,
		BlockCacheStorage:              blockCacheStorage,
		me:                             Candidate{},
		candidates:                     make([]Candidate, 0),
		rng:                            rng,
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

func (bss *BlocksmithStrategyMain) estimatePreviousBlockPersistTime(lastBlock *model.Block) (int64, error) {
	var (
		numberOfSkippedBlocksmith int
		result                    int64
		err                       error
	)

	if lastBlock.GetHeight() < 1 || bss.Chaintype.GetTypeInt() == (&chaintype.SpineChain{}).GetTypeInt() {
		// no need to estimate persist time if previous block is genesis
		return lastBlock.GetTimestamp(), nil
	}
	previousLastBlock, err := func(lastBlock *model.Block) (*model.Block, error) {
		// get previous.height - 1 block to estimate persist time
		previousLastBlockObj, err := util.GetBlockByHeightUseBlocksCache(
			lastBlock.GetHeight()-1,
			bss.QueryExecutor,
			bss.BlockQuery,
			bss.BlockCacheStorage,
		)
		if err != nil {
			return nil, err
		}
		return &model.Block{
			ID:        previousLastBlockObj.ID,
			Height:    previousLastBlockObj.Height,
			Timestamp: previousLastBlockObj.Timestamp,
			BlockHash: previousLastBlockObj.BlockHash,
		}, nil
	}(lastBlock)

	if err != nil {
		return lastBlock.GetTimestamp(), err
	}
	firstBlocksmithExpiryTime := previousLastBlock.GetTimestamp() + bss.Chaintype.GetSmithingPeriod() +
		bss.Chaintype.GetBlocksmithBlockCreationTime() +
		bss.Chaintype.GetBlocksmithNetworkTolerance()

	qry := bss.SkippedBlocksmithQuery.GetNumberOfSkippedBlocksmithsByBlockHeight(lastBlock.GetHeight())
	rows, err := bss.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&numberOfSkippedBlocksmith)
		if err != nil {
			if err != sql.ErrNoRows {
				return result, err
			}
		}
	}

	if numberOfSkippedBlocksmith > 0 {
		result = firstBlocksmithExpiryTime + (int64(numberOfSkippedBlocksmith-1) * bss.Chaintype.GetBlocksmithTimeGap())
	} else {
		result = lastBlock.GetTimestamp()
	}
	return result, nil
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
	round, err := bss.GetSmithingRound(prevBlock, &model.Block{Timestamp: now})
	if err != nil {
		return err
	}
	currCandidateCount := len(bss.candidates)
	newCandidateCount := currCandidateCount
	lastBlockEstimatedPersistTime, err := bss.estimatePreviousBlockPersistTime(prevBlock)
	if err != nil {
		return err
	}
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
		startTime := lastBlockEstimatedPersistTime +
			bss.Chaintype.GetSmithingPeriod() + int64(newCandidateCount)*bss.Chaintype.GetBlocksmithTimeGap()
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

func (bss *BlocksmithStrategyMain) CalculateCumulativeDifficulty(prevBlock, block *model.Block) (string, error) {
	// all blocksmith up to current blocksmith
	blocksmiths, err := bss.GetBlocksBlocksmiths(prevBlock, block)
	if err != nil {
		return "", err
	}
	prevCummulativeDiff, _ := new(big.Int).SetString(prevBlock.GetCumulativeDifficulty(), 10)
	currentCumulativeDifficulty := new(big.Int).SetInt64(constant.CumulativeDifficultyDivisor / int64(len(blocksmiths)))

	newCummulativeDifficulty := new(big.Int).Add(prevCummulativeDiff, currentCumulativeDifficulty)
	return newCummulativeDifficulty.String(), nil
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

	round, err := bss.GetSmithingRound(prevBlock, block)
	if err != nil {
		return err
	}
	rng := crypto.NewRandomNumberGenerator()
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, prevBlock.BlockSeed)
	if err != nil {
		return err
	}
	var (
		validRandomNumbers []int64
		idx                int
	)
	// check for n-previous round also if round > 1, this will check if block come from valid blocksmith
	gap := bss.Chaintype.GetBlocksmithNetworkTolerance() + bss.Chaintype.GetBlocksmithBlockCreationTime()
	validNumberOfRounds := 1 + gap/bss.Chaintype.GetBlocksmithTimeGap()
	for i := 0; i < round; i++ {
		randomNumber := rng.Next()
		if int64(i) >= (int64(round) - validNumberOfRounds) {
			validRandomNumbers = append(validRandomNumbers, randomNumber)
		}
	}
	for i := 0; i < len(validRandomNumbers); i++ {
		idx = bss.convertRandomNumberToIndex(validRandomNumbers[i], int64(len(activeNodeRegistry)))
		if bytes.Equal(activeNodeRegistry[idx].Node.NodePublicKey, block.BlocksmithPublicKey) {
			startTime, endTime, err := bss.getValidBlockCreationTime(prevBlock, round-len(validRandomNumbers)+(i+1))
			if err != nil {
				return err
			}
			// validate block's timestamp within persistable timestamp
			if block.GetTimestamp() >= startTime && block.GetTimestamp() < endTime {
				return nil
			}
		}
	}
	return errors.New("IsBlockValid:Failed-InvalidSmithingTime")
}

// getValidBlockPersistTime calculate the valid starting time (inclusive) and ending time (exclusive) for a block to be persisted
// exception for first blocksmith (1 round) don't need to wait until previous smithing (which do not exist) to be expired
// first
func (bss *BlocksmithStrategyMain) getValidBlockPersistTime(previousBlock *model.Block, round int) (start, end int64, err error) {
	offset := bss.Chaintype.GetBlocksmithBlockCreationTime() + bss.Chaintype.GetBlocksmithNetworkTolerance()
	if round <= 1 {
		startTime := previousBlock.GetTimestamp() + bss.Chaintype.GetSmithingPeriod()
		return startTime, startTime + offset, nil
	}
	firstRoundExpiry := bss.Chaintype.GetSmithingPeriod() + offset
	gaps := int64(round-1) * bss.Chaintype.GetBlocksmithTimeGap()
	estimatedPreviousBlockPersistTime, err := bss.estimatePreviousBlockPersistTime(previousBlock)
	if err != nil {
		return 0, 0, err
	}
	startTime := estimatedPreviousBlockPersistTime + firstRoundExpiry + gaps
	return startTime, startTime + bss.Chaintype.GetBlocksmithTimeGap(), nil
}

// getValidBlockCreationTime return the valid time to create block given previousBlock and round
func (bss *BlocksmithStrategyMain) getValidBlockCreationTime(previousBlock *model.Block, round int) (start, end int64, err error) {
	offset := bss.Chaintype.GetBlocksmithBlockCreationTime() + bss.Chaintype.GetBlocksmithNetworkTolerance()
	if round <= 1 {
		startTime := previousBlock.GetTimestamp() + bss.Chaintype.GetSmithingPeriod()
		return startTime, startTime + offset, nil
	}
	gaps := int64(round-1) * bss.Chaintype.GetBlocksmithTimeGap()
	estimatedPreviousBlockPersistTime, err := bss.estimatePreviousBlockPersistTime(previousBlock)
	if err != nil {
		return 0, 0, err
	}
	startTime := estimatedPreviousBlockPersistTime + bss.Chaintype.GetSmithingPeriod() + gaps
	return startTime, startTime + offset, nil
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

	blocksmithIndex, err := bss.GetSmithingIndex(previousBlock, block, activeNodeRegistry)
	if err != nil {
		return err
	}
	startTime, endTime, err := bss.getValidBlockPersistTime(previousBlock, blocksmithIndex+1)
	if err != nil {
		return err
	}
	if timestamp >= startTime && timestamp < endTime {
		return nil
	}
	return blocker.NewBlocker(blocker.ValidationErr, "%s-PendingPersist", bss.Chaintype.GetName())
}

// GetBlocksBlocksmiths fetch the blocksmiths candidate list up to block.BlocksmithPublicKey, if the block.BlocksmithPublicKey
// is first blocksmith then it returns only a single model.Blocksmith, otherwise it returns n-1 number of skipped blocksmith
// including (possibly) the block.BlocksmithPublicKey with the valid blocksmith at n-th index.
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
	round, err := bss.GetSmithingRound(previousBlock, block)
	if err != nil {
		return nil, err
	}
	rng := crypto.NewRandomNumberGenerator()
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, previousBlock.GetBlockSeed())
	if err != nil {
		return nil, err
	}
	var blocksmithIndex = -1

	for i := 0; i < round; i++ {
		randomNumber := rng.Next()
		skippedNodeIdx := bss.convertRandomNumberToIndex(randomNumber, int64(len(activeNodeRegistry)))
		result = append(result, &model.Blocksmith{
			NodeID:        activeNodeRegistry[skippedNodeIdx].Node.GetNodeID(),
			NodePublicKey: activeNodeRegistry[skippedNodeIdx].Node.GetNodePublicKey(),
			Score:         big.NewInt(activeNodeRegistry[skippedNodeIdx].ParticipationScore),
		})
		isValidBlocksmith := bytes.Equal(activeNodeRegistry[skippedNodeIdx].Node.GetNodePublicKey(), block.GetBlocksmithPublicKey())
		if isValidBlocksmith {
			blocksmithIndex = i
		}
		if i == round-1 && blocksmithIndex < 0 {
			return nil, blocker.NewBlocker(blocker.ValidationErr, "GetBlocksBlocksmith:BlocksmithNotInCandidates")
		}
	}
	return result[:blocksmithIndex+1], nil
}

func (bss *BlocksmithStrategyMain) GetSmithingIndex(
	previousBlock, block *model.Block, activeRegistries []storage.NodeRegistry,
) (int, error) {
	var (
		round = 1 // round start from 1
		err   error
	)
	rng := crypto.NewRandomNumberGenerator()
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, previousBlock.BlockSeed)
	if err != nil {
		return 0, err
	}

	previousBlockEstimatedPersistTime, err := bss.estimatePreviousBlockPersistTime(previousBlock)
	if err != nil {
		return 0, err
	}
	timeGap := block.GetTimestamp() - previousBlockEstimatedPersistTime
	if timeGap < bss.Chaintype.GetSmithingPeriod()+bss.Chaintype.GetBlocksmithTimeGap() {
		// first blocksmith, validate if blocksmith public key is valid
		randomNumber := rng.Next()
		idx := bss.convertRandomNumberToIndex(randomNumber, int64(len(activeRegistries)))
		if !bytes.Equal(activeRegistries[idx].Node.GetNodePublicKey(), block.GetBlocksmithPublicKey()) {
			return 0, blocker.NewBlocker(blocker.ValidationErr, "GetSmithingIndex:InvalidBlocksmithTime")
		}
		return 0, nil // first blocksmith
	}

	afterFirstBlocksmith := math.Floor(float64(timeGap-bss.Chaintype.GetSmithingPeriod()) / float64(bss.Chaintype.GetBlocksmithTimeGap()))
	round += int(afterFirstBlocksmith)
	lastIndex := -1
	for i := 0; i < round; i++ {
		randomNumber := rng.Next()
		idx := bss.convertRandomNumberToIndex(randomNumber, int64(len(activeRegistries)))
		if bytes.Equal(activeRegistries[idx].Node.GetNodePublicKey(), block.GetBlocksmithPublicKey()) {
			lastIndex = i
		}
	}
	if lastIndex > -1 {
		return lastIndex, nil
	}
	return 0, blocker.NewBlocker(blocker.ValidationErr, "GetSmithingIndex:BlocksmithNotFound")
}

func (bss *BlocksmithStrategyMain) GetSmithingRound(previousBlock, block *model.Block) (int, error) {
	var (
		round = 1 // round start from 1
	)
	previousEstimatedTime, err := bss.estimatePreviousBlockPersistTime(previousBlock)
	if err != nil {
		return round, err
	}
	timeGap := block.GetTimestamp() - previousEstimatedTime
	if timeGap < bss.Chaintype.GetSmithingPeriod()+bss.Chaintype.GetBlocksmithTimeGap() {
		return round, nil // first blocksmith
	}
	afterFirstBlocksmith := math.Floor(float64(timeGap-bss.Chaintype.GetSmithingPeriod()) / float64(bss.Chaintype.GetBlocksmithTimeGap()))
	round += int(afterFirstBlocksmith)
	return round, nil
}
