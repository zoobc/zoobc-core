package transaction

import (
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAccountLedgerHelperExecutorAddLedgerRecordFail struct {
		query.ExecutorInterface
	}
	mockAccountLedgerHelperExecutorAddLedgerRecordSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockAccountLedgerHelperExecutorAddLedgerRecordFail) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("mockedError")
}

func (*mockAccountLedgerHelperExecutorAddLedgerRecordSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func TestAccountLedgerHelper_InsertLedgerEntry(t *testing.T) {
	type fields struct {
		AccountLedgerQuery query.AccountLedgerQueryInterface
		QueryExecutor      query.ExecutorInterface
	}
	type args struct {
		accountLedger *model.AccountLedger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "executeTransactionFail",
			fields: fields{
				AccountLedgerQuery: query.NewAccountLedgerQuery(),
				QueryExecutor:      &mockAccountLedgerHelperExecutorAddLedgerRecordFail{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "executeTransactionSuccess",
			fields: fields{
				AccountLedgerQuery: query.NewAccountLedgerQuery(),
				QueryExecutor:      &mockAccountLedgerHelperExecutorAddLedgerRecordSuccess{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alh := &AccountLedgerHelper{
				AccountLedgerQuery: tt.fields.AccountLedgerQuery,
				QueryExecutor:      tt.fields.QueryExecutor,
			}
			if err := alh.InsertLedgerEntry(tt.args.accountLedger); (err != nil) != tt.wantErr {
				t.Errorf("InsertLedgerEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
