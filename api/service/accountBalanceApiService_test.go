package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// mock GetAccountBalance
	mockExecutorGetAccountBalanceFail struct {
		query.Executor
	}
	mockExecutorGetAccountBalanceNotFound struct {
		query.Executor
	}
	mockExecutorGetAccountBalanceSuccess struct {
		query.Executor
	}
)

var mockAccountBalanceQuery = query.NewAccountBalanceQuery()

func (*mockExecutorGetAccountBalanceSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"AccountId", "BlockHeight", "SpendableBalance", "Balance", "PopRevenue", "Latest"}).AddRow(
		[]byte{1}, 1, 10000, 10000, 0, 1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockExecutorGetAccountBalanceFail) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnError(sql.ErrNoRows)
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockExecutorGetAccountBalanceNotFound) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"AccountId", "BlockHeight", "SpendableBalance", "Balance", "PopRevenue", "Latest"}))
	row := db.QueryRow(qe)
	return row, nil
}

func TestNewAccountBalanceService(t *testing.T) {
	type args struct {
		executor            query.ExecutorInterface
		accountBalanceQuery *query.AccountBalanceQuery
	}
	tests := []struct {
		name string
		args args
		want *AccountBalanceService
	}{
		{
			name: "NewAccountBalanceService:success",
			args: args{
				executor:            nil,
				accountBalanceQuery: nil,
			},
			want: &AccountBalanceService{
				AccountBalanceQuery: nil,
				QueryExecutor:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountBalanceService(tt.args.executor, tt.args.accountBalanceQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountBalanceService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountBalanceService_GetAccountBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery *query.AccountBalanceQuery
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountBalanceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalanceResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalance:fail",
			fields: fields{
				AccountBalanceQuery: mockAccountBalanceQuery,
				QueryExecutor:       &mockExecutorGetAccountBalanceFail{},
			},
			args: args{request: &model.GetAccountBalanceRequest{
				AccountAddress: "BCZ000000000000",
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalance:notFound",
			fields: fields{
				AccountBalanceQuery: mockAccountBalanceQuery,
				QueryExecutor:       &mockExecutorGetAccountBalanceNotFound{},
			},
			args: args{request: &model.GetAccountBalanceRequest{
				AccountAddress: "BCZ000000000000",
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalance:success",
			fields: fields{
				AccountBalanceQuery: mockAccountBalanceQuery,
				QueryExecutor:       &mockExecutorGetAccountBalanceSuccess{},
			},
			args: args{request: &model.GetAccountBalanceRequest{
				AccountAddress: "BCZ000000000000",
			}},
			want: &model.GetAccountBalanceResponse{
				AccountBalance: &model.AccountBalance{
					AccountAddress:   "\001",
					BlockHeight:      1,
					SpendableBalance: 10000,
					Balance:          10000,
					PopRevenue:       0,
					Latest:           true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs := &AccountBalanceService{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			got, err := abs.GetAccountBalance(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceService.GetAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceService.GetAccountBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mock GetAccountBalances
type (
	mockGetAccountBalancesExecutorError struct {
		query.ExecutorInterface
	}
	mockGetAccountBalancesQueryError struct {
		query.ExecutorInterface
	}
	mockGetAccountBalancesQuerySuccess struct {
		query.ExecutorInterface
	}
)

func (*mockGetAccountBalancesExecutorError) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("ExecuteSelect Fail")
}

func (*mockGetAccountBalancesQueryError) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.Query("")
}

func (*mockGetAccountBalancesQuerySuccess) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(query.NewAccountBalanceQuery().Fields)
	mockRows.AddRow(
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		0,
		100000000000,
		101666666666,
		0,
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}

func TestAccountBalanceService_GetAccountBalances(t *testing.T) {
	type fields struct {
		AccountBalanceQuery *query.AccountBalanceQuery
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountBalancesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalancesResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalances:ExecutorError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockGetAccountBalancesExecutorError{},
			},
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: []string{
						"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalances:QueryError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockGetAccountBalancesQueryError{},
			},
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: []string{
						"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalances:QuerySuccess",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockGetAccountBalancesQuerySuccess{},
			},
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: []string{
						"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					},
				},
			},
			want: &model.GetAccountBalancesResponse{
				AccountBalances: []*model.AccountBalance{
					{
						AccountAddress:   "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
						BlockHeight:      0,
						SpendableBalance: 100000000000,
						Balance:          101666666666,
						PopRevenue:       0,
						Latest:           true,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs := &AccountBalanceService{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			got, err := abs.GetAccountBalances(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceService.GetAccountBalances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceService.GetAccountBalances() = %v, want %v", got, tt.want)
			}
		})
	}
}
