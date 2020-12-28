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
