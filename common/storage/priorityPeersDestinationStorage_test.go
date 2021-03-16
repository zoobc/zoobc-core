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
)

func TestPriorityPeersDestinationCacheHybridStorage_SetItem(t *testing.T) {
	type args struct {
		key  interface{}
		item interface{}
	}
	type fields struct {
		priorityPeersDestinations map[string][]uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    map[string][]uint32
	}{
		{
			name: "wantErr:wrongKeyType",
			args: args{
				key: 1,
			},
			wantErr: true,
		},
		{
			name: "wantErr:wrongItemType",
			args: args{
				key:  "a",
				item: "a",
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:newPublicKey",
			args: args{
				key:  "a",
				item: uint32(1),
			},
			fields: fields{
				priorityPeersDestinations: make(map[string][]uint32),
			},
			want: func() map[string][]uint32 {
				result := make(map[string][]uint32)
				result["a"] = []uint32{1}
				return result
			}(),
		},
		{
			name: "wantSuccess:removeIrrelevantHeightsButKeepAtLeast5HeightsInSafeRollback",
			args: args{
				key:  "a",
				item: uint32(2*constant.MinRollbackBlocks + 1),
			},
			fields: fields{
				priorityPeersDestinations: func() map[string][]uint32 {
					result := make(map[string][]uint32)
					result["a"] = []uint32{
						0,
						constant.MinRollbackBlocks - 100,
						constant.MinRollbackBlocks - 80,
						constant.MinRollbackBlocks - 60,
						constant.MinRollbackBlocks - 50,
						constant.MinRollbackBlocks - 40,
						constant.MinRollbackBlocks - 10,
						constant.MinRollbackBlocks,
						2*constant.MinRollbackBlocks - 120,
						2*constant.MinRollbackBlocks - 90,
						2*constant.MinRollbackBlocks - 70,
						2*constant.MinRollbackBlocks - 50,
						2*constant.MinRollbackBlocks - 10,
						2*constant.MinRollbackBlocks - 3,
						2 * constant.MinRollbackBlocks,
					}
					return result
				}(),
			},
			want: func() map[string][]uint32 {
				result := make(map[string][]uint32)
				result["a"] = []uint32{1380, 1390, 1400, 1430, 1440, 2760, 2790, 2810, 2830, 2870, 2877, 2880, 2881}
				return result
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  int(constant.PriorityStrategyMaxPriorityPeers),
				safeHeight:                2 * constant.MinRollbackBlocks,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if err := ppdc.SetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.SetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(ppdc.priorityPeersDestinations, tt.want) {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.SetItem() got = %v, want %v", ppdc.priorityPeersDestinations, tt.want)
			}
		})
	}
}

func TestPriorityPeersDestinationCacheHybridStorage_GetItem(t *testing.T) {
	var testResult []uint32
	mockFilledPriorityPeersDestinations := make(map[string][]uint32)
	mockFilledPriorityPeersDestinations["a"] = []uint32{1, 2, 3}
	mockFilledPriorityPeersDestinations["b"] = []uint32{4, 5, 6, 7, 8, 9, 10}

	type fields struct {
		minCount                  int
		safeHeight                uint32
		RWMutex                   sync.RWMutex
		priorityPeersDestinations map[string][]uint32
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
		want    []uint32
	}{
		{
			name: "wantError:wrongKeyType",
			args: args{
				key: 4,
			},
			wantErr: true,
		},
		{
			name: "wantSucces:notFound",
			args: args{
				key:  "c",
				item: &testResult,
			},
			fields: fields{
				priorityPeersDestinations: mockFilledPriorityPeersDestinations,
			},
			want: make([]uint32, 0),
		},
		{
			name: "wantSucces:found",
			args: args{
				key:  "a",
				item: &testResult,
			},
			fields: fields{
				priorityPeersDestinations: mockFilledPriorityPeersDestinations,
			},
			want: mockFilledPriorityPeersDestinations["a"],
		},
	}
	for _, tt := range tests {
		// resetting test result
		testResult = make([]uint32, 0)

		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  tt.fields.minCount,
				safeHeight:                tt.fields.safeHeight,
				RWMutex:                   tt.fields.RWMutex,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if err := ppdc.GetItem(tt.args.key, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.GetItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(testResult, tt.want) {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.GetItem() got = %v, want %v", testResult, tt.want)
			}
		})
	}
}

