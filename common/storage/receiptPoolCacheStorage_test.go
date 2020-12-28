package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewReceiptPoolCacheStorage(t *testing.T) {
	tests := []struct {
		name string
		want *ReceiptPoolCacheStorage
	}{
		{
			name: "TestNewReceiptPoolCacheStorage:Success",
			want: &ReceiptPoolCacheStorage{
				RWMutex:  sync.RWMutex{},
				receipts: []model.Receipt{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewReceiptPoolCacheStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReceiptPoolCacheStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestReceiptPoolCacheStorage_ClearCache:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_GetAllItems(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
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
			name: "TestReceiptPoolCacheStorage_GetAllItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				items: &[]model.Receipt{},
			},
			wantErr: false,
		},
		{
			name: "TestReceiptPoolCacheStorage_GetAllItems:Fail-InvalidBatchReceipt",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.GetAllItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_GetItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	type args struct {
		in0 interface{}
		in1 interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestReceiptPoolCacheStorage_GetItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				in0: nil,
				in1: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.GetItem(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_GetSize(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestReceiptPoolCacheStorage_GetSize:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if got := brs.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestReceiptPoolCacheStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if got := brs.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
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
			name: "TestReceiptPoolCacheStorage_RemoveItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				in0: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.RemoveItem(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	type args struct {
		in0  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestReceiptPoolCacheStorage_SetItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				in0:  nil,
				item: model.Receipt{},
			},
			wantErr: false,
		},
		{
			name: "TestReceiptPoolCacheStorage_SetItem:Fail-InvalidBatchReceiptItem",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				in0:  nil,
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.SetItem(tt.args.in0, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
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
			name: "TestReceiptPoolCacheStorage_SetItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				items: []model.Receipt{},
			},
			wantErr: false,
		},
		{
			name: "TestReceiptPoolCacheStorage_SetItems:Fail-InvalidBatchReceiptItem",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.SetItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_size(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts []model.Receipt
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestReceiptPoolCacheStorage_size:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				receipts: []model.Receipt{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if got := brs.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}
