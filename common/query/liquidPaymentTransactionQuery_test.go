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
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	liquidPayTxAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	liquidPayTxAddress2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
)

func TestLiquidPaymentTransactionQuery_InsertLiquidPaymentTransaction(t *testing.T) {
	liquidPayment := &model.LiquidPayment{
		ID:               1,
		SenderAddress:    liquidPayTxAddress1,
		RecipientAddress: liquidPayTxAddress2,
		Amount:           123456,
		AppliedTime:      1231413,
		CompleteMinutes:  4343,
		Status:           model.LiquidPaymentStatus_LiquidPaymentPending,
		BlockHeight:      24,
	}

	type args struct {
		liquidPayment *model.LiquidPayment
	}
	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{
			name: "wantSuccess",
			args: args{
				liquidPayment: liquidPayment,
			},
			want: [][]interface{}{
				{
					"UPDATE liquid_payment_transaction set latest = ? WHERE id = ?",
					false,
					1,
				},
				append(
					[]interface{}{
						"INSERT INTO liquid_payment_transaction (id,sender_address,recipient_address,amount," +
							"applied_time,complete_minutes,status,block_height,latest) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?)",
					},
					[]interface{}{
						liquidPayment.GetID(),
						liquidPayment.GetSenderAddress(),
						liquidPayment.GetRecipientAddress(),
						liquidPayment.GetAmount(),
						liquidPayment.GetAppliedTime(),
						liquidPayment.GetCompleteMinutes(),
						liquidPayment.GetStatus(),
						liquidPayment.GetBlockHeight(),
						true,
					}...,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			if got := lpt.InsertLiquidPaymentTransaction(tt.args.liquidPayment); fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("LiquidPaymentTransactionQuery.InsertLiquidPaymentTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_CompleteLiquidPaymentTransaction(t *testing.T) {
	type args struct {
		id           int64
		causedFields map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{
			name: "wantSuccess",
			args: args{
				id: 1234,
				causedFields: map[string]interface{}{
					"block_height": 123,
				},
			},
			want: [][]interface{}{
				{
					"INSERT INTO liquid_payment_transaction (id, sender_address, recipient_address, amount, applied_time, complete_minutes, status," +
						" block_height, latest) SELECT id, sender_address, recipient_address, amount, applied_time, complete_minutes, ?, 123, true FROM" +
						" liquid_payment_transaction WHERE id = 1234 AND latest = 1 ON CONFLICT(id, block_height) DO UPDATE SET status = ?",
					model.LiquidPaymentStatus_LiquidPaymentCompleted,
					model.LiquidPaymentStatus_LiquidPaymentCompleted,
				},
				{
					"UPDATE liquid_payment_transaction set latest = ? WHERE id = ? AND block_height != 123 and latest = true",
					false,
					1234,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			if got := lpt.CompleteLiquidPaymentTransaction(tt.args.id, tt.args.causedFields); fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("LiquidPaymentTransactionQuery.CompleteLiquidPaymentTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_GetPendingLiquidPaymentTransactionByID(t *testing.T) {
	type args struct {
		id     int64
		status model.LiquidPaymentStatus
	}
	tests := []struct {
		name     string
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "wantSuccess",
			args: args{
				id:     123,
				status: model.LiquidPaymentStatus_LiquidPaymentPending,
			},
			wantStr: "SELECT id, sender_address, recipient_address, amount, applied_time, complete_minutes, status," +
				" block_height, latest FROM liquid_payment_transaction WHERE id = ? AND status = ? AND latest = ?",
			wantArgs: []interface{}{123, model.LiquidPaymentStatus_LiquidPaymentPending, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			gotStr, gotArgs := lpt.GetPendingLiquidPaymentTransactionByID(tt.args.id, tt.args.status)
			if gotStr != tt.wantStr {
				t.Errorf("LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if fmt.Sprintf("%v", gotArgs) != fmt.Sprintf("%v", tt.wantArgs) {
				t.Errorf("LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_GetPassedTimePendingLiquidPaymentTransactions(t *testing.T) {
	type args struct {
		timestamp int64
	}
	tests := []struct {
		name     string
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name: "wantSuccess",
			args: args{
				timestamp: 123141,
			},
			wantQStr: "SELECT id, sender_address, recipient_address, amount, applied_time, complete_minutes, status," +
				" block_height, latest FROM liquid_payment_transaction WHERE applied_time+(complete_minutes*60) <= ? AND status = ? AND latest = ?",
			wantArgs: []interface{}{123141, model.LiquidPaymentStatus_LiquidPaymentPending, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			gotQStr, gotArgs := lpt.GetPassedTimePendingLiquidPaymentTransactions(tt.args.timestamp)
			if gotQStr != tt.wantQStr {
				t.Errorf("LiquidPaymentTransactionQuery.GetPassedTimePendingLiquidPaymentTransactions() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if fmt.Sprintf("%v", gotArgs) != fmt.Sprintf("%v", tt.wantArgs) {
				t.Errorf("LiquidPaymentTransactionQuery.GetPassedTimePendingLiquidPaymentTransactions() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_ExtractModel(t *testing.T) {
	type args struct {
		liquidPayment *model.LiquidPayment
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{name: "wantSuccess",
			args: args{
				liquidPayment: &model.LiquidPayment{
					ID:               123,
					SenderAddress:    liquidPayTxAddress1,
					RecipientAddress: liquidPayTxAddress2,
					Amount:           1234,
					AppliedTime:      12345,
					CompleteMinutes:  123456,
					Status:           1234567,
					BlockHeight:      12345678,
					Latest:           true,
				},
			},
			want: []interface{}{123,
				liquidPayTxAddress1,
				liquidPayTxAddress2,
				1234,
				12345,
				123456,
				1234567,
				12345678,
				true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			if got := lpt.ExtractModel(tt.args.liquidPayment); fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("LiquidPaymentTransactionQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_BuildModels(t *testing.T) {
	mockLiquidPaymentTransaction := NewLiquidPaymentTransactionQuery()
	mockLiquidPayment := &model.LiquidPayment{
		ID:               123,
		SenderAddress:    liquidPayTxAddress1,
		RecipientAddress: liquidPayTxAddress2,
		Amount:           1234,
		AppliedTime:      12345,
		CompleteMinutes:  123456,
		Status:           1234567,
		BlockHeight:      12345678,
		Latest:           true,
	}
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockLiquidPaymentTransaction.Fields)
	mockRow.AddRow(
		mockLiquidPayment.GetID(),
		mockLiquidPayment.GetSenderAddress(),
		mockLiquidPayment.GetRecipientAddress(),
		mockLiquidPayment.GetAmount(),
		mockLiquidPayment.GetAppliedTime(),
		mockLiquidPayment.GetCompleteMinutes(),
		mockLiquidPayment.GetStatus(),
		mockLiquidPayment.GetBlockHeight(),
		mockLiquidPayment.GetLatest(),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow, _ := db.Query("")

	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		args    args
		want    []*model.LiquidPayment
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{rows: mockedRow},
			want: []*model.LiquidPayment{
				mockLiquidPayment,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := mockLiquidPaymentTransaction
			got, err := lpt.BuildModels(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentTransactionQuery.BuildModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPaymentTransactionQuery.BuildModels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_Scan(t *testing.T) {
	mockLiquidPaymentTransaction := NewLiquidPaymentTransactionQuery()
	mockLiquidPayment := &model.LiquidPayment{
		ID:               123,
		SenderAddress:    liquidPayTxAddress1,
		RecipientAddress: liquidPayTxAddress2,
		Amount:           1234,
		AppliedTime:      12345,
		CompleteMinutes:  123456,
		Status:           1234567,
		BlockHeight:      12345678,
		Latest:           true,
	}
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockLiquidPaymentTransaction.Fields)
	mockRow.AddRow(
		mockLiquidPayment.GetID(),
		mockLiquidPayment.GetSenderAddress(),
		mockLiquidPayment.GetRecipientAddress(),
		mockLiquidPayment.GetAmount(),
		mockLiquidPayment.GetAppliedTime(),
		mockLiquidPayment.GetCompleteMinutes(),
		mockLiquidPayment.GetStatus(),
		mockLiquidPayment.GetBlockHeight(),
		mockLiquidPayment.GetLatest(),
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow := db.QueryRow("")
	type args struct {
		liquidPayment *model.LiquidPayment
		row           *sql.Row
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				liquidPayment: mockLiquidPayment,
				row:           mockedRow,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := mockLiquidPaymentTransaction
			if err := lpt.Scan(tt.args.liquidPayment, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentTransactionQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_Rollback(t *testing.T) {
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name: "wantSuccess",
			args: args{height: 30},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM liquid_payment_transaction WHERE block_height > ?",
					uint32(30),
				},
				{
					`
			UPDATE liquid_payment_transaction SET latest = ?
			WHERE latest = ? AND (id, block_height) IN (
				SELECT t2.id, MAX(t2.block_height)
				FROM liquid_payment_transaction as t2
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
			lpt := NewLiquidPaymentTransactionQuery()
			if gotMultiQueries := lpt.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("LiquidPaymentTransactionQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_SelectDataForSnapshot(t *testing.T) {
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "wantSuccess",
			args: args{
				fromHeight: 720,
				toHeight:   1440,
			},
			want: "SELECT id,sender_address,recipient_address,amount,applied_time,complete_minutes,status," +
				"block_height,latest FROM liquid_payment_transaction WHERE (id, block_height) IN (SELECT t2.id, MAX(" +
				"t2.block_height) FROM liquid_payment_transaction as t2 WHERE t2.block_height >= 720" +
				" AND t2.block_height <= 1440 AND t2.block_height != 0 GROUP BY t2.id) ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			if got := lpt.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("LiquidPaymentTransactionQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentTransactionQuery_TrimDataBeforeSnapshot(t *testing.T) {
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM liquid_payment_transaction WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lpt := NewLiquidPaymentTransactionQuery()
			if got := lpt.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("LiquidPaymentTransactionQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
