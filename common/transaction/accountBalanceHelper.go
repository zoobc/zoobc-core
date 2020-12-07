package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// AccountBalanceHelperInterface methods collection for transaction helper, it for account balance stuff and account ledger also
	// It better to use with QueryExecutor.BeginTX()
	AccountBalanceHelperInterface interface {
		AddAccountSpendableBalance(address []byte, amount int64) error
		AddAccountBalance(address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64,
			blockTimestamp uint64) error
		GetBalanceByAccountAddress(accountBalance *model.AccountBalance, address []byte, dbTx bool) error
		HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error)
		HasEnoughBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error)
	}
	// AccountBalanceHelper fields for AccountBalanceHelperInterface for transaction helper
	AccountBalanceHelper struct {
		// accountBalance cache when get from db, use this for validation only.
		accountBalance      model.AccountBalance
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
)

func NewAccountBalanceHelper(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	accountLedgerQuery query.AccountLedgerQueryInterface,
) *AccountBalanceHelper {
	return &AccountBalanceHelper{
		AccountBalanceQuery: accountBalanceQuery,
		AccountLedgerQuery:  accountLedgerQuery,
		QueryExecutor:       queryExecutor,
	}
}

// AddAccountSpendableBalance add spendable_balance field to the address provided, must be executed inside db transaction
// scope
func (abh *AccountBalanceHelper) AddAccountSpendableBalance(address []byte, amount int64) error {
	accountBalanceSenderQ, accountBalanceSenderQArgs := abh.AccountBalanceQuery.AddAccountSpendableBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
		},
	)
	err := abh.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err == nil {
		abh.accountBalance = model.AccountBalance{}
	}
	return err
}

// AddAccountBalance add balance and spendable_balance field to the address provided at blockHeight, must be executed
// inside db transaction scope, there process is:
//      - Add new record into account_balance
//      - Add new record into account_ledger
func (abh *AccountBalanceHelper) AddAccountBalance(
	address []byte,
	amount int64,
	event model.EventType,
	blockHeight uint32,
	transactionID int64,
	blockTimestamp uint64,
) error {

	var queries [][]interface{}

	addAccountBalanceQ := abh.AccountBalanceQuery.AddAccountBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
			"block_height":    blockHeight,
		},
	)
	queries = append(queries, addAccountBalanceQ...)

	accountLedgerQ, accountLedgerArgs := abh.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: address,
		BalanceChange:  amount,
		TransactionID:  transactionID,
		BlockHeight:    blockHeight,
		EventType:      event,
		Timestamp:      blockTimestamp,
	})
	queries = append(queries, append([]interface{}{accountLedgerQ}, accountLedgerArgs...))
	err := abh.QueryExecutor.ExecuteTransactions(queries)
	if err == nil {
		abh.accountBalance = model.AccountBalance{}
	}
	return err
}

// GetBalanceByAccountAddress fetching the balance of an account from database
func (abh *AccountBalanceHelper) GetBalanceByAccountAddress(accountBalance *model.AccountBalance, address []byte, dbTx bool) error {
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

// HasEnoughSpendableBalance check if account has enough has spendable balance and will save
func (abh *AccountBalanceHelper) HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error) {
	if bytes.Equal(abh.accountBalance.GetAccountAddress(), address) {
		return abh.accountBalance.GetSpendableBalance() >= compareBalance, nil
	}
	var (
		row            *sql.Row
		accountBalance model.AccountBalance
	)
	qry, args := abh.AccountBalanceQuery.GetAccountBalanceByAccountAddress(address)
	row, err = abh.QueryExecutor.ExecuteSelectRow(qry, dbTX, args...)
	if err != nil {
		return enough, err
	}
	err = abh.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return enough, err
	}
	abh.accountBalance = accountBalance
	return accountBalance.GetSpendableBalance() >= compareBalance, nil
}

// check if account has enough balance
func (abh *AccountBalanceHelper) HasEnoughBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error) {
	if bytes.Equal(abh.accountBalance.GetAccountAddress(), address) {
		return abh.accountBalance.GetBalance() >= compareBalance, nil
	}
	var (
		accountBalance model.AccountBalance
	)
	err = abh.GetBalanceByAccountAddress(&accountBalance, address, dbTX)
	if err != nil {
		return false, err
	}
	abh.accountBalance = accountBalance
	return accountBalance.GetBalance() >= compareBalance, nil
}
