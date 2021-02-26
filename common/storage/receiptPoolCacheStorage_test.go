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

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestReceiptPoolCacheStorage_SetItem(t *testing.T) {
	type fields struct {
		RWMutex  sync.RWMutex
		receipts map[string][]model.Receipt
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
			name: "wantError:InvalidKey",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				key: 123,
			},
			wantErr: true,
		},
		{
			name: "wantError:InvalidReceipt",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				key:  "123",
				item: model.Block{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				key:  "123",
				item: model.Receipt{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.SetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("ReceiptPoolCacheStorage.SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_GetItems(t *testing.T) {
	mockReceiptGroups := make(map[string][]model.Receipt)
	mockReceiptGroups["010203"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 3},
		},
	}
	mockReceiptGroups["010204"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 4},
		},
	}
	mockReceiptGroups["010205"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 5},
		},
	}

	successGetItemsResults := make(map[string][]model.Receipt)
	mockReceiptGroups["010203"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 3},
		},
	}
	mockReceiptGroups["010204"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 4},
		},
	}

	type fields struct {
		RWMutex  sync.RWMutex
		receipts map[string][]model.Receipt
	}
	type args struct {
		keys  interface{}
		items interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    map[string][]model.Receipt
	}{
		{
			name: "wantError:InvalidKey",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				keys: 123,
			},
			wantErr: true,
		},
		{
			name: "wantError:InvalidReceipt",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				keys:  []string{"010203"},
				items: model.Block{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess-FileNotFound",
			fields: fields{
				receipts: make(map[string][]model.Receipt),
			},
			args: args{
				keys:  []string{"010203", "010204"},
				items: make(map[string][]model.Receipt),
			},
			want: successGetItemsResults,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.GetItems(tt.args.keys, tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("ReceiptPoolCacheStorage.GetItems() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.args.items, tt.want) {
				t.Errorf("ReceiptPoolCacheStorage.GetItems() got = %v, want %v", tt.args.items, tt.want)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_ClearCache(t *testing.T) {
	mockReceiptGroups := make(map[string][]model.Receipt)
	mockReceiptGroups["010203"] = []model.Receipt{
		{
			DatumHash: []byte{1, 2, 3},
		},
	}

	type fields struct {
		RWMutex  sync.RWMutex
		receipts map[string][]model.Receipt
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    map[string][]model.Receipt
	}{
		{
			name: "wantSuccess",
			fields: fields{
				receipts: mockReceiptGroups,
			},
			want: make(map[string][]model.Receipt),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := &ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			if err := brs.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("ReceiptPoolCacheStorage.ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(brs.receipts, tt.want) {
				t.Errorf("ReceiptPoolCacheStorage.ClearCache() got = %v, want %v", brs.receipts, tt.want)
			}
		})
	}
}

func TestReceiptPoolCacheStorage_CleanExpiredReceipts(t *testing.T) {
	mockReceiptGroups := make(map[string][]model.Receipt)
	mockReceiptGroups["010203"] = []model.Receipt{
		{
			ReferenceBlockHeight: 1,
		},
		{
			ReferenceBlockHeight: 2,
		},
	}
	mockReceiptGroups["010204"] = []model.Receipt{
		{
			ReferenceBlockHeight: 1,
		},
		{
			ReferenceBlockHeight: 4,
		},
	}

	expectedResult := make(map[string][]model.Receipt)
	expectedResult["010203"] = []model.Receipt{
		{
			ReferenceBlockHeight: 2,
		},
	}
	expectedResult["010204"] = []model.Receipt{
		{
			ReferenceBlockHeight: 4,
		},
	}

	type fields struct {
		RWMutex  sync.RWMutex
		receipts map[string][]model.Receipt
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]model.Receipt
	}{
		{
			name: "wantSuccess:no_receipts_are_cleaned_on_block_height_less_than_receipt_life_cut_off",
			fields: fields{
				receipts: mockReceiptGroups,
			},
			args: args{
				blockHeight: constant.ReceiptLifeCutOff - 1,
			},
			want: mockReceiptGroups,
		},
		{
			name: "wantSuccess:expired_receipts_are_deleted",
			fields: fields{
				receipts: mockReceiptGroups,
			},
			args: args{
				blockHeight: constant.ReceiptLifeCutOff + 2,
			},
			want: expectedResult,
		},
		{
			name: "wantSuccess:all_expired_receipts_are_deleted",
			fields: fields{
				receipts: mockReceiptGroups,
			},
			args: args{
				blockHeight: constant.ReceiptLifeCutOff + 5,
			},
			want: make(map[string][]model.Receipt),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brs := ReceiptPoolCacheStorage{
				RWMutex:  tt.fields.RWMutex,
				receipts: tt.fields.receipts,
			}
			brs.CleanExpiredReceipts(tt.args.blockHeight)
			if !reflect.DeepEqual(brs.receipts, tt.want) {
				t.Errorf("ReceiptPoolCacheStorage.ClearCache() got = %v, want %v", brs.receipts, tt.want)
			}
		})
	}
}
