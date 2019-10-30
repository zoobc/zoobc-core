package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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

func (*mockExecutorValidateFailExecuteSelectFailRU) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:selectFail")
}

func (*mockExecutorValidateFailAccountNotNodeOwnerRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateFailNodeDeleted) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey2,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeDeleted,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateFailNodeNotFoundRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address,"+
		" locked_balance, registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateFailNodeAlreadyRegisteredRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey2,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateSuccessUpdateNodePublicKeyRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateSuccessRU) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
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
			int64(1000000000),
			int64(1000000000),
			int64(100000000),
			true,
		))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateSuccessRU) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorValidateSuccessRU) ExecuteTransactions(queries [][]interface{}) error {
	return nil
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
		NodeAddress: &model.NodeAddress{
			Address: "127.0.1",
		},
	}
	txBodyWithValidNodePubKey := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey2,
		NodeAddress: &model.NodeAddress{
			Address: "127.0.0.1",
			Port:    8080,
		},
	}
	txBodyWithInvalidLockedBalance := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(100),
		NodeAddress: &model.NodeAddress{
			Address: "127.0.0.1",
			Port:    8080,
		},
	}
	txBodyWithLockedBalanceTooHigh := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(10000000000),
	}
	txBodyWithValidLockedBalance := &model.UpdateNodeRegistrationTransactionBody{
		Poown:         poown,
		LockedBalance: int64(1000000000),
		NodeAddress: &model.NodeAddress{
			Address: "127.0.0.1",
			Port:    8080,
		},
	}
	txBodyWithInvalidNodeURI := &model.UpdateNodeRegistrationTransactionBody{
		Poown: poown,
		NodeAddress: &model.NodeAddress{
			Address: "http://google.com",
		},
	}

	txBodyWithInvalidNodeAddress := &model.UpdateNodeRegistrationTransactionBody{
		Poown: poown,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.x",
		},
	}
	txBodyWithValidNodeAddress := &model.UpdateNodeRegistrationTransactionBody{
		Poown: poown,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.10",
		},
	}
	txBodyWithValidNodeURI := &model.UpdateNodeRegistrationTransactionBody{
		Poown: poown,
		NodeAddress: &model.NodeAddress{
			Address: "https://google.com",
		},
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
				Body:                  txBodyWithValidPoown,
				SenderAddress:         senderAddress1,
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
				Body:                  txBodyWithValidPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailNodeDeleted{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{SenderAccountNotNodeOwner}",
			fields: fields{
				Body:                  txBodyWithValidPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailAccountNotNodeOwnerRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{UpdateNodePublicKey.NodeAlreadyRegistered}",
			fields: fields{
				Body:                  txBodyWithInvalidNodePubKey,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailNodeAlreadyRegisteredRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:success-{UpdateNodePublicKey}",
			fields: fields{
				Body:                  txBodyWithValidNodePubKey,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessUpdateNodePublicKeyRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{UpdateLockedBalance.NewBalanceLowerThanPrevious}",
			fields: fields{
				Body:                  txBodyWithInvalidLockedBalance,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{UpdateLockedBalance.InsufficientAccountBalance}",
			fields: fields{
				Body:                  txBodyWithLockedBalanceTooHigh,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:success-{UpdateLockedBalance}",
			fields: fields{
				Body:                  txBodyWithValidLockedBalance,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{UpdateNodeAddress.InvalidURI}",
			fields: fields{
				Body:                  txBodyWithInvalidNodeURI,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{UpdateNodeAddress.InvalidIP}",
			fields: fields{
				Body:                  txBodyWithInvalidNodeAddress,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:success-{UpdateNodeAddressValidURI}",
			fields: fields{
				Body:                  txBodyWithValidNodeURI,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:success-{UpdateNodeAddress}",
			fields: fields{
				Body:                  txBodyWithValidNodeAddress,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmed:success",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmed:fail-{PreviousNodeRecordNotFound}",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailNodeNotFoundRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateNodeRegistration_ApplyConfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccessRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: false,
		},
		{
			name: "ApplyConfirmed:fail-{PreviousNodeRecordNotFound}",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailNodeNotFoundRU{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		LockedBalance: int64(10000000000),
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:fail-{executeTransactionsFail}",
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body:                  txBody,
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				Body:                  txBody,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
			want: txBody.LockedBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
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
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		NodeAddress: &model.NodeAddress{
			Address: "11.10.10.10",
		},
	}
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
			},
			want: 199,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &UpdateNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("UpdateNodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNodeRegistration_ParseBodyBytes(t *testing.T) {

	mockNodeRegisryQ := query.NewNodeRegistrationQuery()
	_, _, txBody, txBodyBytes := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey) + 4 + len([]byte(
					mockNodeRegisryQ.ExtractNodeAddress(txBody.GetNodeAddress())),
				))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "UpdateNodeRegistration:error - no poown",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: txBodyBytes[:(len(txBody.NodePublicKey) + 4 +
					len([]byte(mockNodeRegisryQ.ExtractNodeAddress(txBody.GetNodeAddress()))) + int(constant.Balance))],
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateNodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNodeRegistration_GetTransactionBody(t *testing.T) {
	_, _, mockTxBody, _ := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.UpdateNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}
