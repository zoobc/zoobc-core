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

func TestNewScrambleCacheStackStorage(t *testing.T) {
	tests := []struct {
		name string
		want *ScrambleCacheStackStorage
	}{
		{
			name: "TestNewScrambleCacheStackStorage:Success",
			want: &ScrambleCacheStackStorage{
				itemLimit:      36,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScrambleCacheStackStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScrambleCacheStackStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Clear(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_Clear:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetAll(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
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
			name: "TestScrambleCacheStackStorage_GetAll:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				items: &[]model.ScrambledNodes{
					{},
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAll:Fail-ItemIsNotScrambleNodes",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetAll(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetAtIndex(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		index uint32
		item  interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				index: 0,
				item:  &model.ScrambledNodes{},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Fail-IndexOutOfRange",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				index: 0,
				item:  &model.ScrambledNodes{},
			},
			wantErr: true,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Fail-ItemIsNotScrambleNodes",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{{}},
			},
			args: args{
				index: 0,
				item:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetAtIndex(tt.args.index, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAtIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetTop(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
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
			name: "TestScrambleCacheStackStorage_GetTop:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				item: &model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetTop:Fail-EmptyScramble",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestScrambleCacheStackStorage_GetTop:Fail-ItemIsNotScrambleNode",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{{}},
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetTop(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetTop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Pop(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_Pop:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{
						IndexNodes:           nil,
						NodePublicKeyToIDMap: nil,
						AddressNodes:         nil,
						BlockHeight:          0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Pop:Fail-StackEmpty",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Pop(); (err != nil) != tt.wantErr {
				t.Errorf("Pop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_PopTo(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		index uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_PopTo:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{
						IndexNodes:           nil,
						NodePublicKeyToIDMap: nil,
						AddressNodes:         nil,
						BlockHeight:          0,
					},
				},
			},
			args: args{
				index: 0,
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_PopTo:Fail-IndexOutOfRange",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				index: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.PopTo(tt.args.index); (err != nil) != tt.wantErr {
				t.Errorf("PopTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Push(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
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
			name: "TestScrambleCacheStackStorage_Push:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Push:Success-Len>0",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				item: model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Push:Fail-ItemIsNotScrambledNode",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Push(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_copy(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		src model.ScrambledNodes
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.ScrambledNodes
	}{
		{
			name: "TestScrambleCacheStackStorage_copy:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{},
			},
			args: args{
				src: model.ScrambledNodes{
					IndexNodes:           make(map[string]*int),
					NodePublicKeyToIDMap: make(map[string]int64),
					AddressNodes:         make([]*model.Peer, 0),
					BlockHeight:          0,
				},
			},
			want: model.ScrambledNodes{
				IndexNodes:           make(map[string]*int),
				NodePublicKeyToIDMap: make(map[string]int64),
				AddressNodes:         make([]*model.Peer, 0),
				BlockHeight:          0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if got := s.copy(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrambleCacheStackStorage_size(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestScrambleCacheStackStorage_size:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			want: 645,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if got := s.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}
