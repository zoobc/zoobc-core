package handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetEscrowTransactionsFailed struct {
		service.EscrowTransactionServiceInterface
	}
	mockGetEscrowTransactionsSuccess struct {
		service.EscrowTransactionServiceInterface
	}
)

func (*mockGetEscrowTransactionsFailed) GetEscrowTransactions(request *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	return nil, errors.New("Error GetEscrowTransactions")
}

func (*mockGetEscrowTransactionsSuccess) GetEscrowTransactions(request *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	return &model.GetEscrowTransactionsResponse{}, nil
}

func TestEscrowTransactionHandler_GetEscrowTransactions(t *testing.T) {
	type fields struct {
		Service service.EscrowTransactionServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetEscrowTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetEscrowTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetEscrowTransactions:Error",
			fields: fields{
				Service: &mockGetEscrowTransactionsFailed{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetEscrowTransactions:Success",
			fields: fields{
				Service: &mockGetEscrowTransactionsSuccess{},
			},
			want:    &model.GetEscrowTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := &EscrowTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := eh.GetEscrowTransactions(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetEscrowTransactionError struct {
		service.EscrowTransactionServiceInterface
	}
	mockGetEscrowTransactionSuccess struct {
		service.EscrowTransactionServiceInterface
	}
)

func (*mockGetEscrowTransactionError) GetEscrowTransaction(request *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	return nil, errors.New("Error GetEscrowTransaction")
}
func (*mockGetEscrowTransactionSuccess) GetEscrowTransaction(request *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	return &model.Escrow{}, nil
}

func TestEscrowTransactionHandler_GetEscrowTransaction(t *testing.T) {
	type fields struct {
		Service service.EscrowTransactionServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetEscrowTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Escrow
		wantErr bool
	}{
		{
			name: "GetEscrowTransaction:Error",
			fields: fields{
				Service: &mockGetEscrowTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetEscrowTransaction:Success",
			fields: fields{
				Service: &mockGetEscrowTransactionSuccess{},
			},
			want:    &model.Escrow{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := &EscrowTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := eh.GetEscrowTransaction(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
