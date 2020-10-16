package service

import (
	"github.com/montanaflynn/stats"
	"github.com/zoobc/zoobc-core/common/crypto"
	"math"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	CoinbaseServiceInterface interface {
		GetCoinbase(blockTimesatamp, previousBlockTimesatamp int64) int64
		CoinbaseLotteryWinners(
			activeNodeRegistries []storage.NodeRegistry,
			scoreSum, blockTimestamp int64,
			previousBlock *model.Block,
		) ([][]byte, error)
	}

	CoinbaseService struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		Chaintype             chaintype.ChainType
		Rng                   *crypto.RandomNumberGenerator
	}
)

func NewCoinbaseService(
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	queryExecutor query.ExecutorInterface,
	chaintype chaintype.ChainType,
	rng *crypto.RandomNumberGenerator,
) *CoinbaseService {
	return &CoinbaseService{
		NodeRegistrationQuery: nodeRegistrationQuery,
		QueryExecutor:         queryExecutor,
		Chaintype:             chaintype,
		Rng:                   rng,
	}
}

// GetCoinbase return the value of coinbase / new coins that will be created every block.
func (cbs *CoinbaseService) GetCoinbase(blockTimesatamp, previousBlockTimesatamp int64) int64 {
	return cbs.GetTotalDistribution(blockTimesatamp) - cbs.GetTotalDistribution(previousBlockTimesatamp)
}

// GetTotalDistribution get number of token that should be distributed by given timestamp
func (cbs *CoinbaseService) GetTotalDistribution(blockTimestamp int64) int64 {
	var (
		coinbaseSigmoidMin float64 = 1 / (1 + math.Exp(-constant.CoinbaseSigmoidStart))
		coinbaseSigmoidMax float64 = 1 / (1 + math.Exp(-constant.CoinbaseSigmoidEnd))
		// t is ranges from 0.0 at the genesis, to 1.0 after CoinbaseTime
		// err occur only when the length input is 0
		t, _ = stats.Min(stats.Float64Data{
			1,
			float64(blockTimestamp-cbs.Chaintype.GetGenesisBlockTimestamp()) / float64(constant.CoinbaseTime),
		})

		// x ranges from CoinbaseSigmoidStart at the genesis, to CoinbaseSigmoidEnd after coinbaseTime,
		x float64 = (t * (constant.CoinbaseSigmoidEnd - constant.CoinbaseSigmoidStart)) + constant.CoinbaseSigmoidStart
		// y is ranges from 0.0 at the genesis, to 1.0 after coinbaseTime,
		y float64 = ((1 / (1 + (math.Exp(-x)))) - coinbaseSigmoidMin) * (1.0 / (coinbaseSigmoidMax - coinbaseSigmoidMin))
	)
	return int64(math.Floor(y * float64(constant.CoinbaseTotalDistribution)))
}

// CoinbaseLotteryWinners get the current list of blocksmiths, duplicate it (to not change the original one)
// and sort it using the NodeOrder algorithm. The first n (n = constant.MaxNumBlocksmithRewards) in the newly ordered list
// are the coinbase lottery winner (the blocksmiths that will be rewarded for the current block)
func (cbs *CoinbaseService) CoinbaseLotteryWinners(
	activeRegistries []storage.NodeRegistry,
	scoreSum, blockTimestamp int64,
	previousBlock *model.Block,
) ([][]byte, error) {

	var (
		selectedAccounts [][]byte
		numRewards       int64
	)
	err := cbs.Rng.Reset(constant.CoinbaseSelectionSeedPrefix, previousBlock.GetBlockSeed())
	if err != nil {
		return nil, err
	}
	activeRegistryLength := len(activeRegistries)

	// get number of rewards recipients
	numRewards = (blockTimestamp - previousBlock.Timestamp) * constant.CoinbaseNumberRewardsPerSecond
	numRewards = util.MinInt64(numRewards, constant.CoinbaseMaxNumberRewardsPerBlock)
	numRewards = util.MinInt64(numRewards, int64(activeRegistryLength))
	for i := 0; i < int(numRewards); i++ {
		var (
			tempPreviousSum int64
			rawRandomNumber = cbs.Rng.Next()
			// scale down random number to [0-scoreSum]
			rounds             = rawRandomNumber / scoreSum
			roundWithRemainder = rounds * scoreSum
			winnerScore        = rawRandomNumber - roundWithRemainder
		)

		for j := 0; j < activeRegistryLength; j++ {
			participationScore := int64(
				math.Floor(float64(activeRegistries[j].ParticipationScore) / float64(activeRegistryLength)))
			if winnerScore >= tempPreviousSum && winnerScore < tempPreviousSum+participationScore {
				selectedAccounts = append(selectedAccounts, activeRegistries[j].Node.AccountAddress)
				break
			}
			tempPreviousSum += participationScore
		}
	}
	return selectedAccounts, nil
}
