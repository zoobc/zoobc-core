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
package transaction

import (
	"database/sql"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	approvalEscrowAccountAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
		239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	approvalEscrowAccountAddress2 = []byte{0, 0, 0, 0, 33, 130, 42, 143, 177, 97, 43, 208, 76, 119, 240, 91, 41, 170, 240, 161, 55, 224,
		8, 205, 139, 227, 189, 146, 86, 211, 52, 194, 131, 126, 233, 100}
	approvalEscrowAccountAddress3 = []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
)

func TestApprovalEscrowTransaction_GetBodyBytes(t *testing.T) {
	type fields struct {
		ID                 int64
		Fee                int64
		SenderAddress      []byte
		Height             uint32
		Body               *model.ApprovalEscrowTransactionBody
		Escrow             *model.Escrow
		QueryExecutor      query.ExecutorInterface
		EscrowQuery        query.EscrowTransactionQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		TypeActionSwitcher TypeActionSwitcher
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: nil,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 120978123123,
				},
			},
			want: []byte{1, 0, 0, 0, 115, 169, 219, 42, 28, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                 tt.fields.ID,
				Fee:                tt.fields.Fee,
				SenderAddress:      tt.fields.SenderAddress,
				Height:             tt.fields.Height,
				Body:               tt.fields.Body,
				Escrow:             tt.fields.Escrow,
				QueryExecutor:      tt.fields.QueryExecutor,
				EscrowQuery:        tt.fields.EscrowQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				TypeActionSwitcher: tt.fields.TypeActionSwitcher,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApprovalEscrowTransaction_ParseBodyBytes(t *testing.T) {
	type fields struct {
		ID                 int64
		Fee                int64
		SenderAddress      []byte
		Height             uint32
		Body               *model.ApprovalEscrowTransactionBody
		Escrow             *model.Escrow
		QueryExecutor      query.ExecutorInterface
		EscrowQuery        query.EscrowTransactionQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		TypeActionSwitcher TypeActionSwitcher
	}
	type args struct {
		bodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: nil,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 120978123123,
				},
			},
			args: args{bodyBytes: []byte{1, 0, 0, 0, 115, 169, 219, 42, 28, 0, 0, 0}},
			want: &model.ApprovalEscrowTransactionBody{
				Approval:      1,
				TransactionID: 120978123123,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                 tt.fields.ID,
				Fee:                tt.fields.Fee,
				SenderAddress:      tt.fields.SenderAddress,
				Height:             tt.fields.Height,
				Body:               tt.fields.Body,
				Escrow:             tt.fields.Escrow,
				QueryExecutor:      tt.fields.QueryExecutor,
				EscrowQuery:        tt.fields.EscrowQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				TypeActionSwitcher: tt.fields.TypeActionSwitcher,
			}
			got, err := tx.ParseBodyBytes(tt.args.bodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBodyBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorValidate struct {
		query.Executor
	}
	mockQueryExecutorValidateNotFound struct {
		query.Executor
	}
	mockAccountBalanceQueryValidateNotFound struct {
		query.AccountBalanceQuery
	}
	mockAccountBalanceQueryValidateFound struct {
		query.AccountBalanceQuery
	}
)

func (*mockQueryExecutorValidate) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mockRow.AddRow(
		120978123123,
		approvalEscrowAccountAddress1,
		approvalEscrowAccountAddress2,
		approvalEscrowAccountAddress3,
		1,
		10,
		100,
		0,
		1,
		true,
		"",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow := db.QueryRow("")
	return mockedRow, nil
}
func (*mockQueryExecutorValidateNotFound) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	mockedRow := db.QueryRow("")
	return mockedRow, nil
}

func (*mockAccountBalanceQueryValidateNotFound) GetAccountBalanceByAccountAddress(sender string) (qStr string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockAccountBalanceQueryValidateNotFound) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	return sql.ErrNoRows
}
func (*mockAccountBalanceQueryValidateFound) GetAccountBalanceByAccountAddress(sender string) (qStr string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockAccountBalanceQueryValidateFound) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	accountBalance.AccountAddress = approvalEscrowAccountAddress1
	accountBalance.Balance = 1000
	accountBalance.Latest = true

	return nil
}

type (
	mockAccountApprovalEscrowTransactionAccountBalanceHelperAccountBalanceNotFound struct {
		AccountBalanceHelper
	}
	mockAccountBalanceApprovalEscrowTransactionAccountBalanceHelperWantSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountApprovalEscrowTransactionAccountBalanceHelperAccountBalanceNotFound) HasEnoughSpendableBalance(
	dbTX bool,
	address []byte,
	compareBalance int64) (enough bool, err error) {
	return false, sql.ErrNoRows
}
func (*mockAccountBalanceApprovalEscrowTransactionAccountBalanceHelperWantSuccess) HasEnoughSpendableBalance(
	dbTX bool,
	address []byte,
	compareBalance int64) (enough bool, err error) {
	return true, nil
}
func TestApprovalEscrowTransaction_Validate(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		Height               uint32
		Body                 *model.ApprovalEscrowTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	type args struct {
		dbTx bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantError:NotFound",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: approvalEscrowAccountAddress1,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 120978123123,
				},
				QueryExecutor: &mockQueryExecutorValidateNotFound{},
				EscrowQuery:   query.NewEscrowTransactionQuery(),
			},
			args:    args{dbTx: false},
			wantErr: true,
		},
		{
			name: "wantError:InvalidTransactionID",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: approvalEscrowAccountAddress1,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 0,
				},
				QueryExecutor:    &mockQueryExecutorValidate{},
				EscrowQuery:      query.NewEscrowTransactionQuery(),
				TransactionQuery: nil,
			},
			args:    args{dbTx: false},
			wantErr: true,
		},
		{
			name: "wantError:AccountBalanceNotFound",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: approvalEscrowAccountAddress2,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 120978123123,
				},
				QueryExecutor:        &mockQueryExecutorValidate{},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: &mockAccountApprovalEscrowTransactionAccountBalanceHelperAccountBalanceNotFound{},
			},
			args:    args{dbTx: false},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				ID:            0,
				Fee:           0,
				SenderAddress: approvalEscrowAccountAddress3,
				Height:        0,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 120978123123,
				},
				QueryExecutor:        &mockQueryExecutorValidate{},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				AccountBalanceHelper: &mockAccountBalanceApprovalEscrowTransactionAccountBalanceHelperWantSuccess{},
			},
			args: args{dbTx: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				TransactionQuery:     tt.fields.TransactionQuery,
				TypeActionSwitcher:   tt.fields.TypeActionSwitcher,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorUnconfirmed struct {
		query.Executor
	}
)

