package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountQueryInterface interface {
		GetAccountByID(accountID []byte) (string, []interface{})
		ExtractModel(account *model.Account) []interface{}
		BuildModel(accounts []*model.Account, rows *sql.Rows) []*model.Account
	}

	AccountQuery struct {
		Fields    []string
		TableName string
	}
)

// NewAccountQuery returns AccountQuery instance
func NewAccountQuery() *AccountQuery {
	return &AccountQuery{
		Fields:    []string{"id", "account_type", "address"},
		TableName: "account",
	}
}

func (aq *AccountQuery) getTableName() string {
	return aq.TableName
}

// GetAccountByID returns query string to get account by ID
func (aq *AccountQuery) GetAccountByID(accountID []byte) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(aq.Fields, ", "), aq.getTableName()),
		[]interface{}{accountID}
}

func (*AccountQuery) ExtractModel(account *model.Account) []interface{} {
	return []interface{}{
		account.ID,
		account.AccountType,
		account.Address,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*AccountQuery) BuildModel(accounts []*model.Account, rows *sql.Rows) []*model.Account {
	for rows.Next() {
		var account model.Account
		_ = rows.Scan(
			&account.ID,
			&account.AccountType,
			&account.Address)
		accounts = append(accounts, &account)
	}
	return accounts
}
