package handler

import (
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
)

type (
	mockGetNodeHardwareError struct {
		service.NodeHardwareServiceInterface
	}
	mockGetNodeHardwareSuccess struct {
		service.NodeHardwareServiceInterface
	}
	mockGetNodeHardwareServer struct {
		rpcService.NodeHardwareService_GetNodeHardwareServer
	}
)

func (*mockGetNodeHardwareError) GetNodeHardware(request *model.GetNodeHardwareRequest) (*model.GetNodeHardwareResponse, error) {
	return nil, errors.New("Error GetNodeHardware")
}

func (*mockGetNodeHardwareSuccess) GetNodeHardware(request *model.GetNodeHardwareRequest) (*model.GetNodeHardwareResponse, error) {
	return &model.GetNodeHardwareResponse{}, nil
}

func (*mockGetNodeHardwareServer) Recv() (*model.GetNodeHardwareRequest, error) {
	return &model.GetNodeHardwareRequest{}, nil
}

func (*mockGetNodeHardwareServer) Send(*model.GetNodeHardwareResponse) error {
	return errors.New("Error Send")
}

func TestNodeHardwareHandler_GetNodeHardware(t *testing.T) {
	type fields struct {
		Service service.NodeHardwareServiceInterface
	}
	type args struct {
		stream rpcService.NodeHardwareService_GetNodeHardwareServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetNodeHardware:Error",
			fields: fields{
				Service: &mockGetNodeHardwareError{},
			},
			args: args{
				stream: &mockGetNodeHardwareServer{},
			},
			wantErr: true,
		},
		{
			name: "GetNodeHardware:Success",
			fields: fields{
				Service: &mockGetNodeHardwareSuccess{},
			},
			args: args{
				stream: &mockGetNodeHardwareServer{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nhh := &NodeHardwareHandler{
				Service: tt.fields.Service,
			}
			if err := nhh.GetNodeHardware(tt.args.stream); (err != nil) != tt.wantErr {
				t.Errorf("NodeHardwareHandler.GetNodeHardware() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
