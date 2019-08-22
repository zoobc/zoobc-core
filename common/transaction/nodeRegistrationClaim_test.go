package transaction

import (
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAuthPoownClaimNR struct {
		success bool
		auth.ProofOfOwnershipValidation
	}
	mockExecutorValidateFailExecuteSelectDuplicateAccountClaimNR struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectDuplicateNodePubKeyClaimNR struct {
		query.Executor
	}
	mockExecutorValidateSuccessClaimNR struct {
		query.Executor
	}
)

func (mk *mockAuthPoownClaimNR) ValidateProofOfOwnership(
	poown *model.ProofOfOwnership,
	nodePublicKey []byte,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) error {
	if mk.success {
		return nil
	}
	return errors.New("MockedError")
}

func (*mockExecutorValidateFailExecuteSelectDuplicateAccountClaimNR) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address,"+
		" locked_balance, queued, latest, height FROM node_registry WHERE account_address = "+senderAddress2+
		" AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"queued",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			false,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateFailExecuteSelectDuplicateNodePubKeyClaimNR) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address,"+
		" locked_balance, queued, latest, height FROM node_registry WHERE account_address = "+senderAddress2+
		" AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" queued, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"queued",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			false,
			true,
			uint32(1),
		))
		return db.Query("")
	}
	return nil, nil
}

func (*mockExecutorValidateSuccessClaimNR) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address,"+
		" locked_balance, queued, latest, height FROM node_registry WHERE account_address = "+senderAddress2+
		" AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.Query("")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" queued, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{}))
		return db.Query("")
	}
	return nil, nil
}

func TestClaimNodeRegistration_Validate(t *testing.T) {
	poown, _, _ := GetFixturesForClaimNoderegistration()
	txBodyWithoutPoown := &model.ClaimNodeRegistrationTransactionBody{}
	txBodyWithPoown := &model.ClaimNodeRegistrationTransactionBody{
		Poown: poown,
	}
	txBodyFull := &model.ClaimNodeRegistrationTransactionBody{
		AccountAddress: senderAddress2,
		Poown:          poown,
	}

	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
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
		errText string
	}{
		{
			name: "Validate:fail-{PoownRequired}",
			fields: fields{
				Body: txBodyWithoutPoown,
			},
			wantErr: true,
			errText: "ValidationErr: PoownRequired",
		},
		{
			name: "Validate:fail-{InvalidPoown}",
			fields: fields{
				Body:      txBodyWithPoown,
				AuthPoown: &mockAuthPoown{success: false},
			},
			wantErr: true,
			errText: "MockedError",
		},
		{
			name: "Validate:fail-{AccountAddressRequired}",
			fields: fields{
				Body:      txBodyWithPoown,
				AuthPoown: &mockAuthPoown{success: true},
			},
			wantErr: true,
			errText: "ValidationErr: AccountAddressRequired",
		},
		{
			name: "Validate:fail-{AccountAddressAlreadyRegistered}",
			fields: fields{
				Body:                  txBodyFull,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectDuplicateAccountClaimNR{},
			},
			wantErr: true,
			errText: "ValidationErr: AccountAlreadyNodeOwner",
		},
		{
			name: "Validate:success",
			fields: fields{
				Body:                  txBodyFull,
				AuthPoown:             &mockAuthPoown{success: true},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorValidateSuccessClaimNR{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
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
			err := tx.Validate()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("ProofOfOwnershipValidation.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err.Error() != tt.errText {
					t.Errorf("ProofOfOwnershipValidation.ValidateProofOfOwnership() error text = %s, wantErr text %s", err.Error(), tt.errText)
				}
			}
		})
	}
}
