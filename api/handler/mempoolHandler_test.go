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
	MockGetMempoolTransactionError struct {
		service.MempoolTransactionServiceInterface
	}
	MockGetMempoolTransactionSuccess struct {
		service.MempoolTransactionServiceInterface
	}
)

func (*MockGetMempoolTransactionError) GetMempoolTransaction(chainType chaintype.ChainType, params *model.GetMempoolTransactionRequest) (*model.GetMempoolTransactionResponse, error) {
	return nil, errors.New("Error GetMempoolTransaction")
}

func (*MockGetMempoolTransactionSuccess) GetMempoolTransaction(chainType chaintype.ChainType, params *model.GetMempoolTransactionRequest) (*model.GetMempoolTransactionResponse, error) {
	return &model.GetMempoolTransactionResponse{}, nil
}

func TestMempoolTransactionHandler_GetMempoolTransaction(t *testing.T) {
	type fields struct {
		Service service.MempoolTransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetMempoolTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMempoolTransactionResponse
		wantErr bool
	}{
		{
			name: "GetMempoolTransaction:Error",
			fields: fields{
				Service: &MockGetMempoolTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMempoolTransaction:Success",
			fields: fields{
				Service: &MockGetMempoolTransactionSuccess{},
			},
			want:    &model.GetMempoolTransactionResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uth := &MempoolTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := uth.GetMempoolTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolTransactionHandler.GetMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolTransactionHandler.GetMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	MockGetMempoolTransactionsError struct {
		service.MempoolTransactionServiceInterface
	}
	MockGetMempoolTransactionsSuccess struct {
		service.MempoolTransactionServiceInterface
	}
)

func (*MockGetMempoolTransactionsError) GetMempoolTransactions(chainType chaintype.ChainType, params *model.GetMempoolTransactionsRequest) (*model.GetMempoolTransactionsResponse, error) {
	return nil, errors.New("Error GetMempoolTransactions")
}

func (*MockGetMempoolTransactionsSuccess) GetMempoolTransactions(chainType chaintype.ChainType, params *model.GetMempoolTransactionsRequest) (*model.GetMempoolTransactionsResponse, error) {
	return &model.GetMempoolTransactionsResponse{}, nil
}

func TestMempoolTransactionHandler_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Service service.MempoolTransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetMempoolTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMempoolTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetMempoolTransactions:Error",
			fields: fields{
				Service: &MockGetMempoolTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMempoolTransactions:Success",
			fields: fields{
				Service: &MockGetMempoolTransactionsSuccess{},
			},
			want:    &model.GetMempoolTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uth := &MempoolTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := uth.GetMempoolTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolTransactionHandler.GetMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolTransactionHandler.GetMempoolTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
