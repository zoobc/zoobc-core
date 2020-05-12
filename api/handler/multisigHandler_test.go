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
	mockGetPendingTransactionsError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionsSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionsError) GetPendingTransactions(param *model.GetPendingTransactionsRequest) (*model.GetPendingTransactionsResponse, error) {
	return nil, errors.New("Error GetPendingTransactions")
}

func (*mockGetPendingTransactionsSuccess) GetPendingTransactions(param *model.GetPendingTransactionsRequest) (*model.GetPendingTransactionsResponse, error) {
	return &model.GetPendingTransactionsResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactions(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactions:ErrorPageLessThanOne",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Error",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Success",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsSuccess{},
			},
			want:    &model.GetPendingTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetPendingTransactionDetailByTransactionHashError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionDetailByTransactionHashSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionDetailByTransactionHashError) GetPendingTransactionDetailByTransactionHash(param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return nil, errors.New("Error GetPendingTransactionDetailByTransactionHash")
}

func (*mockGetPendingTransactionDetailByTransactionHashSuccess) GetPendingTransactionDetailByTransactionHash(param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return &model.GetPendingTransactionDetailByTransactionHashResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactionDetailByTransactionHash(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionDetailByTransactionHashRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionDetailByTransactionHashResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionDetailByTransactionHash:Error",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash:Success",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashSuccess{},
			},
			want:    &model.GetPendingTransactionDetailByTransactionHashResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactionDetailByTransactionHash(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
