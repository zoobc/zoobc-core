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
	mockGetProofOfOwnershipError struct {
		service.NodeAdminServiceInterface
	}
	mockGetProofOfOwnershipSuccess struct {
		service.NodeAdminServiceInterface
	}
)

func (*mockGetProofOfOwnershipError) GetProofOfOwnership() (*model.ProofOfOwnership, error) {
	return nil, errors.New("Error GetProofOfOwnership")
}

func (*mockGetProofOfOwnershipSuccess) GetProofOfOwnership() (*model.ProofOfOwnership, error) {
	return &model.ProofOfOwnership{}, nil
}

func TestNodeAdminHandler_GetProofOfOwnership(t *testing.T) {
	type fields struct {
		Service service.NodeAdminServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetProofOfOwnershipRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "GetProofOfOwnership:Error",
			fields: fields{
				Service: &mockGetProofOfOwnershipError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetProofOfOwnership:Success",
			fields: fields{
				Service: &mockGetProofOfOwnershipSuccess{},
			},
			want:    &model.ProofOfOwnership{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gp := &NodeAdminHandler{
				Service: tt.fields.Service,
			}
			got, err := gp.GetProofOfOwnership(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminHandler.GetProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminHandler.GetProofOfOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGenerateNodeKeyError struct {
		service.NodeAdminServiceInterface
	}
	mockGenerateNodeKeySuccess struct {
		service.NodeAdminServiceInterface
	}
)

func (*mockGenerateNodeKeyError) GenerateNodeKey(seed string) ([]byte, error) {
	return nil, errors.New("Error GenerateNodeKey")
}

func (*mockGenerateNodeKeySuccess) GenerateNodeKey(seed string) ([]byte, error) {
	return []byte(""), nil
}

func TestNodeAdminHandler_GenerateNodeKey(t *testing.T) {
	type fields struct {
		Service service.NodeAdminServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GenerateNodeKeyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GenerateNodeKeyResponse
		wantErr bool
	}{
		{
			name: "GenerateNodeKey:Error",
			fields: fields{
				Service: &mockGenerateNodeKeyError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GenerateNodeKey:Success",
			fields: fields{
				Service: &mockGenerateNodeKeySuccess{},
			},
			want: &model.GenerateNodeKeyResponse{
				NodePublicKey: []byte(""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gp := &NodeAdminHandler{
				Service: tt.fields.Service,
			}
			got, err := gp.GenerateNodeKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminHandler.GenerateNodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminHandler.GenerateNodeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
