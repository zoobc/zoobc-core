package service

import (
	"database/sql"
	"errors"
	"fmt"
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
	}
	type args struct {
		activeRegistries []storage.NodeRegistry
		scoreSum         float64
		blockTimestamp   int64
		previousBlock    *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		args    args
		wantErr bool
	}{
		{
			name: "WantFail:selectRowFail",
			fields: fields{
				QueryExecutor:         &mockCoinbaseLotteryWinnersQueryExecutorSelectFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				activeRegistries: []storage.NodeRegistry{
					{
						Node:               model.NodeRegistration{},
						ParticipationScore: 1,
					},
				},
				scoreSum:       100,
				blockTimestamp: 10,
				previousBlock: &model.Block{
					Timestamp: 1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "WantFail:ScanFail",
			fields: fields{
				QueryExecutor:         &mockCoinbaseLotteryWinnersQueryExecutorSuccess{},
				NodeRegistrationQuery: &mockCoinbaseLotteryWinnersNodeRegistrationQueryScanFail{},
			},
			args: args{
				activeRegistries: []storage.NodeRegistry{
					{
						Node:               model.NodeRegistration{},
						ParticipationScore: 1,
					},
				},
				scoreSum:       100,
				blockTimestamp: 10,
				previousBlock: &model.Block{
					Timestamp: 1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "CoinbaseLotteryWinners:success",
			fields: fields{
				QueryExecutor:         &mockCoinbaseLotteryWinnersQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
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
				scoreSum:       100,
				blockTimestamp: 10,
				previousBlock: &model.Block{
					Timestamp: 1,
				},
			},
			want: []string{
				bcsAddress1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &CoinbaseService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			got, err := bs.CoinbaseLotteryWinners(tt.args.activeRegistries, tt.args.scoreSum, tt.args.blockTimestamp, tt.args.previousBlock)
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
			want: 86041924,
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
			},
			want: &CoinbaseService{
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCoinbaseService(
				tt.args.nodeRegistrationQuery,
				tt.args.queryExecutor,
				tt.args.chaintype,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCoinbaseService() = %v, want %v", got, tt.want)
			}
		})
	}
}
