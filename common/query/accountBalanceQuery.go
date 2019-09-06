package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// AccountBalanceQuery is struct will implemented AccountBalanceInterface
	AccountBalanceQuery struct {
		Fields    []string
		TableName string
	}
	// AccountBalanceQueryInterface interface that implemented by AccountBalanceQuery
	AccountBalanceQueryInterface interface {
		GetAccountBalanceByAccountAddress(accountAddress string) (string, interface{})
		InsertAccountBalance(accountBalance *model.AccountBalance) (str string, args []interface{})
		AddAccountBalance(balance int64, causedFields map[string]interface{}) [][]interface{}
		AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{})
		ExtractModel(accountBalance *model.AccountBalance) []interface{}
		BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) []*model.AccountBalance
		Scan(accountBalance *model.AccountBalance, row *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}
)

// NewAccountBalanceQuery will create a new AccountBalanceQuery
func NewAccountBalanceQuery() *AccountBalanceQuery {
	return &AccountBalanceQuery{
		Fields: []string{
			"account_address",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
			"latest",
		},
		TableName: "account_balance",
	}
}
func (q *AccountBalanceQuery) GetAccountBalanceByAccountAddress(accountAddress string) (query string, args interface{}) {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE account_address = ? AND latest = 1`,
		strings.Join(q.Fields, ","), q.TableName), accountAddress
}

func (q *AccountBalanceQuery) AddAccountBalance(balance int64, causedFields map[string]interface{}) [][]interface{} {
	var (
		queries            [][]interface{}
		updateVersionQuery string
	)
	// insert account if account not in account balance yet
	insertBalanceQuery := fmt.Sprintf("INSERT INTO %s (account_address, block_height, spendable_balance, balance, pop_revenue, latest) "+
		"SELECT ?, %d, 0, 0, 0, 1 WHERE NOT EXISTS (SELECT account_address FROM %s WHERE account_address = ?)", q.TableName,
		causedFields["block_height"], q.TableName)
	// update or insert new account_balance row
	updateBalanceQuery := fmt.Sprintf("INSERT INTO %s (account_address, block_height, spendable_balance, balance, pop_revenue, latest) "+
		"SELECT account_address, %d, spendable_balance + %d, balance + %d, pop_revenue, latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1 ON CONFLICT(account_address, block_height) "+
		"DO UPDATE SET (spendable_balance, balance) = (SELECT "+
		"spendable_balance + %d, balance + %d FROM %s WHERE account_address = ? AND latest = 1)",
		q.TableName, causedFields["block_height"], balance, balance, balance, balance, q.TableName)

	queries = append(queries,
		[]interface{}{
			insertBalanceQuery, causedFields["account_address"], causedFields["account_address"],
		},
		[]interface{}{
			updateBalanceQuery, causedFields["account_address"], causedFields["account_address"],
		},
	)
	if causedFields["block_height"].(uint32) != 0 {
		// set previous version record to latest = false
		updateVersionQuery = fmt.Sprintf("UPDATE %s SET latest = false WHERE account_address = ? AND block_height != %d AND latest = true",
			q.TableName, causedFields["block_height"])
		queries = append(queries,
			[]interface{}{
				updateVersionQuery, causedFields["account_address"],
			},
		)
	}
	return queries
}

func (q *AccountBalanceQuery) AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (
	str string, args []interface{}) {
	return fmt.Sprintf("UPDATE %s SET spendable_balance = spendable_balance + (%d) WHERE account_address = ?",
		q.TableName, balance), []interface{}{causedFields["account_address"]}
}

func (q *AccountBalanceQuery) InsertAccountBalance(accountBalance *model.AccountBalance) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		q.TableName,
		strings.Join(q.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(q.Fields)-1)),
	), q.ExtractModel(accountBalance)
}

func (*AccountBalanceQuery) ExtractModel(account *model.AccountBalance) []interface{} {
	return []interface{}{
		account.AccountAddress,
		account.BlockHeight,
		account.SpendableBalance,
		account.Balance,
		account.PopRevenue,
		account.Latest,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*AccountBalanceQuery) BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) []*model.AccountBalance {
	for rows.Next() {
		var accountBalance model.AccountBalance
		_ = rows.Scan(
			&accountBalance.AccountAddress,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest)
		accountBalances = append(accountBalances, &accountBalance)
	}
	return accountBalances
}

// Scan similar with `sql.Scan`
func (*AccountBalanceQuery) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	err := row.Scan(
		&accountBalance.AccountAddress,
		&accountBalance.BlockHeight,
		&accountBalance.SpendableBalance,
		&accountBalance.Balance,
		&accountBalance.PopRevenue,
		&accountBalance.Latest,
	)
	return err
}

// Rollback delete records `WHERE block_height > "height"
// and UPDATE latest of the `account_address` clause by `block_height`
func (q *AccountBalanceQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", q.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE (block_height || '_' || account_address) IN (
				SELECT (MAX(block_height) || '_' || account_address) as con
				FROM %s
				WHERE latest = 0
				GROUP BY account_address
			)`,
				q.TableName,
				q.TableName,
			),
			1,
		},
	}
}
