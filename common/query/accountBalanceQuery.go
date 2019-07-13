package query

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// AccountBalanceQuery is struct will implemented AccountBalanceInt
	AccountBalanceQuery struct {
		Fields    []string
		TableName string
	}
	// AccountBalanceInt interface that implemented by AccountBalanceQuery
	AccountBalanceInt interface {
		GetAccountBalanceByAccountID() string
		UpdateAccountBalance(fields, causedFields map[string]interface{}) (str string, args []interface{})
		InsertAccountBalance(accountBalance *model.AccountBalance) (str string, args []interface{})
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

func (q *AccountBalanceQuery) UpdateAccountBalance(fields, causedFields map[string]interface{}) (str string, args []interface{}) {

	var (
		buff *bytes.Buffer
		i, j int
	)

	buff = bytes.NewBufferString(fmt.Sprintf(`
		UPDATE %s SET 
	`, q.TableName))

	for k, v := range fields {
		buff.WriteString(fmt.Sprintf("%s = ?", k))
		if i < len(fields) {
			buff.WriteString(",")
		}
		args = append(args, v)
		i++
	}

	buff.WriteString("WHERE")
	for k, v := range causedFields {
		buff.WriteString(fmt.Sprintf("%s = ?", k))
		if j < len(causedFields) {
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

func (q *AccountBalanceQuery) ExtractModel(account *model.AccountBalance) []interface{} {
	return []interface{}{
		account.AccountID,
		account.Balance,
		account.SpendableBalance,
		account.BlockHeight,
		account.PopRevenue,
	}
}
