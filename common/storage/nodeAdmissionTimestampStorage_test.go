package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewNodeAdmissionTimestampStorage(t *testing.T) {
	tests := []struct {
		name string
		want *NodeAdmissionTimestampStorage
	}{
		{
			name: "TestNewNodeAdmissionTimestampStorage:Success",
			want: &NodeAdmissionTimestampStorage{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeAdmissionTimestampStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeAdmissionTimestampStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestNodeAdmissionTimestampStorage_ClearCache:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_GetAllItems(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
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
			name: "TestNodeAdmissionTimestampStorage_GetAllItems:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				item: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.GetAllItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_GetItem(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
	}
	type args struct {
		lastChange interface{}
		item       interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeAdmissionTimestampStorage_GetItem:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            1,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				lastChange: nil,
				item: &model.NodeAdmissionTimestamp{
					Timestamp:   1,
					BlockHeight: 0,
					Latest:      false,
				},
			},
			wantErr: false,
		},
		{
			name: "TestNodeAdmissionTimestampStorage_GetItem:Fail-EmptyNodeAdmissionTimestampStorage",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				lastChange: nil,
				item:       nil,
			},
			wantErr: true,
		},
		{
			name: "TestNodeAdmissionTimestampStorage_GetItem:Fail-WrongTypeItem",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            1,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				lastChange: nil,
				item:       nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.GetItem(tt.args.lastChange, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestNodeAdmissionTimestampStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if got := ns.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
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
			name: "TestNodeAdmissionTimestampStorage_RemoveItem:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				key: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
	}
	type args struct {
		lastChange interface{}
		item       interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeAdmissionTimestampStorage_SetItem:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				lastChange: nil,
				item:       model.NodeAdmissionTimestamp{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeAdmissionTimestampStorage_SetItem:Fail-WrongTypeItem",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				lastChange: nil,
				item:       nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.SetItem(tt.args.lastChange, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAdmissionTimestampStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex                    sync.RWMutex
		nextNodeAdmissionTimestamp model.NodeAdmissionTimestamp
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
			name: "TestNodeAdmissionTimestampStorage_SetItems:Success",
			fields: fields{
				RWMutex: sync.RWMutex{},
				nextNodeAdmissionTimestamp: model.NodeAdmissionTimestamp{
					Timestamp:            0,
					BlockHeight:          0,
					Latest:               false,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
			},
			args: args{
				in0: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NodeAdmissionTimestampStorage{
				RWMutex:                    tt.fields.RWMutex,
				nextNodeAdmissionTimestamp: tt.fields.nextNodeAdmissionTimestamp,
			}
			if err := ns.SetItems(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
