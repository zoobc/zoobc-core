package service

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
)

// type(
// 	MockBlockWithMetaData struct{

// 	}
// )

// func (*MockBlockWithMetaData) GetLastBlock() (*model.Block, error) {
// 	return &model.Block{
// 		Height: 0,
// 	}, nil
// }

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