func TestPriorityPeersDestinationCacheHybridStorage_RemoveItems(t *testing.T) {
	mockFilledPriorityPeersDestinations := make(map[string][]uint32)
	mockFilledPriorityPeersDestinations["a"] = []uint32{1, 2, 3}
	mockFilledPriorityPeersDestinations["b"] = []uint32{4, 5, 6, 7, 8, 9, 10}

	mockFAfterRemovePriorityPeersDestinations := make(map[string][]uint32)
	mockFAfterRemovePriorityPeersDestinations["a"] = []uint32{1, 2, 3}

	type fields struct {
		minCount                  int
		safeHeight                uint32
		RWMutex                   sync.RWMutex
		priorityPeersDestinations map[string][]uint32
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    map[string][]uint32
	}{
		{
			name: "wantError:wrongKeyType",
			args: args{
				key: 4,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				key: "b",
			},
			fields: fields{
				priorityPeersDestinations: mockFilledPriorityPeersDestinations,
			},
			want: mockFAfterRemovePriorityPeersDestinations,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  tt.fields.minCount,
				safeHeight:                tt.fields.safeHeight,
				RWMutex:                   tt.fields.RWMutex,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if err := ppdc.RemoveItems(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.RemoveItems() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(ppdc.priorityPeersDestinations, tt.want) {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.RemoveItems() got = %v, want %v", ppdc.priorityPeersDestinations, tt.want)
			}
		})
	}
}

func TestPriorityPeersDestinationCacheHybridStorage_RemoveItem(t *testing.T) {
	type fields struct {
		minCount                  int
		safeHeight                uint32
		RWMutex                   sync.RWMutex
		priorityPeersDestinations map[string][]uint32
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    map[string][]uint32
	}{
		{
			name: "wantError:wrongKeyType",
			args: args{
				key: "adf",
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				key: uint32(10),
			},
			fields: fields{
				priorityPeersDestinations: func() map[string][]uint32 {
					result := make(map[string][]uint32)
					result["a"] = []uint32{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
					result["b"] = []uint32{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
					return result
				}(),
			},
			want: func() map[string][]uint32 {
				result := make(map[string][]uint32)
				result["a"] = []uint32{1, 3, 5, 7, 9}
				result["b"] = []uint32{2, 4, 6, 8, 10}
				return result
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  tt.fields.minCount,
				safeHeight:                tt.fields.safeHeight,
				RWMutex:                   tt.fields.RWMutex,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if err := ppdc.RemoveItem(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.RemoveItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(ppdc.priorityPeersDestinations, tt.want) {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.RemoveItem() got = %v, want %v", ppdc.priorityPeersDestinations, tt.want)
			}
		})
	}
}

func TestPriorityPeersDestinationCacheHybridStorage_GetSize(t *testing.T) {
	mockEmptyPriorityPeersDestinations := make(map[string][]uint32)
	mockFilledPriorityPeersDestinations := make(map[string][]uint32)
	mockFilledPriorityPeersDestinations["a"] = []uint32{1, 2, 3}
	mockFilledPriorityPeersDestinations["b"] = []uint32{4, 5, 6, 7, 8, 9, 10}

	type fields struct {
		minCount                  int
		safeHeight                uint32
		RWMutex                   sync.RWMutex
		priorityPeersDestinations map[string][]uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "wantSuccess:EmptyMap",
			fields: fields{
				priorityPeersDestinations: mockEmptyPriorityPeersDestinations,
			},
			want: 0,
		},
		{
			name: "wantSuccess:FilledMap",
			fields: fields{
				priorityPeersDestinations: mockFilledPriorityPeersDestinations,
			},
			want: 10 * 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  tt.fields.minCount,
				safeHeight:                tt.fields.safeHeight,
				RWMutex:                   tt.fields.RWMutex,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if got := ppdc.GetSize(); got != tt.want {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityPeersDestinationCacheHybridStorage_ClearCache(t *testing.T) {
	mockFilledPriorityPeersDestinations := make(map[string][]uint32)
	mockFilledPriorityPeersDestinations["a"] = []uint32{1, 2, 3}
	mockFilledPriorityPeersDestinations["b"] = []uint32{4, 5, 6, 7, 8, 9, 10}

	type fields struct {
		minCount                  int
		safeHeight                uint32
		RWMutex                   sync.RWMutex
		priorityPeersDestinations map[string][]uint32
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    map[string][]uint32
	}{
		{
			name: "wantSuccess",
			fields: fields{
				priorityPeersDestinations: mockFilledPriorityPeersDestinations,
			},
			want: make(map[string][]uint32),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppdc := &PriorityPeersDestinationCacheHybridStorage{
				minCount:                  tt.fields.minCount,
				safeHeight:                tt.fields.safeHeight,
				RWMutex:                   tt.fields.RWMutex,
				priorityPeersDestinations: tt.fields.priorityPeersDestinations,
			}
			if err := ppdc.ClearCache(); (err != nil) != tt.wantErr {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.ClearCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(ppdc.priorityPeersDestinations, tt.want) {
				t.Errorf("PriorityPeersDestinationCacheHybridStorage.ClearCache() got = %v, want %v", ppdc.priorityPeersDestinations, tt.want)
			}
		})

	}
}
