package query

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBatchQuery   = NewBatchReceiptQuery()
	mockBatchReceipt = &model.Receipt{
		SenderPublicKey:    []byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		RecipientPublicKey: []byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
	}
)

func TestNewBatchReceiptQuery(t *testing.T) {
	tests := []struct {
		name string
		want *BatchReceiptQuery
	}{
		{
			name: "wantSuccess",
			want: mockBatchQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBatchReceiptQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBatchReceiptQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchReceiptQuery_getTableName(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			want:   "batch_receipt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := br.getTableName(); got != tt.want {
				t.Errorf("BatchReceiptQuery.getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchReceiptQuery_InsertBatchReceipt(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipt *model.Receipt
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			args: args{
				receipt: mockBatchReceipt,
			},
			wantQStr: fmt.Sprintf(
				"INSERT INTO batch_receipt (%s) VALUES(%s)",
				strings.Join(mockBatchQuery.Fields, ", "),
				fmt.Sprintf("? %s", strings.Repeat(", ?", len(mockBatchQuery.Fields)-1)),
			),
			wantArgs: mockBatchQuery.ExtractModel(mockBatchReceipt),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := br.InsertBatchReceipt(tt.args.receipt)
			if gotQStr != tt.wantQStr {
				t.Errorf("BatchReceiptQuery.InsertBatchReceipt() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.InsertBatchReceipt() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBatchReceiptQuery_GetBatchReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			want:   fmt.Sprintf("SELECT %s FROM %s", strings.Join(mockBatchQuery.Fields, ", "), mockBatchQuery.getTableName()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := br.GetBatchReceipts(); got != tt.want {
				t.Errorf("BatchReceiptQuery.GetBatchReceipts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchReceiptQuery_RemoveBatchReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			want:   fmt.Sprintf("DELETE FROM %s", mockBatchQuery.getTableName()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := br.RemoveBatchReceipts(); got != tt.want {
				t.Errorf("BatchReceiptQuery.RemoveBatchReceipts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchReceiptQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipt *model.Receipt
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			args: args{
				receipt: mockBatchReceipt,
			},
			want: mockBatchQuery.ExtractModel(mockBatchReceipt),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := b.ExtractModel(tt.args.receipt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchReceiptQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
