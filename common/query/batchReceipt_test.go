package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockReceiptQuery = NewBatchReceiptQuery()
	mockBatchReceipt = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey:      []byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
			RecipientPublicKey:   []byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
			DatumType:            uint32(1),
			DatumHash:            []byte{1, 2, 3, 4, 5, 6},
			ReferenceBlockHeight: uint32(1),
			ReferenceBlockHash:   []byte{1, 2, 3, 4, 5, 6},
			RMRLinked:            []byte{1, 2, 3, 4, 5, 6},
			RecipientSignature:   []byte{1, 2, 3, 4, 5, 6},
		},
		RMR:      []byte{1, 2, 3, 4, 5, 6},
		RMRIndex: uint32(4),
	}
)

func TestReceiptQuery_InsertReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipts []*model.BatchReceipt
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
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipts: []*model.BatchReceipt{mockBatchReceipt},
			},
			wantQStr: "INSERT INTO node_receipt " +
				"(sender_public_key, recipient_public_key, " +
				"datum_type, datum_hash, reference_block_height, " +
				"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index) " +
				"VALUES(?,? ,? ,? ,? ,? ,? ,? ,? ,? )",
			wantArgs: []interface{}{
				&mockBatchReceipt.Receipt.SenderPublicKey,
				&mockBatchReceipt.Receipt.RecipientPublicKey,
				&mockBatchReceipt.Receipt.DatumType,
				&mockBatchReceipt.Receipt.DatumHash,
				&mockBatchReceipt.Receipt.ReferenceBlockHeight,
				&mockBatchReceipt.Receipt.ReferenceBlockHash,
				&mockBatchReceipt.Receipt.RMRLinked,
				&mockBatchReceipt.Receipt.RecipientSignature,
				&mockBatchReceipt.RMR,
				&mockBatchReceipt.RMRIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := rq.InsertReceipts(tt.args.batchReceipts)
			if gotQStr != tt.wantQStr {
				t.Errorf("BatchReceiptQuery.InsertReceipts() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.InsertReceipts() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestReceiptQuery_InsertReceipt(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipt *model.BatchReceipt
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipt: mockBatchReceipt,
			},
			wantStr: "INSERT INTO node_receipt " +
				"(sender_public_key, recipient_public_key, datum_type, datum_hash, " +
				"reference_block_height, reference_block_hash, rmr_linked, " +
				"recipient_signature, rmr, rmr_index) VALUES(? , ? , ? , ? , ? , ? , ? , ? , ? , ? )",
			wantArgs: []interface{}{
				&mockBatchReceipt.Receipt.SenderPublicKey,
				&mockBatchReceipt.Receipt.RecipientPublicKey,
				&mockBatchReceipt.Receipt.DatumType,
				&mockBatchReceipt.Receipt.DatumHash,
				&mockBatchReceipt.Receipt.ReferenceBlockHeight,
				&mockBatchReceipt.Receipt.ReferenceBlockHash,
				&mockBatchReceipt.Receipt.RMRLinked,
				&mockBatchReceipt.Receipt.RecipientSignature,
				&mockBatchReceipt.RMR,
				&mockBatchReceipt.RMRIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := rq.InsertReceipt(tt.args.batchReceipt)
			if gotStr != tt.wantStr {
				t.Errorf("BatchReceiptQuery.InsertReceipt() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.InsertReceipt() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestReceiptQuery_GetReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		paginate model.Pagination
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{paginate: model.Pagination{
				OrderBy: model.OrderBy_ASC,
			}},
			want: "SELECT sender_public_key, recipient_public_key, datum_type, " +
				"datum_hash, reference_block_height, reference_block_hash, rmr_linked, " +
				"recipient_signature, rmr, rmr_index FROM node_receipt ORDER BY reference_block_height " +
				"ASC LIMIT 256 OFFSET 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := rq.GetReceipts(tt.args.paginate); got != tt.want {
				t.Errorf("BatchReceiptQuery.GetReceipts() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorNodeReceiptScan struct {
		Executor
	}
)

func (*mockQueryExecutorNodeReceiptScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockReceiptQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(4),
	))
	return db.QueryRow("")
}

func TestReceiptQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipt *model.BatchReceipt
		row          *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipt: mockBatchReceipt,
				row:          (&mockQueryExecutorNodeReceiptScan{}).ExecuteSelectRow("", ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := r.Scan(tt.args.batchReceipt, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("BatchReceiptQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorNodeReceiptBuildModel struct {
		Executor
	}
)

func (*mockQueryExecutorNodeReceiptBuildModel) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockReceiptQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(4),
	))
	return db.Query("")
}

func TestReceiptQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipts []*model.BatchReceipt
		rows          *sql.Rows
	}
	rows, err := (&mockQueryExecutorNodeReceiptBuildModel{}).ExecuteSelect("", false, "")
	if err != nil {
		t.Errorf("Rows Failed: %s", err.Error())
		return
	}
	defer rows.Close()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.BatchReceipt
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipts: []*model.BatchReceipt{},
				rows:          rows,
			},
			want: []*model.BatchReceipt{mockBatchReceipt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := re.BuildModel(tt.args.batchReceipts, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeReceiptQuery_PruneData(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockHeight uint32
		limit       uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "WantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				blockHeight: 2000,
				limit:       500,
			},
			want: "DELETE FROM node_receipt WHERE reference_block_height IN(" +
				"SELECT reference_block_height FROM node_receipt " +
				"WHERE reference_block_height <? ORDER BY reference_block_height ASC LIMIT ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, args := rq.PruneData(tt.args.blockHeight, tt.args.limit)
			if got != tt.want {
				t.Errorf("PruneData() = \n%v, want \n%v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(args, []interface{}{tt.args.blockHeight, tt.args.limit}) {
				t.Errorf("PruneData() = \n%v, want \n%v", args, []interface{}{tt.args.blockHeight, tt.args.limit})
			}
		})
	}
}
