package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestNewReceiptReminderStorage(t *testing.T) {
	tests := []struct {
		name string
		want *ReceiptReminderStorage
	}{
		{
			name: "TestNewReceiptReminderStorage:Success",
			want: &ReceiptReminderStorage{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewReceiptReminderStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReceiptReminderStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptReminderStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestReceiptReminderStorage_ClearCache:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptReminderStorage_GetAllItems(t *testing.T) {
	mockKey := map[string]chaintype.ChainType{}
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
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
			name: "TestReceiptReminderStorage_GetAllItems:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key: &mockKey,
			},
			wantErr: false,
		},
		{
			name: "TestReceiptReminderStorage_GetAllItems:Fail-WrongKey",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: nil,
			},
			args: args{
				key: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.GetAllItems(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptReminderStorage_GetItem(t *testing.T) {
	mockItem := chaintype.GetChainType(1)
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
	}
	type args struct {
		key  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestReceiptReminderStorage_GetItem:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  "test",
				item: &mockItem,
			},
			wantErr: false,
		},
		{
			name: "TestReceiptReminderStorage_GetItem:Fail-WrongTypeKey",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  nil,
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestReceiptReminderStorage_GetItem:Fail-WrongTypeItem",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  "test",
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.GetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptReminderStorage_GetSize(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestReceiptReminderStorage_GetSize",
			fields: fields{
				RWMutex: sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{
					"1": chaintype.GetChainType(1),
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if got := rs.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptReminderStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestReceiptReminderStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if got := rs.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiptReminderStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
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
			name: "TestReceiptReminderStorage_RemoveItem:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: make(map[string]chaintype.ChainType),
			},
			args: args{
				key: "test",
			},
			wantErr: false,
		},
		{
			name: "TestReceiptReminderStorage_RemoveItem:Fail-WrongTypeKey",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: nil,
			},
			args: args{
				key: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptReminderStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
	}
	type args struct {
		key  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestReceiptReminderStorage_SetItem:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  "test",
				item: chaintype.GetChainType(1),
			},
			wantErr: false,
		},
		{
			name: "TestReceiptReminderStorage_SetItem:Fail-WrongTypeKey",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  1,
				item: chaintype.GetChainType(1),
			},
			wantErr: true,
		},
		{
			name: "TestReceiptReminderStorage_SetItem:Fail-WrongTypeItem",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: map[string]chaintype.ChainType{},
			},
			args: args{
				key:  "1",
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.SetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptReminderStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex   sync.RWMutex
		reminders map[string]chaintype.ChainType
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
			name: "TestReceiptReminderStorage_SetItems:Success",
			fields: fields{
				RWMutex:   sync.RWMutex{},
				reminders: nil,
			},
			args: args{
				in0: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptReminderStorage{
				RWMutex:   tt.fields.RWMutex,
				reminders: tt.fields.reminders,
			}
			if err := rs.SetItems(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
