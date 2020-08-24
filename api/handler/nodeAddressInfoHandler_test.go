package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetNodeAddressesInfoFailed struct {
		service.NodeAddressInfoAPIServiceInterface
	}
	mockGetNodeAddressesInfoSuccess struct {
		service.NodeAddressInfoAPIServiceInterface
	}
)

func (*mockGetNodeAddressesInfoFailed) GetNodeAddressesInfo(request *model.GetNodeAddressesInfoRequest,
) (*model.GetNodeAddressesInfoResponse, error) {
	return nil, errors.New("Error GetNodeAddressesInfo")
}

func (*mockGetNodeAddressesInfoSuccess) GetNodeAddressesInfo(request *model.GetNodeAddressesInfoRequest,
) (*model.GetNodeAddressesInfoResponse, error) {
	return &model.GetNodeAddressesInfoResponse{}, nil
}

func TestNodeAddressInfoHandler_GetNodeAddressInfo(t *testing.T) {
	type fields struct {
		Service service.NodeAddressInfoAPIServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeAddressesInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeAddressesInfoResponse
		wantErr bool
	}{
		{
			name: "GetNodeAddressInfo:Error",
			args: args{
				req: &model.GetNodeAddressesInfoRequest{},
			},
			fields: fields{
				Service: &mockGetNodeAddressesInfoFailed{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeAddressInfo:NoNodeIDsFound",
			args: args{
				req: &model.GetNodeAddressesInfoRequest{},
			},
			fields: fields{
				Service: &mockGetNodeAddressesInfoSuccess{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeAddressInfo:Success",
			args: args{
				req: &model.GetNodeAddressesInfoRequest{
					NodeIDs: []int64{
						1, 2, 3,
					},
				},
			},
			fields: fields{
				Service: &mockGetNodeAddressesInfoSuccess{},
			},
			want:    &model.GetNodeAddressesInfoResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			naih := &NodeAddressInfoHandler{
				Service: tt.fields.Service,
			}
			got, err := naih.GetNodeAddressInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAddressInfoHandler.GetNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAddressInfoHandler.GetNodeAddressInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
