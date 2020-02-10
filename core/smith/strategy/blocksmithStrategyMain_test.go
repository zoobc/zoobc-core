package strategy

import (
	"database/sql"
	"errors"
	"math"
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/DATA-DOG/go-sqlmock"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockBlock = &model.Block{
		BlockSeed: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		Height:    1,
	}
	bssNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bssNodePubKey2 = []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bssMockBlocksmiths = []*model.Blocksmith{
		{
			NodePublicKey: bssNodePubKey1,
			NodeID:        2,
			NodeOrder:     new(big.Int).SetInt64(1000),
			Score:         new(big.Int).SetInt64(1000),
		},
		{
			NodePublicKey: bssNodePubKey2,
			NodeID:        3,
			NodeOrder:     new(big.Int).SetInt64(2000),
			Score:         new(big.Int).SetInt64(2000),
		},
		{
			NodePublicKey: bssMockBlockData.BlocksmithPublicKey,
			NodeID:        4,
			NodeOrder:     new(big.Int).SetInt64(3000),
			Score:         new(big.Int).SetInt64(3000),
		},
	}
	bssMockBlockData = model.Block{
		ID:        constant.MainchainGenesisBlockID,
		BlockHash: make([]byte, 32),
		PreviousBlockHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
			97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
		Height:    1,
		Timestamp: 1,
		BlockSeed: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
		CumulativeDifficulty: "1000",
		PayloadLength:        1,
		PayloadHash:          []byte{},
		BlocksmithPublicKey: []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		TotalAmount:   1000,
		TotalFee:      0,
		TotalCoinBase: 1,
		Version:       0,
	}
)

