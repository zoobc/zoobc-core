package query

import (
	"database/sql"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockEscrowQuery = NewEscrowTransactionQuery()
	mockEscrow      = &model.Escrow{
		ID:               1,
		SenderAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		RecipientAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		ApproverAddress:  "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		Amount:           10,
		Commission:       1,
		Timeout:          120,
		Status:           1,
		BlockHeight:      0,
		Latest:           true,
	}
	mockEscrowValues = []interface{}{
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	}
)

func TestEscrowTransactionQuery_InsertEscrowTransaction(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		escrow *model.Escrow
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockEscrowQuery),
			args: args{
				escrow: &model.Escrow{
					ID:               0,
					SenderAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					RecipientAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					ApproverAddress:  "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Amount:           10,
					Commission:       1,
					Timeout:          120,
					Status:           1,
					BlockHeight:      0,
					Latest:           true,
				},
			},
			want: [][]interface{}{
				{
					"UPDATE escrow_transaction set latest = ? WHERE id = ?",
					false,
					int64(0),
				},
				{
					"INSERT INTO escrow_transaction (id,sender_address,recipient_address,approver_address,amount,commission,timeout,status," +
						"block_height,latest,instruction) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					int64(0),
					"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					int64(10),
					int64(1),
					uint64(120),
					model.EscrowStatus_Approved,
					uint32(0),
					true,
					"",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := et.InsertEscrowTransaction(tt.args.escrow); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertEscrowTransaction() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestEscrowTransactionQuery_GetLatestEscrowTransactionByID(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		id int64
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
			fields: fields(*mockEscrowQuery),
			args:   args{id: 1},
			wantQStr: "SELECT id, sender_address, recipient_address, approver_address, amount, commission, timeout, " +
				"status, block_height, latest, instruction FROM escrow_transaction WHERE id = ? AND latest = ?",
			wantArgs: []interface{}{int64(1), true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := et.GetLatestEscrowTransactionByID(tt.args.id)
			if gotQStr != tt.wantQStr {
				t.Errorf("GetLatestEscrowTransactionByID() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetLatestEscrowTransactionByID() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestEscrowTransactionQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		escrow *model.Escrow
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockEscrowQuery),
			args:   args{escrow: mockEscrow},
			want:   mockEscrowValues,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := et.ExtractModel(tt.args.escrow); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestEscrowTransactionQuery_BuildModels(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockEscrowQuery.Fields)
	mockRow.AddRow(
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow, _ := db.Query("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Escrow
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockEscrowQuery),
			args:   args{rows: mockedRow},
			want: []*model.Escrow{
				mockEscrow,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := et.BuildModels(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModels() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestEscrowTransactionQuery_Scan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockEscrowQuery.Fields)
	mockRow.AddRow(
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow := db.QueryRow("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		escrow *model.Escrow
		row    *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockEscrowQuery),
			args:   args{escrow: mockEscrow, row: mockedRow},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := et.Scan(tt.args.escrow, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEscrowTransactionQuery_ExpiringEscrowTransactions(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*mockEscrowQuery),
			args: args{
				blockHeight: 1,
			},
			wantQStr: "UPDATE escrow_transaction SET latest = ?, status = ? WHERE timeout < ? AND status = 0",
			wantArgs: []interface{}{1, model.EscrowStatus_Expired, uint32(1)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := et.ExpiringEscrowTransactions(tt.args.blockHeight)
			if gotQStr != tt.wantQStr {
				t.Errorf("ExpiringEscrowTransactions() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ExpiringEscrowTransactions() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestEscrowTransactionQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*mockEscrowQuery),
			args:   args{height: 1},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM escrow_transaction WHERE block_height > ?",
					uint32(1),
				},
				{
					`
			UPDATE escrow_transaction SET latest = ?
			WHERE latest = ? AND (id, block_height) IN (
				SELECT t2.id, MAX(t2.block_height)
				FROM escrow_transaction as t2
				GROUP BY t2.id
			)`,
					1,
					0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := et.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = \n%v, want \n%v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestEscrowTransactionQuery_SelectDataForSnapshot(t *testing.T) {
	qry := NewEscrowTransactionQuery()
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			args: args{
				fromHeight: 0,
				toHeight:   1,
			},
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
			},
			want: "SELECT id,sender_address,recipient_address,approver_address,amount,commission,timeout,status,block_height,latest," +
				"instruction FROM escrow_transaction WHERE (id, block_height) IN (SELECT t2.id, " +
				"MAX(t2.block_height) FROM escrow_transaction as t2 WHERE t2." +
				"block_height >= 0 AND t2.block_height <= 1 GROUP BY t2.id) ORDER BY" +
				" block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := et.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("EscrowTransactionQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscrowTransactionQuery_GetEscrowTransactions(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fields map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*mockEscrowQuery),
			args: args{
				fields: map[string]interface{}{
					"height": 100,
					"latest": 1,
				},
			},
			want: "SELECT id, sender_address, recipient_address, approver_address, amount, commission, timeout, " +
				"status, block_height, latest, instruction FROM escrow_transaction WHERE height = ? AND latest = ? ",
			want1: []interface{}{
				100,
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, got1 := et.GetEscrowTransactions(tt.args.fields)
			if !strings.Contains(got, "height = ?") || !strings.Contains(got, "latest = ?") {
				t.Errorf("GetEscrowTransactions() got = \n%v, want \n%v", got, tt.want)
				return
			}

			// perhaps tt.want1 is []int without string or any other types and sort it
			if !cmp.Equal(got1, tt.want1, cmp.Transformer("Sort", func(in []interface{}) []interface{} {
				sort.Slice(in, func(i, j int) bool {
					return in[i].(int) < in[j].(int)
				})
				return in
			})) {
				t.Errorf("GetEscrowTransactions() got1 = \n%v, want \n%v", got1, tt.want1)
			}

		})
	}
}

func TestEscrowTransactionQuery_TrimDataBeforeSnapshot(t *testing.T) {
	qry := NewEscrowTransactionQuery()
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "TrimDataBeforeSnapshot",
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
			},
			want: "DELETE FROM escrow_transaction WHERE block_height >= 0 AND block_height <= 10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := et.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("EscrowTransactionQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
