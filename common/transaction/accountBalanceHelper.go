package transaction

import (
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	AccountBalanceHelperInterface interface {
		AddAccountSpendableBalance(address string, amount int64) error
		AddAccountBalance(address string, amount int64, blockHeight uint32) error
	}

	AccountBalanceHelper struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
)

func NewAccountBalanceHelper(
	accountBalanceQuery query.AccountBalanceQueryInterface,
	queryExecutor query.ExecutorInterface,
) *AccountBalanceHelper {
	return &AccountBalanceHelper{
		AccountBalanceQuery: accountBalanceQuery,
		QueryExecutor:       queryExecutor,
	}
}

// AddAccountSpendableBalance add spendable_balance field to the address provided, must be executed inside db transaction
// scope
func (abh *AccountBalanceHelper) AddAccountSpendableBalance(address string, amount int64) error {
	accountBalanceSenderQ, accountBalanceSenderQArgs := abh.AccountBalanceQuery.AddAccountSpendableBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
		},
	)
	return abh.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)

}

// AddAccountBalance add balance and spendable_balance field to the address provided at blockHeight, must be executed
// inside db transaction scope
func (abh *AccountBalanceHelper) AddAccountBalance(address string, amount int64, blockHeight uint32) error {
	addAccountBalanceQ := abh.AccountBalanceQuery.AddAccountBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
			"block_height":    blockHeight,
		},
	)
	return abh.QueryExecutor.ExecuteTransactions(addAccountBalanceQ)
}
