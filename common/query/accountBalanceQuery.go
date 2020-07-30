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
		GetAccountBalanceByAccountAddress(accountAddress string) (str string, args []interface{})
		GetAccountBalances() string
		InsertAccountBalance(accountBalance *model.AccountBalance) (str string, args []interface{})
		InsertAccountBalances(accountBalances []*model.AccountBalance) (str string, args []interface{})
		AddAccountBalance(balance int64, causedFields map[string]interface{}) [][]interface{}
		AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{})
		ExtractModel(accountBalance *model.AccountBalance) []interface{}
		BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) ([]*model.AccountBalance, error)
		Scan(accountBalance *model.AccountBalance, row *sql.Row) error
		GetFields() []string
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

func (q *AccountBalanceQuery) GetAccountBalanceByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE account_address = ? AND latest = 1 ORDER BY block_height DESC`,
		strings.Join(q.Fields, ","), q.TableName), []interface{}{accountAddress}
}

func (q *AccountBalanceQuery) GetAccountBalances() string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE latest = 1`,
		strings.Join(q.Fields, ","), q.TableName)
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
	return fmt.Sprintf("UPDATE %s SET spendable_balance = spendable_balance + (%d) WHERE account_address = ?"+
		" AND latest = 1",
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

// InsertAccountBalances represents query builder to insert multiple record in single query
func (q *AccountBalanceQuery) InsertAccountBalances(accountBalances []*model.AccountBalance) (str string, args []interface{}) {
	if len(accountBalances) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			q.getTableName(),
			strings.Join(q.Fields, ", "),
		)
		for k, accBalance := range accountBalances {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(q.Fields)-1),
			)
			if k < len(accountBalances)-1 {
				str += ","
			}
			args = append(args, q.ExtractModel(accBalance)...)
		}
	}
	return str, args
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
func (*AccountBalanceQuery) BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) ([]*model.AccountBalance, error) {
	for rows.Next() {
		var (
			accountBalance model.AccountBalance
			err            error
		)
		err = rows.Scan(
			&accountBalance.AccountAddress,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest,
		)
		if err != nil {
			return nil, err
		}
		accountBalances = append(accountBalances, &accountBalance)
	}
	return accountBalances, nil
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
func (q *AccountBalanceQuery) getTableName() string {
	return q.TableName
}

func (q *AccountBalanceQuery) GetFields() []string {
	return q.Fields
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
			WHERE latest = ? AND (account_address, block_height) IN (
				SELECT t2.account_address, MAX(t2.block_height)
				FROM %s as t2
				GROUP BY t2.account_address
			)`,
				q.TableName,
				q.TableName,
			),
			1,
			0,
		},
	}
}

func (q *AccountBalanceQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	snapshotField := []string{
		"account_address",
		"block_height",
		"balance",
		"balance",
		"pop_revenue",
		"latest",
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE (account_address, block_height) IN (SELECT t2.account_address, "+
		"MAX(t2.block_height) FROM %s as t2 WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 "+
		"GROUP BY t2.account_address) ORDER BY block_height",
		strings.Join(snapshotField, ","), q.getTableName(), q.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (q *AccountBalanceQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		q.getTableName(), fromHeight, toHeight)
}
