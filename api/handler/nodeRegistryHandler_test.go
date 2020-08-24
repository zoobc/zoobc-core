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
	mockGetNodeRegistrationsFailed struct {
		service.NodeRegistryServiceInterface
	}
	mockGetNodeRegistrationsSuccess struct {
		service.NodeRegistryServiceInterface
	}
)

func (*mockGetNodeRegistrationsFailed) GetNodeRegistrations(*model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	return nil, errors.New("Error GetNodeRegistrations")
}
func (*mockGetNodeRegistrationsSuccess) GetNodeRegistrations(*model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	return &model.GetNodeRegistrationsResponse{}, nil
}

func TestNodeRegistryHandler_GetNodeRegistrations(t *testing.T) {
	type fields struct {
		Service service.NodeRegistryServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeRegistrationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationsResponse
		wantErr bool
	}{
		{
			name: "GetNodeRegistrations:Error",
			fields: fields{
				Service: &mockGetNodeRegistrationsFailed{},
			},
			args: args{
				req: &model.GetNodeRegistrationsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistrations:Success",
			fields: fields{
				Service: &mockGetNodeRegistrationsSuccess{},
			},
			args: args{
				req: &model.GetNodeRegistrationsRequest{},
			},
			want:    &model.GetNodeRegistrationsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrh := NodeRegistryHandler{
				Service: tt.fields.Service,
			}
			got, err := nrh.GetNodeRegistrations(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNodeRegistrationFailed struct {
		service.NodeRegistryServiceInterface
	}
	mockGetNodeRegistrationSuccess struct {
		service.NodeRegistryServiceInterface
	}
)

func (*mockGetNodeRegistrationFailed) GetNodeRegistration(*model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	return nil, errors.New("Error GetNodeRegistration")
}

func (*mockGetNodeRegistrationSuccess) GetNodeRegistration(*model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	return &model.GetNodeRegistrationResponse{}, nil
}

func TestNodeRegistryHandler_GetNodeRegistration(t *testing.T) {
	type fields struct {
		Service service.NodeRegistryServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeRegistrationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationResponse
		wantErr bool
	}{
		{
			name: "GetNodeRegistration:Failed",
			fields: fields{
				Service: &mockGetNodeRegistrationFailed{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistration:Success",
			fields: fields{
				Service: &mockGetNodeRegistrationSuccess{},
			},
			args: args{
				req: &model.GetNodeRegistrationRequest{},
			},
			want:    &model.GetNodeRegistrationResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrh := NodeRegistryHandler{
				Service: tt.fields.Service,
			}
			got, err := nrh.GetNodeRegistration(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryHandler.GetNodeRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryHandler.GetNodeRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}
