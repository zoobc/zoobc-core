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

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
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

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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

func (*mockExecutorValidateRemoveNodeRegistrationFailNodeAlreadyDeleted) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			0,
			body.NodePublicKey,
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			1,
			"10.10.10.10",
			1,
			uint32(model.NodeRegistrationState_NodeDeleted),
			1,
			1,
		))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorValidateRemoveNodeRegistrationFailGetRNode) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status,"+
		" latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		return nil, errors.New("MockedError")
	}
	return nil, nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	body, _ := GetFixturesForRemoveNoderegistration()
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockedRows := mock.NewRows(query.NewNodeRegistrationQuery().Fields)
	mockedRows.AddRow(
		0,
		body.NodePublicKey,
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		1,
		"10.10.10.10",
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

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) BeginTx() error {
	return nil
}

func (*mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess) CommitTx() error {
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
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
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Height:                1,
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{NodeAlreadyDeleted}",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:                1,
				QueryExecutor:         &mockExecutorValidateRemoveNodeRegistrationFailNodeAlreadyDeleted{},
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
		AccountLedgerQuery    query.AccountLedgerQueryInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
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
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body:                  body,
				Fee:                   1,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorApplyConfirmedRemoveNodeRegistrationSuccess{},
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
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
				AccountLedgerQuery:    tt.fields.AccountLedgerQuery,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestRemoveNodeRegistration_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		ID                      int64
		Body                    *model.NodeRegistrationTransactionBody
		Fee                     int64
		SenderAddress           string
		Height                  uint32
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
		AuthPoown               auth.ProofOfOwnershipValidationInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
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
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
			want: true,
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_DifferentSenders}",
			fields: fields{
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tAAAA",
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tAAAA",
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_NoOtherRecordsFound}",
			fields: fields{
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_SetupAccountDatasetTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_SendMoneyTransaction),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				ID:                      tt.fields.ID,
				Body:                    tt.fields.Body,
				Fee:                     tt.fields.Fee,
				SenderAddress:           tt.fields.SenderAddress,
				Height:                  tt.fields.Height,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				AuthPoown:               tt.fields.AuthPoown,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions)
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
