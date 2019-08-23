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
		ID:                      -1273123123,
		BlockID:                 -123123123123,
		Version:                 1,
		Height:                  1,
		SenderAccountAddress:    "senderAccountAddress",
		RecipientAccountAddress: "recipientAccountAddress",
		TransactionType:         binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
		Fee:                     1,
		Timestamp:               10000,
		TransactionHash:         make([]byte, 200),
		TransactionBodyLength:   88,
		TransactionBodyBytes:    make([]byte, 88),
		Signature:               make([]byte, 68),
		TransactionIndex:        1,
	}
	// mockTransactionRow represent a transaction row for test purpose only
	// copy just the values only,
	mockTransactionRow = []interface{}{
		-1273123123,
		-123123123123,
		1,
		"senderAccountAddress",
		"recipientAccountAddress",
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
			name:   "transaction query without condition",
			params: &paramsStruct{},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index from \"transaction\"",
		},
		{
			name: "transaction query with ID param only",
			params: &paramsStruct{
				ID: 1,
			},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index from \"transaction\" " +
				"WHERE id = 1",
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

func TestGetTransactions(t *testing.T) {
	transactionQuery := NewTransactionQuery(chaintype.GetChainType(0))

	type paramsStruct struct {
		Limit  uint32
		Offset uint64
	}

	tests := []struct {
		name   string
		params *paramsStruct
		want   string
	}{
		{
			name:   "transactions query without condition",
			params: &paramsStruct{},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version," +
				" transaction_index from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 0,10",
		},
		{
			name: "transactions query with limit",
			params: &paramsStruct{
				Limit: 10,
			},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 0,10",
		},
		{
			name: "transactions query with offset",
			params: &paramsStruct{
				Offset: 20,
			},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 20,10",
		},
		{
			name: "transactions query with all the params",
			params: &paramsStruct{
				Limit:  10,
				Offset: 20,
			},
			want: "SELECT id, block_id, block_height, sender_account_address, " +
				"recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
				"transaction_index from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 20,10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := transactionQuery.GetTransactions(tt.params.Limit, tt.params.Offset)
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
		name     string
		fields   fields
		args     args
		wantStr  []string
		wantArgs uint32
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockTransactionQuery),
			args:   args{height: uint32(1)},
			wantStr: []string{
				"DELETE FROM \"transaction\" WHERE block_height > 1",
			},
			wantArgs: uint32(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tq := &TransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := tq.Rollback(tt.args.height)
			if !reflect.DeepEqual(gotStr, tt.wantStr) {
				t.Errorf("Rollback() = %v, want %v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Rollback() = %v, want %v", gotArgs, tt.wantArgs)
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
			wantStr: fmt.Sprintf("SELECT %s FROM \"transaction\" WHERE block_id = ?",
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

type (
	mockQueryExecutorBuildMmodel struct {
		Executor
	}
)

func (*mockQueryExecutorBuildMmodel) ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockTransactionQuery.Fields).AddRow(
			-1273123123,
			-123123123123,
			1,
			"senderAccountAddress",
			"recipientAccountAddress",
			binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
			1,
			10000,
			make([]byte, 200),
			88,
			make([]byte, 88),
			make([]byte, 68),
			1,
			1,
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
	rows, _ := (&mockQueryExecutorBuildMmodel{}).ExecuteSelect("", "")
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
			if got := tr.BuildModel(tt.args.txs, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
