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
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				},
			},
			want: &model.GetAccountBalanceResponse{
				AccountBalance: &model.AccountBalance{
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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
