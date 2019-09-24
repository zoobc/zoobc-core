package transaction

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorValidateRemoveNodeRegistrationSuccess struct {
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

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			1,
			"10.10.10.10",
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

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			1,
			"10.10.10.10",
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

func (*mockExecutorValidateRemoveNodeRegistrationSuccess) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			1,
			"10.10.10.10",
			1,
			1,
			1,
			1,
		))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorValidateRemoveNodeRegistrationFailGetRNode) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		return nil, errors.New("MockedError")
	}
	return nil, nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			1,
			"10.10.10.10",
			1,
			1,
			1,
			1,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address,"+
		" locked_balance, queued, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			1,
			"10.10.10.10",
			1,
			1,
			1,
			1,
		))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) BeginTx() error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) CommitTx() error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationFail) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		return nil, errors.New("MockedError")
	}
	return nil, nil
}

func TestRemoveNodeRegistration_GetBodyBytes(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(body.NodePublicKey)
	bodyBytes := buffer.Bytes()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.TransactionBodyInterface
	}{
		{
			name: "ParseBodyBytes:success",
			args: args{
				txBodyBytes: bodyBytes,
			},
			want: txBody,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got, _ := r.ParseBodyBytes(tt.args.txBodyBytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveNodeRegistration.ParseBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveNodeRegistration_GetSize(t *testing.T) {
	tx := &RemoveNodeRegistration{}
	want := constant.NodePublicKey
	if got := tx.GetSize(); got != want {
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

func TestRemoveNodeRegistration_Validate(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:success",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Height:                1,
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{GetNodeQuery}",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Height:                1,
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationFailGetRNode{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{AccountNotNodeOwner}",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:                1,
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:fail-{executeTransactionsFail}",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body:                  body,
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body:                  body,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmed:success",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyUnconfirmedRemoveNodeRegistrationSuccess{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmed:fail",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyUnconfirmedRemoveNodeRegistrationFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveNodeRegistration_ApplyConfirmed(t *testing.T) {
	body, _ := GetFixturesForRemoveNoderegistration()
	type fields struct {
		Body                  *model.RemoveNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess{},
			},
			wantErr: false,
		},
		{
			name: "ApplyConfirmed:fail-{nodeNotExist}",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyConfirmedRemoveNodeRegistrationFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
