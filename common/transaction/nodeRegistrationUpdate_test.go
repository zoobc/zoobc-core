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
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	mockExecutorValidateFailExecuteSelectFailRU struct {
		query.Executor
	}
	mockExecutorValidateFailAccountNotNodeOwnerRU struct {
		query.Executor
	}
	mockExecutorValidateFailNodeDeleted struct {
		query.Executor
	}
	mockExecutorValidateFailNodeAlreadyRegisteredRU struct {
		query.Executor
	}
	mockExecutorValidateSuccessUpdateNodePublicKeyRU struct {
		query.Executor
	}
	mockExecutorValidateFailNodeNotFoundRU struct {
		query.Executor
	}
	mockExecutorValidateSuccessRU struct {
		query.Executor
	}
)

func (*mockExecutorValidateFailExecuteSelectFailRU) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("mockError:selectFail")
}

func (*mockExecutorValidateFailAccountNotNodeOwnerRU) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorValidateFailNodeDeleted) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
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
			nodePubKey2,
			senderAddress1,
			uint32(1),
			int64(1000),
			model.NodeRegistrationState_NodeDeleted,
			true,
			uint32(1),
		))
		return db.QueryRow(""), nil
	}
	return nil, nil
}

func (*mockExecutorValidateFailNodeNotFoundRU) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewNodeRegistrationQuery().Fields))
	return db.QueryRow(qe), nil
}

func (*mockExecutorValidateFailNodeAlreadyRegisteredRU) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if strings.Contains(qe, "WHERE account_address = ? AND latest=1") {
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
			nodePubKey2,
			senderAddress1,
			uint32(1),
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
	} else {
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
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
	}
	return db.QueryRow(qe), nil
}

func (*mockExecutorValidateSuccessUpdateNodePublicKeyRU) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
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
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.QueryRow(""), nil
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow(""), nil
	}
	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM"+
		" account_balance WHERE account_address = ? AND latest = 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"account_address",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
			"latest",
		}).AddRow(
			senderAddress1,
			uint32(1),
			int64(1000000000),
			int64(1000000000),
			int64(100000000),
			true,
		))
		return db.QueryRow(""), nil
	}
	return nil, errors.New("mocked select failed, query  not found")
}

func (*mockExecutorValidateSuccessUpdateNodePublicKeyRU) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows([]string{
			"account_address",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
			"latest",
		}).AddRow(
			senderAddress1,
			uint32(1),
			int64(1000000000),
			int64(1000000000),
			int64(100000000),
			true,
		))
	return db.Query(qStr)
}

func (*mockExecutorValidateSuccessRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
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
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("")
	}
	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM"+
		" account_balance WHERE account_address = ? AND latest = 1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"account_address",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
			"latest",
		}).AddRow(
			senderAddress1,
			uint32(1),
			int64(1000000000000),
			int64(1000000000000),
			int64(100000000),
			true,
		))
		return db.Query("")
	}
	return nil, errors.New("mocked select failed, query  not found")
}

func (*mockExecutorValidateSuccessRU) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	var mockedRows *sqlmock.Rows
	switch {
	case strings.Contains(qStr, "account_balance"):
		mockedRows = mock.NewRows(query.NewAccountBalanceQuery().Fields)
		mockedRows.AddRow(
			senderAddress1,
			uint32(1),
			int64(900000000),
			int64(1000000000000),
			int64(100000000),
			true,
		)
	default:
		mockedRows = mock.NewRows(query.NewNodeRegistrationQuery().Fields)
		mockedRows.AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		)
	}
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(mockedRows)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorValidateSuccessRU) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorValidateSuccessRU) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

