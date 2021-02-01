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
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	executorSetupAccountDatasetApplyConfirmedSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetApplyConfirmedFail struct {
		query.Executor
	}
)

func (*executorSetupAccountDatasetApplyConfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupAccountDatasetApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSetupAccountDatasetApplyConfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupAccountDatasetApplyConfirmedFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockedError")
}

func TestSetupAccountDataset_ApplyConfirmed(t *testing.T) {
	mockSetupAccountDatasetTransactionBody, _ := GetFixturesForSetupAccountDataset()

	type fields struct {
		Body                 *model.SetupAccountDatasetTransactionBody
		TransactionObject    *model.Transaction
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Body: mockSetupAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorSetupAccountDatasetApplyConfirmedSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:UndoUnconfirmedFail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  3,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorSetupAccountDatasetApplyConfirmedFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionsFail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  0,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorSetupAccountDatasetApplyConfirmedFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("SetupAccountDataset.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorSetupAccountDatasetApplyUnconfirmedSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetApplyUnconfirmedFail struct {
		query.Executor
	}
)

func (*executorSetupAccountDatasetApplyUnconfirmedSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewAccountBalanceQuery().Fields,
	).AddRow(1, 2, 50, 50, 0, 1))
	return db.Query(qStr, 1)
}

func (*executorSetupAccountDatasetApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupAccountDatasetApplyUnconfirmedFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewAccountBalanceQuery().Fields,
	).AddRow(1, 2, 50, 50, 0, 1))
	return db.Query(qStr, 1)
}

func (*executorSetupAccountDatasetApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestSetupAccountDataset_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SetupAccountDatasetTransactionBody
		TransactionObject    *model.Transaction
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "Admin",
					Value:    "Welcome",
				},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  nil,
				QueryExecutor:        &executorSetupAccountDatasetApplyUnconfirmedSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ExecuteSpendableBalanceFail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "Admin",
					Value:    "Welcome",
				},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  nil,
				QueryExecutor:        &executorSetupAccountDatasetApplyUnconfirmedFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupAccountDataset.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorSetupAccountDatasetUndoUnconfirmSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetUndoUnconfirmFail struct {
		query.Executor
	}
)

func (*executorSetupAccountDatasetUndoUnconfirmSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupAccountDatasetUndoUnconfirmFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestSetupAccountDataset_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SetupAccountDatasetTransactionBody
		TransactionObject    *model.Transaction
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  nil,
				QueryExecutor:        &executorSetupAccountDatasetUndoUnconfirmSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:fail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  nil,
				QueryExecutor:        &executorSetupAccountDatasetUndoUnconfirmFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupAccountDataset.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorSetupAccountDatasetValidateSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetValidateAlreadyExists struct {
		query.Executor
	}
)

func (*executorSetupAccountDatasetValidateSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "account_balance") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewAccountBalanceQuery().Fields).AddRow(
				senderAddress1,
				1,
				1,
				1,
				0,
				true,
			),
		)
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewAccountDatasetsQuery().Fields).AddRow(
				senderAddress1,
				recipientAddress1,
				"Admin",
				"You're Welcome",
				false,
				true,
				5,
			),
		)
	}

	return db.QueryRow(qStr), nil
}

func (*executorSetupAccountDatasetValidateAlreadyExists) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "account_balance") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewAccountBalanceQuery().Fields).AddRow(
				senderAddress1,
				1,
				1,
				1,
				0,
				true,
			),
		)
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewAccountDatasetsQuery().Fields).AddRow(
				senderAddress1,
				recipientAddress1,
				"Admin",
				"You're Welcome",
				true,
				true,
				5,
			),
		)
	}

	return db.QueryRow(qStr), nil
}

func TestSetupAccountDataset_Validate(t *testing.T) {
	type fields struct {
		Body                 *model.SetupAccountDatasetTransactionBody
		TransactionObject    *model.Transaction
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantErr:BalanceNotEnough",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee: 60,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorSetupAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:AlreadyExists",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "Admin",
					Value:    "Welcome",
				},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetValidateAlreadyExists{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:Success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "Admin",
					Value:    "Welcome",
				},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorSetupAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("SetupAccountDataset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupAccountDataset_GetAmount(t *testing.T) {
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		TransactionObject   *model.Transaction
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: nil,
					Height:               5,
				},
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("SetupAccountDataset.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccountDataset_GetSize(t *testing.T) {
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		TransactionObject   *model.Transaction
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "Admin",
					Value:    "Welcome",
				},
				TransactionObject: &model.Transaction{
					Fee: 1,
					SenderAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
						239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
					RecipientAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 5,
				},
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("SetupAccountDataset.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccountDataset_GetBodyBytes(t *testing.T) {

	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		TransactionObject   *model.Transaction
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					Property: "AccountDatasetEscrowApproval",
					Value:    "Happy birthday",
				},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  5,
				},
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: []byte{
				28, 0, 0, 0, 65, 99, 99, 111, 117, 110, 116, 68, 97, 116, 97, 115, 101, 116, 69, 115, 99,
				114, 111, 119, 65, 112, 112, 114, 111, 118, 97, 108, 14, 0, 0, 0, 72, 97, 112, 112, 121,
				32, 98, 105, 114, 116, 104, 100, 97, 121,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAccountDataset.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccountDataset_GetTransactionBody(t *testing.T) {
	mockTxBody, _ := GetFixturesForSetupAccountDataset()
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		TransactionObject   *model.Transaction
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: mockTxBody,
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}
