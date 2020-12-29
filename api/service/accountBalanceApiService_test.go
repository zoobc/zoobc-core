// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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

var (
	accBalanceAccountAddress1 = []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183,
		229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162}
	mockAccountBalanceQuery = query.NewAccountBalanceQuery()
)

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
				AccountAddress: accBalanceAccountAddress1,
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
				AccountAddress: accBalanceAccountAddress1,
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
				AccountAddress: accBalanceAccountAddress1,
			}},
			want: &model.GetAccountBalanceResponse{
				AccountBalance: &model.AccountBalance{
					AccountAddress:   []byte{1},
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
		accBalanceAccountAddress1,
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
					AccountAddresses: [][]byte{
						{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
							116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
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
					AccountAddresses: [][]byte{
						{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
							116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
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
					AccountAddresses: [][]byte{
						{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
							116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
					},
				},
			},
			want: &model.GetAccountBalancesResponse{
				AccountBalances: []*model.AccountBalance{
					{
						AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113,
							140, 183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
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
