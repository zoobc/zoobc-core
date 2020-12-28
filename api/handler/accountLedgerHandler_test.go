package handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type mockGetAccountLedgersSuccess struct {
	service.AccountLedgerServiceInterface
}

func (*mockGetAccountLedgersSuccess) GetAccountLedgers(request *model.GetAccountLedgersRequest) (*model.GetAccountLedgersResponse, error) {
	return &model.GetAccountLedgersResponse{}, nil
}

func TestAccountLedgerHandler_GetAccountLedgers(t *testing.T) {
	type fields struct {
		Service service.AccountLedgerServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetAccountLedgersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountLedgersResponse
		wantErr bool
	}{
		{
			name: "GetAccountLedgers:Success",
			fields: fields{
				Service: &mockGetAccountLedgersSuccess{},
			},
			args: args{
				request: &model.GetAccountLedgersRequest{},
			},
			want:    &model.GetAccountLedgersResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AccountLedgerHandler{
				Service: tt.fields.Service,
			}
			got, err := al.GetAccountLedgers(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountLedgers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountLedgers() got = %v, want %v", got, tt.want)
			}
		})
	}
}
