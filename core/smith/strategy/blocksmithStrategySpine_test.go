package strategy

import (
	"bytes"
	"database/sql"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockQueryGetBlocksmithsSpineSuccessNoBlocksmith struct {
		query.Executor
	}
	mockQueryGetBlocksmithsSpineSuccessWithBlocksmith struct {
		query.Executor
	}

	mockQuerySortBlocksmithSpineSuccessWithBlocksmiths struct {
		query.Executor
	}
	mockQueryGetBlocksmithsSpineFail struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksmithsSpineFail) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockQueryGetBlocksmithsSpineSuccessWithBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		rows *sql.Rows
	)
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height <= 1 AND public_key_action=0 AND latest=1":
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows(
			[]string{
				"node_public_key",
				"public_key_action",
				"latest",
				"height",
			},
		).AddRow(
			bssMockBlocksmiths[0].NodePublicKey,
			uint32(model.SpinePublicKeyAction_AddKey),
			true,
			uint32(1),
		))
		rows, _ = db.Query("A")
	default:
		return nil, errors.New("MockErr")
	}
	return rows, nil
}

func (*mockQuerySortBlocksmithSpineSuccessWithBlocksmiths) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		rows *sql.Rows
	)
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height <= 1 AND public_key_action=0 AND latest=1":
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows(
			[]string{
				"node_public_key",
				"public_key_action",
				"latest",
				"height",
			},
		).AddRow(
			bssMockBlocksmiths[0].NodePublicKey,
			uint32(model.SpinePublicKeyAction_AddKey),
			true,
			uint32(1),
		).AddRow(
			bssMockBlocksmiths[1].NodePublicKey,
			uint32(model.SpinePublicKeyAction_AddKey),
			true,
			uint32(1),
		))
		rows, _ = db.Query("A")
	default:
		return nil, errors.New("MockErr")
	}
	return rows, nil
}

func (*mockQueryGetBlocksmithsSpineSuccessNoBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows(
		[]string{
			"node_public_key",
			"public_key_action",
			"latest",
			"height",
		},
	))
	rows, _ := db.Query("A")
	return rows, nil
}

func TestBlocksmithStrategySpine_GetSmithTime(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
	}
	type args struct {
		blocksmithIndex int64
		block           *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "GetSmithTime:0",
			fields: fields{
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args: args{
				blocksmithIndex: 0,
				block: &model.Block{
					Timestamp: 0,
				},
			},
			want: 30,
		},
		{
			name: "GetSmithTime:1",
			fields: fields{
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: 120000 + 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
			}
			if got := bss.GetSmithTime(tt.args.blocksmithIndex, tt.args.block); got != tt.want {
				t.Errorf("BlocksmithStrategySpine.GetSmithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_GetBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Blocksmith
		wantErr bool
	}{
		{
			name: "GetBlocksmiths:success",
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessWithBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args:    args{mockBlock},
			wantErr: false,
			want: []*model.Blocksmith{
				{
					Chaintype:     &chaintype.SpineChain{},
					NodeID:        0,
					BlockSeed:     -1965565459747201754,
					NodeOrder:     new(big.Int).SetInt64(28),
					Score:         big.NewInt(constant.DefaultParticipationScore),
					NodePublicKey: bssMockBlocksmiths[0].NodePublicKey,
				},
			},
		},
		{
			name: "GetBlocksmiths:fail-{sqlSelectErr}",
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineFail{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args:    args{mockBlock},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlocksmiths:fail-{noSpinePublicKeyFound}",
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessNoBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args:    args{mockBlock},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
			}
			got, err := bss.GetBlocksmiths(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlocksmithStrategySpine.GetBlocksmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.GetBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_SortBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Success",
			fields: fields{
				QueryExecutor:        &mockQuerySortBlocksmithSpineSuccessWithBlocksmiths{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args: args{
				block: mockBlock,
			},
		},
		{
			name: "GetBlocksmiths:fail-{noSpinePublicKeyFound}",
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineFail{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
			},
			args: args{mockBlock},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
			}
			bss.SortedBlocksmiths = make([]*model.Blocksmith, 0)
			bss.SortBlocksmiths(tt.args.block)
			if len(bss.SortedBlocksmiths) > 0 && !bytes.Equal(bss.SortedBlocksmiths[0].NodePublicKey,
				bssMockBlocksmiths[1].NodePublicKey) && !bytes.Equal(bss.SortedBlocksmiths[1].NodePublicKey,
				bssMockBlocksmiths[0].NodePublicKey) {
				t.Errorf("sorting fail")
			}
		})
	}
}

func TestBlocksmithStrategySpine_GetSortedBlocksmithsMap(t *testing.T) {
	var mockBlocksmithMap = make(map[string]*int64)
	for index, mockBlocksmith := range bssMockBlocksmiths {
		mockIndex := int64(index)
		mockBlocksmithMap[string(mockBlocksmith.NodePublicKey)] = &mockIndex
	}
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*int64
	}{
		{
			name: "success:noSorting",
			args: args{
				block: &model.Block{ID: 0},
			},
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessNoBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    bssMockBlocksmiths,
				SortedBlocksmithsMap: mockBlocksmithMap,
			},
			want: mockBlocksmithMap,
		},
		{
			name: "success",
			args: args{
				block: &model.Block{ID: 1},
			},
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessNoBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    bssMockBlocksmiths,
				SortedBlocksmithsMap: mockBlocksmithMap,
			},
			want: mockBlocksmithMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
			}
			if got := bss.GetSortedBlocksmithsMap(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.GetSortedBlocksmithsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_GetSortedBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.Blocksmith
	}{
		{
			name: "success : last sorted block id = incoming block id",
			args: args{
				block: &model.Block{ID: 0},
			},
			fields: fields{
				QueryExecutor:         nil,
				SpinePublicKeyQuery:   nil,
				Logger:                log.New(),
				SortedBlocksmiths:     bssMockBlocksmiths,
				SortedBlocksmithsMap:  nil,
				SortedBlocksmithsLock: sync.RWMutex{},
			},
			want: bssMockBlocksmiths,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
			}
			if got := bss.GetSortedBlocksmiths(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.GetSortedBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_CalculateSmith(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
	}
	type args struct {
		lastBlock       *model.Block
		blocksmithIndex int64
		generator       *model.Blocksmith
		score           int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Blocksmith
		wantErr bool
	}{
		{
			name: "CalculateSmith:success",
			args: args{
				lastBlock:       mockBlock,
				blocksmithIndex: 0,
				generator: &model.Blocksmith{
					NodePublicKey: bssNodePubKey1,
					NodeID:        1,
				},
				score: constant.DefaultParticipationScore,
			},
			want: &model.Blocksmith{
				NodePublicKey: bssNodePubKey1,
				NodeID:        1,
				Score:         big.NewInt(1000000000),
				SmithTime:     30,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
			}
			err := bss.CalculateSmith(tt.args.lastBlock, tt.args.blocksmithIndex, tt.args.generator, tt.args.score)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlocksmithStrategySpine.CalculateSmith() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := tt.args.generator
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.CalculateSmith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlocksmithStrategySpine(t *testing.T) {
	type args struct {
		queryExecutor       query.ExecutorInterface
		spinePublicKeyQuery query.SpinePublicKeyQueryInterface
		logger              *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategySpine
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategySpine(nil, nil, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategySpine(tt.args.queryExecutor, tt.args.spinePublicKeyQuery,
				tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategySpine() = %v, want %v", got, tt.want)
			}
		})
	}
}
