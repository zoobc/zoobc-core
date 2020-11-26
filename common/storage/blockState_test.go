package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestBlockStateStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestBlockStateStorage_ClearCache:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockStateStorage_GetAllItems(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
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
			name: "TestBlockStateStorage_GetAllItems:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.GetAllItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockStateStorage_GetItem(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	type args struct {
		lastUpdate interface{}
		block      interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlockStateStorage_GetItem:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: make([]byte, 32),
			},
			args: args{
				lastUpdate: nil,
				block: &model.Block{
					ID:                   0,
					BlockHash:            make([]byte, 32),
					PreviousBlockHash:    make([]byte, 32),
					Height:               0,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "",
					BlocksmithPublicKey:  nil,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Version:              0,
					PayloadLength:        0,
					PayloadHash:          nil,
					MerkleRoot:           nil,
					MerkleTree:           nil,
					ReferenceBlockHeight: 0,
					Transactions:         nil,
					PublishedReceipts:    nil,
					SpinePublicKeys:      nil,
					SpineBlockManifests:  nil,
					TransactionIDs:       nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.GetItem(tt.args.lastUpdate, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockStateStorage_GetSize(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestBlockStateStorage_GetSize:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: make([]byte, 32),
			},
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if got := bs.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockStateStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestBlockStateStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if got := bs.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockStateStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlockStateStorage_RemoveItem:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			args: args{
				key: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockStateStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	type args struct {
		lastUpdate interface{}
		block      interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlockStateStorage_SetItem:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: make([]byte, 32),
			},
			args: args{
				lastUpdate: nil,
				block:      model.Block{},
			},
			wantErr: false,
		},
		{
			name: "TestBlockStateStorage_SetItem:Fail-ErrorWrongTypeItem",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: make([]byte, 32),
			},
			args: args{
				lastUpdate: nil,
				block:      nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.SetItem(tt.args.lastUpdate, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockStateStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		lastBlockBytes []byte
	}
	type args struct {
		in0 interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestBlockStateStorage_SetItems:Success",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			args: args{
				in0: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:        tt.fields.RWMutex,
				lastBlockBytes: tt.fields.lastBlockBytes,
			}
			if err := bs.SetItems(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBlockStateStorage(t *testing.T) {
	tests := []struct {
		name string
		want *BlockStateStorage
	}{
		{
			name: "TestNewBlockStateStorage:Success",
			want: &BlockStateStorage{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockStateStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockStateStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
