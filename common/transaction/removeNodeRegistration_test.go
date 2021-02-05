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
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	mockExecutorValidateRemoveNodeRegistrationSuccess struct {
		query.Executor
	}
	mockExecutorValidateRemoveNodeRegistrationFailNodeAlreadyDeleted struct {
		query.Executor
	}
	mockExecutorValidateRemoveNodeRegistrationFailGetRNode struct {
		query.Executor
	}
	mockExecutorApplyUnconfirmedRemoveNodeRegistrationSuccess struct {
		query.Executor
	}
	mockExecutorApplyUnconfirmedRemoveNodeRegistrationFail struct {
		query.Executor
	}
	mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess struct {
		query.Executor
	}
	mockExecutorApplyConfirmedRemoveNodeRegistrationFail struct {
		query.Executor
	}
)

func (*mockExecutorApplyUnconfirmedRemoveNodeRegistrationSuccess) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			1,
			1,
			1,
			1,
			1,
		))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorApplyUnconfirmedRemoveNodeRegistrationSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorApplyUnconfirmedRemoveNodeRegistrationFail) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			senderAddress1,
			1,
			1,
			1,
			1,
			1,
		))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorApplyUnconfirmedRemoveNodeRegistrationFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockdeError")
}

func (*mockExecutorValidateRemoveNodeRegistrationSuccess) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewNodeRegistrationQuery().Fields).AddRow(
		0,
		body.NodePublicKey,
		senderAddress1,
		1,
		1,
		1,
		1,
		1,
	))

	return db.QueryRow(qe), nil
}

func (*mockExecutorValidateRemoveNodeRegistrationFailNodeAlreadyDeleted) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewNodeRegistrationQuery().Fields).AddRow(
		0,
		body.NodePublicKey,
		senderAddress1,
		1,
		1,
		uint32(model.NodeRegistrationState_NodeDeleted),
		1,
		1,
	))
	return db.QueryRow(qe), nil
}

func (*mockExecutorValidateRemoveNodeRegistrationFailGetRNode) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockedError")
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockedRows := mock.NewRows(query.NewNodeRegistrationQuery().Fields)
	mockedRows.AddRow(
		0,
		body.NodePublicKey,
		senderAddress1,
		1,
		1,
		1,
		1,
		1,
	)
	mock.ExpectQuery("SELECT").WillReturnRows(mockedRows)
	return db.QueryRow(qe), nil

}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) BeginTx(bool, int) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) CommitTx(bool) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationFail) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewNodeRegistrationQuery().Fields))
	return db.QueryRow(qe), nil
}

func TestRemoveNodeRegistration_GetBodyBytes(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(body.NodePublicKey)
	bodyBytes := buffer.Bytes()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: body,
			},
			want: bodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveNodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveNodeRegistration_ParseBodyBytes(t *testing.T) {
	_, bodyBytes := GetFixturesForRemoveNoderegistration()
	txBody := &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
	}
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "ParseBodyBytes:fail - no body",
			args: args{
				txBodyBytes: []byte{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseBodyBytes:fail - wrong public key length",
			args: args{
				txBodyBytes: []byte{1, 2, 3, 4},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseBodyBytes:success",
			args: args{
				txBodyBytes: bodyBytes,
			},
			want:    txBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := r.ParseBodyBytes(tt.args.txBodyBytes)
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

func TestRemoveNodeRegistration_GetSize(t *testing.T) {
	tx := &RemoveNodeRegistration{}
	want := constant.NodePublicKey
	if got, _ := tx.GetSize(); got != want {
		t.Errorf("TestRemoveNodeRegistration.GetSize() = %v, want %v", got, want)
	}
}

func TestRemoveNodeRegistration_GetAmount(t *testing.T) {
	tx := &RemoveNodeRegistration{}
	want := int64(0)
	if got := tx.GetAmount(); got != want {
		t.Errorf("TestRemoveNodeRegistration.GetAmount() = %v, want %v", got, want)
	}
}

type (
	mockAccountBalanceHelperRemoveNodeRegistrationValidateFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperRemoveNodeRegistrationValidateNotEnoughSpendable struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperRemoveNodeRegistrationValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeRemoveNodeRegistrationValidate int64 = 10
)

func (*mockAccountBalanceHelperRemoveNodeRegistrationValidateFail) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	return errors.New("MockedError")
}

func (*mockAccountBalanceHelperRemoveNodeRegistrationValidateNotEnoughSpendable) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeRemoveNodeRegistrationValidate - 1
	return nil
}

func (*mockAccountBalanceHelperRemoveNodeRegistrationValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeRemoveNodeRegistrationValidate + 1
	return nil
}

func TestRemoveNodeRegistration_Validate(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:success",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceHelper:  &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{GetNodeQuery}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationFailGetRNode{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{AccountNotNodeOwner}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress2,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{NodeAlreadyDeleted}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationFailNodeAlreadyDeleted{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{GetAccountBalanceByAccountAddressFail}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  mockFeeRemoveNodeRegistrationValidate,
					SenderAccountAddress: senderAddress1,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceHelper:  &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{GetAccountBalanceByAccountAddressNotEnoughSpendable}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  mockFeeRemoveNodeRegistrationValidate,
					SenderAccountAddress: senderAddress1,
					Height:               1,
				},
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceHelper:  &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:fail-{executeTransactionsFail}",
			fields: fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
					Fee:                  1,
				},
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Body:                  body,
				AccountBalanceHelper:  &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
					Fee:                  1,
				},
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Body:                  body,
				AccountBalanceHelper:  &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmed:success",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyUnconfirmedRemoveNodeRegistrationSuccess{},
				AccountBalanceHelper:  &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmed:fail",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyUnconfirmedRemoveNodeRegistrationFail{},
				AccountBalanceHelper:  &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockRemoveNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess struct {
		storage.TransactionalCache
	}
)

