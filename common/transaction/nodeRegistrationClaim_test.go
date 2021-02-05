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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	mockExecutorValidateSuccessClaimNR struct {
		query.Executor
	}
	mockExecutorValidateFailClaimNRNodeNotRegistered struct {
		query.Executor
	}
	mockExecutorValidateFailClaimNRNodeAlreadyDeleted struct {
		query.Executor
	}
	mockExecutorApplyConfirmedSuccessClaimNR struct {
		query.Executor
	}
	mockExecutorApplyConfirmedFailNodeNotFoundClaimNR struct {
		query.Executor
	}
	mockAccountBalanceHelperClaimNRValidateFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperClaimNRValidateNotEnoughSpendable struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperClaimNRValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeClaimNodeRegistrationValidate int64 = 10
)

func (*mockExecutorApplyConfirmedFailNodeNotFoundClaimNR) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorApplyConfirmedSuccessClaimNR) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockExecutorApplyConfirmedSuccessClaimNR) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			int64(1000),
			uint32(model.NodeRegistrationState_NodeRegistered),
			true,
			uint32(1),
		))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorValidateSuccessClaimNR) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			int64(1000),
			uint32(model.NodeRegistrationState_NodeRegistered),
			true,
			uint32(1),
		))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorValidateFailClaimNRNodeAlreadyDeleted) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			int64(0),
			uint32(model.NodeRegistrationState_NodeDeleted),
			true,
			uint32(1),
		))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorValidateFailClaimNRNodeNotRegistered) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockAccountBalanceHelperClaimNRValidateFail) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance,
	address []byte,
	dbTx bool,
) error {
	return errors.New("MockedError")
}

func (*mockAccountBalanceHelperClaimNRValidateNotEnoughSpendable) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeClaimNodeRegistrationValidate - 1
	return nil
}

func (*mockAccountBalanceHelperClaimNRValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeClaimNodeRegistrationValidate + 1
	return nil
}

func TestClaimNodeRegistration_Validate(t *testing.T) {
	poown, _, _ := GetFixturesForClaimNoderegistration()
	txBodyWithoutPoown := &model.ClaimNodeRegistrationTransactionBody{}
	txBodyWithPoown := &model.ClaimNodeRegistrationTransactionBody{
		Poown: poown,
	}
	txBodyFull := &model.ClaimNodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: []byte{1, 1, 1, 1},
	}

	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errText string
	}{
		{
			name: "Validate:fail-{PoownRequired}",
			fields: fields{
				TransactionObject: &model.Transaction{},
				Body:              txBodyWithoutPoown,
			},
			wantErr: true,
			errText: "ValidationErr: PoownRequired",
		},
		{
			name: "Validate:fail-{InvalidPoown}",
			fields: fields{
				TransactionObject: &model.Transaction{},
				Body:              txBodyWithPoown,
				AuthPoown:         &mockAuthPoown{success: false},
			},
			wantErr: true,
			errText: "MockedError",
		},
		{
			name: "Validate:fail-{ClaimedNodeNotRegistered}",
			fields: fields{
				TransactionObject:     &model.Transaction{},
				Body:                  txBodyWithPoown,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateFailClaimNRNodeNotRegistered{},
			},
			wantErr: true,
			errText: blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyNotRegistered").Error(),
		},
		{
			name: "Validate:fail-{ClaimedNodeAlreadyClaimedOrDeleted}",
			fields: fields{
				TransactionObject:     &model.Transaction{},
				Body:                  txBodyWithPoown,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateFailClaimNRNodeAlreadyDeleted{},
			},
			wantErr: true,
			errText: blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyClaimedOrDeleted").Error(),
		},
		{
			name: "Validate:fail-{GetAccountBalanceByAccountAddressFail}",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeClaimNodeRegistrationValidate,
				},
				Body:                  txBodyFull,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateSuccessClaimNR{},
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRValidateFail{},
			},
			wantErr: true,
			errText: "ValidationErr: BalanceNotEnough",
		},
		{
			name: "Validate:fail-{GetAccountBalanceByAccountAddressNotEnoughSpendable}",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeClaimNodeRegistrationValidate,
				},
				Body:                  txBodyFull,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateSuccessClaimNR{},
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRValidateNotEnoughSpendable{},
			},
			wantErr: true,
			errText: blocker.NewBlocker(blocker.ValidationErr, "BalanceNotEnough").Error(),
		},
		{
			name: "Validate:success",
			fields: fields{
				TransactionObject:     &model.Transaction{},
				Body:                  txBodyFull,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateSuccessClaimNR{},
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRValidateSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			err := tx.Validate(false)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NodeAuthValidation.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err.Error() != tt.errText {
					t.Errorf("NodeAuthValidation.ValidateProofOfOwnership() error text = %s, wantErr text %s", err.Error(), tt.errText)
				}
			}
		})
	}
}

