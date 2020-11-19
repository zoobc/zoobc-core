package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	// AccountBalanceHelperInterface methods collection for transaction helper, it for account balance stuff and account ledger also
	// It better to use with QueryExecutor.BeginTX()
	AccountBalanceHelperInterface interface {
		AddAccountSpendableBalance(address []byte, amount int64) error
		AddAccountSpendableBalanceInCache(address []byte, amount int64) error
		UpdateAccountSpendableBalanceInCache(address []byte, amount int64) error
		AddAccountBalance(address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64,
			blockTimestamp uint64) error
		GetBalanceByAccountAddress(accountBalance *model.AccountBalance, address []byte, dbTx bool) error
		HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error)
	}
	// AccountBalanceHelper fields for AccountBalanceHelperInterface for transaction helper
	AccountBalanceHelper struct {
		// accountBalance cache when get from db, use this for validation only.
		accountBalance          model.AccountBalance
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		AccountLedgerQuery      query.AccountLedgerQueryInterface
		QueryExecutor           query.ExecutorInterface
		SpendableBalanceStorage storage.CacheStorageInterface
	}
)

func NewAccountBalanceHelper(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	accountLedgerQuery query.AccountLedgerQueryInterface,
	spendableBalanceStorage storage.CacheStorageInterface,
) *AccountBalanceHelper {
	return &AccountBalanceHelper{
		AccountBalanceQuery:     accountBalanceQuery,
		AccountLedgerQuery:      accountLedgerQuery,
		QueryExecutor:           queryExecutor,
		SpendableBalanceStorage: spendableBalanceStorage,
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
	return err
}

// AddAccountSpendableBalanceInCache to add or update spendable balance in cache storage
func (abh *AccountBalanceHelper) AddAccountSpendableBalanceInCache(address []byte, amount int64) error {
	var (
		currentSpendAbleBalance int64
		err                     = abh.SpendableBalanceStorage.GetItem(address, &currentSpendAbleBalance)
	)
	if err != nil {
		return err
	}
	if currentSpendAbleBalance == 0 {
		// get spendable balace from DB
		var accountBalance model.AccountBalance
		err = abh.GetBalanceByAccountAddress(&accountBalance, address, false)
		if err != nil {
			return err
		}
		currentSpendAbleBalance = accountBalance.GetSpendableBalance()
	}
	newSpendableBalance := currentSpendAbleBalance + amount
	return abh.SpendableBalanceStorage.SetItem(address, newSpendableBalance)
}

/*
 UpdateAccountSpendableBalanceInCache to update existing spendable balance in cache storage
 - update existing cache spendable balance should be in transactional process
*/
func (abh *AccountBalanceHelper) UpdateAccountSpendableBalanceInCache(address []byte, amount int64) error {
	var (
		spendAbleBalance int64
		err              = abh.SpendableBalanceStorage.GetItem(address, &spendAbleBalance)
	)
	if err != nil {
		return err
	}
	if spendAbleBalance == 0 {
		return nil
	}
	spendAbleBalance = spendAbleBalance + amount
	txSpendableBalanceStorage, ok := abh.SpendableBalanceStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastSpendableBalanceStorageAsTransactionalCacheInterface")
	}
	return txSpendableBalanceStorage.TxSetItem(address, spendAbleBalance)
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
	if err != nil {
		return err
	}

	// check if spendable balance is cached
	var spendableBalanceCache int64
	err = abh.SpendableBalanceStorage.GetItem(address, &spendableBalanceCache)
	if err != nil {
		return err
	}
	if spendableBalanceCache == 0 {
		return nil
	}
	// updating spendable balance in cache
	spendableBalanceCache = spendableBalanceCache + amount
	txSpendableBalanceStorage, ok := abh.SpendableBalanceStorage.(storage.TransactionalCache)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "FailToCastSpendableBalanceStorageAsTransactionalCacheInterface")
	}
	err = txSpendableBalanceStorage.TxSetItem(address, spendableBalanceCache)
	if err != nil {
		return err
	}
	return nil
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

	// check into spendable balance cache
	var spendableBalance int64
	err = abh.SpendableBalanceStorage.GetItem(address, &spendableBalance)
	if err != nil {
		return err
	}
	if spendableBalance == 0 {
		return nil
	}
	// use spendable cache when spendable account is cached
	accountBalance.SpendableBalance = spendableBalance
	return nil
}

// HasEnoughSpendableBalance check if account has enough has spendable balance and will save
func (abh *AccountBalanceHelper) HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error) {
	if bytes.Equal(abh.accountBalance.GetAccountAddress(), address) {
		return abh.accountBalance.GetSpendableBalance() >= compareBalance, nil
	}
	var (
		accountBalance   model.AccountBalance
		spendableBalance int64
	)
	// check first into spendable balance cache
	err = abh.SpendableBalanceStorage.GetItem(address, &spendableBalance)
	if err != nil {
		return enough, err
	}
	if spendableBalance == 0 {
		// check into Database if not cached
		err = abh.GetBalanceByAccountAddress(&accountBalance, address, dbTX)
		if err != nil {
			return enough, err
		}
		abh.accountBalance = accountBalance
		spendableBalance = accountBalance.GetSpendableBalance()
	}
	return spendableBalance >= compareBalance, nil
}
