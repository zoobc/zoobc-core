package service

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

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
		blocksmiths []*model.Blocksmith
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		args    args
		wantErr bool
	}{
		{
			name: "CoinbaseLotteryWinners:success",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
			args: args{
				blocksmiths: mockCoinbaseLotteryWinnersBlocksmiths,
			},
			want: []string{
				bcsAddress2,
				bcsAddress3,
				bcsAddress1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &CoinbaseService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			got, err := bs.CoinbaseLotteryWinners(tt.args.blocksmiths)
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
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name:   "Success-50 Zoobc",
			fields: fields{},
			want:   50 * constant.OneZBC,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := &CoinbaseService{
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := co.GetCoinbase(); got != tt.want {
				t.Errorf("GetCoinbase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCoinbaseService(t *testing.T) {
	type args struct {
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		queryExecutor         query.ExecutorInterface
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
			if got := NewCoinbaseService(tt.args.nodeRegistrationQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCoinbaseService() = %v, want %v", got, tt.want)
			}
		})
	}
}
