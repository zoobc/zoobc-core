package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockAccountLedgerQuery = NewAccountLedgerQuery()
	mockAccountLedger      = &model.AccountLedger{
		AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		BalanceChange: 10000,
		BlockHeight:   1,
		TransactionID: -123123123123,
		EventType:     model.EventType_EventNodeRegistrationTransaction,
		Timestamp:     1562117271,
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
			wantQStr: "INSERT INTO account_ledger (account_address, balance_change, block_height, transaction_id, event_type, timestamp) " +
				"VALUES(? , ?, ?, ?, ?, ?)",
			wantArgs: []interface{}{
				mockAccountLedger.GetAccountAddress(),
				mockAccountLedger.GetBalanceChange(),
				mockAccountLedger.GetBlockHeight(),
				mockAccountLedger.GetTransactionID(),
				mockAccountLedger.GetEventType(),
				mockAccountLedger.GetTimestamp(),
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

func TestAccountLedgerQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedger *model.AccountLedger
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedger: mockAccountLedger,
			},
			want: []interface{}{
				mockAccountLedger.GetAccountAddress(),
				mockAccountLedger.GetBalanceChange(),
				mockAccountLedger.GetBlockHeight(),
				mockAccountLedger.GetTransactionID(),
				mockAccountLedger.GetEventType(),
				mockAccountLedger.GetTimestamp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := a.ExtractModel(tt.args.accountLedger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountLedgerQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountLedgerQuery_BuildModel(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rowsMock := sqlmock.NewRows(mockAccountLedgerQuery.Fields)
	rowsMock.AddRow(
		mockAccountLedger.GetAccountAddress(),
		mockAccountLedger.GetBalanceChange(),
		mockAccountLedger.GetBlockHeight(),
		mockAccountLedger.GetTransactionID(),
		mockAccountLedger.GetEventType(),
		mockAccountLedger.GetTimestamp(),
	)
	mock.ExpectQuery("").WillReturnRows(rowsMock)
	rows, _ := db.Query("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedgers []*model.AccountLedger
		rows           *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.AccountLedger
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedgers: []*model.AccountLedger{},
				rows:           rows,
			},
			want: []*model.AccountLedger{mockAccountLedger},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := a.BuildModel(tt.args.accountLedgers, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountLedgerQuery.BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountLedgerQuery.BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
