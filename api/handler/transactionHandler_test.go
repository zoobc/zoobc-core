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
	mockGetTransactionError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionError) GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error) {
	return nil, errors.New("Error GetTransaction")
}
func (*mockGetTransactionSuccess) GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}

func TestTransactionHandler_GetTransaction(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransaction:Failed",
			fields: fields{
				Service: &mockGetTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransaction:Success",
			fields: fields{
				Service: &mockGetTransactionSuccess{},
			},
			want:    &model.Transaction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetTransactionsError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionsSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionsError) GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	return nil, errors.New("Error GetTransactions")
}

func (*mockGetTransactionsSuccess) GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	return &model.GetTransactionsResponse{}, nil
}

func TestTransactionHandler_GetTransactions(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetTransactions:Failed",
			fields: fields{
				Service: &mockGetTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactions:Success",
			fields: fields{
				Service: &mockGetTransactionsSuccess{},
			},
			want:    &model.GetTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockPostTransactionError struct {
		service.TransactionServiceInterface
	}
	mockPostTransactionSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockPostTransactionError) PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error) {
	return nil, errors.New("Error PostTransaction")
}
func (*mockPostTransactionSuccess) PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}

func TestTransactionHandler_PostTransaction(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.PostTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.PostTransactionResponse
		wantErr bool
	}{
		{
			name: "PostTransaction:Failed",
			fields: fields{
				Service: &mockPostTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "PostTransaction:Success",
			fields: fields{
				Service: &mockPostTransactionSuccess{},
			},
			want: &model.PostTransactionResponse{
				Transaction: &model.Transaction{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.PostTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.PostTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.PostTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetTransactionMinimumFeeError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionMinimumFeeSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionMinimumFeeError) GetTransactionMinimumFee(request *model.GetTransactionMinimumFeeRequest,
) (*model.GetTransactionMinimumFeeResponse, error) {
	return nil, errors.New("Error GetTransactionMinimumFee")
}
func (*mockGetTransactionMinimumFeeSuccess) GetTransactionMinimumFee(request *model.GetTransactionMinimumFeeRequest,
) (*model.GetTransactionMinimumFeeResponse, error) {
	return &model.GetTransactionMinimumFeeResponse{}, nil
}

func TestTransactionHandler_GetTransactionMinimumFee(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionMinimumFeeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetTransactionMinimumFeeResponse
		wantErr bool
	}{

		{
			name: "GetTransactionMinimumFee:Failed",
			fields: fields{
				Service: &mockGetTransactionMinimumFeeError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionMinimumFee:Success",
			fields: fields{
				Service: &mockGetTransactionMinimumFeeSuccess{},
			},
			want:    &model.GetTransactionMinimumFeeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransactionMinimumFee(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransactionMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransactionMinimumFee() = %v, want %v", got, tt.want)
			}
		})
	}
}
