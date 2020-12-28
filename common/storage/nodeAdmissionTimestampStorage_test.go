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
