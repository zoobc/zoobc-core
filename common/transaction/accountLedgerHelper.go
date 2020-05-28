package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	AccountLedgerHelperInterface interface {
		InsertLedgerEntry(
			accountLedger *model.AccountLedger,
		) error
	}
	AccountLedgerHelper struct {
		AccountLedgerQuery query.AccountLedgerQueryInterface
		QueryExecutor      query.ExecutorInterface
	}
)

func NewAccountLedgerHelper(
	accountLedgerQuery query.AccountLedgerQueryInterface,
	queryExecutor query.ExecutorInterface,
) *AccountLedgerHelper {
	return &AccountLedgerHelper{
		AccountLedgerQuery: accountLedgerQuery,
		QueryExecutor:      queryExecutor,
	}
}

func (alh *AccountLedgerHelper) InsertLedgerEntry(
	accountLedger *model.AccountLedger,
) error {
	senderAccountLedgerQ, senderAccountLedgerArgs := alh.AccountLedgerQuery.InsertAccountLedger(accountLedger)
	err := alh.QueryExecutor.ExecuteTransaction(senderAccountLedgerQ, senderAccountLedgerArgs...)
	if err != nil {
		return err
	}
	return nil
}
