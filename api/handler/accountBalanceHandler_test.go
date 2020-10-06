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
	mockGetAccountBalanceError struct {
		service.AccountBalanceServiceInterface
	}
	mockGetAccountBalanceSuccess struct {
		service.AccountBalanceServiceInterface
	}
)

func (*mockGetAccountBalanceError) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	return nil, errors.New("error GetAccountBalance")
}
func (*mockGetAccountBalanceSuccess) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	return &model.GetAccountBalanceResponse{
		AccountBalance: &model.AccountBalance{
			AccountAddress: request.AccountAddress,
		},
	}, nil
}

func TestAccountBalanceHandler_GetAccountBalance(t *testing.T) {
	type fields struct {
		Service service.AccountBalanceServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetAccountBalanceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalanceResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalance:fail",
			fields: fields{
				Service: &mockGetAccountBalanceError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalance:success",
			fields: fields{
				Service: &mockGetAccountBalanceSuccess{},
			},
			args: args{
				request: &model.GetAccountBalanceRequest{
					AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140,
						183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want: &model.GetAccountBalanceResponse{
				AccountBalance: &model.AccountBalance{
					AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140,
						183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHandler{
				Service: tt.fields.Service,
			}
			got, err := abh.GetAccountBalance(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceHandler.GetAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceHandler.GetAccountBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetAccountBalancesSuccess struct {
		service.AccountBalanceServiceInterface
	}
)

func (*mockGetAccountBalancesSuccess) GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error) {
	return &model.GetAccountBalancesResponse{
		AccountBalances: []*model.AccountBalance{},
	}, nil
}

func TestAccountBalanceHandler_GetAccountBalances(t *testing.T) {
	type fields struct {
		Service service.AccountBalanceServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetAccountBalancesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalancesResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalancesHandler:fail",
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: [][]byte{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalancesHandler:success",
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: [][]byte{
						[]byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140,
							183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
					},
				},
			},
			fields: fields{
				Service: &mockGetAccountBalancesSuccess{},
			},
			want: &model.GetAccountBalancesResponse{
				AccountBalances: []*model.AccountBalance{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHandler{
				Service: tt.fields.Service,
			}
			got, err := abh.GetAccountBalances(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceHandler.GetAccountBalances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceHandler.GetAccountBalances() = %v, want %v", got, tt.want)
			}
		})
	}
}