func (*mockQueryExecutorUnconfirmed) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

type (
	mockAccountBalanceHelperApprovalEscrowTransactionApplyUnconfirmed struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperApprovalEscrowTransactionApplyUnconfirmed) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}
func TestApprovalEscrowTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		Height               uint32
		Body                 *model.ApprovalEscrowTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:                   0,
				Fee:                  1,
				SenderAddress:        nil,
				Height:               0,
				Body:                 nil,
				Escrow:               nil,
				QueryExecutor:        &mockQueryExecutorUnconfirmed{},
				EscrowQuery:          nil,
				TransactionQuery:     nil,
				AccountBalanceHelper: &mockAccountBalanceHelperApprovalEscrowTransactionApplyUnconfirmed{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				TransactionQuery:     tt.fields.TransactionQuery,
				TypeActionSwitcher:   tt.fields.TypeActionSwitcher,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperApprovalEscrowTransactionUndoApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperApprovalEscrowTransactionUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}

func TestApprovalEscrowTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		Height               uint32
		Body                 *model.ApprovalEscrowTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:                   0,
				Fee:                  1,
				SenderAddress:        nil,
				Height:               0,
				Body:                 nil,
				Escrow:               nil,
				QueryExecutor:        &mockQueryExecutorUnconfirmed{},
				EscrowQuery:          nil,
				TransactionQuery:     nil,
				AccountBalanceHelper: &mockAccountBalanceHelperApprovalEscrowTransactionUndoApplyUnconfirmedSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				TransactionQuery:     tt.fields.TransactionQuery,
				TypeActionSwitcher:   tt.fields.TypeActionSwitcher,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockEscrowQueryApplyConfirmedOK struct {
		query.EscrowTransactionQuery
	}
	mockTransactionQueryApplyConfirmedOK struct {
		query.TransactionQuery
	}
	mockEscrowQueryExecutorApplyConfirmedOK struct {
		query.Executor
	}
)

func (*mockEscrowQueryExecutorApplyConfirmedOK) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, nil
}
func (*mockEscrowQueryExecutorApplyConfirmedOK) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockEscrowQueryApplyConfirmedOK) GetLatestEscrowTransactionByID(int64) (qStr string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockTransactionQueryApplyConfirmedOK) GetTransaction(int64) string {
	return ""
}
func (*mockTransactionQueryApplyConfirmedOK) Scan(tx *model.Transaction, row *sql.Row) error {
	tx.ID = -1273123123
	tx.BlockID = -123123123123
	tx.Version = 1
	tx.Height = 1
	tx.SenderAccountAddress = approvalEscrowAccountAddress1
	tx.RecipientAccountAddress = approvalEscrowAccountAddress2
	tx.TransactionType = binary.LittleEndian.Uint32([]byte{4, 0, 0, 0})
	tx.Fee = 1
	tx.Timestamp = 10000
	tx.TransactionHash = make([]byte, 200)
	tx.TransactionBodyLength = 88
	tx.TransactionBodyBytes = make([]byte, 88)
	tx.Signature = make([]byte, 68)
	tx.TransactionIndex = 1

	return nil
}
func (*mockEscrowQueryApplyConfirmedOK) Scan(escrow *model.Escrow, _ *sql.Row) error {
	escrow.ID = 1
	escrow.SenderAddress = approvalEscrowAccountAddress1
	escrow.RecipientAddress = approvalEscrowAccountAddress2
	escrow.ApproverAddress = approvalEscrowAccountAddress3
	escrow.Amount = 10
	escrow.Commission = 1
	escrow.Timeout = 120
	escrow.Status = 1
	escrow.BlockHeight = 0
	escrow.Latest = true
	return nil
}

type (
	mockAccountBalanceHelperApprovalEscrowTransactionApplyConfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperApprovalEscrowTransactionApplyConfirmedSuccess) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return nil
}

func TestApprovalEscrowTransaction_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		Height               uint32
		Body                 *model.ApprovalEscrowTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:            1234567890,
				Fee:           1,
				SenderAddress: approvalEscrowAccountAddress3,
				Height:        1,
				Body: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 1234567890,
				},
				EscrowQuery:      &mockEscrowQueryApplyConfirmedOK{},
				QueryExecutor:    &mockEscrowQueryExecutorApplyConfirmedOK{},
				TransactionQuery: &mockTransactionQueryApplyConfirmedOK{},
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &mockEscrowQueryExecutorApplyConfirmedOK{},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperApprovalEscrowTransactionApplyConfirmedSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ApprovalEscrowTransaction{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				TransactionQuery:     tt.fields.TransactionQuery,
				TypeActionSwitcher:   tt.fields.TypeActionSwitcher,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
