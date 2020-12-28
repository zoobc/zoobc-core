// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/monitoring"
)

func TestNewNodeRegistryCacheStorage(t *testing.T) {
	type args struct {
		metricLabel monitoring.CacheStorageType
		sortFunc    func([]NodeRegistry)
	}
	tests := []struct {
		name string
		args args
		want *NodeRegistryCacheStorage
	}{
		{
			name: "TestNewNodeRegistryCacheStorage:Success",
			args: args{
				metricLabel: "testing",
				sortFunc:    nil,
			},
			want: &NodeRegistryCacheStorage{
				isInTransaction: false,
				nodeRegistries:  []NodeRegistry{},
				nodeIDIndexes:   map[int64]int{},
				metricLabel:     "testing",
				sortItems:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeRegistryCacheStorage(tt.args.metricLabel, tt.args.sortFunc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeRegistryCacheStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_Begin(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_Begin:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.Begin(); (err != nil) != tt.wantErr {
				t.Errorf("Begin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_ClearCache(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_ClearCache:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_Commit(t *testing.T) {
	mock := sync.RWMutex{}
	mock.Lock()
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_Commit:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     mock,
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.Commit(); (err != nil) != tt.wantErr {
				t.Errorf("Commit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_GetAllItems(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
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
			name: "TestNodeRegistryCacheStorage_GetAllItems:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				item: &[]NodeRegistry{{}},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetAllItems:Success",
			fields: fields{
				isInTransaction:             true,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				item: &[]NodeRegistry{{}},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetAllItems:Fail-WrongTypeItem",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.GetAllItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_GetItem(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		idx  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  int64(1),
				item: &NodeRegistry{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Fail-WrongTypeItem",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  nil,
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Fail-KeyNil",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  nil,
				item: &NodeRegistry{},
			},
			wantErr: true,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Success-ID",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  1,
				item: &NodeRegistry{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Fail-NodeRegistryNotFound",
			fields: fields{
				isInTransaction:             true,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  int64(11),
				item: &NodeRegistry{},
			},
			wantErr: true,
		},
		{
			name: "TestNodeRegistryCacheStorage_GetItem:Fail-UnknownType",
			fields: fields{
				isInTransaction:             true,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}, {}},
				nodeIDIndexes:               map[int64]int{1: 1},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  "",
				item: &NodeRegistry{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.GetItem(tt.args.idx, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_GetSize(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestNodeRegistryCacheStorage_GetSize:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			want: 539,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if got := n.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestNodeRegistryCacheStorage_GetTotalItems:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if got := n.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_RemoveItem(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		idx interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_RemoveItem:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx: 0,
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_RemoveItem:Success-ID-INT64",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  map[int64]int{0: 0},
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx: int64(0),
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_RemoveItem:Fail-IdCannotBeNil",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx: nil,
			},
			wantErr: true,
		},
		{
			name: "TestNodeRegistryCacheStorage_RemoveItem:Fail-UnknownType",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx: "",
			},
			wantErr: true,
		},
		{
			name: "TestNodeRegistryCacheStorage_RemoveItem:Fail-IndexOutOfRange",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{},
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx: int64(11),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.RemoveItem(tt.args.idx); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_Rollback(t *testing.T) {
	mock := sync.RWMutex{}
	mock.Lock()
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_Rollback:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     mock,
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.Rollback(); (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_SetItem(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		idx  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_SetItem:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  0,
				item: NodeRegistry{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_SetItem:Fail-WrongTypeItem",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				idx:  nil,
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.SetItem(tt.args.idx, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_SetItems(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
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
			name: "TestNodeRegistryCacheStorage_SetItems:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   func(slice []NodeRegistry) {},
			},
			args: args{
				items: []NodeRegistry{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_SetItems:Fail-WrongTypeItem",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   func(slice []NodeRegistry) {},
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.SetItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_TxRemoveItem(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		idx interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_TxRemoveItem:Success",
			fields: fields{
				isInTransaction:   false,
				transactionalLock: sync.RWMutex{},
				RWMutex:           sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{
					{
						Node: model.NodeRegistration{
							NodeID: 1,
							NodeAddressInfo: &model.NodeAddressInfo{
								NodeID: 1,
							},
						},
						ParticipationScore: 0,
					},
					{},
				},
				transactionalNodeIDIndexes: map[int64]int{1: 1},
				nodeIDIndexes:              map[int64]int{1: 1},
				metricLabel:                "",
				sortItems:                  func(slice []NodeRegistry) {},
			},
			args: args{
				idx: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.TxRemoveItem(tt.args.idx); (err != nil) != tt.wantErr {
				t.Errorf("TxRemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_TxSetItem(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		idx  interface{}
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNodeRegistryCacheStorage_TxSetItem:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  map[int64]int{},
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   func(slice []NodeRegistry) {},
			},
			args: args{
				idx:  nil,
				item: NodeRegistry{},
			},
			wantErr: false,
		},
		{
			name: "TestNodeRegistryCacheStorage_TxSetItem:Fail-WrongTypeItem",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  map[int64]int{},
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   func(slice []NodeRegistry) {},
			},
			args: args{
				idx:  nil,
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.TxSetItem(tt.args.idx, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("TxSetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_TxSetItems(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
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
			name: "TestNodeRegistryCacheStorage_TxSetItems:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   func(slice []NodeRegistry) {},
			},
			args: args{
				items: []NodeRegistry{{}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if err := n.TxSetItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("TxSetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_copy(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	type args struct {
		src NodeRegistry
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   NodeRegistry
	}{
		{
			name: "TestNodeRegistryCacheStorage_copy:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: []NodeRegistry{{}},
				transactionalNodeIDIndexes:  map[int64]int{},
				nodeRegistries:              []NodeRegistry{{}},
				nodeIDIndexes:               map[int64]int{},
				metricLabel:                 "",
				sortItems:                   nil,
			},
			args: args{
				src: NodeRegistry{
					Node: model.NodeRegistration{
						NodeID:             0,
						NodePublicKey:      make([]byte, 1),
						AccountAddress:     nil,
						RegistrationHeight: 0,
						LockedBalance:      0,
						RegistrationStatus: 0,
						Latest:             false,
						Height:             0,
					},
					ParticipationScore: 0,
				},
			},
			want: NodeRegistry{
				Node: model.NodeRegistration{
					NodeID:             0,
					NodePublicKey:      make([]byte, 1),
					AccountAddress:     nil,
					RegistrationHeight: 0,
					LockedBalance:      0,
					RegistrationStatus: 0,
					Latest:             false,
					Height:             0,
				},
				ParticipationScore: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if got := n.copy(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistryCacheStorage_size(t *testing.T) {
	type fields struct {
		isInTransaction             bool
		transactionalLock           sync.RWMutex
		RWMutex                     sync.RWMutex
		transactionalNodeRegistries []NodeRegistry
		transactionalNodeIDIndexes  map[int64]int
		nodeRegistries              []NodeRegistry
		nodeIDIndexes               map[int64]int
		metricLabel                 monitoring.CacheStorageType
		sortItems                   func(slice []NodeRegistry)
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestNodeRegistryCacheStorage_size:Success",
			fields: fields{
				isInTransaction:             false,
				transactionalLock:           sync.RWMutex{},
				RWMutex:                     sync.RWMutex{},
				transactionalNodeRegistries: nil,
				transactionalNodeIDIndexes:  nil,
				nodeRegistries:              nil,
				nodeIDIndexes:               nil,
				metricLabel:                 "",
				sortItems:                   nil,
			},
			want: 539,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistryCacheStorage{
				isInTransaction:             tt.fields.isInTransaction,
				transactionalLock:           tt.fields.transactionalLock,
				RWMutex:                     tt.fields.RWMutex,
				transactionalNodeRegistries: tt.fields.transactionalNodeRegistries,
				transactionalNodeIDIndexes:  tt.fields.transactionalNodeIDIndexes,
				nodeRegistries:              tt.fields.nodeRegistries,
				nodeIDIndexes:               tt.fields.nodeIDIndexes,
				metricLabel:                 tt.fields.metricLabel,
				sortItems:                   tt.fields.sortItems,
			}
			if got := n.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}
