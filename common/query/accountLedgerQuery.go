package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountLedgerQuery struct {
		Fields    []string
		TableName string
	}
	// AccountLedgerQueryInterface includes interface methods for AccountLedgerQuery
	AccountLedgerQueryInterface interface {
		ExtractModel(accountLedger *model.AccountLedger) []interface{}
		InsertAccountLedger(accountLedger *model.AccountLedger) (qStr string, args []interface{})
	}
)

func NewAccountLegderQuery() *AccountLedgerQuery {
	return &AccountLedgerQuery{
		Fields: []string{
			"account_address",
			"account_balance",
			"block_height",
			"transaction_id",
			"event_type",
		},
		TableName: "account_ledger",
	}
}

func (q *AccountLedgerQuery) getTableName() interface{} {
	return q.TableName
}

// InsertAccountLedger represents insert query for AccountLedger
func (q *AccountLedgerQuery) InsertAccountLedger(accountLedger *model.AccountLedger) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			q.getTableName(),
			strings.Join(q.Fields, ", "),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(q.Fields)-1)),
		),
		q.ExtractModel(accountLedger)
}

// ExtractModel will extract accountLedger model to []interface
func (*AccountLedgerQuery) ExtractModel(accountLedger *model.AccountLedger) []interface{} {
	return []interface{}{
		accountLedger.GetAccountAddress(),
		accountLedger.GetAccountBalance(),
		accountLedger.GetBlockHeight(),
		accountLedger.GetTransactionID(),
		accountLedger.GetEventType(),
	}
}
