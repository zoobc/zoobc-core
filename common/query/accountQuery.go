package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountQuery struct {
		Fields    []string
		TableName string
	}

	AccountQueryInterface interface {
		GetAccountByID(accountID []byte) (str string, args []interface{})
		GetAccountByIDs(ids [][]byte) (str string, args [][]byte)
		InsertAccount(account *model.Account) (str string, args []interface{})
		ExtractModel(account *model.Account) []interface{}
		BuildModel(accounts []*model.Account, rows *sql.Rows) []*model.Account
		GetTableName() string
	}
)

// NewAccountQuery returns AccountQuery instance
func NewAccountQuery() *AccountQuery {
	return &AccountQuery{
		Fields:    []string{"id", "account_type", "address"},
		TableName: "account",
	}
}

// GetAccountByID returns query string to get account by ID
func (aq *AccountQuery) GetAccountByID(accountID []byte) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(aq.Fields, ", "), aq.TableName),
		[]interface{}{accountID}
}

// GetAccountByIDs return query string to get accounts by multiple IDs
func (aq *AccountQuery) GetAccountByIDs(ids [][]byte) (str string, args [][]byte) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE id in (%s)",
		strings.Join(aq.Fields, ","),
		aq.TableName,
		fmt.Sprintf("? %s", strings.Repeat(",?", len(ids)-1)),
	), args
}

func (aq *AccountQuery) InsertAccount(account *model.Account) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT OR IGNORE INTO %s (%s) VALUES(%s)",
		aq.TableName,
		strings.Join(aq.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(aq.Fields)-1)),
	), aq.ExtractModel(account)
}

func (aq *AccountQuery) ExtractModel(account *model.Account) []interface{} {
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

// GetTableName is func to get account table name
func (aq *AccountQuery) GetTableName() string {
	return aq.TableName
}
