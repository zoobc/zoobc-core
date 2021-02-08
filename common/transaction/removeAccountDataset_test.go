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
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	executorRemoveAccountDatasetApplyConfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetApplyConfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetApplyConfirmedSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyConfirmedFail) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorRemoveAccountDatasetApplyConfirmedFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_ApplyConfirmed(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
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
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:UndoUnconfirmedFail",
			fields: fields{
				Body: &model.RemoveAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  3,
				},
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionsFail",
			fields: fields{
				Body: &model.RemoveAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  0,
				},
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetApplyUnconfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetApplyUnconfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetApplyUnconfirmedSuccess) ExecuteSelect(qStr string, _ bool,
	_ ...interface{}) (*sql.Rows, error) {
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

func (*executorRemoveAccountDatasetApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyUnconfirmedFail) ExecuteSelect(qStr string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
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

func (*executorRemoveAccountDatasetApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_ApplyUnconfirmed(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()
	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
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
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetApplyUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ExecuteSpendableBalanceFail",
			fields: fields{
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetApplyUnconfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetUndoUnconfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetUndoUnconfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetUndoUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*executorRemoveAccountDatasetUndoUnconfirmedFail) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
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
				Body: &model.RemoveAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: nil,
				},
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetUndoUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:fail",
			fields: fields{
				Body: &model.RemoveAccountDatasetTransactionBody{},
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: nil,
				},
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetUndoUnconfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetValidateSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetValidateFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetValidateSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
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

func (*executorRemoveAccountDatasetValidateFail) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*executorRemoveAccountDatasetValidateFail) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewAccountDatasetsQuery().Fields),
	)

	return db.QueryRow(qStr), nil
}

func TestRemoveAccountDataset_Validate(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
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
			name: "Validate:success",
			fields: fields{
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorRemoveAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "Validate:BalanceNotEnough",
			fields: fields{
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     60,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorRemoveAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "Validate:noRow",
			fields: fields{
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
				},
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorRemoveAccountDatasetValidateFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				TransactionObject:    tt.fields.TransactionObject,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveAccountDataset_GetSize(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                *model.RemoveAccountDatasetTransactionBody
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
				Body: mockRemoveAccountDatasetTransactionBody,
				TransactionObject: &model.Transaction{
					Fee:                     1,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: recipientAddress1,
					Height:                  5,
				},
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 21,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("RemoveAccountDataset.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveAccountDataset_GetTransactionBody(t *testing.T) {
	mockTxBody, _ := GetFixturesForRemoveAccountDataset()
	type fields struct {
		Body                *model.RemoveAccountDatasetTransactionBody
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
			tx := &RemoveAccountDataset{
				Body:                tt.fields.Body,
				TransactionObject:   tt.fields.TransactionObject,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}
