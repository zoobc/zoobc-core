package query

import (
	"bytes"
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
		GetAccountBalanceByAccountID() string
		UpdateAccountBalance(fields, causedFields map[string]interface{}) (str string, args []interface{})
		InsertAccountBalance(accountBalance *model.AccountBalance) (str string, args []interface{})
		AddAccountBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{})
		AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{})
		ExtractModel(accountBalance *model.AccountBalance) []interface{}
		BuildModel(accountBalances []*model.AccountBalance, rows *sql.Rows) []*model.AccountBalance
	}
)

// NewAccountBalanceQuery will create a new AccountBalanceQuery
func NewAccountBalanceQuery() *AccountBalanceQuery {
	return &AccountBalanceQuery{
		Fields: []string{
			"account_id",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
			"latest",
		},
		TableName: "account_balance",
	}
}
func (q *AccountBalanceQuery) GetAccountBalanceByAccountID() string {
	return fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE account_id = ? 
		AND latest = 1 
	`, strings.Join(q.Fields, ","), q.TableName)
}

func (q *AccountBalanceQuery) AddAccountBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{}) {
	return fmt.Sprintf("UPDATE %s SET balance = balance + (%d), spendable_balance = spendable_balance + (%d) WHERE account_id = ?",
		q.TableName, balance, balance), []interface{}{causedFields["account_id"]}
}

func (q *AccountBalanceQuery) AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (
	str string, args []interface{}) {
	return fmt.Sprintf("UPDATE %s SET spendable_balance = spendable_balance + (%d) WHERE account_id = ?",
		q.TableName, balance), []interface{}{causedFields["account_id"]}
}

func (q *AccountBalanceQuery) UpdateAccountBalance(fields, causedFields map[string]interface{}) (str string, args []interface{}) {

	var (
		buff *bytes.Buffer
		i, j int
	)

	buff = bytes.NewBufferString(fmt.Sprintf(`
		UPDATE %s SET 
	`, q.TableName))

	for k, v := range fields {
		buff.WriteString(fmt.Sprintf("%s = ? ", k))
		if i < len(fields) && len(fields) > 1 {
			buff.WriteString(",")
		}
		args = append(args, v)
		i++
	}

	buff.WriteString("WHERE ")
	for k, v := range causedFields {
		buff.WriteString(fmt.Sprintf("%s = ?", k))
		if j < len(causedFields) && len(causedFields) > 1 {
			buff.WriteString(" AND")
		}
		j++
		args = append(args, v)
	}

	return buff.String(), args
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
		account.AccountID,
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
			&accountBalance.AccountID,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest)
		accountBalances = append(accountBalances, &accountBalance)
	}
	return accountBalances
}
