package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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

var (
	// CoinbaseLotteryWinners mock
	mockCoinbaseLotteryWinnersBlocksmiths = []*model.Blocksmith{
		{
			NodeID:        1,
			NodeOrder:     new(big.Int).SetInt64(8000),
			NodePublicKey: []byte{1, 3, 4, 5, 6},
		},
		{
			NodeID:    2,
			NodeOrder: new(big.Int).SetInt64(1000),
		},
		{
			NodeID:    3,
			NodeOrder: new(big.Int).SetInt64(5000),
		},
	}
	// CoinbaseLotteryWinners mock
)

func TestBlockService_CoinbaseLotteryWinners(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		blocksmiths            []*model.Blocksmith
		blockTimestamp         int64
		previousBlockTimestamp int64
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
				blocksmiths:            mockCoinbaseLotteryWinnersBlocksmiths,
				blockTimestamp:         4,
				previousBlockTimestamp: 1,
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
				blocksmiths:            mockCoinbaseLotteryWinnersBlocksmiths,
				blockTimestamp:         4,
				previousBlockTimestamp: 1,
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
				blocksmiths:            mockCoinbaseLotteryWinnersBlocksmiths,
				blockTimestamp:         4,
				previousBlockTimestamp: 1,
			},
			want: []string{
				bcsAddress2,
				bcsAddress3,
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
			got, err := bs.CoinbaseLotteryWinners(tt.args.blocksmiths, tt.args.blockTimestamp, tt.args.previousBlockTimestamp)
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
		blockTimesatamp         int64
		previousBlockTimesatamp int64
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
				blockTimesatamp:         (&chaintype.MainChain{}).GetGenesisBlockTimestamp() + 15,
				previousBlockTimesatamp: (&chaintype.MainChain{}).GetGenesisBlockTimestamp(),
			},
			want: 5234176702,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbs := &CoinbaseService{
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				Chaintype:             tt.fields.Chaintype,
			}
			if got := cbs.GetCoinbase(tt.args.blockTimesatamp, tt.args.previousBlockTimesatamp); got != tt.want {
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
