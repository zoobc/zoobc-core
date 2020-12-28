package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewNodeAddressInfoStorage(t *testing.T) {
	tests := []struct {
		name string
		want *NodeAddressInfoStorage
	}{
		{
			name: "NewNodeAddressInfoStorage:Success",
			want: &NodeAddressInfoStorage{
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeAddressInfoStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeAddressInfoStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoStorage_Begin(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "NodeAddressInfoStorage_Begin:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.Begin(); (err != nil) != tt.wantErr {
				t.Errorf("Begin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "NodeAddressInfoStorage_ClearCache:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_Commit(t *testing.T) {
	mock := sync.RWMutex{}
	mock.Lock()
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "NodeAddressInfoStorage_Commit:Success",
			fields: fields{
				RWMutex:                                    mock,
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            true,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_Commit:FailNotInTransaction",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.Commit(); (err != nil) != tt.wantErr {
				t.Errorf("Commit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_GetAllItems(t *testing.T) {
	mockNaiMapByID := make(map[int64]map[string]model.NodeAddressInfo)
	mockNaiMapModel := map[string]model.NodeAddressInfo{
		"127.0.0.1": {
			NodeID:      1,
			Address:     "127.0.0.1",
			Port:        3001,
			BlockHeight: 10,
			BlockHash:   make([]byte, 32),
			Status:      0,
			Signature:   make([]byte, 64),
		},
	}
	mockNaiMapByID[1] = mockNaiMapModel
	mockItem := []*model.NodeAddressInfo{
		{
			NodeID:      111,
			Address:     "127.0.0.1",
			Port:        3001,
			BlockHeight: 10,
			BlockHash:   make([]byte, 32),
			Status:      2,
			Signature:   make([]byte, 64),
		},
	}

	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_GetAllItems:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            true,
				nodeAddressInfoMapByID:                     mockNaiMapByID,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				item: &mockItem,
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_GetAllItems:Fail-ItemError",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.GetAllItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_GetItem(t *testing.T) {
	mockNaiMapByID := make(map[int64]map[string]model.NodeAddressInfo)
	mockNaiMapModel := map[string]model.NodeAddressInfo{
		"127.0.0.1": {
			NodeID:      1,
			Address:     "127.0.0.1",
			Port:        3001,
			BlockHeight: 10,
			BlockHash:   make([]byte, 32),
			Status:      0,
			Signature:   make([]byte, 64),
		},
	}
	mockNaiMapByID[1] = mockNaiMapModel
	mockNaiMapByStatus := make(map[model.NodeAddressStatus]map[int64]map[string]bool)
	mockNaiMapStatusBool := map[int64]map[string]bool{1: {"127.0.0.1": true}}
	mockNaiMapByStatus[model.NodeAddressStatus_Unset] = mockNaiMapStatusBool
	var mockItem []*model.NodeAddressInfo
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_GetItem:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            true,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 mockNaiMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      111,
					AddressPort: "127.0.0.1",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
				item: &mockItem,
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_GetItem:Fail-NilKey",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key:  nil,
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_GetItem:Fail-NilItem",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key:  NodeAddressInfoStorageKey{},
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_GetItem:Fail-StatusNil",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      0,
					AddressPort: "",
					Statuses:    nil,
				},
				item: mockItem,
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_GetItem:Success-StorageKeyStatus\"\"",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     mockNaiMapByID,
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 mockNaiMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      0,
					AddressPort: "",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_Unset,
					},
				},
				item: &mockItem,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.GetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_GetSize(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "NodeAddressInfoStorage_GetSize:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			want: 318,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if got := nas.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "NodeAddressInfoStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if got := nas.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_RemoveItem:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      111,
					AddressPort: "127.0.0.1",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_RemoveItem:KeyError",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				key: 1,
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_RemoveItem:StatusNil",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      111,
					AddressPort: "127.0.0.1",
					Statuses:    []model.NodeAddressStatus{},
				},
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_RemoveItem:NodeId:0",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      0,
					AddressPort: "127.0.0.1",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_Rollback(t *testing.T) {
	mock := sync.RWMutex{}
	mock.Lock()
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "NodeAddressInfoStorage_Rollback:Success",
			fields: fields{
				RWMutex:                                    mock,
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            true,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_Rollback:Fail",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.Rollback(); (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_SetItem:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: nil,
				item: model.NodeAddressInfo{
					NodeID:      111,
					Address:     "127.0.0.1",
					Port:        3001,
					BlockHeight: 10,
					BlockHash:   make([]byte, 32),
					Status:      0,
					Signature:   make([]byte, 64),
				},
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_SetItem:Fail",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				key:  nil,
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.SetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_SetItems:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				item: []*model.NodeAddressInfo{
					{
						NodeID:      111,
						Address:     "127.0.0.1",
						Port:        3000,
						BlockHeight: 10,
						BlockHash:   make([]byte, 32),
						Status:      0,
						Signature:   make([]byte, 64),
					},
					{
						NodeID:      222,
						Address:     "127.0.0.2",
						Port:        3002,
						BlockHeight: 20,
						BlockHash:   make([]byte, 32),
						Status:      0,
						Signature:   make([]byte, 64),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.SetItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_TxRemoveItem(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_TxRemoveItem:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      111,
					AddressPort: "127.0.0.1",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "NodeAddressInfoStorage_TxRemoveItem:Error-Key",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: nil,
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_TxRemoveItem:Error-StatusNil",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      111,
					AddressPort: "127.0.0.1",
					Statuses:    []model.NodeAddressStatus{},
				},
			},
			wantErr: true,
		},
		{
			name: "NodeAddressInfoStorage_TxRemoveItem:Error-NodeId:0",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     make(map[int64]map[string]model.NodeAddressInfo),
				nodeAddressInfoMapByAddressPort:            make(map[string]map[int64]bool),
				nodeAddressInfoMapByStatus:                 make(map[model.NodeAddressStatus]map[int64]map[string]bool),
				transactionalRemovedNodeAddressInfoMapByID: make(map[int64]map[string]bool),
			},
			args: args{
				key: NodeAddressInfoStorageKey{
					NodeID:      0,
					AddressPort: "127.0.0.1",
					Statuses: []model.NodeAddressStatus{
						model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.TxRemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("TxRemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_TxSetItem(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	type args struct {
		id   interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				id:   nil,
				item: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.TxSetItem(tt.args.id, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("TxSetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_TxSetItems(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
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
			name: "NodeAddressInfoStorage_TxSetItems:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if err := nas.TxSetItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("TxSetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoStorage_append(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	type args struct {
		nodeAddresses []*model.NodeAddressInfo
		nodeAddress   model.NodeAddressInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.NodeAddressInfo
	}{
		{
			name: "NodeAddressInfoStorage_append:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			args: args{
				nodeAddresses: []*model.NodeAddressInfo{},
				nodeAddress: model.NodeAddressInfo{
					NodeID:      111,
					Address:     "127.0.0.1",
					Port:        3001,
					BlockHeight: 10,
					BlockHash:   make([]byte, 32),
					Status:      2,
					Signature:   make([]byte, 64),
				},
			},
			want: []*model.NodeAddressInfo{
				{
					NodeID:      111,
					Address:     "127.0.0.1",
					Port:        3001,
					BlockHeight: 10,
					BlockHash:   make([]byte, 32),
					Status:      2,
					Signature:   make([]byte, 64),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if got := nas.append(tt.args.nodeAddresses, tt.args.nodeAddress); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoStorage_size(t *testing.T) {
	type fields struct {
		RWMutex                                    sync.RWMutex
		transactionalLock                          sync.RWMutex
		isInTransaction                            bool
		nodeAddressInfoMapByID                     map[int64]map[string]model.NodeAddressInfo
		nodeAddressInfoMapByAddressPort            map[string]map[int64]bool
		nodeAddressInfoMapByStatus                 map[model.NodeAddressStatus]map[int64]map[string]bool
		transactionalRemovedNodeAddressInfoMapByID map[int64]map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "NodeAddressInfoStorage_size:Success",
			fields: fields{
				RWMutex:                                    sync.RWMutex{},
				transactionalLock:                          sync.RWMutex{},
				isInTransaction:                            false,
				nodeAddressInfoMapByID:                     nil,
				nodeAddressInfoMapByAddressPort:            nil,
				nodeAddressInfoMapByStatus:                 nil,
				transactionalRemovedNodeAddressInfoMapByID: nil,
			},
			want: 318,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAddressInfoStorage{
				RWMutex:                                    tt.fields.RWMutex,
				transactionalLock:                          tt.fields.transactionalLock,
				isInTransaction:                            tt.fields.isInTransaction,
				nodeAddressInfoMapByID:                     tt.fields.nodeAddressInfoMapByID,
				nodeAddressInfoMapByAddressPort:            tt.fields.nodeAddressInfoMapByAddressPort,
				nodeAddressInfoMapByStatus:                 tt.fields.nodeAddressInfoMapByStatus,
				transactionalRemovedNodeAddressInfoMapByID: tt.fields.transactionalRemovedNodeAddressInfoMapByID,
			}
			if got := nas.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}
