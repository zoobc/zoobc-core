package service

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewBlockPoolService(t *testing.T) {
	tests := []struct {
		name string
		want *BlockPoolService
	}{
		// TODO: Add test cases.
		{
			name: "NewBlockPoolService:success",
			want: &BlockPoolService{
				BlockQueue: make(map[int64]*model.Block),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockPoolService(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockPoolService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_GetBlocks(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int64]*model.Block
	}{
		// TODO: Add test cases.
		{
			name: "GetBlocks:success",
			fields: fields{
				// BlockQueueLock: sync.RWMutex,
				BlockQueue: map[int64]*model.Block{
					0: &model.Block{},
				},
			},
			want: map[int64]*model.Block{
				0: &model.Block{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			if got := bps.GetBlocks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockPoolService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_GetBlock(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	type args struct {
		index int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		// TODO: Add test cases.
		{
			name: "GetBlock:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: &model.Block{},
				},
			},
			want: &model.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			if got := bps.GetBlock(tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockPoolService.GetBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_InsertBlock(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	type args struct {
		block *model.Block
		index int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "InsertBlock:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: &model.Block{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			bps.InsertBlock(tt.args.block, tt.args.index)
		})
	}
}

func TestBlockPoolService_ClearBlockPool(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "ClearBlockPool:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: &model.Block{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			bps.ClearBlockPool()
		})
	}
}
