package strategy

import (
	"database/sql"
	"errors"
	"github.com/zoobc/zoobc-core/common/storage"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

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

type (
	mockActiveNodeRegistryCacheSuccess struct {
		storage.NodeRegistryCacheStorage
	}
	mockActiveNodeRegistryCacheSuccessWithContent struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockActiveNodeRegistryCacheSuccess) GetAllItems(item interface{}) error {
	castedItem := item.(*[]storage.NodeRegistry)
	*castedItem = make([]storage.NodeRegistry, 0)
	return nil
}

func (*mockActiveNodeRegistryCacheSuccessWithContent) GetAllItems(item interface{}) error {
	castedItem := item.(*[]storage.NodeRegistry)
	*castedItem = []storage.NodeRegistry{
		{
			Node: model.NodeRegistration{
				NodeID:        bssMockBlocksmiths[0].NodeID,
				NodePublicKey: bssMockBlocksmiths[0].NodePublicKey,
				Latest:        true,
				Height:        0,
			},
			ParticipationScore: bssMockBlocksmiths[0].Score.Int64(),
		},
	}
	return nil
}

func TestBlocksmithService_GetBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		SortedBlocksmithsMap     map[string]*int64
		SortedBlocksmithsMapLock sync.RWMutex
		ActiveNodeRegistryCache  storage.CacheStorageInterface
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
			name: "success - no blocksmiths",
			fields: fields{
				QueryExecutor:           &mockQueryGetBlocksmithsMainSuccessNoBlocksmith{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSuccess{},

				Logger: log.New(),
			},
			args:    args{&model.Block{}},
			wantErr: false,
			want:    nil,
		},
		{
			name: "success - with blocksmiths",
			fields: fields{
				QueryExecutor:           &mockQueryGetBlocksmithsMainSuccessWithBlocksmith{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSuccessWithContent{},
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
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Logger:                  tt.fields.Logger,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:    tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsLock:   tt.fields.SortedBlocksmithsMapLock,
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
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Logger                  *log.Logger
		SortedBlocksmiths       []*model.Blocksmith
		SortedBlocksmithsMap    map[string]*int64
		ActiveNodeRegistryCache storage.CacheStorageInterface
		LastSortedBlockID       int64
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
				QueryExecutor:           &mockQuerySortBlocksmithMainSuccessWithBlocksmiths{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				Logger:                  log.New(),
				SortedBlocksmiths:       nil,
				SortedBlocksmithsMap:    make(map[string]*int64),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSuccessWithContent{},
				LastSortedBlockID:       1,
			},
			args: args{
				block: mockBlock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Logger:                  tt.fields.Logger,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:    tt.fields.SortedBlocksmithsMap,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				LastSortedBlockID:       tt.fields.LastSortedBlockID,
			}
			bss.SortBlocksmiths(tt.args.block, true)
			if bss.SortedBlocksmiths[0].NodeID != bssMockBlocksmiths[0].NodeID &&
				bss.SortedBlocksmiths[1].NodeID != bssMockBlocksmiths[1].NodeID {
				t.Errorf("sorting fail")
			}
		})
	}
}

