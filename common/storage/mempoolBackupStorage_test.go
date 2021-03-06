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
)

func TestMempoolBackupStorage_ClearCache(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestMempoolBackupStorage_ClearCache:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_GetAllItems(t *testing.T) {
	mockItem := make(map[int64][]byte)
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
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
			name: "TestMempoolBackupStorage_GetAllItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				item: &mockItem,
			},
			wantErr: false,
		},
		{
			name: "TestMempoolBackupStorage_GetAllItems:Fail-WrongTypeItem",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.GetAllItems(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAllItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_GetItem(t *testing.T) {
	mockItem := make([]byte, 1)
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
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
			name: "TestMempoolBackupStorage_GetItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				key:  int64(1),
				item: &mockItem,
			},
			wantErr: false,
		},
		{
			name: "TestMempoolBackupStorage_GetItem:Fail-WrongKey",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				key:  nil,
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestMempoolBackupStorage_GetItem:Fail-WrongItem",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				key:  int64(1),
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.GetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_GetSize(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestMempoolBackupStorage_GetSize:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if got := m.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMempoolBackupStorage_GetTotalItems(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestMempoolBackupStorage_GetTotalItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if got := m.GetTotalItems(); got != tt.want {
				t.Errorf("GetTotalItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMempoolBackupStorage_RemoveItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
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
			name: "TestMempoolBackupStorage_RemoveItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				key: int64(1),
			},
			wantErr: false,
		},
		{
			name: "TestMempoolBackupStorage_RemoveItem:Fail-WrongKey",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				key: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
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
			name: "TestMempoolBackupStorage_SetItem:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: map[int64][]byte{},
			},
			args: args{
				key:  int64(1),
				item: make([]byte, 1),
			},
			wantErr: false,
		},
		{
			name: "TestMempoolBackupStorage_SetItem:Fail-WrongKey",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: map[int64][]byte{},
			},
			args: args{
				key:  nil,
				item: nil,
			},
			wantErr: true,
		}, {
			name: "TestMempoolBackupStorage_SetItem:Fail-WrongItem",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: map[int64][]byte{},
			},
			args: args{
				key:  int64(1),
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.SetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_SetItems(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
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
			name: "TestMempoolBackupStorage_SetItems:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				items: make(map[int64][]byte),
			},
			wantErr: false,
		},
		{
			name: "TestMempoolBackupStorage_SetItems:Fail-WrongItems",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if err := m.SetItems(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("SetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolBackupStorage_size(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		mempools map[int64][]byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "TestMempoolBackupStorage_size:Success",
			fields: fields{
				RWMutex:  sync.RWMutex{},
				mempools: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MempoolBackupStorage{
				RWMutex:  tt.fields.RWMutex,
				mempools: tt.fields.mempools,
			}
			if got := m.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMempoolBackupStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MempoolBackupStorage
	}{
		{
			name: "TestNewMempoolBackupStorage:Success",
			want: &MempoolBackupStorage{
				RWMutex:  sync.RWMutex{},
				mempools: map[int64][]byte{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMempoolBackupStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMempoolBackupStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
