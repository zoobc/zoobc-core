// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
		Message:               []byte{1, 2, 3},
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
		"",
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
				"transaction_index, child_type, message from \"transaction\"" +
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
			wantStr: fmt.Sprintf("SELECT %s FROM \"transaction\" WHERE block_id = ? AND child_type = ?"+
				" ORDER BY transaction_index ASC",
				strings.Join(mockTransactionQuery.Fields, ", "),
			),
			wantArgs: []interface{}{int64(1), uint32(model.TransactionChildType_NoneChild)},
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
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, transaction_index, " +
				"child_type, message FROM \"transaction\" WHERE child_type = ? AND id IN(?, ?, ?, ?)",
			wantArgs: []interface{}{
				uint32(model.TransactionChildType_NoneChild),
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
			model.TransactionChildType_NoneChild,
			[]byte{1, 2, 3},
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
			model.TransactionChildType_MultiSignatureChild,
			[]byte{},
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
