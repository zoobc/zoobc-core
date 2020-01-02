package service

import (
	"database/sql"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockBlock = &model.Block{
		BlockSeed: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		Height:    1,
	}
)

type (
	mockQueryExecutorGetBlocksmithsSuccessNoBlocksmith struct {
		query.Executor
	}
	mockQueryExecutorGetBlocksmithsSuccessWithBlocksmith struct {
		query.Executor
	}

	mockQueryExecutorSortBlocksmithSuccessWithBlocksmiths struct {
		query.Executor
	}
	mockQueryExecutorGetBlocksmithsFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetBlocksmithsFail) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockQueryExecutorGetBlocksmithsSuccessNoBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockNodeRegistrationQuery := query.NewNodeRegistrationQuery()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockNodeRegistrationQuery.Fields))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQueryExecutorSortBlocksmithSuccessWithBlocksmiths) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		mockBlocksmiths[0].NodeID,
		mockBlocksmiths[0].NodePublicKey,
		mockBlocksmiths[0].Score.String(),
		uint32(1),
	).AddRow(
		mockBlocksmiths[1].NodeID,
		mockBlocksmiths[1].NodePublicKey,
		mockBlocksmiths[1].Score.String(),
		uint32(1),
	))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQueryExecutorGetBlocksmithsSuccessWithBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		mockBlocksmiths[0].NodeID,
		mockBlocksmiths[0].NodePublicKey,
		mockBlocksmiths[0].Score.String(),
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
				QueryExecutor:         &mockQueryExecutorGetBlocksmithsFail{},
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
				QueryExecutor:         &mockQueryExecutorGetBlocksmithsSuccessNoBlocksmith{},
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
				QueryExecutor:         &mockQueryExecutorGetBlocksmithsSuccessWithBlocksmith{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args:    args{mockBlock},
			wantErr: false,
			want: []*model.Blocksmith{
				{
					NodeID:        mockBlocksmiths[0].NodeID,
					BlockSeed:     -7765827254621503546,
					NodeOrder:     new(big.Int).SetInt64(13195850646937615),
					Score:         mockBlocksmiths[0].Score,
					NodePublicKey: mockBlocksmiths[0].NodePublicKey,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
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
				SortedBlocksmiths:        mockBlocksmiths,
				SortedBlocksmithsMap:     nil,
				SortedBlocksmithsMapLock: sync.RWMutex{},
			},
			want: mockBlocksmiths,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
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
	for index, mockBlocksmith := range mockBlocksmiths {
		mockIndex := int64(index)
		mockBlocksmithMap[string(mockBlocksmith.NodePublicKey)] = &mockIndex
	}
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
		want   map[string]*int64
	}{
		{
			name: "success",
			fields: fields{
				QueryExecutor:            nil,
				NodeRegistrationQuery:    nil,
				Logger:                   nil,
				SortedBlocksmiths:        mockBlocksmiths,
				SortedBlocksmithsMap:     mockBlocksmithMap,
				SortedBlocksmithsMapLock: sync.RWMutex{},
			},
			want: mockBlocksmithMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     1,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsMapLock,
			}
			if got := bss.GetSortedBlocksmithsMap(&model.Block{ID: 1}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSortedBlocksmithsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithService_SortBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		SortedBlocksmithsMap     map[string]*int64
		SortedBlocksmithsMapLock sync.RWMutex
		LastSortedBlockID        int64
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
				QueryExecutor:            &mockQueryExecutorSortBlocksmithSuccessWithBlocksmiths{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				Logger:                   log.New(),
				SortedBlocksmiths:        nil,
				SortedBlocksmithsMap:     make(map[string]*int64),
				SortedBlocksmithsMapLock: sync.RWMutex{},
				LastSortedBlockID:        1,
			},
			args: args{
				block: mockBlock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsMapLock,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
			}
			bss.SortBlocksmiths(tt.args.block)
			if bss.SortedBlocksmiths[0].NodeID != mockBlocksmiths[0].NodeID &&
				bss.SortedBlocksmiths[1].NodeID != mockBlocksmiths[1].NodeID {
				t.Errorf("sorting fail")
			}
		})
	}
}

func TestNewBlocksmithService(t *testing.T) {
	type args struct {
		queryExecutor         query.ExecutorInterface
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		logger                *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithService
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithService(nil, nil, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithService(tt.args.queryExecutor, tt.args.nodeRegistrationQuery, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}
