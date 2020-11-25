package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

func TestBlocksStorage_Clear(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_Clear:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_GetAll(t *testing.T) {
	mockItems := []BlockCacheObject{
		{
			ID:        1,
			Height:    1,
			Timestamp: 1,
			BlockHash: make([]byte, 32),
		},
		{
			ID:        2,
			Height:    2,
			Timestamp: 2,
			BlockHash: make([]byte, 32),
		},
	}
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		items interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_GetAll:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        1,
						Height:    1,
						Timestamp: 1,
						BlockHash: make([]byte, 32),
					},
					{
						ID:        2,
						Height:    2,
						Timestamp: 2,
						BlockHash: make([]byte, 32),
					},
				},
			},
			args: args{
				items: &mockItems,
			},
			wantErr: false,
		},
		{
			name: "TestBlocksStorage_GetAll:Fail-ItemError",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        1,
						Height:    1,
						Timestamp: 1,
						BlockHash: make([]byte, 32),
					},
					{
						ID:        2,
						Height:    2,
						Timestamp: 2,
						BlockHash: make([]byte, 32),
					},
				},
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.GetAll(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_GetAtIndex(t *testing.T) {
	mockBlockCacheObject := []BlockCacheObject{
		{
			ID:        9,
			Height:    9,
			Timestamp: 9,
			BlockHash: make([]byte, 32),
		},
	}
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		height uint32
		item   interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_GetAtIndex:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 9,
				blocks:          mockBlockCacheObject,
			},
			args: args{
				height: 9,
				item: &BlockCacheObject{
					ID:        9,
					Height:    9,
					Timestamp: 9,
					BlockHash: make([]byte, 32),
				},
			},
			wantErr: false,
		},
		{
			name: "TestBlocksStorage_GetAtIndex:Fail-ErrorBlockCacheObject",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 9,
				blocks:          mockBlockCacheObject,
			},
			args: args{
				height: 9,
				item:   nil,
			},
			wantErr: true,
		},
		{
			name: "TestBlocksStorage_GetAtIndex:Fail-IndexOutOfRange",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 9,
				blocks:          mockBlockCacheObject,
			},
			args: args{
				height: 10,
				item: &BlockCacheObject{
					ID:        10,
					Height:    9,
					Timestamp: 9,
					BlockHash: make([]byte, 32),
				},
			},
			wantErr: true,
		},
		{
			name: "TestBlocksStorage_GetAtIndex:Fail-IndexOutOfRange",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 10,
				blocks:          mockBlockCacheObject,
			},
			args: args{
				height: 9,
				item: &BlockCacheObject{
					ID:        9,
					Height:    9,
					Timestamp: 9,
					BlockHash: make([]byte, 32),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.GetAtIndex(tt.args.height, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAtIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_GetTop(t *testing.T) {
	mockBlockCacheObject := []BlockCacheObject{
		{
			ID:        9,
			Height:    9,
			Timestamp: 9,
			BlockHash: make([]byte, 32),
		},
	}
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_GetTop:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          mockBlockCacheObject,
			},
			args: args{
				item: &BlockCacheObject{
					ID:        10,
					Height:    10,
					Timestamp: 10,
					BlockHash: make([]byte, 32),
				},
			},
			wantErr: false,
		},
		{
			name: "TestBlocksStorage_GetTop:Fail-EmptyBlockCache",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          []BlockCacheObject{},
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestBlocksStorage_GetTop:Fail-ErrorNotBlockCacheObject",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        0,
						Height:    0,
						Timestamp: 0,
						BlockHash: nil,
					},
				},
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.GetTop(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetTop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_Pop(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_Pop:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.Pop(); (err != nil) != tt.wantErr {
				t.Errorf("Pop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_PopTo(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_PopTo:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        0,
						Height:    0,
						Timestamp: 0,
						BlockHash: nil,
					},
				},
			},
			args: args{
				height: 0,
			},
			wantErr: false,
		},
		{
			name: "TestBlocksStorage_PopTo:Fail-HeightOutOfRange",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        0,
						Height:    0,
						Timestamp: 0,
						BlockHash: nil,
					},
				},
			},
			args: args{
				height: 1,
			},
			wantErr: true,
		},
		{
			name: "TestBlocksStorage_PopTo:Fail-HeightOutOfRange",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 2,
				blocks: []BlockCacheObject{
					{
						ID:        0,
						Height:    0,
						Timestamp: 0,
						BlockHash: nil,
					},
				},
			},
			args: args{
				height: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.PopTo(tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("PopTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_Push(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlocksStorage_Push:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			args: args{
				item: BlockCacheObject{
					ID:        0,
					Height:    0,
					Timestamp: 0,
					BlockHash: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "TestBlocksStorage_Push:Fail-NotBlockCacheObject",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestBlocksStorage_Push:Success:RemoveFirstCache",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks: []BlockCacheObject{
					{
						ID:        0,
						Height:    0,
						Timestamp: 0,
						BlockHash: nil,
					},
					{
						ID:        1,
						Height:    1,
						Timestamp: 1,
						BlockHash: nil,
					},
				},
			},
			args: args{
				item: BlockCacheObject{
					ID:        1,
					Height:    1,
					Timestamp: 1,
					BlockHash: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if err := b.Push(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_copy(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	type args struct {
		blockCacheObject BlockCacheObject
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantBlockCacheObjectCopy BlockCacheObject
	}{
		{
			name: "TestBlocksStorage_copy:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			args: args{
				blockCacheObject: BlockCacheObject{
					ID:        0,
					Height:    0,
					Timestamp: 0,
					BlockHash: make([]byte, 32),
				},
			},
			wantBlockCacheObjectCopy: BlockCacheObject{
				ID:        0,
				Height:    0,
				Timestamp: 0,
				BlockHash: make([]byte, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if gotBlockCacheObjectCopy := b.copy(tt.args.blockCacheObject); !reflect.DeepEqual(gotBlockCacheObjectCopy, tt.wantBlockCacheObjectCopy) {
				t.Errorf("copy() = %v, want %v", gotBlockCacheObjectCopy, tt.wantBlockCacheObjectCopy)
			}
		})
	}
}

func TestBlocksStorage_size(t *testing.T) {
	type fields struct {
		RWMutex         sync.RWMutex
		itemLimit       int
		lastBlockHeight uint32
		blocks          []BlockCacheObject
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestBlocksStorage_size:Success",
			fields: fields{
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
			},
			want: 104,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlocksStorage{
				RWMutex:         tt.fields.RWMutex,
				itemLimit:       tt.fields.itemLimit,
				lastBlockHeight: tt.fields.lastBlockHeight,
				blocks:          tt.fields.blocks,
			}
			if got := b.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlocksStorage(t *testing.T) {
	tests := []struct {
		name string
		want *BlocksStorage
	}{
		{
			name: "TestNewBlocksStorage:Success",
			want: &BlocksStorage{
				itemLimit:       int(constant.MaxBlocksCacheStorage),
				lastBlockHeight: 0,
				blocks:          make([]BlockCacheObject, 0, constant.MinRollbackBlocks),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
