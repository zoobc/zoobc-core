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
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// GetBlocksmithAccountAddress mocks
	mockExecutorGetBlocksmithAccountAddressExecuteSelectFail struct {
		query.Executor
	}

	mockGetBlocksmithAccountAddressExecutorSuccess struct {
		query.Executor
	}
	mockGetBlocksmithAccountAddressNodeRegistrationFail struct {
		query.NodeRegistrationQuery
	}
	mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows struct {
		query.NodeRegistrationQuery
	}
	mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows struct {
		query.NodeRegistrationQuery
	}
	// GetBlocksmithAccountAddress mocks
	// RewardBlocksmithAccountAddresses mocks
	mockRewardBlocksmithAccountAddressesExecutorFail struct {
		query.Executor
	}
	mockRewardBlocksmithAccountAddressesExecutorSuccess struct {
		query.Executor
	}
	// RewardBlocksmithAccountAddresses mocks
)

var (
	// GetBlocksmithAccountAddress mocks
	mockGetBlocksmithAccountAddressNodeRegistry = &model.NodeRegistration{AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88,
		220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}}
)

func (*mockExecutorGetBlocksmithAccountAddressExecuteSelectFail) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetBlocksmithAccountAddressExecutorSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationFail) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{}, nil
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		mockGetBlocksmithAccountAddressNodeRegistry,
	}, nil
}

func (*mockRewardBlocksmithAccountAddressesExecutorFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockedError")
}

func (*mockRewardBlocksmithAccountAddressesExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestBlocksmithService_GetBlocksmithAccountAddress(t *testing.T) {
	type fields struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetBlocksmithAccountAddress-ExecuteSelectFail",
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorGetBlocksmithAccountAddressExecuteSelectFail{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-BuildModelFail-IncorrectColumn",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationFail{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-VersionedNodeRegistrationNotFound",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-VersionedNodeRegistrationNotFound",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    mockGetBlocksmithAccountAddressNodeRegistry.AccountAddress,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlocksmithService{
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := bs.GetBlocksmithAccountAddress(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("GetBlocksmithAccountAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_RewardBlocksmithAccountAddresses(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		AccountLedgerQuery    query.AccountLedgerQueryInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		blocksmithAccountAddresses [][]byte
		totalReward                int64
		timestamp                  int64
		height                     uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "RewardBlocksmithAccountAddress:NoAccountToBeRewarded",
			args: args{
				blocksmithAccountAddresses: [][]byte{},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       nil,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
		{
			name: "RewardBlocksmithAccountAddress:executorFailExecuteTransactions",
			args: args{
				blocksmithAccountAddresses: [][]byte{bcsAddress1},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       &mockRewardBlocksmithAccountAddressesExecutorFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
		{
			name: "RewardBlocksmithAccountAddress:success",
			args: args{
				blocksmithAccountAddresses: [][]byte{bcsAddress1},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       &mockRewardBlocksmithAccountAddressesExecutorSuccess{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlocksmithService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountLedgerQuery:    tt.fields.AccountLedgerQuery,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			if err := bs.RewardBlocksmithAccountAddresses(
				tt.args.blocksmithAccountAddresses,
				tt.args.totalReward,
				tt.args.timestamp,
				tt.args.height,
			); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.RewardBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBlocksmithService(t *testing.T) {
	type args struct {
		accountBalanceQuery   query.AccountBalanceQueryInterface
		accountLedgerQuery    query.AccountLedgerQueryInterface
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		queryExecutor         query.ExecutorInterface
		chaintype             chaintype.ChainType
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithService
	}{
		{
			name: "NewBlocksmithServiceSuccess",
			args: args{
				accountLedgerQuery:    nil,
				accountBalanceQuery:   nil,
				nodeRegistrationQuery: nil,
				queryExecutor:         nil,
			},
			want: &BlocksmithService{
				AccountBalanceQuery:   nil,
				AccountLedgerQuery:    nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithService(tt.args.accountBalanceQuery, tt.args.accountLedgerQuery,
				tt.args.nodeRegistrationQuery, tt.args.queryExecutor, tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}
