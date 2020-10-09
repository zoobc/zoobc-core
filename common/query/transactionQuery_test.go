package query

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockTransactionQuery = NewTransactionQuery(chaintype.GetChainType(0))
	mockTransaction      = &model.Transaction{
		ID:      -1273123123,
		BlockID: -123123123123,
		Version: 1,
		Height:  1,
		SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
			28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		TransactionType:       binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
		Fee:                   1,
		Timestamp:             10000,
		TransactionHash:       make([]byte, 200),
		TransactionBodyLength: 88,
		TransactionBodyBytes:  make([]byte, 88),
		Signature:             make([]byte, 68),
		TransactionIndex:      1,
	}
	// mockTransactionRow represent a transaction row for test purpose only
	// copy just the values only,
	mockTransactionRow = []interface{}{
		-1273123123,
		-123123123123,
		1,
		mockTransaction.SenderAccountAddress,
		mockTransaction.RecipientAccountAddress,
		binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
		1,
		10000,
		make([]byte, 200),
		88,
		make([]byte, 88),
		make([]byte, 68),
		1,
		1,
	}
)
var _ = mockTransactionRow

func TestGetTransaction(t *testing.T) {
	transactionQuery := NewTransactionQuery(chaintype.GetChainType(0))

	type paramsStruct struct {
		ID int64
	}

	tests := []struct {
		name   string
		params *paramsStruct
		want   string
	}{
		{
			name: "transaction query with ID param only",
			params: &paramsStruct{
				ID: 1,
			},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index, multisig_child from \"transaction\"" +
				" WHERE id = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := transactionQuery.GetTransaction(tt.params.ID)
			if query != tt.want {
				t.Errorf("GetTransactionError() \ngot = %v \nwant = %v", query, tt.want)
				return
			}
		})
	}
}

func TestTransactionQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
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
			name:   "wantSuccess",
			fields: fields(*mockTransactionQuery),
			args:   args{height: uint32(1)},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM \"transaction\" WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tq := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotMultiQueries := tq.Rollback(tt.args.height)
			if !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
				return
			}
		})
	}
}

func TestTransactionQuery_InsertTransaction(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		tx *model.Transaction
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
			fields: fields(*mockTransactionQuery),
			args:   args{tx: mockTransaction},
			wantStr: fmt.Sprintf("INSERT INTO \"transaction\" (%s) VALUES(?%s)",
				strings.Join(mockTransactionQuery.Fields, ", "),
				strings.Repeat(", ?", len(mockTransactionQuery.Fields)-1),
			),
			wantArgs: mockTransactionQuery.ExtractModel(mockTransaction),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tq := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := tq.InsertTransaction(tt.args.tx)
			if ok := strings.Compare(regexp.QuoteMeta(gotStr), regexp.QuoteMeta(tt.wantStr)); ok != 0 {
				t.Errorf("InsertTransaction() gotStr = %v, want %v", gotStr, tt.wantStr)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertTransaction() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestTransactionQuery_GetTransactionsByBlockID(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		blockID int64
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
			fields: fields(*mockTransactionQuery),
			args:   args{blockID: int64(1)},
			wantStr: fmt.Sprintf("SELECT %s FROM \"transaction\" WHERE block_id = ? AND multisig_child = false"+
				" ORDER BY transaction_index ASC",
				strings.Join(mockTransactionQuery.Fields, ", "),
			),
			wantArgs: []interface{}{int64(1)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tq := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := tq.GetTransactionsByBlockID(tt.args.blockID)
			if gotStr != tt.wantStr {
				t.Errorf("GetTransactionsByBlockID() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetTransactionsByBlockID() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestTransactionQuery_GetTransactionsByIds(t *testing.T) {

	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		txIds []int64
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
			fields: fields(*mockTransactionQuery),
			args:   args{txIds: []int64{1, 2, 3, 4}},
			wantStr: "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, transaction_index, multisig_child " +
				"FROM \"transaction\" WHERE multisig_child = false AND id IN(?, ?, ?, ?)",
			wantArgs: []interface{}{
				int64(1),
				int64(2),
				int64(3),
				int64(4),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tq := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := tq.GetTransactionsByIds(tt.args.txIds)
			if gotStr != tt.wantStr {
				t.Errorf("GetTransactionsByIds() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetTransactionsByIds() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

type (
	mockQueryExecutorBuildModel struct {
		Executor
	}
)

func (*mockQueryExecutorBuildModel) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockTransactionQuery.Fields).AddRow(
			-1273123123,
			-123123123123,
			1,
			mockTransaction.SenderAccountAddress,
			mockTransaction.RecipientAccountAddress,
			binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
			1,
			10000,
			make([]byte, 200),
			88,
			make([]byte, 88),
			make([]byte, 68),
			1,
			1,
			false,
		),
	)
	return db.Query("")
}
func TestTransactionQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		txs  []*model.Transaction
		rows *sql.Rows
	}
	rows, _ := (&mockQueryExecutorBuildModel{}).ExecuteSelect("", false)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.Transaction
	}{
		{
			name:   "wantTransaction",
			fields: fields(*mockTransactionQuery),
			args: args{
				txs:  []*model.Transaction{},
				rows: rows,
			},
			want: []*model.Transaction{
				mockTransaction,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got, _ := tr.BuildModel(tt.args.txs, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockRowTransactionQueryScan struct {
		Executor
	}
)

func (*mockRowTransactionQueryScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockTransactionQuery.Fields).AddRow(
			-1273123123,
			-123123123123,
			1,
			mockTransaction.SenderAccountAddress,
			mockTransaction.RecipientAccountAddress,
			binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
			1,
			10000,
			make([]byte, 200),
			88,
			make([]byte, 88),
			make([]byte, 68),
			1,
			1,
			false,
		),
	)
	return db.QueryRow("")
}

func TestTransactionQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		tx  *model.Transaction
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockTransactionQuery),
			args: args{
				tx:  &model.Transaction{},
				row: (&mockRowTransactionQueryScan{}).ExecuteSelectRow("", ""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if err := tr.Scan(tt.args.tx, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("TransactionQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
