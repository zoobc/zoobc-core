package storage

import (
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestBlockStateStorage_GetItem(t *testing.T) {
	type fields struct {
		RWMutex sync.RWMutex
		block   model.Block
	}
	type args struct {
		chaintypeInt interface{}
		block        interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				RWMutex: sync.RWMutex{},
				block:   model.Block{Height: 100, BlockHash: []byte{0, 0, 0}},
			},
			args: args{
				chaintypeInt: int32(0),
				block:        &model.Block{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockStateStorage{
				RWMutex:   tt.fields.RWMutex,
				lastBlock: tt.fields.block,
			}
			if err := bs.GetItem(tt.args.chaintypeInt, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.args.block == nil {
				t.Error("GetItem() got nil")
			}
		})
	}
}
