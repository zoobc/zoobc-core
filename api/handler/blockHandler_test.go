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

type (
	mockGetBlocksError struct {
		service.BlockServiceInterface
	}
	mockGetBlocksSucess struct {
		service.BlockServiceInterface
	}
)

func (*mockGetBlocksError) GetBlocks(chainType chaintype.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error) {
	return nil, errors.New("Error GetBlocks")
}

func (*mockGetBlocksSucess) GetBlocks(chainType chaintype.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error) {
	return &model.GetBlocksResponse{}, nil
}

func TestBlockHandler_GetBlocks(t *testing.T) {
	type fields struct {
		Service service.BlockServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetBlocksRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlocksResponse
		wantErr bool
	}{
		{
			name: "GetBlocks:LimitExceeded",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 1000,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocks:Error",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 500,
				},
			},
			fields: fields{
				Service: &mockGetBlocksError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocks:Success",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 500,
				},
			},
			fields: fields{
				Service: &mockGetBlocksSucess{},
			},
			want:    &model.GetBlocksResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockHandler{
				Service: tt.fields.Service,
			}
			got, err := bs.GetBlocks(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHandler.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockHandler.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
