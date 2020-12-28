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
package service

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewBlockPoolService(t *testing.T) {
	tests := []struct {
		name string
		want *BlockPoolService
	}{
		// TODO: Add test cases.
		{
			name: "NewBlockPoolService:success",
			want: &BlockPoolService{
				BlockQueue: make(map[int64]*model.Block),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockPoolService(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockPoolService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_GetBlocks(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int64]*model.Block
	}{
		// TODO: Add test cases.
		{
			name: "GetBlocks:success",
			fields: fields{
				// BlockQueueLock: sync.RWMutex,
				BlockQueue: map[int64]*model.Block{
					0: {},
				},
			},
			want: map[int64]*model.Block{
				0: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			if got := bps.GetBlocks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockPoolService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_GetBlock(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	type args struct {
		index int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		// TODO: Add test cases.
		{
			name: "GetBlock:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: {},
				},
			},
			want: &model.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			if got := bps.GetBlock(tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockPoolService.GetBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockPoolService_InsertBlock(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	type args struct {
		block *model.Block
		index int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "InsertBlock:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			bps.InsertBlock(tt.args.block, tt.args.index)
			if got := bps.GetBlock(tt.args.index); !reflect.DeepEqual(got, tt.fields.BlockQueue[tt.args.index]) {
				t.Errorf("BlockPoolService.InsertBlock() = %v, want %v", got, tt.fields.BlockQueue[tt.args.index])
			}
		})
	}
}

func TestBlockPoolService_ClearBlockPool(t *testing.T) {
	type fields struct {
		BlockQueueLock sync.RWMutex
		BlockQueue     map[int64]*model.Block
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "ClearBlockPool:success",
			fields: fields{
				BlockQueue: map[int64]*model.Block{
					0: {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bps := &BlockPoolService{
				BlockQueueLock: tt.fields.BlockQueueLock,
				BlockQueue:     tt.fields.BlockQueue,
			}
			bps.ClearBlockPool()
			if len(tt.fields.BlockQueue) > 0 {
				t.Errorf("BlockPoolService.ClearBlockPool() = %v, want %v", tt.fields.BlockQueue, make(map[int64]*model.Block))
			}
		})
	}
}
