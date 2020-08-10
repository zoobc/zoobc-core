package transaction

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	AccountBalanceHelperInterface interface {
		AddAccountSpendableBalance(address string, amount int64) error
		AddAccountBalance(address string, amount int64, blockHeight uint32) error
		GetBalanceByAccountID(accountBalance *model.AccountBalance, address string, dbTx bool) error
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

// GetBalanceByAccountID fetching the balance of an account from database
func (abh *AccountBalanceHelper) GetBalanceByAccountID(accountBalance *model.AccountBalance, address string, dbTx bool) error {
	var (
		row *sql.Row
		err error
	)

	qry, args := abh.AccountBalanceQuery.GetAccountBalanceByAccountAddress(address)
	row, err = abh.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = abh.AccountBalanceQuery.Scan(accountBalance, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "TXSenderNotFound")
	}
	return nil
}