func TestClaimNodeRegistration_GetAmount(t *testing.T) {
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("ClaimNodeRegistration.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaimNodeRegistration_GetSize(t *testing.T) {
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields *fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			fields: &fields{
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
			},
			want: 204,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			got, err := tx.GetSize()
			if err != nil {
				t.Errorf("ClaimNodeRegistration.GetSize() = err %s", err)
			}
			if got != tt.want {
				t.Errorf("ClaimNodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaimNodeRegistration_GetBodyBytes(t *testing.T) {
	_, txBody, txBodyBytes := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
				Body: txBody,
			},
			want: txBodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClaimNodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockAccountBalanceHelperClaimNodeRegistrationApplyUnconfirmedSuccess struct {
	AccountBalanceHelper
}

func (*mockAccountBalanceHelperClaimNodeRegistrationApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}
func TestClaimNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
				Body: txBody,
				TransactionObject: &model.Transaction{
					Fee:                  1,
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNodeRegistrationApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("ClaimNodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedFail) AddAccountSpendableBalance(address []byte, amount int64) error {
	return sql.ErrTxDone
}
func (*mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}
func TestClaimNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
				Body:                  txBody,
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedFail{},
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
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Body:                  txBody,
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("ClaimNodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockClaimNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess struct {
		storage.TransactionalCache
	}
)

func (*mockClaimNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess) TxRemoveItem(interface{}) error {
	return nil
}

type (
	mockNodeRegistryCacheSuccess struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockNodeRegistryCacheSuccess) GetItem(index, item interface{}) error {
	castedItem := item.(*storage.NodeRegistry)
	*castedItem = storage.NodeRegistry{
		ParticipationScore: 10,
	}
	return nil
}

func (*mockNodeRegistryCacheSuccess) TxRemoveItem(interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheSuccess) TxSetItem(id, item interface{}) error {
	return nil
}

type (
	mockNodeRegistryCacheNotFound struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockNodeRegistryCacheNotFound) GetItem(index, item interface{}) error {

	return blocker.NewBlocker(blocker.NotFound, "mockedError")
}

func (*mockNodeRegistryCacheNotFound) TxRemoveItem(interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheNotFound) TxSetItem(id, item interface{}) error {
	return nil
}

func TestClaimNodeRegistration_ApplyConfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                    *model.ClaimNodeRegistrationTransactionBody
		TransactionObject       *model.Transaction
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		QueryExecutor           query.ExecutorInterface
		AuthPoown               auth.NodeAuthValidationInterface
		NodeAddressInfoStorage  storage.TransactionalCache
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		ActiveNodeRegistryCache storage.TransactionalCache
		AccountBalanceHelper    AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errText string
	}{
		{
			name: "ApplyConfirmed:fail-{NodePublicKeyNotRegistered}",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
					Fee:                  1,
				},
				QueryExecutor:           &mockExecutorApplyConfirmedFailNodeNotFoundClaimNR{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				NodeAddressInfoStorage:  &mockClaimNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess{},
				ActiveNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
			},
			wantErr: true,
			errText: "AppErr: NodePublicKeyNotRegistered",
		},
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
					Fee:                  1,
				},
				QueryExecutor:           &mockExecutorApplyConfirmedSuccessClaimNR{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				NodeAddressInfoQuery:    query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage:  &mockClaimNodeRegistrationApplyConfirmedNodeAddressInfoStorageSuccess{},
				ActiveNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
				AccountBalanceHelper:    &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                    tt.fields.Body,
				TransactionObject:       tt.fields.TransactionObject,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				BlockQuery:              tt.fields.BlockQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				AuthPoown:               tt.fields.AuthPoown,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				AccountBalanceHelper:    tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				if (err == nil) != tt.wantErr {
					t.Errorf("NodeAuthValidation.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
		})
	}
}

func TestClaimNodeRegistration_ParseBodyBytes(t *testing.T) {
	_, txBody, txBodyBytes := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
			name: "ClaimNodeRegistration:error - empty body bytes",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ClaimNodeRegistration:error - wrong public key length",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: txBodyBytes[:10]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ClaimNodeRegistration:error - no account address length",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey))]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no account address",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey) + 4)]},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "ClaimNodeRegistration:ParseBodyBytes - success",
			fields: fields{},
			args: args{
				txBodyBytes: txBodyBytes,
			},
			want:    txBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
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

func TestClaimNodeRegistration_GetTransactionBody(t *testing.T) {
	_, mockTxBody, _ := GetFixturesForClaimNoderegistration()

	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestClaimNodeRegistration_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		ID                      int64
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
