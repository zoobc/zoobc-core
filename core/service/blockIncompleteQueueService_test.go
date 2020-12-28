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

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
)

func TestBlockIncompleteQueueService_GetBlockQueue(t *testing.T) {
	mockBlockWithMetaData := make(map[int64]*BlockWithMetaData)
	mockBlockWithMetaData[int64(0)] = &BlockWithMetaData{Block: &model.Block{}}

	type fields struct {
		BlocksQueue                   map[int64]*BlockWithMetaData
		BlockRequiringTransactionsMap map[int64]TransactionIDsMap
		TransactionsRequiredMap       map[int64]BlockIDsMap
		Chaintype                     chaintype.ChainType
		BlockQueueLock                sync.Mutex
		Observer                      *observer.Observer
	}
	type args struct {
		blockID int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		// TODO: Add test cases.
		{
			name: "GetBlockQueue:error",
			fields: fields{
				BlocksQueue: make(map[int64]*BlockWithMetaData),
			},
			args: args{},
			want: nil,
		},
		{
			name: "GetBlockQueue:success",
			fields: fields{
				BlocksQueue: mockBlockWithMetaData,
			},
			args: args{
				blockID: int64(0),
			},
			want: &model.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buqs := &BlockIncompleteQueueService{
				BlocksQueue:                   tt.fields.BlocksQueue,
				BlockRequiringTransactionsMap: tt.fields.BlockRequiringTransactionsMap,
				TransactionsRequiredMap:       tt.fields.TransactionsRequiredMap,
				Chaintype:                     tt.fields.Chaintype,
				BlockQueueLock:                tt.fields.BlockQueueLock,
				Observer:                      tt.fields.Observer,
			}
			if got := buqs.GetBlockQueue(tt.args.blockID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockIncompleteQueueService.GetBlockQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockIncompleteQueueService_AddTransaction(t *testing.T) {
	type fields struct {
		BlocksQueue                   map[int64]*BlockWithMetaData
		BlockRequiringTransactionsMap map[int64]TransactionIDsMap
		TransactionsRequiredMap       map[int64]BlockIDsMap
		Chaintype                     chaintype.ChainType
		BlockQueueLock                sync.Mutex
		Observer                      *observer.Observer
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.Block
	}{
		// TODO: Add test cases.
		{
			name: "AddTransaction:errorBlocksQueueIsNil",
			fields: fields{
				TransactionsRequiredMap: make(map[int64]BlockIDsMap),
			},
			args: args{
				transaction: &model.Transaction{},
			},
			want: nil,
		},
		{
			name: "AddTransaction:errorLenTxsIsEmpty",
			fields: fields{
				BlocksQueue: map[int64]*BlockWithMetaData{
					0: {Block: &model.Block{}},
				},
				TransactionsRequiredMap: map[int64]BlockIDsMap{
					0: {0: true},
				},
				BlockRequiringTransactionsMap: map[int64]TransactionIDsMap{
					0: {0: 1},
				},
			},
			args: args{
				transaction: &model.Transaction{
					BlockID: int64(0),
				},
			},
			want: nil,
		},
		{
			name: "AddTransaction:success",
			fields: fields{
				BlocksQueue: map[int64]*BlockWithMetaData{
					15: {Block: &model.Block{
						Transactions: []*model.Transaction{
							{},
							{},
						},
					}},
				},
				TransactionsRequiredMap: map[int64]BlockIDsMap{
					21: {15: true},
					22: {15: true},
				},
				BlockRequiringTransactionsMap: map[int64]TransactionIDsMap{
					15: {21: 0, 22: 1},
				},
			},
			args: args{
				transaction: &model.Transaction{
					ID: 21,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buqs := &BlockIncompleteQueueService{
				BlocksQueue:                   tt.fields.BlocksQueue,
				BlockRequiringTransactionsMap: tt.fields.BlockRequiringTransactionsMap,
				TransactionsRequiredMap:       tt.fields.TransactionsRequiredMap,
				Chaintype:                     tt.fields.Chaintype,
				BlockQueueLock:                tt.fields.BlockQueueLock,
				Observer:                      tt.fields.Observer,
			}
			if got := buqs.AddTransaction(tt.args.transaction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockIncompleteQueueService.AddTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