type (
	mockQueryGetBlocksmithsMainSuccessNoBlocksmith struct {
		query.Executor
	}
	mockQueryGetBlocksmithsMainSuccessWithBlocksmith struct {
		query.Executor
	}

	mockQuerySortBlocksmithMainSuccessWithBlocksmiths struct {
		query.Executor
	}
	mockQueryGetBlocksmithsMainFail struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksmithsMainFail) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockQueryGetBlocksmithsMainSuccessNoBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockNodeRegistrationQuery := query.NewNodeRegistrationQuery()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockNodeRegistrationQuery.Fields))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQuerySortBlocksmithMainSuccessWithBlocksmiths) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		bssMockBlocksmiths[0].NodeID,
		bssMockBlocksmiths[0].NodePublicKey,
		bssMockBlocksmiths[0].Score.String(),
		uint32(1),
	).AddRow(
		bssMockBlocksmiths[1].NodeID,
		bssMockBlocksmiths[1].NodePublicKey,
		bssMockBlocksmiths[1].Score.String(),
		uint32(1),
	))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQueryGetBlocksmithsMainSuccessWithBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		bssMockBlocksmiths[0].NodeID,
		bssMockBlocksmiths[0].NodePublicKey,
		bssMockBlocksmiths[0].Score.String(),
		uint32(1),
	))
	rows, _ := db.Query("foo")
	return rows, nil
}
func TestBlocksmithService_GetBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		SortedBlocksmithsMap     map[string]*int64
		SortedBlocksmithsMapLock sync.RWMutex
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
			name: "fail - ExecuteSelect Fail",
			fields: fields{
				QueryExecutor:         &mockQueryGetBlocksmithsMainFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args:    args{&model.Block{}},
			wantErr: true,
			want:    nil,
		},
		{
			name: "success - no blocksmiths",
			fields: fields{
				QueryExecutor:         &mockQueryGetBlocksmithsMainSuccessNoBlocksmith{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args:    args{&model.Block{}},
			wantErr: false,
			want:    nil,
		},
		{
			name: "success - with blocksmiths",
			fields: fields{
				QueryExecutor:         &mockQueryGetBlocksmithsMainSuccessWithBlocksmith{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args:    args{mockBlock},
			wantErr: false,
			want: []*model.Blocksmith{
				{
					NodeID:        bssMockBlocksmiths[0].NodeID,
					BlockSeed:     -7765827254621503546,
					NodeOrder:     new(big.Int).SetInt64(13195850646937615),
					Score:         bssMockBlocksmiths[0].Score,
					NodePublicKey: bssMockBlocksmiths[0].NodePublicKey,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsMapLock,
			}
			got, err := bss.GetBlocksmiths(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlocksmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlocksmiths() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithService_GetSortedBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		SortedBlocksmithsMap     map[string]*int64
		SortedBlocksmithsMapLock sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []*model.Blocksmith
	}{
		{
			name: "success : last sorted block id = incoming block id",
			fields: fields{
				QueryExecutor:            nil,
				NodeRegistrationQuery:    nil,
				Logger:                   nil,
				SortedBlocksmiths:        bssMockBlocksmiths,
				SortedBlocksmithsMap:     nil,
				SortedBlocksmithsMapLock: sync.RWMutex{},
			},
			want: bssMockBlocksmiths,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     1,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsMapLock,
			}
			if got := bss.GetSortedBlocksmiths(&model.Block{ID: 1}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSortedBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithService_GetSortedBlocksmithsMap(t *testing.T) {
	var mockBlocksmithMap = make(map[string]*int64)
	for index, mockBlocksmith := range bssMockBlocksmiths {
		mockIndex := int64(index)
		mockBlocksmithMap[string(mockBlocksmith.NodePublicKey)] = &mockIndex
	}
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		SortedBlocksmithsMap  map[string]*int64
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*int64
	}{
		{
			name: "success",
			fields: fields{
				QueryExecutor:         nil,
				NodeRegistrationQuery: nil,
				Logger:                nil,
				SortedBlocksmiths:     bssMockBlocksmiths,
				SortedBlocksmithsMap:  mockBlocksmithMap,
			},
			want: mockBlocksmithMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     1,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
			}
			if got := bss.GetSortedBlocksmithsMap(&model.Block{ID: 1}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSortedBlocksmithsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithService_SortBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		SortedBlocksmithsMap  map[string]*int64
		LastSortedBlockID     int64
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
				QueryExecutor:         &mockQuerySortBlocksmithMainSuccessWithBlocksmiths{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
				SortedBlocksmiths:     nil,
				SortedBlocksmithsMap:  make(map[string]*int64),
				LastSortedBlockID:     1,
			},
			args: args{
				block: mockBlock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
			}
			bss.SortBlocksmiths(tt.args.block, true)
			if bss.SortedBlocksmiths[0].NodeID != bssMockBlocksmiths[0].NodeID &&
				bss.SortedBlocksmiths[1].NodeID != bssMockBlocksmiths[1].NodeID {
				t.Errorf("sorting fail")
			}
		})
	}
}

type (
	mockQueryExecutorGetSmithTimeExecuteSelectFail struct {
		query.Executor
	}
	mockQueryExecutorGetSmithTimeBuildModelFail struct {
		query.Executor
	}
	mockQueryExecutorGetSmithTimeBuildModelSuccess struct {
		query.Executor
	}
	mockSkippedBlocksmithQueryReturnZero struct {
		query.SkippedBlocksmithQuery
	}
	mockSkippedBlocksmithQueryReturnTwo struct {
		query.SkippedBlocksmithQuery
	}
)

func (*mockQueryExecutorGetSmithTimeExecuteSelectFail) ExecuteSelect(
	query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedErr")
}

func (*mockQueryExecutorGetSmithTimeBuildModelFail) ExecuteSelect(
	q string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		[]string{"invalidColumn"},
	).AddRow(-11))
	return db.Query("")
}

func (*mockQueryExecutorGetSmithTimeBuildModelSuccess) ExecuteSelect(
	q string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		[]string{"invalidColumn"},
	).AddRow(-11))
	return db.Query("")
}

func (*mockSkippedBlocksmithQueryReturnZero) BuildModel(
	skippedBlocksmiths []*model.SkippedBlocksmith, rows *sql.Rows) ([]*model.SkippedBlocksmith, error) {
	return make([]*model.SkippedBlocksmith, 0), nil
}

func (*mockSkippedBlocksmithQueryReturnTwo) BuildModel(
	skippedBlocksmiths []*model.SkippedBlocksmith, rows *sql.Rows) ([]*model.SkippedBlocksmith, error) {
	return make([]*model.SkippedBlocksmith, 2), nil
}

func TestGetSmithTime(t *testing.T) {
	type args struct {
		blocksmithIndex int64
		block           *model.Block
	}
	type fields struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		SortedBlocksmithsMap                   map[string]*int64
		LastSortedBlockID                      int64
		LastEstimatedPersistedTimestampBlockID int64
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
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
				SortedBlocksmiths:     nil,
				SortedBlocksmithsMap:  make(map[string]*int64),
				LastSortedBlockID:     1,
			},
			args: args{
				blocksmithIndex: 0,
				block: &model.Block{
					Timestamp: 0,
				},
			},
			want: 15,
		},
		{
			name: "GetSmithTime:1-cached",
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
				SortedBlocksmiths:     nil,
				SortedBlocksmithsMap:  make(map[string]*int64),
				LastSortedBlockID:     1,
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: (&chaintype.MainChain{}).GetSmithingPeriod() + constant.SmithingBlocksmithTimeGap,
		},
		{
			name: "GetSmithTime:1-no-cache : get skipped blocksmith fail",
			fields: fields{
				NodeRegistrationQuery:                  query.NewNodeRegistrationQuery(),
				SkippedBlocksmithQuery:                 query.NewSkippedBlocksmithQuery(),
				Logger:                                 log.New(),
				SortedBlocksmiths:                      nil,
				SortedBlocksmithsMap:                   make(map[string]*int64),
				LastSortedBlockID:                      1,
				LastEstimatedPersistedTimestampBlockID: 1000,
				QueryExecutor:                          &mockQueryExecutorGetSmithTimeExecuteSelectFail{},
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: math.MaxInt64,
		},
		{
			name: "GetSmithTime:1-no-cache : build skipped blocksmith fail",
			fields: fields{
				NodeRegistrationQuery:                  query.NewNodeRegistrationQuery(),
				SkippedBlocksmithQuery:                 query.NewSkippedBlocksmithQuery(),
				Logger:                                 log.New(),
				SortedBlocksmiths:                      nil,
				SortedBlocksmithsMap:                   make(map[string]*int64),
				LastSortedBlockID:                      1,
				LastEstimatedPersistedTimestampBlockID: 1000,
				QueryExecutor:                          &mockQueryExecutorGetSmithTimeBuildModelFail{},
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: math.MaxInt64,
		},
		{
			name: "GetSmithTime:1-no-cache : no previous skipped blocksmith",
			fields: fields{
				NodeRegistrationQuery:                  query.NewNodeRegistrationQuery(),
				SkippedBlocksmithQuery:                 &mockSkippedBlocksmithQueryReturnZero{},
				Logger:                                 log.New(),
				SortedBlocksmiths:                      nil,
				SortedBlocksmithsMap:                   make(map[string]*int64),
				LastSortedBlockID:                      1,
				LastEstimatedPersistedTimestampBlockID: 1000,
				QueryExecutor:                          &mockQueryExecutorGetSmithTimeBuildModelSuccess{},
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: 120000 + (&chaintype.MainChain{}).GetSmithingPeriod() + constant.SmithingBlocksmithTimeGap,
		},
		{
			name: "GetSmithTime:1-no-cache : 2 previous skipped blocksmiths",
			fields: fields{
				NodeRegistrationQuery:                  query.NewNodeRegistrationQuery(),
				SkippedBlocksmithQuery:                 &mockSkippedBlocksmithQueryReturnTwo{},
				Logger:                                 log.New(),
				SortedBlocksmiths:                      nil,
				SortedBlocksmithsMap:                   make(map[string]*int64),
				LastSortedBlockID:                      1,
				LastEstimatedPersistedTimestampBlockID: 1000,
				QueryExecutor:                          &mockQueryExecutorGetSmithTimeBuildModelSuccess{},
			},
			args: args{
				blocksmithIndex: 1,
				block: &model.Block{
					Timestamp: 120000,
				},
			},
			want: (120000 - constant.SmithingBlocksmithTimeGap + constant.SmithingBlockCreationTime +
				constant.SmithingNetworkTolerance) +
				(&chaintype.MainChain{}).GetSmithingPeriod() + constant.SmithingBlocksmithTimeGap,
		},
	}
	for _, tt := range tests {
		bss := &BlocksmithStrategyMain{
			QueryExecutor:                          tt.fields.QueryExecutor,
			NodeRegistrationQuery:                  tt.fields.NodeRegistrationQuery,
			SkippedBlocksmithQuery:                 tt.fields.SkippedBlocksmithQuery,
			Logger:                                 tt.fields.Logger,
			SortedBlocksmiths:                      tt.fields.SortedBlocksmiths,
			SortedBlocksmithsMap:                   tt.fields.SortedBlocksmithsMap,
			LastSortedBlockID:                      tt.fields.LastSortedBlockID,
			LastEstimatedPersistedTimestampBlockID: tt.fields.LastEstimatedPersistedTimestampBlockID,
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := bss.GetSmithTime(tt.args.blocksmithIndex, tt.args.block); got != tt.want {
				t.Errorf("GetSmithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlocksmithService(t *testing.T) {
	type args struct {
		queryExecutor          query.ExecutorInterface
		nodeRegistrationQuery  query.NodeRegistrationQueryInterface
		skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface
		logger                 *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategyMain
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategyMain(nil, nil, nil, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategyMain(tt.args.queryExecutor, tt.args.nodeRegistrationQuery,
				tt.args.skippedBlocksmithQuery, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategyMain() = %v, want %v", got, tt.want)
			}
		})
	}
}
