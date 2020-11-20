package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

func TestNewPublishedReceiptService(t *testing.T) {
	type args struct {
		publishedReceiptQuery query.PublishedReceiptQueryInterface
		receiptUtil           util.ReceiptUtilInterface
		publishedReceiptUtil  util.PublishedReceiptUtilInterface
		receiptService        ReceiptServiceInterface
		queryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *PublishedReceiptService
	}{
		{
			name: "NewPublishedReceiptService-Success",
			args: args{
				publishedReceiptQuery: nil,
				receiptUtil:           nil,
				publishedReceiptUtil:  nil,
				receiptService:        nil,
				queryExecutor:         nil,
			},
			want: &PublishedReceiptService{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  nil,
				ReceiptService:        nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublishedReceiptService(tt.args.publishedReceiptQuery, tt.args.receiptUtil, tt.args.publishedReceiptUtil,
				tt.args.receiptService, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPublishedReceiptService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// ProcessPublishedReceipts mocks
	mockProcessPublishedReceiptsReceiptServiceFail struct {
		ReceiptService
	}
	mockProcessPublishedReceiptsReceiptServiceSuccess struct {
		ReceiptService
	}
	mockProcessPublishedReceiptPublishedReceiptUtilFail struct {
		util.PublishedReceiptUtil
	}
	mockProcessPublishedReceiptPublishedReceiptUtilSuccess struct {
		util.PublishedReceiptUtil
	}
	// ProcessPublishedReceipts mocks
)

func (*mockProcessPublishedReceiptsReceiptServiceFail) ValidateReceipt(
	_ *model.Receipt,
) error {
	return errors.New("mockedError")
}

func (*mockProcessPublishedReceiptsReceiptServiceSuccess) ValidateReceipt(
	_ *model.Receipt,
) error {
	return nil
}

func (*mockProcessPublishedReceiptPublishedReceiptUtilFail) InsertPublishedReceipt(
	_ *model.PublishedReceipt, _ bool,
) error {
	return errors.New("mockedError")
}

func (*mockProcessPublishedReceiptPublishedReceiptUtilSuccess) InsertPublishedReceipt(
	_ *model.PublishedReceipt, _ bool,
) error {
	return nil
}

func TestPublishedReceiptService_ProcessPublishedReceipts(t *testing.T) {
	dummyPublishedReceipts := []*model.PublishedReceipt{
		{
			Receipt: &model.Receipt{},
		},
	}
	type fields struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		ReceiptUtil           util.ReceiptUtilInterface
		PublishedReceiptUtil  util.PublishedReceiptUtilInterface
		ReceiptService        ReceiptServiceInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "ProcessPublishedReceipt-NoReceipt",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  nil,
				ReceiptService:        nil,
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					FreeReceipts: make([]*model.PublishedReceipt, 0),
				},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "ProcessPublishedReceipt-ValidateReceiptFail",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  nil,
				ReceiptService:        &mockProcessPublishedReceiptsReceiptServiceFail{},
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					FreeReceipts: dummyPublishedReceipts,
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "ProcessPublishedReceipt-ValidateReceiptFail",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  &mockProcessPublishedReceiptPublishedReceiptUtilFail{},
				ReceiptService:        &mockProcessPublishedReceiptsReceiptServiceSuccess{},
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					FreeReceipts: dummyPublishedReceipts,
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "ProcessPublishedReceipt-ValidateReceiptFail",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  &mockProcessPublishedReceiptPublishedReceiptUtilSuccess{},
				ReceiptService:        &mockProcessPublishedReceiptsReceiptServiceSuccess{},
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					PublishedReceipts: dummyPublishedReceipts,
				},
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PublishedReceiptService{
				PublishedReceiptQuery: tt.fields.PublishedReceiptQuery,
				ReceiptUtil:           tt.fields.ReceiptUtil,
				PublishedReceiptUtil:  tt.fields.PublishedReceiptUtil,
				ReceiptService:        tt.fields.ReceiptService,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := ps.ProcessPublishedReceipts(nil, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPublishedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProcessPublishedReceipts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
