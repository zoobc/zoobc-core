package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewMempoolTransactionsService(t *testing.T) {
	type args struct {
		queryExecutor query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *MempoolTransactionService
	}{
		{
			name: "NewMempoolTranscationService",
			args: args{
				queryExecutor: &query.Executor{},
			},
			want: &MempoolTransactionService{
				Query: &query.Executor{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMempoolTransactionsService(tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMempoolTransactionsService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorGetMempoolTXsFail struct {
		query.Executor
	}
	mockQueryExecutorGetMempoolTXsScanFail struct {
		query.Executor
	}
	mockQueryExecutorGetMempoolTXs struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetMempoolTXsFail) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}
func (*mockQueryExecutorGetMempoolTXsScanFail) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2))
	return db.Query(qStr)
}
func (*mockQueryExecutorGetMempoolTXs) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "total_record") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(query.NewMempoolQuery(&chaintype.MainChain{}).Fields).
				AddRow(
					1,
					1,
					1000,
					make([]byte, 88),
					"accountA",
					"accountA",
				))
	}
	return db.Query(qStr)
}
func TestMempoolTransactionService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType contract.ChainType
		params    *model.GetMempoolTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMempoolTransactionsResponse
		wantErr bool
	}{
		{
			name: "wantFail:ExecuteFail",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXsFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params:    &model.GetMempoolTransactionsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ExecuteScanFail",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXsScanFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params:    &model.GetMempoolTransactionsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Success",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXs{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params:    &model.GetMempoolTransactionsRequest{},
			},
			want: &model.GetMempoolTransactionsResponse{
				Total: 1,
				MempoolTransactions: []*model.MempoolTransaction{
					{
						ID:                      1,
						FeePerByte:              1,
						ArrivalTimestamp:        1000,
						SenderAccountAddress:    "accountA",
						RecipientAccountAddress: "accountA",
						TransactionBytes:        make([]byte, 88),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := &MempoolTransactionService{
				Query: tt.fields.Query,
			}
			got, err := ut.GetMempoolTransactions(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolTransactionService.GetMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolTransactionService.GetMempoolTransactions() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
