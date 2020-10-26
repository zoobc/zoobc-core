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
	mockGetAccountDatasetsError struct {
		service.AccountDatasetServiceInterface
	}
	mockGetAccountDatasetsSuccess struct {
		service.AccountDatasetServiceInterface
	}
)

func (*mockGetAccountDatasetsError) GetAccountDatasets(request *model.GetAccountDatasetsRequest) (*model.GetAccountDatasetsResponse, error) {
	return nil, errors.New("Error GetAccountDatasets")
}
func (*mockGetAccountDatasetsSuccess) GetAccountDatasets(request *model.GetAccountDatasetsRequest) (*model.GetAccountDatasetsResponse, error) {
	return &model.GetAccountDatasetsResponse{}, nil
}

func TestAccountDatasetHandler_GetAccountDatasets(t *testing.T) {
	type fields struct {
		Service service.AccountDatasetServiceInterface
	}
	type args struct {
		in0     context.Context
		request *model.GetAccountDatasetsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountDatasetsResponse
		wantErr bool
	}{
		{
			name: "GetAccountDatasets:LimitExceeded",
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(600),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDatasets:Error",
			fields: fields{
				Service: &mockGetAccountDatasetsError{},
			},
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(250),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDatasets:Success",
			fields: fields{
				Service: &mockGetAccountDatasetsSuccess{},
			},
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(250),
					},
				},
			},
			want:    &model.GetAccountDatasetsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adh := &AccountDatasetHandler{
				Service: tt.fields.Service,
			}
			got, err := adh.GetAccountDatasets(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountDatasetHandler.GetAccountDatasets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetHandler.GetAccountDatasets() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetAccountDatasetError struct {
		service.AccountDatasetServiceInterface
	}
	mockGetAccountDatasetSuccess struct {
		service.AccountDatasetServiceInterface
	}
)

func (*mockGetAccountDatasetError) GetAccountDataset(request *model.GetAccountDatasetRequest) (*model.AccountDataset, error) {
	return nil, errors.New("Error GetAccountDataset")
}
func (*mockGetAccountDatasetSuccess) GetAccountDataset(request *model.GetAccountDatasetRequest) (*model.AccountDataset, error) {
	return &model.AccountDataset{}, nil
}

func TestAccountDatasetHandler_GetAccountDataset(t *testing.T) {
	type fields struct {
		Service service.AccountDatasetServiceInterface
	}
	type args struct {
		in0     context.Context
		request *model.GetAccountDatasetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.AccountDataset
		wantErr bool
	}{
		{
			name: "GetAccountDataset:InvalidRequest",
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: nil,
					Property:                "",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDataset:Error",
			fields: fields{
				Service: &mockGetAccountDatasetError{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112,
						113, 140, 183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDataset:Success",
			fields: fields{
				Service: &mockGetAccountDatasetSuccess{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112,
						113, 140, 183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want:    &model.AccountDataset{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adh := &AccountDatasetHandler{
				Service: tt.fields.Service,
			}
			got, err := adh.GetAccountDataset(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountDatasetHandler.GetAccountDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetHandler.GetAccountDataset() = %v, want %v", got, tt.want)
			}
		})
	}
}
