package service

import (
	"database/sql"
	"math"
	"math/big"
	"sort"

	"github.com/montanaflynn/stats"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	CoinbaseServiceInterface interface {
		GetCoinbase(blockTimesatamp, previousBlockTimesatamp int64) int64
		CoinbaseLotteryWinners(
			blocksmiths []*model.Blocksmith,
			blockTimestamp int64,
			previousBlockTimestamp int64,
		) ([]string, error)
	}

	CoinbaseService struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		Chaintype             chaintype.ChainType
	}
)

func NewCoinbaseService(
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	queryExecutor query.ExecutorInterface,
	chaintype chaintype.ChainType,
) *CoinbaseService {
	return &CoinbaseService{
		NodeRegistrationQuery: nodeRegistrationQuery,
		QueryExecutor:         queryExecutor,
		Chaintype:             chaintype,
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
	blocksmiths []*model.Blocksmith,
	blockTimestamp,
	previousBlockTimestamp int64,
) ([]string, error) {
	var (
		selectedAccounts []string
		numRewards       int64
		qry              string
		qryArgs          []interface{}
		row              *sql.Row
		err              error
		nodeRegistration model.NodeRegistration
	)
	// copy the pointer array to not change original order

	// sort blocksmiths by NodeOrder
	sort.SliceStable(blocksmiths, func(i, j int) bool {
		bi, bj := blocksmiths[i], blocksmiths[j]
		res := bi.NodeOrder.Cmp(bj.NodeOrder)
		if res == 0 {
			// compare node ID
			nodePKI := new(big.Int).SetUint64(uint64(bi.NodeID))
			nodePKJ := new(big.Int).SetUint64(uint64(bj.NodeID))
			res = nodePKI.Cmp(nodePKJ)
		}
		// ascending sort
		return res < 0
	})

	// get number of rewards recipients
	numRewards = (blockTimestamp - previousBlockTimestamp) * constant.CoinbaseNumberRewardsPerSecond
	numRewards = util.MinInt64(numRewards, constant.CoinbaseMaxNumberRewardsPerBlock)
	numRewards = util.MinInt64(numRewards, int64(len(blocksmiths)))

	for idx, sortedBlockSmith := range blocksmiths {
		if idx >= int(numRewards) {
			break
		}
		// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
		qry, qryArgs = cbs.NodeRegistrationQuery.GetNodeRegistrationByID(sortedBlockSmith.NodeID)
		row, err = cbs.QueryExecutor.ExecuteSelectRow(qry, false, qryArgs...)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		err = cbs.NodeRegistrationQuery.Scan(&nodeRegistration, row)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, blocker.NewBlocker(blocker.DBErr, "CoinbaseLotteryNodeRegistrationNotFound")
			}
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		selectedAccounts = append(selectedAccounts, nodeRegistration.AccountAddress)
	}
	return selectedAccounts, nil
}
