package transaction

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAccountBalanceHelperExecutorAddSpendableFail struct {
		query.ExecutorInterface
	}
	mockAccountBalanceHelperExecutorAddSpendableSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockAccountBalanceHelperExecutorAddSpendableFail) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperExecutorAddSpendableFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperExecutorAddSpendableSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockAccountBalanceHelperExecutorAddSpendableSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestAccountBalanceHelper_AddAccountSpendableBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		address string
		amount  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "executorError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableFail{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "executeSuccess",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableSuccess{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHelper{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := abh.AddAccountSpendableBalance(tt.args.address, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("AddAccountSpendableBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountBalanceHelper_AddAccountBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		address     string
		amount      int64
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "executorError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableFail{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "executeSuccess",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableSuccess{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHelper{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountLedgerQuery:  tt.fields.AccountLedgerQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := abh.AddAccountBalance(tt.args.address, tt.args.amount, 0, tt.args.blockHeight, 0, 0); (err != nil) != tt.wantErr {
				t.Errorf("AddAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperSuccess struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperFail struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}
func (*mockAccountBalanceHelperSuccess) HasEnoughSpendableBalance(
	dbTX bool, address string, compareBalance int64,
) (enough bool, err error) {
	return true, nil
}
func (*mockAccountBalanceHelperFail) AddAccountSpendableBalance(address string, amount int64) error {
	return sql.ErrTxDone
}
func (*mockAccountBalanceHelperFail) HasEnoughSpendableBalance(
	dbTX bool, address string, compareBalance int64,
) (enough bool, err error) {
	return false, nil
}
func (*mockAccountBalanceHelperSuccess) AddAccountBalance(
	address string, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return nil
}
func (*mockAccountBalanceHelperFail) AddAccountBalance(
	address string, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return sql.ErrTxDone
}
