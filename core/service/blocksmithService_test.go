package service

import (
	"reflect"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				Logger:                   tt.fields.Logger,
				SortedBlocksmiths:        tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:     tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsMapLock: tt.fields.SortedBlocksmithsMapLock,
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				Logger:                   tt.fields.Logger,
				SortedBlocksmiths:        tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:     tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsMapLock: tt.fields.SortedBlocksmithsMapLock,
			}
			if got := bss.GetSortedBlocksmiths(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSortedBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithService_GetSortedBlocksmithsMap(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				Logger:                   tt.fields.Logger,
				SortedBlocksmiths:        tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:     tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsMapLock: tt.fields.SortedBlocksmithsMapLock,
			}
			if got := bss.GetSortedBlocksmithsMap(); !reflect.DeepEqual(got, tt.want) {
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
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				Logger:                   tt.fields.Logger,
				SortedBlocksmiths:        tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:     tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsMapLock: tt.fields.SortedBlocksmithsMapLock,
			}
			bss.SortBlocksmiths(tt.args.block)
		})
	}
}

func TestBlocksmithService_copyBlocksmithsToMap(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		Logger                   *log.Logger
		SortedBlocksmiths        []*model.Blocksmith
		SortedBlocksmithsMap     map[string]*int64
		SortedBlocksmithsMapLock sync.RWMutex
	}
	type args struct {
		blocksmiths []*model.Blocksmith
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithService{
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				Logger:                   tt.fields.Logger,
				SortedBlocksmiths:        tt.fields.SortedBlocksmiths,
				SortedBlocksmithsMap:     tt.fields.SortedBlocksmithsMap,
				SortedBlocksmithsMapLock: tt.fields.SortedBlocksmithsMapLock,
			}
			bss.copyBlocksmithsToMap(tt.args.blocksmiths)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithService(tt.args.queryExecutor, tt.args.nodeRegistrationQuery, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}
