package service

import (
	"math/big"
	"sort"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	CoinbaseServiceInterface interface {
		GetCoinbase() int64
		CoinbaseLotteryWinners(blocksmiths []*model.Blocksmith) ([]string, error)
	}

	CoinbaseService struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewCoinbaseService(
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	queryExecutor query.ExecutorInterface,
) *CoinbaseService {
	return &CoinbaseService{
		NodeRegistrationQuery: nodeRegistrationQuery,
		QueryExecutor:         queryExecutor,
	}
}

// GetCoinbase return the value of coinbase / new coins that will be created every block.
func (*CoinbaseService) GetCoinbase() int64 {
	return 50 * constant.OneZBC
}

// CoinbaseLotteryWinners get the current list of blocksmiths, duplicate it (to not change the original one)
// and sort it using the NodeOrder algorithm. The first n (n = constant.MaxNumBlocksmithRewards) in the newly ordered list
// are the coinbase lottery winner (the blocksmiths that will be rewarded for the current block)
func (cbs *CoinbaseService) CoinbaseLotteryWinners(blocksmiths []*model.Blocksmith) ([]string, error) {
	var (
		selectedAccounts []string
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

	for idx, sortedBlockSmith := range blocksmiths {
		if idx > constant.MaxNumBlocksmithRewards-1 {
			break
		}
		// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
		qry, args := cbs.NodeRegistrationQuery.GetNodeRegistrationByID(sortedBlockSmith.NodeID)
		rows, err := cbs.QueryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		nr, err := cbs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if (err != nil) || len(nr) == 0 {
			rows.Close()
			return nil, blocker.NewBlocker(blocker.DBErr, "CoinbaseLotteryNodeRegistrationNotFound")
		}
		selectedAccounts = append(selectedAccounts, nr[0].AccountAddress)
		rows.Close()
	}
	return selectedAccounts, nil
}
