package service

import (
	"errors"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/util"
	"reflect"
	"testing"
)

type (
	mockPublishedReceiptsUtilSuccess struct {
		util.PublishedReceiptUtil
	}
	mockPublishedReceiptsUtilError struct {
		util.PublishedReceiptUtil
	}
)

func (*mockPublishedReceiptsUtilSuccess) GetPublishedReceiptsByBlockHeightRange(
	fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}

func (*mockPublishedReceiptsUtilError) GetPublishedReceiptsByBlockHeightRange(
	fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockedError")
}

func TestPublishedReceiptService_GetPublishedReceipts(t *testing.T) {
	type fields struct {
		PublishedReceiptUtil util.PublishedReceiptUtilInterface
	}
	type args struct {
		params *model.GetPublishedReceiptsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPublishedReceiptsResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				PublishedReceiptUtil: &mockPublishedReceiptsUtilSuccess{},
			},
			args: args{
				params: &model.GetPublishedReceiptsRequest{
					FromHeight: 0,
					ToHeight:   100,
				},
			},
			want: &model.GetPublishedReceiptsResponse{
				PublishedReceipts: []*model.PublishedReceipt{},
			},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				PublishedReceiptUtil: &mockPublishedReceiptsUtilError{},
			},
			args: args{
				params: &model.GetPublishedReceiptsRequest{
					FromHeight: 0,
					ToHeight:   100,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prs := &PublishedReceiptService{
				PublishedReceiptUtil: tt.fields.PublishedReceiptUtil,
			}
			got, err := prs.GetPublishedReceipts(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublishedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublishedReceipts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
