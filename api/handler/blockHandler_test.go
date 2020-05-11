package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetBlockError struct {
		service.BlockServiceInterface
	}
	mockGetBlockSucess struct {
		service.BlockServiceInterface
	}
)

func (*mockGetBlockError) GetBlockByID(chainType chaintype.ChainType, ID int64) (*model.BlockExtendedInfo, error) {
	return nil, errors.New("Error GetBlockByID")
}

func (*mockGetBlockError) GetBlockByHeight(chainType chaintype.ChainType, Height uint32) (*model.BlockExtendedInfo, error) {
	return nil, errors.New("Error GetBlockByHeight")
}

func (*mockGetBlockSucess) GetBlockByID(chainType chaintype.ChainType, ID int64) (*model.BlockExtendedInfo, error) {
	return &model.BlockExtendedInfo{}, nil
}

func (*mockGetBlockSucess) GetBlockByHeight(chainType chaintype.ChainType, Height uint32) (*model.BlockExtendedInfo, error) {
	return &model.BlockExtendedInfo{}, nil
}

func TestBlockHandler_GetBlock(t *testing.T) {
	type fields struct {
		Service service.BlockServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetBlockRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlockExtendedInfo
		wantErr bool
	}{
		{
			name: "GetBlock:Error",
			fields: fields{
				Service: &mockGetBlockError{},
			},
			args: args{
				req: &model.GetBlockRequest{
					ID:     1,
					Height: 1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlock:Sucess",
			fields: fields{
				Service: &mockGetBlockSucess{},
			},
			args: args{
				req: &model.GetBlockRequest{
					ID:     1,
					Height: 1,
				},
			},
			want:    &model.BlockExtendedInfo{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockHandler{
				Service: tt.fields.Service,
			}
			got, err := bs.GetBlock(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHandler.GetBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockHandler.GetBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
