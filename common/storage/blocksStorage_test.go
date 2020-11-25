package storage

import (
	"reflect"
	"sync"
	"testing"
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
				blocks:          nil,
			},
			args: args{
				items: nil,
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
			if err := b.GetAll(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_GetAtIndex(t *testing.T) {
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
				lastBlockHeight: 0,
				blocks:          nil,
			},
			args: args{
				height: 0,
				item:   nil,
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
			if err := b.GetAtIndex(tt.args.height, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAtIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksStorage_GetTop(t *testing.T) {
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
				blocks:          nil,
			},
			args: args{
				item: nil,
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
				blocks:          nil,
			},
			args: args{
				height: 0,
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
				item: nil,
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
					BlockHash: nil,
				},
			},
			wantBlockCacheObjectCopy: BlockCacheObject{
				ID:        0,
				Height:    0,
				Timestamp: 0,
				BlockHash: nil,
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
			want: 0,
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
				RWMutex:         sync.RWMutex{},
				itemLimit:       0,
				lastBlockHeight: 0,
				blocks:          nil,
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
