package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockAccountLedgerQuery = NewAccountLedgerQuery()
	mockAccountLedger      = &model.AccountLedger{
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		AccountBalance: 10000,
		BlockHeight:    1,
		TransactionID:  -123123123123,
		EventType:      model.EventType_EventNodeRegistrationTransaction,
	}
)

func TestAccountLedgerQuery_InsertAccountLedger(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedger *model.AccountLedger
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedger: mockAccountLedger,
			},
			wantQStr: "INSERT INTO account_ledger (account_address, account_balance, block_height, transaction_id, event_type) " +
				"VALUES(? , ?, ?, ?, ?)",
			wantArgs: []interface{}{
				mockAccountLedger.GetAccountAddress(),
				mockAccountLedger.GetAccountBalance(),
				mockAccountLedger.GetBlockHeight(),
				mockAccountLedger.GetTransactionID(),
				mockAccountLedger.GetEventType(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := q.InsertAccountLedger(tt.args.accountLedger)
			if gotQStr != tt.wantQStr {
				t.Errorf("InsertAccountLedger() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertAccountLedger() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountLedgerQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args:   args{height: 1},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM account_ledger WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := q.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}
