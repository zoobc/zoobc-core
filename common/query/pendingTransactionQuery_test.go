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
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockPendingTransactionQueryInstance = NewPendingTransactionQuery()
)

func TestNewPendingTransactionQuery(t *testing.T) {
	tests := []struct {
		name string
		want *PendingTransactionQuery
	}{
		{
			name: "NewPendingTransactionQuery-Success",
			want: &PendingTransactionQuery{
				Fields: []string{
					"sender_address",
					"transaction_hash",
					"transaction_bytes",
					"status",
					"block_height",
					"latest",
				},
				TableName: "pending_transaction",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPendingTransactionQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPendingTransactionQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mock PendingTransactionQueryBuildModel
func getPendingTransactionQueryBuildModelFailRow() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"RandomField"})
	mockRow.AddRow(
		make([]byte, 32),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}
func getPendingTransactionQueryBuildModelSuccessRow() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingTransactionQueryInstance.Fields)
	mockRow.AddRow(
		nil,
		make([]byte, 32),
		make([]byte, 100),
		model.PendingTransactionStatus_PendingTransactionExecuted,
		uint32(10),
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

// mock PendingTransactionQueryBuildModel

func TestPendingTransactionQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pts  []*model.PendingTransaction
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PendingTransaction
		wantErr bool
	}{
		{
			name: "BuildModel-Fail",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pts:  []*model.PendingTransaction{},
				rows: getPendingTransactionQueryBuildModelFailRow(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pts:  []*model.PendingTransaction{},
				rows: getPendingTransactionQueryBuildModelSuccessRow(),
			},
			want: []*model.PendingTransaction{
				{
					SenderAddress:    nil,
					TransactionHash:  make([]byte, 32),
					TransactionBytes: make([]byte, 100),
					Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
					BlockHeight:      10,
					Latest:           true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := ptq.BuildModel(tt.args.pts, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockPendingTransactionExtractModel = &model.PendingTransaction{
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
		BlockHeight:      10,
	}
)

func TestPendingTransactionQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: mockPendingTransactionExtractModel,
			},
			want: []interface{}{
				&mockPendingTransactionExtractModel.SenderAddress,
				&mockPendingTransactionExtractModel.TransactionHash,
				&mockPendingTransactionExtractModel.TransactionBytes,
				&mockPendingTransactionExtractModel.Status,
				&mockPendingTransactionExtractModel.BlockHeight,
				&mockPendingTransactionExtractModel.Latest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := pe.ExtractModel(tt.args.pendingTx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_GetPendingTransactionByHash(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		txHash               []byte
		status               []model.PendingTransactionStatus
		currentHeight, limit uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetPendingTransactionByHash-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				txHash: make([]byte, 32),
				status: []model.PendingTransactionStatus{
					model.PendingTransactionStatus_PendingTransactionPending,
					model.PendingTransactionStatus_PendingTransactionExecuted,
				},
				currentHeight: 0,
				limit:         constant.MinRollbackBlocks,
			},
			wantStr: "SELECT sender_address, transaction_hash, transaction_bytes, status, block_height, latest FROM pending_transaction " +
				"WHERE transaction_hash = ? AND status IN (?, ?) AND block_height >= ? AND latest = true",
			wantArgs: []interface{}{
				make([]byte, 32),
				model.PendingTransactionStatus_PendingTransactionPending,
				model.PendingTransactionStatus_PendingTransactionExecuted,
				uint32(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ptq.GetPendingTransactionByHash(
				tt.args.txHash,
				tt.args.status,
				tt.args.currentHeight,
				tt.args.limit,
			)
			if gotStr != tt.wantStr {
				t.Errorf("GetPendingTransactionByHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetPendingTransactionByHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

var (
	mockInsertPendingTransaction = &model.PendingTransaction{
		SenderAddress:    nil,
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
		BlockHeight:      10,
	}
)

func TestPendingTransactionQuery_InsertPendingTransaction(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name: "InsertPendingTransaction-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: mockInsertPendingTransaction,
			},
			want: [][]interface{}{
				append([]interface{}{
					"INSERT OR REPLACE INTO pending_transaction (sender_address, transaction_hash, " +
						"transaction_bytes, status, block_height, latest) VALUES(? , ? , ? , ? , ? , ? )",
				}, mockPendingTransactionQueryInstance.ExtractModel(mockInsertPendingTransaction)...),
				{
					"UPDATE pending_transaction SET latest = false WHERE transaction_hash = ? AND block_height " +
						"!= 10 AND latest = true",
					mockInsertPendingTransaction.TransactionHash,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got := ptq.InsertPendingTransaction(tt.args.pendingTx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertPendingTransaction() gotArgs = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_Rollback(t *testing.T) {
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
			name: "Rollback",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				height: 10,
			},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM pending_transaction WHERE block_height > ?",
					uint32(10),
				},
				{
					"UPDATE pending_transaction SET latest = ? WHERE latest = ? AND (transaction_hash, " +
						"block_height) IN (SELECT t2.transaction_hash, MAX(t2.block_height) FROM pending_transaction as t2 GROUP BY t2." +
						"transaction_hash)",
					1, 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := ptq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

// mock PendingTransactionQuery Scan
func getPendingTransactionQueryScanFailRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}
func getPendingTransactionQueryScanSuccessRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockPendingTransactionQueryInstance.Fields)
	mockRow.AddRow(
		"",
		make([]byte, 32),
		make([]byte, 100),
		uint32(0),
		uint32(10),
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

// mock PendingTransactionQuery Scan

func TestPendingTransactionQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTx *model.PendingTransaction
		row       *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Scan-Fail",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: &model.PendingTransaction{},
				row:       getPendingTransactionQueryScanFailRow(),
			},
			wantErr: true,
		},
		{
			name: "Scan-Success",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				pendingTx: &model.PendingTransaction{},
				row:       getPendingTransactionQueryScanSuccessRow(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := pe.Scan(tt.args.pendingTx, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPendingTransactionQuery_getTableName(t *testing.T) {
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
			name: "getTableName",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			want: mockPendingTransactionQueryInstance.TableName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := ptq.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_SelectDataForSnapshot(t *testing.T) {
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
			name: "SelectDataForSnapshot",
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "SELECT sender_address,transaction_hash,transaction_bytes,status,block_height," +
				"latest FROM pending_transaction WHERE (transaction_hash, block_height) IN (SELECT t2.transaction_hash, " +
				"MAX(t2.block_height) FROM pending_transaction as t2 WHERE t2.block_height >= 1 AND t2.block_height <= 10 AND t2.block_height != 0 GROUP BY t2." +
				"transaction_hash) ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := ptq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PendingTransactionQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			fields: fields{
				Fields:    mockPendingTransactionQueryInstance.Fields,
				TableName: mockPendingTransactionQueryInstance.TableName,
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM pending_transaction WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := ptq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PendingTransactionQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingTransactionQuery_GetPendingTransactionsExpireByHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		currentHeight uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewPendingTransactionQuery()),
			args: args{
				currentHeight: 1000,
			},
			wantStr: "SELECT sender_address, transaction_hash, transaction_bytes, status, block_height, latest " +
				"FROM pending_transaction WHERE block_height = ? AND status = ? AND latest = ?",
			wantArgs: []interface{}{
				uint32(1720) - constant.MinRollbackBlocks,
				model.PendingTransactionStatus_PendingTransactionPending,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ptq.GetPendingTransactionsExpireByHeight(tt.args.currentHeight)
			if gotStr != tt.wantStr {
				t.Errorf("GetPendingTransactionsExpireByHeight() gotStr = %v, want %v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetPendingTransactionsExpireByHeight() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPendingTransactionQuery_InsertPendingTransactions(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		pendingTXs []*model.PendingTransaction
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewPendingTransactionQuery()),
			args: args{
				pendingTXs: []*model.PendingTransaction{
					mockInsertPendingTransaction,
				},
			},
			wantArgs: NewPendingTransactionQuery().ExtractModel(mockInsertPendingTransaction),
			wantStr: "INSERT OR REPLACE INTO pending_transaction (sender_address, transaction_hash, transaction_bytes, status, block_height, latest) " +
				"VALUES (?, ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptq := &PendingTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ptq.InsertPendingTransactions(tt.args.pendingTXs)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPendingTransactions() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPendingTransactions() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
