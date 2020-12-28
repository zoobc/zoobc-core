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
	"fmt"
	"github.com/zoobc/zoobc-core/common/crypto"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	mockCoinbaseLotteryWinnersQueryExecutorSuccess struct {
		query.Executor
	}
	mockCoinbaseLotteryWinnersQueryExecutorSelectFail struct {
		query.Executor
	}
	mockCoinbaseLotteryWinnersNodeRegistrationQueryScanFail struct {
		query.NodeRegistrationQuery
	}
)

func (*mockCoinbaseLotteryWinnersQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	for idx, arg := range args {
		if idx == 0 {
			nodeID := fmt.Sprintf("%d", arg)
			switch nodeID {
			case "1":
				mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
					"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
				}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 100))
			case "2":
				mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
					"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
				}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
			case "3":
				mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
					"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
				}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
			}
		}
	}
	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockCoinbaseLotteryWinnersQueryExecutorSelectFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("mocked error")
}

func (*mockCoinbaseLotteryWinnersNodeRegistrationQueryScanFail) Scan(
	nr *model.NodeRegistration, row *sql.Row,
) error {
	return sql.ErrNoRows
}

func TestBlockService_CoinbaseLotteryWinners(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		rng                   *crypto.RandomNumberGenerator
	}
	type args struct {
		activeRegistries []storage.NodeRegistry
		scoreSum         int64
		blockTimestamp   int64
		previousBlock    *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		want    [][]byte
		args    args
		wantErr bool
	}{
		{
			name: "CoinbaseLotteryWinners:success",
			fields: fields{
				QueryExecutor:         &mockCoinbaseLotteryWinnersQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				rng:                   crypto.NewRandomNumberGenerator(),
			},
			args: args{
				activeRegistries: []storage.NodeRegistry{
					{
						Node: model.NodeRegistration{
							NodeID:         1,
							NodePublicKey:  bcsNodePubKey1,
							AccountAddress: bcsAddress1,
						},
						ParticipationScore: 1,
					},
					{
						Node: model.NodeRegistration{
							NodeID:         2,
							NodePublicKey:  bcsNodePubKey2,
							AccountAddress: bcsAddress2,
						},
						ParticipationScore: 10,
					},
					{
						Node: model.NodeRegistration{
							NodeID:         3,
							NodePublicKey:  bcsNodePubKey3,
							AccountAddress: bcsAddress3,
						},
						ParticipationScore: 5,
					},
				},
				scoreSum:       (1 / 3) + (10 / 3) + (5 / 3),
				blockTimestamp: 10,
				previousBlock: &model.Block{
					Timestamp: 1,
				},
			},
			want: [][]byte{
				bcsAddress2,
				bcsAddress2,
				bcsAddress2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &CoinbaseService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Rng:                   tt.fields.rng,
			}
			got, err := bs.CoinbaseLotteryWinners(
				tt.args.activeRegistries, tt.args.scoreSum, tt.args.blockTimestamp, tt.args.previousBlock,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.CoinbaseLotteryWinners() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.CoinbaseLotteryWinners() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoinbaseService_GetCoinbase(t *testing.T) {
	type fields struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		Chaintype             chaintype.ChainType
	}
	type args struct {
		blockTimestamp         int64
		previousBlockTimestamp int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				blockTimestamp:         (&chaintype.MainChain{}).GetGenesisBlockTimestamp() + 15,
				previousBlockTimestamp: (&chaintype.MainChain{}).GetGenesisBlockTimestamp(),
			},
			want: 430209406,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbs := &CoinbaseService{
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				Chaintype:             tt.fields.Chaintype,
			}
			if got := cbs.GetCoinbase(tt.args.blockTimestamp, tt.args.previousBlockTimestamp); got != tt.want {
				t.Errorf("CoinbaseService.GetCoinbase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCoinbaseService(t *testing.T) {
	type args struct {
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		queryExecutor         query.ExecutorInterface
		chaintype             chaintype.ChainType
		rng                   *crypto.RandomNumberGenerator
	}
	tests := []struct {
		name string
		args args
		want *CoinbaseService
	}{
		{
			name: "NewCoinbaseService-success",
			args: args{
				nodeRegistrationQuery: nil,
				queryExecutor:         nil,
				rng:                   crypto.NewRandomNumberGenerator(),
			},
			want: &CoinbaseService{
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
				Rng:                   crypto.NewRandomNumberGenerator(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCoinbaseService(
				tt.args.nodeRegistrationQuery,
				tt.args.queryExecutor,
				tt.args.chaintype,
				tt.args.rng,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCoinbaseService() = %v, want %v", got, tt.want)
			}
		})
	}
}
