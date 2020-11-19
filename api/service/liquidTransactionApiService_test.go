package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewLiquidTransactionService(t *testing.T) {
	type args struct {
		executor                      query.ExecutorInterface
		liquidPaymentTransactionQuery *query.LiquidPaymentTransactionQuery
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	tests := []struct {
		name string
		args args
		want *LiquidTransactionService
	}{
		{
			name: "wantSuccess",
			args: args{
				executor: query.NewQueryExecutor(db),
			},
			want: &LiquidTransactionService{
				QueryExecutor: query.NewQueryExecutor(db),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLiquidTransactionService(tt.args.executor, tt.args.liquidPaymentTransactionQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLiquidTransactionService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetLiquidTransactionsFail struct {
		query.Executor
	}
	mockQueryGetLiquidTransactionsSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetLiquidTransactionsFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}

func (*mockQueryGetLiquidTransactionsFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("want error")
}

func (*mockQueryGetLiquidTransactionsSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows(query.NewLiquidPaymentTransactionQuery().Fields).
			AddRow(
				1,
				[]byte{0, 1, 2, 3},
				[]byte{0, 1, 2, 3},
				100,
				2,
				100,
				1,
				1,
				true,
			),
		)
	return db.Query("")
}

func (*mockQueryGetLiquidTransactionsSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch strings.Contains(qStr, "total_record") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	default:
		return nil, nil
	}
	return db.QueryRow(qStr), nil
}

func TestLiquidTransactionService_GetLiquidTransactions(t *testing.T) {
	type fields struct {
		LiquidPaymentTransactionQuery *query.LiquidPaymentTransactionQuery
		QueryExecutor                 query.ExecutorInterface
	}
	type args struct {
		request *model.GetLiquidTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetLiquidTransactionsResponse
		wantErr bool
	}{
		{
			name: "wantFail",
			fields: fields{
				LiquidPaymentTransactionQuery: &query.LiquidPaymentTransactionQuery{},
				QueryExecutor:                 &mockQueryGetLiquidTransactionsFail{},
			},
			args: args{
				request: &model.GetLiquidTransactionsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				LiquidPaymentTransactionQuery: &query.LiquidPaymentTransactionQuery{},
				QueryExecutor:                 &mockQueryGetLiquidTransactionsSuccess{},
			},
			args: args{
				request: &model.GetLiquidTransactionsRequest{},
			},
			want: &model.GetLiquidTransactionsResponse{
				Total: 1,
				LiquidTransactions: []*model.LiquidPayment{
					{
						ID:               1,
						SenderAddress:    []byte{0, 1, 2, 3},
						RecipientAddress: []byte{0, 1, 2, 3},
						Amount:           100,
						AppliedTime:      2,
						CompleteMinutes:  100,
						Status:           1,
						BlockHeight:      1,
						Latest:           true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lts := &LiquidTransactionService{
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				QueryExecutor:                 tt.fields.QueryExecutor,
			}
			got, err := lts.GetLiquidTransactions(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidTransactionService.GetLiquidTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidTransactionService.GetLiquidTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