func (*mockRemoveNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess) TxRemoveItem(interface{}) error {
	return nil
}

func TestRemoveNodeRegistration_ApplyConfirmed(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                     *model.RemoveNodeRegistrationTransactionBody
		TransactionObject        *model.Transaction
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		QueryExecutor            query.ExecutorInterface
		NodeAddressInfoQuery     query.NodeAddressInfoQueryInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		NodeAddressInfoStorage   storage.TransactionalCache
		ActiveNodeRegistryCache  storage.TransactionalCache
		PendingNodeRegistryCache storage.TransactionalCache
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyConfirmed:fail-{nodeNotExist}",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
				},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				QueryExecutor:            &mockExecutorApplyConfirmedRemoveNodeRegistrationFail{},
				AccountBalanceHelper:     &mockAccountBalanceHelperSuccess{},
				ActiveNodeRegistryCache:  &mockNodeRegistryCacheSuccess{},
				PendingNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body: body,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
				},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				QueryExecutor:            &mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess{},
				NodeAddressInfoQuery:     query.NewNodeAddressInfoQuery(),
				AccountBalanceHelper:     &mockAccountBalanceHelperSuccess{},
				NodeAddressInfoStorage:   &mockRemoveNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess{},
				ActiveNodeRegistryCache:  &mockNodeRegistryCacheSuccess{},
				PendingNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                     tt.fields.Body,
				TransactionObject:        tt.fields.TransactionObject,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeAddressInfoQuery:     tt.fields.NodeAddressInfoQuery,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				NodeAddressInfoStorage:   tt.fields.NodeAddressInfoStorage,
				ActiveNodeRegistryCache:  tt.fields.ActiveNodeRegistryCache,
				PendingNodeRegistryCache: tt.fields.PendingNodeRegistryCache,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveNodeRegistration_GetTransactionBody(t *testing.T) {
	mockTxBody, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
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
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestRemoveNodeRegistration_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		Body                    *model.NodeRegistrationTransactionBody
		TransactionObject       *model.Transaction
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
		AuthPoown               auth.NodeAuthValidationInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
		newBlockTimestamp    int64
		newBlockHeight       uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "SkipMempoolTransaction:success-{Filtered}",
			fields: fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: senderAddress1,
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: senderAddress1,
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: senderAddress1,
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
			want: true,
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_DifferentSenders}",
			fields: fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: senderAddress2,
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: senderAddress3,
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: senderAddress4,
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_NoOtherRecordsFound}",
			fields: fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: senderAddress2,
						TransactionType:      uint32(model.TransactionType_SetupAccountDatasetTransaction),
					},
					{
						SenderAccountAddress: senderAddress3,
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: senderAddress4,
						TransactionType:      uint32(model.TransactionType_sendZBCTransaction),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                    tt.fields.Body,
				TransactionObject:       tt.fields.TransactionObject,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				AuthPoown:               tt.fields.AuthPoown,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.newBlockTimestamp, tt.args.newBlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NodeRegistration.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
