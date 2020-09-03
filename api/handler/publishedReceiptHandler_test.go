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
	mockGetPublishedReceiptsFailed struct {
		service.PublishedReceiptServiceInterface
	}
	mockGetPublishedReceiptsSuccess struct {
		service.PublishedReceiptServiceInterface
	}
)

func (*mockGetPublishedReceiptsFailed) GetPublishedReceipts(*model.GetPublishedReceiptsRequest,
) (*model.GetPublishedReceiptsResponse, error) {
	return nil, errors.New("Error GetPublishedReceipts")
}

func (*mockGetPublishedReceiptsSuccess) GetPublishedReceipts(*model.GetPublishedReceiptsRequest,
) (*model.GetPublishedReceiptsResponse, error) {
	return &model.GetPublishedReceiptsResponse{}, nil
}

func TestPublishedReceiptHandler_GetPublishedReceipts(t *testing.T) {
	type fields struct {
		Service service.PublishedReceiptServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPublishedReceiptsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPublishedReceiptsResponse
		wantErr bool
	}{
		{
			name:   "GetPublishedReceipts:FailedPrecondition",
			fields: fields{},
			args: args{
				req: &model.GetPublishedReceiptsRequest{
					FromHeight: uint32(2),
					ToHeight:   uint32(1),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "GetPublishedReceipts:OutOfRange",
			fields: fields{},
			args: args{
				req: &model.GetPublishedReceiptsRequest{
					FromHeight: uint32(100),
					ToHeight:   uint32(1000),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublishedReceipts:Failed",
			fields: fields{
				Service: &mockGetPublishedReceiptsFailed{},
			},
			args: args{
				req: &model.GetPublishedReceiptsRequest{
					FromHeight: uint32(600),
					ToHeight:   uint32(1000),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublishedReceipts:Success",
			fields: fields{
				Service: &mockGetPublishedReceiptsSuccess{},
			},
			args: args{
				req: &model.GetPublishedReceiptsRequest{
					FromHeight: uint32(600),
					ToHeight:   uint32(1000),
				},
			},
			want:    &model.GetPublishedReceiptsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prh := &PublishedReceiptHandler{
				Service: tt.fields.Service,
			}
			got, err := prh.GetPublishedReceipts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishedReceiptHandler.GetPublishedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublishedReceiptHandler.GetPublishedReceipts() = %v, want %v", got, tt.want)
			}
		})
	}
}
