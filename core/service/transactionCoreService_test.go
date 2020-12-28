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
package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	// GetTransactionsByIds mocks
	mockGetTransactionsByIdsExecutorFail struct {
		query.Executor
	}
	mockGetTransactionsByIdsExecutorSuccess struct {
		query.Executor
	}
	mockGetTransactionsByIdsTransactionQueryBuildFail struct {
		query.TransactionQuery
	}
	mockGetTransactionsByIdsTransactionQueryBuildSuccess struct {
		query.TransactionQuery
	}
	mockGetTransactionsByIdsExecutorSelectWithEscrowSuccess struct {
		query.Executor
	}
	// GetTransactionsByIds mocks
	// GetTransactionsByBlockID mocks
	mockGetTransactionsByBlockIDExecutorFail struct {
		query.Executor
	}
	mockGetTransactionsByBlockIDExecutorSuccess struct {
		query.Executor
	}
	mockGetTransactionsByBlockIDTransactionQueryBuildFail struct {
		query.TransactionQuery
	}
	mockGetTransactionsByBlockIDTransactionQueryBuildSuccess struct {
		query.TransactionQuery
	}
	mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessOne struct {
		query.EscrowTransactionQuery
	}
	mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessEmpty struct {
		query.EscrowTransactionQuery
	}
	// GetTransactionsByBlockID mocks
)

var (
	// GetTransactionByIds mocks
	mockGetTransactionByIdsResult = []*model.Transaction{
		{
			TransactionHash: make([]byte, 32),
		},
	}
	mockGetTransactionsByBlockIDResult = []*model.Transaction{
		{
			TransactionHash: make([]byte, 32),
		},
	}
	mockGetTransactionsByBlockIDResultWithEscrow = []*model.Transaction{
		{
			TransactionHash: make([]byte, 32),
			Escrow:          mockGetTransactionByBlockIDEscrowTransactionResultOne[0],
		},
	}
	mockGetTransactionByBlockIDEscrowTransactionResultOne = []*model.Escrow{
		{
			ID: 0,
		},
	}
	mockGetTransactionByBlockIDEscrowTransactionResultEmpty = make([]*model.Escrow, 0)

	address1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	address2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	address3 = []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75, 213, 137, 66, 236, 188, 43,
		59, 241, 146, 243, 147, 58, 161, 35, 229, 54}
)

func (*mockGetTransactionsByIdsExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByIdsExecutorSuccess) ExecuteSelect(q string, _ bool, _ ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch {
	case strings.Contains(q, "FROM \"transaction\""):
		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(sqlmock.NewRows([]string{
			"dummyColumn"}).AddRow(
			[]byte{1}))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(mock.NewRows(query.NewEscrowTransactionQuery().Fields))
	}
	rows, _ := db.Query(q)
	return rows, nil
}

func (*mockGetTransactionsByIdsTransactionQueryBuildFail) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByIdsTransactionQueryBuildSuccess) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return mockGetTransactionByIdsResult, nil
}

func (*mockGetTransactionsByIdsExecutorSelectWithEscrowSuccess) ExecuteSelect(q string, _ bool, _ ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockedTX := transaction.GetFixturesForTransaction(
		12345678,
		address1,
		address2,
		true)
	switch {
	case strings.Contains(q, "FROM \"transaction\""):
		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(mock.NewRows(query.NewTransactionQuery(chaintype.GetChainType(0)).Fields).AddRow(
			mockedTX.GetID(),
			mockedTX.GetBlockID(),
			mockedTX.GetHeight(),
			mockedTX.GetSenderAccountAddress(),
			mockedTX.GetRecipientAccountAddress(),
			mockedTX.GetTransactionType(),
			mockedTX.GetFee(),
			mockedTX.GetTimestamp(),
			mockedTX.GetTransactionHash(),
			mockedTX.GetTransactionBodyLength(),
			mockedTX.GetTransactionBodyBytes(),
			mockedTX.GetSignature(),
			mockedTX.GetVersion(),
			mockedTX.GetTransactionIndex(),
			mockedTX.GetMultisigChild(),
			mockedTX.GetMessage(),
		))
	default:
		mockedEscrow := mockedTX.GetEscrow()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(mock.NewRows(query.NewEscrowTransactionQuery().Fields).AddRow(
			mockedEscrow.GetID(),
			mockedEscrow.GetSenderAddress(),
			mockedEscrow.GetRecipientAddress(),
			mockedEscrow.GetApproverAddress(),
			mockedEscrow.GetAmount(),
			mockedEscrow.GetCommission(),
			mockedEscrow.GetTimeout(),
			mockedEscrow.GetStatus(),
			mockedEscrow.GetBlockHeight(),
			mockedEscrow.GetLatest(),
			mockedEscrow.GetInstruction(),
		))
	}
	rows, _ := db.Query(q)
	return rows, nil
}

func (*mockGetTransactionsByBlockIDExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByBlockIDExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetTransactionsByBlockIDTransactionQueryBuildFail) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessOne) BuildModels(
	rows *sql.Rows) ([]*model.Escrow, error) {
	return mockGetTransactionByBlockIDEscrowTransactionResultOne, nil
}

func (*mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessEmpty) BuildModels(
	rows *sql.Rows) ([]*model.Escrow, error) {
	return mockGetTransactionByBlockIDEscrowTransactionResultEmpty, nil
}

func (*mockGetTransactionsByBlockIDTransactionQueryBuildSuccess) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return mockGetTransactionsByBlockIDResult, nil
}

func TestTransactionCoreService_GetTransactionsByIds(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		transactionIds []int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransactionByIds-ExecuteSelect-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByIdsTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByIdsExecutorFail{},
			},
			args: args{
				transactionIds: []int64{1, 2, 3},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionByIds-BuildModel-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByIdsTransactionQueryBuildFail{},
				QueryExecutor:    &mockGetTransactionsByIdsExecutorSuccess{},
			},
			args: args{
				transactionIds: []int64{1, 2, 3},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionsByIds-Escrow",
			fields: fields{
				TransactionQuery:       query.NewTransactionQuery(chaintype.GetChainType(0)),
				QueryExecutor:          &mockGetTransactionsByIdsExecutorSelectWithEscrowSuccess{},
				EscrowTransactionQuery: query.NewEscrowTransactionQuery(),
			},
			args: args{
				transactionIds: []int64{1},
			},
			want: []*model.Transaction{
				transaction.GetFixturesForTransaction(
					12345678,
					address1,
					address2,
					true),
			},
		},
		{
			name: "GetTransactionByIds-BuildModel-Success",
			fields: fields{
				TransactionQuery:       &mockGetTransactionsByIdsTransactionQueryBuildSuccess{},
				QueryExecutor:          &mockGetTransactionsByIdsExecutorSuccess{},
				EscrowTransactionQuery: query.NewEscrowTransactionQuery(),
			},
			args: args{
				transactionIds: []int64{1},
			},
			want:    mockGetTransactionByIdsResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
			}
			got, err := tg.GetTransactionsByIds(tt.args.transactionIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionsByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionsByIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionCoreService_GetTransactionsByBlockID(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		blockID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransactionsByBlockID-ExecuteSelect-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByBlockIDTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByBlockIDExecutorFail{},
			},
			args: args{
				blockID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionsByBlockID-BuildModel-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByBlockIDTransactionQueryBuildFail{},
				QueryExecutor:    &mockGetTransactionsByBlockIDExecutorSuccess{},
			},
			args: args{
				blockID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionsByBlockID-BuildModel-Success-EscrowOneResult",
			fields: fields{
				TransactionQuery:       &mockGetTransactionsByBlockIDTransactionQueryBuildSuccess{},
				EscrowTransactionQuery: &mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessOne{},
				QueryExecutor:          &mockGetTransactionsByBlockIDExecutorSuccess{},
			},
			args: args{
				blockID: 1,
			},
			want:    mockGetTransactionsByBlockIDResultWithEscrow,
			wantErr: false,
		},
		{
			name: "GetTransactionsByBlockID-BuildModel-Success-EscrowEmptyResult",
			fields: fields{
				TransactionQuery:       &mockGetTransactionsByBlockIDTransactionQueryBuildSuccess{},
				EscrowTransactionQuery: &mockGetTransactionsByBlockIDEscrowTransactionQueryBuildSuccessEmpty{},
				QueryExecutor:          &mockGetTransactionsByBlockIDExecutorSuccess{},
			},
			args: args{
				blockID: 1,
			},
			want:    mockGetTransactionsByBlockIDResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			got, err := tg.GetTransactionsByBlockID(tt.args.blockID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionsByBlockID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionsByBlockID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorExpiringEscrowSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockQueryExecutorExpiringEscrowSuccess) BeginTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) CommitTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) RollbackTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mockRows := sqlmock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mockRows.AddRow(
		int64(1),
		address1,
		address2,
		address3,
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRows)
	return db.Query(qStr)
}
func (*mockQueryExecutorExpiringEscrowSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {

	db, mock, _ := sqlmock.New()
	mockedRows := sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields)
	tx, _ := transaction.GetFixtureForSpecificTransaction(
		1234567890,
		12345678901,
		address1,
		address2,
		8,
		model.TransactionType_SendMoneyTransaction,
		&model.SendMoneyTransactionBody{
			Amount: 10,
		},
		true,
		true,
	)
	mockedRows.AddRow(
		tx.GetID(),
		tx.GetBlockID(),
		tx.GetHeight(),
		tx.GetSenderAccountAddress(),
		tx.GetRecipientAccountAddress(),
		tx.GetTransactionType(),
		tx.GetFee(),
		tx.GetTimestamp(),
		tx.GetTransactionHash(),
		tx.GetTransactionBodyLength(),
		tx.GetTransactionBodyBytes(),
		tx.GetSignature(),
		tx.GetVersion(),
		tx.GetTransactionIndex(),
		tx.GetMultisigChild(),
		tx.GetMessage(),
	)
	mock.ExpectQuery(qStr).WillReturnRows(mockedRows)
	return db.QueryRow(qStr), nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestTransactionCoreService_ExpiringEscrowTransactions(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
		TypeActionSwitcher     transaction.TypeActionSwitcher
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				EscrowTransactionQuery: query.NewEscrowTransactionQuery(),
				QueryExecutor:          &mockQueryExecutorExpiringEscrowSuccess{},
				TypeActionSwitcher:     &transaction.TypeSwitcher{Executor: &mockQueryExecutorExpiringEscrowSuccess{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
				TypeActionSwitcher:     tt.fields.TypeActionSwitcher,
			}
			if err := tg.ExpiringEscrowTransactions(tt.args.blockHeight, 100, false); (err != nil) != tt.wantErr {
				t.Errorf("ExpiringEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockUndoApplyUnconfirmedTransactionEscrowFalse struct {
	transaction.TXEmpty
}

func (*mockUndoApplyUnconfirmedTransactionEscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockUndoApplyUnconfirmedTransactionEscrowFalse) UndoApplyUnconfirmed() error {
	return nil
}

type mockUndoApplyUnconfirmedTransactionEscrowUndoApplyUnconfirmed struct {
	transaction.NodeRegistration
}

func (*mockUndoApplyUnconfirmedTransactionEscrowUndoApplyUnconfirmed) EscrowUndoApplyUnconfirmed() error {
	return nil
}

type mockUndoApplyUnconfirmedTransactionEscrowTrue struct {
	transaction.TXEmpty
}

func (*mockUndoApplyUnconfirmedTransactionEscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockUndoApplyUnconfirmedTransactionEscrowUndoApplyUnconfirmed{}, true
}

func TestTransactionCoreService_UndoApplyUnconfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "UndoApplyUnconfirmedTransaction:EscrowFalse",
			args: args{
				txAction: &mockUndoApplyUnconfirmedTransactionEscrowFalse{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockUndoApplyUnconfirmedTransactionEscrowTrue{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.UndoApplyUnconfirmedTransaction(tt.args.txAction); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.UndoApplyUnconfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockApplyConfirmedTransactionEscrowFalse struct {
	transaction.TXEmpty
}

func (*mockApplyConfirmedTransactionEscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockApplyConfirmedTransactionEscrowFalse) ApplyConfirmed(blockTimestamp int64) error {
	return nil
}

type mockApplyConfirmedTransactionEscrowApplyConfirmed struct {
	transaction.EscrowTypeAction
}

func (*mockApplyConfirmedTransactionEscrowApplyConfirmed) EscrowApplyConfirmed(blockTimestamp int64) error {
	return nil
}

type mockApplyConfirmedTransactionEscrowTrue struct {
	transaction.TXEmpty
}

func (*mockApplyConfirmedTransactionEscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockApplyConfirmedTransactionEscrowApplyConfirmed{}, true
}

func TestTransactionCoreService_ApplyConfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction       transaction.TypeAction
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ApplyConfirmedTransaction:EscrowFalse",
			args: args{
				txAction:       &mockApplyConfirmedTransactionEscrowFalse{},
				blockTimestamp: 0,
			},
			wantErr: false,
		},
		{
			name: "ApplyConfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockApplyConfirmedTransactionEscrowTrue{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ApplyConfirmedTransaction(tt.args.txAction, tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ApplyConfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockApplyUnconfirmedTransactionEscrowApplyUnconfirmed struct {
	transaction.EscrowTypeAction
}

func (*mockApplyUnconfirmedTransactionEscrowApplyUnconfirmed) EscrowApplyUnconfirmed() error {
	return nil
}

type mockApplyUnconfirmedTransactionEscrowTrue struct {
	transaction.TypeAction
}

func (*mockApplyUnconfirmedTransactionEscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockApplyUnconfirmedTransactionEscrowApplyUnconfirmed{}, true
}

type mockApplyUnconfirmedTransactionEscrowFalse struct {
	transaction.TypeAction
}

func (*mockApplyUnconfirmedTransactionEscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockApplyUnconfirmedTransactionEscrowFalse) ApplyUnconfirmed() error {
	return nil
}

func TestTransactionCoreService_ApplyUnconfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ApplyUnconfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockApplyUnconfirmedTransactionEscrowTrue{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmedTransaction:EscrowFalse",
			args: args{
				txAction: &mockApplyUnconfirmedTransactionEscrowFalse{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ApplyUnconfirmedTransaction(tt.args.txAction); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ApplyUnconfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockValidateTransactionEscrowValidate struct {
	transaction.EscrowTypeAction
}

func (*mockValidateTransactionEscrowValidate) EscrowValidate(dbTx bool) error {
	return nil
}

type mockValidateTransactionEscrowableTrue struct {
	transaction.TypeAction
}

func (*mockValidateTransactionEscrowableTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockValidateTransactionEscrowValidate{}, true
}

type mockValidateTransactionEscrowableFalse struct {
	transaction.TypeAction
}

func (*mockValidateTransactionEscrowableFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}

func (*mockValidateTransactionEscrowableFalse) Validate(dbTx bool) error {
	return nil
}

func TestTransactionCoreService_ValidateTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
		useTX    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ValidateTransaction:EscrowableTrue",
			args: args{
				txAction: &mockValidateTransactionEscrowableTrue{},
			},
			wantErr: false,
		},
		{
			name: "ValidateTransaction:EscrowableFalse",
			args: args{
				txAction: &mockValidateTransactionEscrowableFalse{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ValidateTransaction(tt.args.txAction, tt.args.useTX); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockCompletePassedLiquidPaymentExecutor struct {
		isExecuteSelectError    bool
		isExecuteSelectRowError bool
		query.ExecutorInterface
	}

	mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery struct {
		isBuildModelsError bool
		returnModels       []*model.LiquidPayment
		query.LiquidPaymentTransactionQuery
	}

	mockCompletePassedLiquidPaymentTransactionQuery struct {
		isScanError bool
		query.TransactionQuery
	}

	mockTypeActionSwitcher struct {
		isError  bool
		returnTx transaction.TypeAction
		transaction.TypeActionSwitcher
	}

	mockLiquidPaymentTransaction struct {
		isError bool
		transaction.LiquidPaymentTransaction
	}
)

func (m *mockCompletePassedLiquidPaymentExecutor) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	if m.isExecuteSelectError {
		return nil, errors.New("mockError ExecuteSelect")
	}
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(query.NewLiquidPaymentTransactionQuery().Fields)
	mock.ExpectQuery("").WillReturnRows(mockRows)

	return db.Query("")
}

func (m *mockCompletePassedLiquidPaymentExecutor) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	if m.isExecuteSelectRowError {
		return nil, errors.New("mockError ExecuteSelectRow")
	}
	return nil, nil
}

func (m *mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery) BuildModels(*sql.Rows) ([]*model.LiquidPayment, error) {
	if m.isBuildModelsError {
		return nil, errors.New("mockError BuildModels")
	}
	return m.returnModels, nil
}

func (m *mockCompletePassedLiquidPaymentTransactionQuery) Scan(tx *model.Transaction, row *sql.Row) error {
	if m.isScanError {
		return errors.New("mockError Scan")
	}
	return nil
}

func (m *mockTypeActionSwitcher) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	if m.isError {
		return nil, errors.New("mock error GetTransactionType")
	}
	return m.returnTx, nil
}

func (m *mockLiquidPaymentTransaction) CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error {
	if m.isError {
		return errors.New("mock error CompletePayment")
	}
	return nil
}

func TestTransactionCoreService_CompletePassedLiquidPayment(t *testing.T) {
	type fields struct {
		Log                           *logrus.Logger
		QueryExecutor                 query.ExecutorInterface
		TypeActionSwitcher            transaction.TypeActionSwitcher
		TransactionUtil               transaction.UtilInterface
		TransactionQuery              query.TransactionQueryInterface
		EscrowTransactionQuery        query.EscrowTransactionQueryInterface
		PendingTransactionQuery       query.PendingTransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantErr:ExecuteSelect_error",
			fields: fields{
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{
					isExecuteSelectError: true,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:BuildModels_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					isBuildModelsError: true,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:ExecuteSelectRow_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{
					isExecuteSelectRowError: true,
				},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionQuery.Scan_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: &mockCompletePassedLiquidPaymentTransactionQuery{
					isScanError: true,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:TypeActionSwitcher.GetTransactionType_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: &mockCompletePassedLiquidPaymentTransactionQuery{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					isError: true,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:LiquidPaymentTransaction_casting_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: &mockCompletePassedLiquidPaymentTransactionQuery{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &transaction.TXEmpty{},
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:CompletePayment_error",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: &mockCompletePassedLiquidPaymentTransactionQuery{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &mockLiquidPaymentTransaction{
						isError: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor: &mockCompletePassedLiquidPaymentExecutor{},
				LiquidPaymentTransactionQuery: &mockCompletePassedLiquidPaymentLiquidPaymentTransactionQuery{
					returnModels: []*model.LiquidPayment{
						{},
					},
				},
				TransactionQuery: &mockCompletePassedLiquidPaymentTransactionQuery{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &mockLiquidPaymentTransaction{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				Log:                           tt.fields.Log,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
				TransactionUtil:               tt.fields.TransactionUtil,
				TransactionQuery:              tt.fields.TransactionQuery,
				EscrowTransactionQuery:        tt.fields.EscrowTransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
			}
			if err := tg.CompletePassedLiquidPayment(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.CompletePassedLiquidPayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
