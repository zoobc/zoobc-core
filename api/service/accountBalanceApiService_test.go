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

func (*mockExecutorGetAccountBalanceSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"AccountId", "BlockHeight", "SpendableBalance", "Balance", "PopRevenue", "Latest"}).AddRow(
		[]byte{1}, 1, 10000, 10000, 0, 1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockExecutorGetAccountBalanceFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func (*mockExecutorGetAccountBalanceNotFound) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"AccountId", "BlockHeight", "SpendableBalance", "Balance", "PopRevenue", "Latest"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

func TestAccountBalanceService_GetAccountBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Executor            query.ExecutorInterface
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
				Executor:            &mockExecutorGetAccountBalanceFail{},
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
				Executor:            &mockExecutorGetAccountBalanceNotFound{},
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
				Executor:            &mockExecutorGetAccountBalanceSuccess{},
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
				Executor:            tt.fields.Executor,
			}
			got, err := abs.GetAccountBalance(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAccountBalanceService(t *testing.T) {
	type args struct {
		executor            query.ExecutorInterface
		accountBalanceQuery query.AccountBalanceQueryInterface
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
				Executor:            nil,
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
