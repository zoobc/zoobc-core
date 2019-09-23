package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBatchQuery   = NewBatchReceiptQuery()
	mockBatchReceipt = &model.Receipt{
		SenderPublicKey:      []byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		RecipientPublicKey:   []byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		DatumType:            uint32(1),
		DatumHash:            []byte{1, 2, 3, 4, 5, 6},
		ReferenceBlockHeight: uint32(1),
		ReferenceBlockHash:   []byte{1, 2, 3, 4, 5, 6},
		ReceiptMerkleRoot:    []byte{1, 2, 3, 4, 5, 6},
		RecipientSignature:   []byte{1, 2, 3, 4, 5, 6},
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
			want: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, " +
				"reference_block_height, reference_block_hash, receipt_merkle_root, recipient_signature " +
				"FROM batch_receipt ORDER BY reference_block_height LIMIT 10 OFFSET 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := br.GetBatchReceipts(10, 0); got != tt.want {
				t.Errorf("BatchReceiptQuery.GetBatchReceipts() = %v, want %v", got, tt.want)
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

func TestBatchReceiptQuery_RemoveBatchReceiptByRoot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root []byte
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
				root: []byte{1, 2, 3},
			},
			wantQStr: "DELETE FROM batch_receipt WHERE receipt_merkle_root = ?",
			wantArgs: []interface{}{[]byte{1, 2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := br.RemoveBatchReceiptByRoot(tt.args.root)
			if gotQStr != tt.wantQStr {
				t.Errorf("BatchReceiptQuery.RemoveBatchReceiptByRoot() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.RemoveBatchReceiptByRoot() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

type (
	mockQueryExecutorBatchReceiptBuildModel struct {
		Executor
	}
)

func (*mockQueryExecutorBatchReceiptBuildModel) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockBatchQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
	))
	return db.Query("")
}
func TestBatchReceiptQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipts []*model.Receipt
		rows     *sql.Rows
	}
	rows, _ := (&mockQueryExecutorBatchReceiptBuildModel{}).ExecuteSelect("", false, "")
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.Receipt
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			args: args{
				receipts: []*model.Receipt{},
				rows:     rows,
			},
			want: []*model.Receipt{mockBatchReceipt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := b.BuildModel(tt.args.receipts, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchReceiptQuery.BuildModel() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorBatchReceiptScan struct {
		Executor
	}
)

func (*mockQueryExecutorBatchReceiptScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockBatchQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
	))
	return db.QueryRow("")
}

func TestBatchReceiptQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipt *model.Receipt
		row     *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockBatchQuery),
			args: args{
				receipt: mockBatchReceipt,
				row:     (&mockQueryExecutorBatchReceiptScan{}).ExecuteSelectRow("", ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := b.Scan(tt.args.receipt, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("BatchReceiptQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