func TestNewBlocksmithService(t *testing.T) {
	type args struct {
		queryExecutor           query.ExecutorInterface
		nodeRegistrationQuery   query.NodeRegistrationQueryInterface
		skippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		logger                  *log.Logger
		activeNodeRegistryCache storage.CacheStorageInterface
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
			want: NewBlocksmithStrategyMain(nil, nil, nil, nil, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategyMain(tt.args.queryExecutor, tt.args.nodeRegistrationQuery,
				tt.args.skippedBlocksmithQuery, tt.args.activeNodeRegistryCache, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategyMain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategyMain_IsBlockTimestampValid(t *testing.T) {
	type fields struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		LastSortedBlockID                      int64
		LastEstimatedBlockPersistedTimestamp   int64
		LastEstimatedPersistedTimestampBlockID int64
		SortedBlocksmithsLock                  sync.RWMutex
		SortedBlocksmithsMap                   map[string]*int64
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
		currentBlock        *model.Block
	}
	mainchain := &chaintype.MainChain{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "blocksmithIndex=0 && timeSinceLast > 15",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   0,
				LastEstimatedPersistedTimestampBlockID: 1,
				SortedBlocksmithsLock:                  sync.RWMutex{},
				SortedBlocksmithsMap:                   nil,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 10,
				previousBlock: &model.Block{
					ID: int64(1),
				},
				currentBlock: &model.Block{
					Timestamp: 16,
				},
			},
			wantErr: false,
		},
		{
			name: "blocksmithIndex=1 && blocksmith expired",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   0,
				LastEstimatedPersistedTimestampBlockID: 1,
				SortedBlocksmithsLock:                  sync.RWMutex{},
				SortedBlocksmithsMap:                   nil,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 10,
				previousBlock: &model.Block{
					ID: int64(1),
				},
				currentBlock: &model.Block{
					Timestamp: 26 + mainchain.GetBlocksmithTimeGap() + mainchain.GetBlocksmithNetworkTolerance() +
						mainchain.GetBlocksmithBlockCreationTime(),
				},
			},
			wantErr: true,
		},
		{
			name: "blocksmithIndex=1 && blocksmith pending",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   0,
				LastEstimatedPersistedTimestampBlockID: 1,
				SortedBlocksmithsLock:                  sync.RWMutex{},
				SortedBlocksmithsMap:                   nil,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 10,
				previousBlock: &model.Block{
					ID: int64(1),
				},
				currentBlock: &model.Block{
					Timestamp: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "blocksmithIndex=1 && blocksmith valid time",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   0,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 10,
				previousBlock: &model.Block{
					ID: int64(1),
				},
				currentBlock: &model.Block{
					Timestamp: 26,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:                          tt.fields.QueryExecutor,
				NodeRegistrationQuery:                  tt.fields.NodeRegistrationQuery,
				SkippedBlocksmithQuery:                 tt.fields.SkippedBlocksmithQuery,
				Logger:                                 log.New(),
				SortedBlocksmiths:                      tt.fields.SortedBlocksmiths,
				LastSortedBlockID:                      tt.fields.LastSortedBlockID,
				LastEstimatedBlockPersistedTimestamp:   tt.fields.LastEstimatedBlockPersistedTimestamp,
				LastEstimatedPersistedTimestampBlockID: tt.fields.LastEstimatedPersistedTimestampBlockID,
				SortedBlocksmithsLock:                  tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:                   tt.fields.SortedBlocksmithsMap,
			}
			if err := bss.IsBlockTimestampValid(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock,
				tt.args.currentBlock); (err != nil) != tt.wantErr {
				t.Errorf("IsBlockTimestampValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksmithStrategyMain_CanPersistBlock(t *testing.T) {
	type fields struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		LastSortedBlockID                      int64
		LastEstimatedBlockPersistedTimestamp   int64
		LastEstimatedPersistedTimestampBlockID int64
		SortedBlocksmithsLock                  sync.RWMutex
		SortedBlocksmithsMap                   map[string]*int64
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "previousBlock-genesis",
			fields: fields{},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Height: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "timeSinceLastBlock < smithing period",
			fields: fields{
				LastSortedBlockID:                      0,
				LastEstimatedBlockPersistedTimestamp:   time.Now().Unix(),
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Height: 1,
					ID:     1,
				},
			},
			wantErr: true,
		},
		{
			name: "can persist block-first round",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   time.Now().Unix() - 65,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 3,
				previousBlock: &model.Block{
					Height: 1,
					ID:     1,
				},
			},
			wantErr: false,
		},
		{
			name: "can persist block-multiple round",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   time.Now().Unix() - 95,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 3,
				previousBlock: &model.Block{
					Height: 1,
					ID:     1,
				},
			},
			wantErr: false,
		},
		{
			name: "canPersist",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   time.Now().Unix() - 1000,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 10,
				previousBlock: &model.Block{
					Height: 1,
					ID:     1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:                          tt.fields.QueryExecutor,
				NodeRegistrationQuery:                  tt.fields.NodeRegistrationQuery,
				SkippedBlocksmithQuery:                 tt.fields.SkippedBlocksmithQuery,
				Logger:                                 tt.fields.Logger,
				SortedBlocksmiths:                      tt.fields.SortedBlocksmiths,
				LastSortedBlockID:                      tt.fields.LastSortedBlockID,
				LastEstimatedBlockPersistedTimestamp:   tt.fields.LastEstimatedBlockPersistedTimestamp,
				LastEstimatedPersistedTimestampBlockID: tt.fields.LastEstimatedPersistedTimestampBlockID,
				SortedBlocksmithsLock:                  tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:                   tt.fields.SortedBlocksmithsMap,
			}
			if err := bss.CanPersistBlock(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock); (err != nil) != tt.wantErr {
				t.Errorf("CanPersistBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksmithStrategyMain_IsValidSmithTime(t *testing.T) {
	type fields struct {
		QueryExecutor                          query.ExecutorInterface
		NodeRegistrationQuery                  query.NodeRegistrationQueryInterface
		SkippedBlocksmithQuery                 query.SkippedBlocksmithQueryInterface
		Logger                                 *log.Logger
		SortedBlocksmiths                      []*model.Blocksmith
		LastSortedBlockID                      int64
		LastEstimatedBlockPersistedTimestamp   int64
		LastEstimatedPersistedTimestampBlockID int64
		SortedBlocksmithsLock                  sync.RWMutex
		SortedBlocksmithsMap                   map[string]*int64
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
	}
	mainchain := &chaintype.MainChain{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TimeSinceLastBlock < SmithPeriod",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp:   time.Now().Unix() - (mainchain.GetSmithingPeriod() - 1),
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 3,
				previousBlock:       &model.Block{ID: 1},
			},
			wantErr: true,
		},
		{
			name: "SmithingPending",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp: time.Now().Unix() - mainchain.GetSmithingPeriod() -
					(mainchain.GetBlocksmithBlockCreationTime() + mainchain.GetBlocksmithNetworkTolerance() + 1),
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 6,
				previousBlock:       &model.Block{ID: 1},
			},
			wantErr: true,
		},
		{
			name: "allowedBegin-one round",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp: time.Now().Unix() - mainchain.GetSmithingPeriod() -
					mainchain.GetBlocksmithTimeGap() - 1,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 6,
				previousBlock:       &model.Block{ID: 1},
			},
			wantErr: false,
		},
		{
			name: "allowedBegin-multiple round",
			fields: fields{
				LastEstimatedBlockPersistedTimestamp: time.Now().Unix() - mainchain.GetSmithingPeriod() -
					(11 * mainchain.GetBlocksmithTimeGap()) - 1,
				LastEstimatedPersistedTimestampBlockID: 1,
			},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 6,
				previousBlock:       &model.Block{ID: 1},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				QueryExecutor:                          tt.fields.QueryExecutor,
				NodeRegistrationQuery:                  tt.fields.NodeRegistrationQuery,
				SkippedBlocksmithQuery:                 tt.fields.SkippedBlocksmithQuery,
				Logger:                                 tt.fields.Logger,
				SortedBlocksmiths:                      tt.fields.SortedBlocksmiths,
				LastSortedBlockID:                      tt.fields.LastSortedBlockID,
				LastEstimatedBlockPersistedTimestamp:   tt.fields.LastEstimatedBlockPersistedTimestamp,
				LastEstimatedPersistedTimestampBlockID: tt.fields.LastEstimatedPersistedTimestampBlockID,
				SortedBlocksmithsLock:                  tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:                   tt.fields.SortedBlocksmithsMap,
			}
			if err := bss.IsValidSmithTime(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock); (err != nil) != tt.wantErr {
				t.Errorf("IsValidSmithTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
