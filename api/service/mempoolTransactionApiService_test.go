package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
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
			name: "NewMempoolTransactionService",
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

func (*mockQueryExecutorGetMempoolTXsFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("want error")
}
func (*mockQueryExecutorGetMempoolTXsScanFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2))
	return db.QueryRow(qStr), nil
}
func (*mockQueryExecutorGetMempoolTXs) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "total_record") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(query.NewMempoolQuery(&chaintype.MainChain{}).Fields).
				AddRow(
					1,
					0,
					1,
					1000,
					make([]byte, 88),
					"accountA",
					"accountA",
				))
	}
	return db.Query(qStr)
}
func (*mockQueryExecutorGetMempoolTXs) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	return db.QueryRow(qStr), nil
}
func TestMempoolTransactionService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType chaintype.ChainType
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
			name: "SuccessWithoutAccountAddress",
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
						ID:               1,
						FeePerByte:       1,
						ArrivalTimestamp: 1000,
						SenderAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
							28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						RecipientAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
							45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
						TransactionBytes: make([]byte, 88),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessWithAccountAddress",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXs{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetMempoolTransactionsRequest{
					Address: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				},
			},
			want: &model.GetMempoolTransactionsResponse{
				Total: 1,
				MempoolTransactions: []*model.MempoolTransaction{
					{
						ID:               1,
						FeePerByte:       1,
						ArrivalTimestamp: 1000,
						SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
							45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
						RecipientAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
							28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						TransactionBytes: make([]byte, 88),
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

type (
	mockQueryExecutorGetMempoolTXFail struct {
		query.Executor
	}
	mockQueryExecutorGetMempoolTXScanFail struct {
		query.Executor
	}
	mockQueryExecutorGetMempoolTXSuccess struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetMempoolTXFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockQueryExecutorGetMempoolTXScanFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows([]string{"foo", "bar"}).AddRow(1, 2))
	return db.QueryRow(qStr), nil
}
func (*mockQueryExecutorGetMempoolTXSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewMempoolQuery(&chaintype.MainChain{}).Fields).AddRow(
			1,
			0,
			1,
			1000,
			make([]byte, 88),
			"accountA",
			"accountB",
		))
	return db.QueryRow(qStr), nil
}
func TestMempoolTransactionService_GetMempoolTransaction(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType chaintype.ChainType
		params    *model.GetMempoolTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMempoolTransactionResponse
		wantErr bool
	}{
		{
			name: "wantFail:Error",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ScanError",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXScanFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:Success",
			fields: fields{
				Query: &mockQueryExecutorGetMempoolTXSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
			},
			want: &model.GetMempoolTransactionResponse{
				Transaction: &model.MempoolTransaction{
					ID:               1,
					FeePerByte:       1,
					ArrivalTimestamp: 1000,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					RecipientAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
						28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
					TransactionBytes: make([]byte, 88),
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
			got, err := ut.GetMempoolTransaction(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolTransactionService.GetMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolTransactionService.GetMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
