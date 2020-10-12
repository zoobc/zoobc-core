package service

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockAccountLedgerQuery = query.NewAccountLedgerQuery()
	mockAccountLedger      = &model.AccountLedger{
		AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BalanceChange: 10,
		BlockHeight:   2,
		TransactionID: -9127118158999748858,
		EventType:     model.EventType_EventClaimNodeRegistrationTransaction,
		Timestamp:     1562117271,
	}
)

type (
	mockQueryAccountLedgersSuccess struct {
		query.Executor
	}
)

func (*mockQueryAccountLedgersSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("1"))
	return db.QueryRow(qStr), nil
}
func (*mockQueryAccountLedgersSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
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
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(rowsMock)
	return db.Query(qStr, args...)
}

func TestAccountLedgerService_GetAccountLedgers(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountLedgersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountLedgersResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryAccountLedgersSuccess{},
			},
			args: args{
				request: &model.GetAccountLedgersRequest{
					AccountAddress: mockAccountLedger.GetAccountAddress(),
					Pagination: &model.Pagination{
						Limit:      30,
						OrderField: "account_address",
						OrderBy:    model.OrderBy_DESC,
					},
				},
			},
			want: &model.GetAccountLedgersResponse{
				Total:          1,
				AccountLedgers: []*model.AccountLedger{mockAccountLedger},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AccountLedgerService{
				Query: tt.fields.Query,
			}
			got, err := al.GetAccountLedgers(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountLedgers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountLedgers() got = %v, want %v", got, tt.want)
			}
		})
	}
}
