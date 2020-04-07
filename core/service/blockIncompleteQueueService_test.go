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
					0: &BlockWithMetaData{Block: &model.Block{}},
				},
				TransactionsRequiredMap: map[int64]BlockIDsMap{
					0: BlockIDsMap{0: true},
				},
				BlockRequiringTransactionsMap: map[int64]TransactionIDsMap{
					0: TransactionIDsMap{0: 1},
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
					15: &BlockWithMetaData{Block: &model.Block{
						Transactions: []*model.Transaction{
							{},
							{},
						},
					}},
				},
				TransactionsRequiredMap: map[int64]BlockIDsMap{
					21: BlockIDsMap{15: true},
					22: BlockIDsMap{15: true},
				},
				BlockRequiringTransactionsMap: map[int64]TransactionIDsMap{
					15: TransactionIDsMap{21: 0, 22: 1},
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
