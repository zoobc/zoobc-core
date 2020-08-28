package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	mockGetNodeAddressesInfoFromDbError struct {
		coreService.NodeRegistrationServiceInterface
	}
	mockGetNodeAddressesInfoFromDbSuccess struct {
		coreService.NodeRegistrationServiceInterface
	}
)

func (*mockGetNodeAddressesInfoFromDbError) GetNodeAddressesInfoFromDb(nodeIDs []int64,
	addressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	return nil, errors.New("Error GetNodeAddressesInfoFromDb")
}
func (*mockGetNodeAddressesInfoFromDbSuccess) GetNodeAddressesInfoFromDb(nodeIDs []int64,
	addressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	return make([]*model.NodeAddressInfo, 0), nil
}

func TestNodeAddressInfoAPIService_GetNodeAddressesInfo(t *testing.T) {
	type fields struct {
		NodeRegistrationCoreService coreService.NodeRegistrationServiceInterface
	}
	type args struct {
		request *model.GetNodeAddressesInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeAddressesInfoResponse
		wantErr bool
	}{
		{
			name: "GetNodeAddressesInfo:InternalError",
			fields: fields{
				NodeRegistrationCoreService: &mockGetNodeAddressesInfoFromDbError{},
			},
			args: args{
				request: &model.GetNodeAddressesInfoRequest{
					NodeIDs: []int64{1, 2, 3},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeAddressesInfo:Success",
			fields: fields{
				NodeRegistrationCoreService: &mockGetNodeAddressesInfoFromDbSuccess{},
			},
			args: args{
				request: &model.GetNodeAddressesInfoRequest{
					NodeIDs: []int64{1, 2, 3},
				},
			},
			want:    &model.GetNodeAddressesInfoResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nhs := &NodeAddressInfoAPIService{
				NodeRegistrationCoreService: tt.fields.NodeRegistrationCoreService,
			}
			got, err := nhs.GetNodeAddressesInfo(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAddressInfoAPIService.GetNodeAddressesInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAddressInfoAPIService.GetNodeAddressesInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
