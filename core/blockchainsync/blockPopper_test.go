package blockchainsync

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestService_PopOffToBlock(t *testing.T) {
	type fields struct {
		ChainType     chaintype.ChainType
		BlockService  service.BlockServiceInterface
		QueryExecutor query.ExecutorInterface
	}
	type args struct {
		commonBlock *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "want:TestService_PopOffToBlock successfully return common block",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 3,
				},
			},
			want: func() []*model.Block {
				blocks := []*model.Block{}
				for i := 66; i >= 1; i-- {
					// pop off should allow node to pop off block to before genesis, since the first block may already be different
					blocks = append(blocks, &model.Block{ID: 58, Height: uint32(i),
						Transactions: []*model.Transaction{{Version: 1, ID: 789, BlockID: 40, Height: 69}}})
				}
				return blocks
			}(),
			wantErr: false,
		},
		{
			name: "want:TestService_PopOffToBlock error on getting LastBlock",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetLastBlock{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error on getting BlockByHeight",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetBlockByHeight{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error on getting BlockByID",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetBlockByID{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		// SERVICE QUERY SERVICES FAIL
		{
			name: "want:TestService_PopOffToBlock error on BeginTx function",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutorBeginTXFail{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error when committing transaction",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutorCommitTXFail{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error when executing Transactions",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutorExecuteTransFail{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockPopper{
				ChainType:     tt.fields.ChainType,
				BlockService:  tt.fields.BlockService,
				QueryExecutor: tt.fields.QueryExecutor,
			}
			got, err := bp.PopOffToBlock(tt.args.commonBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.PopOffToBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.PopOffToBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