type (
	mockAccountBalanceHelperUpdateNRValidateSuccess struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperUpdateNRValidateFail struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperUpdateNRValidateSuccess) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return true, nil
}
func (*mockAccountBalanceHelperUpdateNRValidateFail) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return false, sql.ErrNoRows
}
func TestUpdateNodeRegistration_Validate(t *testing.T) {
	_, poown, _, _ := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	txBodyInvalidPoown := &model.UpdateNodeRegistrationTransactionBody{
		Poown: poown,
	}
	txBodyWithoutPoown := &model.UpdateNodeRegistrationTransactionBody{}
	txBodyWithValidPoown := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey1,
	}
	txBodyWithInvalidNodePubKey := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey1,
	}
	txBodyWithValidNodePubKey := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey2,
		LockedBalance: int64(1000000),
	}
	txBodyWithInvalidLockedBalance := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(100),
	}
	txBodyWithLockedBalanceTooHigh := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(10000000000),
	}
	txBodyWithValidLockedBalance := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(100000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
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
			name: "Validate:fail-{PoownRequired}",
			fields: fields{
				Body: txBodyWithoutPoown,
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{InvalidPoown}",
			fields: fields{
				Body:      txBodyInvalidPoown,
				AuthPoown: &mockAuthPoown{success: false},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{executeSelectFail}",
			fields: fields{
				Body: txBodyWithValidPoown,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectFailRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{NodeDeleted}",
			fields: fields{
				Body: txBodyWithValidPoown,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateFailNodeDeleted{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{SenderAccountNotNodeOwner}",
			fields: fields{
				Body: txBodyWithValidPoown,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateFailAccountNotNodeOwnerRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{UpdateNodePublicKey.NodeAlreadyRegistered}",
			fields: fields{
				Body: txBodyWithInvalidNodePubKey,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateFailNodeAlreadyRegisteredRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:success-{UpdateNodePublicKey}",
			fields: fields{
				Body: txBodyWithValidNodePubKey,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessUpdateNodePublicKeyRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRValidateSuccess{},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{UpdateLockedBalance.NewBalanceLowerThanPrevious}",
			fields: fields{
				Body: txBodyWithInvalidLockedBalance,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{UpdateLockedBalance.InsufficientAccountBalance}",
			fields: fields{
				Body: txBodyWithLockedBalanceTooHigh,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRValidateFail{},
			},
			wantErr: true,
		},
		{
			name: "Validate:success-{UpdateLockedBalance}",
			fields: fields{
				Body: txBodyWithValidLockedBalance,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRValidateSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperUpdateNRApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperUpdateNRApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}

func TestUpdateNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
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
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmed:fail-{PreviousNodeRecordNotFound}",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:         &mockExecutorValidateFailNodeNotFoundRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorUpdateNodeRegApplyConfirmedNodeNotFound struct {
		query.Executor
	}
	mockQueryExecutorUpdateNodeRegApplyConfirmedSuccess struct {
		query.Executor
	}
)

func (*mockQueryExecutorUpdateNodeRegApplyConfirmedNodeNotFound) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewAccountBalanceQuery().Fields))
	return db.QueryRow(qe), nil
}

func (*mockQueryExecutorUpdateNodeRegApplyConfirmedSuccess) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockedRows := mock.NewRows(query.NewNodeRegistrationQuery().Fields)
	mockedRows.AddRow(
		int64(10000),
		nodePubKey1,
		senderAddress1,
		uint32(1),
		int64(1000),
		model.NodeRegistrationState_NodeRegistered,
		true,
		uint32(1),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(mockedRows)
	return db.QueryRow(qe), nil
}
func (*mockQueryExecutorUpdateNodeRegApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

type (
	mockAccountBalanceHelperUpdateNRApplyConfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperUpdateNRApplyConfirmedSuccess) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return nil
}
func TestUpdateNodeRegistration_ApplyConfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                     *model.UpdateNodeRegistrationTransactionBody
		TransactionObject        *model.Transaction
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		BlockQuery               query.BlockQueryInterface
		QueryExecutor            query.ExecutorInterface
		AuthPoown                auth.NodeAuthValidationInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		PendingNodeRegistryCache storage.TransactionalCache
		ActiveNodeRegistryCache  storage.TransactionalCache
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyConfirmed:fail-{PreviousNodeRecordNotFound}",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:            &mockQueryExecutorUpdateNodeRegApplyConfirmedNodeNotFound{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				PendingNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryCache:  &mockNodeRegistryCacheSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ApplyConfirmed:errorNotFound",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:            &mockQueryExecutorUpdateNodeRegApplyConfirmedSuccess{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper:     &mockAccountBalanceHelperUpdateNRApplyConfirmedSuccess{},
				PendingNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryCache:  &mockNodeRegistryCacheNotFound{},
			},
			wantErr: true,
		},
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body: txBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
				QueryExecutor:            &mockQueryExecutorUpdateNodeRegApplyConfirmedSuccess{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper:     &mockAccountBalanceHelperUpdateNRApplyConfirmedSuccess{},
				PendingNodeRegistryCache: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryCache:  &mockNodeRegistryCacheSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                         tt.fields.Body,
				TransactionObject:            tt.fields.TransactionObject,
				NodeRegistrationQuery:        tt.fields.NodeRegistrationQuery,
				BlockQuery:                   tt.fields.BlockQuery,
				QueryExecutor:                tt.fields.QueryExecutor,
				AuthPoown:                    tt.fields.AuthPoown,
				AccountBalanceHelper:         tt.fields.AccountBalanceHelper,
				PendingNodeRegistrationCache: tt.fields.PendingNodeRegistryCache,
				ActiveNodeRegistrationCache:  tt.fields.ActiveNodeRegistryCache,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedFail struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}
func (*mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedFail) AddAccountSpendableBalance(address []byte, amount int64) error {
	return sql.ErrNoRows
}

func TestUpdateNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
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
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedFail{},
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
				AccountBalanceHelper:  &mockAccountBalanceHelperUpdateNRUndoApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateNodeRegistration_GetAmount(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				Body: txBody,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("UpdateNodeRegistration.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNodeRegistration_GetSize(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			fields: fields{
				Body:                  txBody,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
				},
			},
			want: 176,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			got, err := tx.GetSize()
			if err != nil {
				t.Errorf("UpdateNodeRegistration.GetSize() = err %s", err)
			}
			if got != tt.want {
				t.Errorf("UpdateNodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNodeRegistration_ParseBodyBytes(t *testing.T) {

	_, _, txBody, txBodyBytes := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
			name: "UpdateNodeRegistration:error - empty body bytes",
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
			name: "UpdateNodeRegistration:error - wrong public key length",
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
			name: "UpdateNodeRegistration:error - no node address length",
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
			name: "UpdateNodeRegistration:error - no node address",
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
			args: args{
				txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey) + 4)],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "UpdateNodeRegistration:error - no locked balance",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "UpdateNodeRegistration:error - no poown",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                  0,
					SenderAccountAddress: nil,
					Height:               0,
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey) + int(constant.Balance))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "UpdateNodeRegistration:ParseBodyBytes - success",
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				txBodyBytes: txBodyBytes,
			},
			want:    txBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			up := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			got, err := up.ParseBodyBytes(tt.args.txBodyBytes)
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

func TestUpdateNodeRegistration_GetBodyBytes(t *testing.T) {
	_, _, txBody, txBodyBytes := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytesBytes:success",
			fields: fields{
				Body:                  txBody,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			want: txBodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateNodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNodeRegistration_GetTransactionBody(t *testing.T) {
	_, _, mockTxBody, _ := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		TransactionObject     *model.Transaction
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
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
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				TransactionObject:     tt.fields.TransactionObject,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestUpdateNodeRegistration_SkipMempoolTransaction(t *testing.T) {
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
						TransactionType:      uint32(model.TransactionType_SendMoneyTransaction),
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
