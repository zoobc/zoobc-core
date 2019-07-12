package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountQueryInterface interface {
		GetAccountByID() string
		ExtractModel(accountBalance *model.AccountBalance) []interface{}
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
func (aq *AccountQuery) GetAccountByID() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(aq.Fields, ", "), aq.getTableName())
}

func (*AccountQuery) ExtractModel(account *model.Account) []interface{} {
	return []interface{}{
		account.ID,
		account.AccountType,
		account.Address,
	}
}

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
