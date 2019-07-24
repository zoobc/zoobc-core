package query

import (
	"bytes"
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
		AddAccountBalance(balance int64, causedFields map[string]interface{}) [][]interface{}
		AddAccountSpendableBalance(balance int64, causedFields map[string]interface{}) (str string, args []interface{})
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
	`, strings.Join(q.Fields, ","), q.TableName)
}

func (q *AccountBalanceQuery) AddAccountBalance(balance int64, causedFields map[string]interface{}) [][]interface{} {
	var queries [][]interface{}
	updateVersionQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE account_id = ? AND block_height = %d - 1 AND latest = true",
		q.TableName, causedFields["block_height"])
	updateBalanceQuery := fmt.Sprintf("INSERT INTO %s (account_id, block_height, spendable_balance, balance, pop_revenue, latest) "+
		"VALUES (?, %d, %d, %d, 0, true) ON CONFLICT(account_id, block_height) DO UPDATE SET spendable_balance = spendable_balance + %d, "+
		"balance = balance + %d", q.TableName, causedFields["block_height"], balance, balance, balance, balance)
	queries = append(queries,
		[]interface{}{
			updateVersionQuery, causedFields["account_id"],
		},
		[]interface{}{
			updateBalanceQuery, causedFields["account_id"],
		},
	)
	return queries
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
