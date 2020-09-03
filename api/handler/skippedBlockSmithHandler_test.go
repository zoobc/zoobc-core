package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetSkippedBlockSmithsError struct {
		service.SkippedBlockSmithServiceInterface
	}
	mockGetSkippedBlockSmithsSuccess struct {
		service.SkippedBlockSmithServiceInterface
	}
)

func (*mockGetSkippedBlockSmithsError) GetSkippedBlockSmiths(*model.GetSkippedBlocksmithsRequest,
) (*model.GetSkippedBlocksmithsResponse, error) {
	return nil, errors.New("Error GetTransaction")
}

func (*mockGetSkippedBlockSmithsSuccess) GetSkippedBlockSmiths(*model.GetSkippedBlocksmithsRequest,
) (*model.GetSkippedBlocksmithsResponse, error) {
	return &model.GetSkippedBlocksmithsResponse{}, nil
}

func TestSkippedBlockSmithHandler_GetSkippedBlockSmiths(t *testing.T) {
	type fields struct {
		Service service.SkippedBlockSmithServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetSkippedBlocksmithsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetSkippedBlocksmithsResponse
		wantErr bool
	}{
		{
			name: "GetSkippedBlockSmiths:startBigger",
			args: args{
				request: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 10,
					BlockHeightEnd:   5,
				},
			},
			fields: fields{
				Service: &mockGetSkippedBlockSmithsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetSkippedBlockSmiths:maxLimit",
			args: args{
				request: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 1,
					BlockHeightEnd:   2 + constant.MaxAPILimitPerPage,
				},
			},
			fields: fields{
				Service: &mockGetSkippedBlockSmithsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetSkippedBlockSmiths:Failed",
			fields: fields{
				Service: &mockGetSkippedBlockSmithsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetSkippedBlockSmiths:Success",
			fields: fields{
				Service: &mockGetSkippedBlockSmithsSuccess{},
			},
			want:    &model.GetSkippedBlocksmithsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbh := &SkippedBlockSmithHandler{
				Service: tt.fields.Service,
			}
			got, err := sbh.GetSkippedBlockSmiths(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkippedBlockSmithHandler.GetSkippedBlockSmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SkippedBlockSmithHandler.GetSkippedBlockSmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}
