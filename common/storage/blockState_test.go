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
				lastBlockBytes: []byte(`{"byte":["a"]}`),
			},
			args: args{
				lastUpdate: nil,
				block:      &model.Block{},
			},
			wantErr: false,
		},
		{
			name: "TestBlockStateStorage_GetItem:Fail-EmptyCache",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: nil,
			},
			args: args{
				lastUpdate: nil,
				block:      &model.Block{},
			},
			wantErr: true,
		},
		{
			name: "TestBlockStateStorage_GetItem:Fail-WrongTypeItem",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: []byte(`{"byte":["a"]}`),
			},
			args: args{
				lastUpdate: nil,
				block:      nil,
			},
			wantErr: true,
		},
		{
			name: "TestBlockStateStorage_GetItem:Fail-Unmarshal",
			fields: fields{
				RWMutex:        sync.RWMutex{},
				lastBlockBytes: make([]byte, 32),
			},
			args: args{
				lastUpdate: nil,
				block:      &model.Block{},
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
