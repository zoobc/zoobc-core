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
	address string,
	dbTx bool,
) error {
	return errors.New("MockedError")
}

func (*mockAccountBalanceHelperClaimNRValidateNotEnoughSpendable) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address string, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeClaimNodeRegistrationValidate - 1
	return nil
}

func (*mockAccountBalanceHelperClaimNRValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address string, dbTx bool,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
			name: "Validate:fail-{ClaimedNodeNotRegistered}",
			fields: fields{
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
				Fee:                   mockFeeClaimNodeRegistrationValidate,
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
				Fee:                   mockFeeClaimNodeRegistrationValidate,
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.NodeAuthValidationInterface
		AccountBalanceHelper  AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			want: 264,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &ClaimNodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("ClaimNodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaimNodeRegistration_GetBodyBytes(t *testing.T) {
	_, txBody, txBodyBytes := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClaimNodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockAccountBalanceHelperClaimNodeRegistrationApplyUnconfirmedSuccess struct {
	AccountBalanceHelper
}

func (*mockAccountBalanceHelperClaimNodeRegistrationApplyUnconfirmedSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}
func TestClaimNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				Fee:                   1,
				Body:                  txBody,
				SenderAddress:         senderAddress1,
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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

func (*mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedFail) AddAccountSpendableBalance(address string, amount int64) error {
	return sql.ErrTxDone
}
func (*mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}
func TestClaimNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Fee:                   1,
				Body:                  txBody,
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRUndoApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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
	mockAccountBalanceHelperClaimNRApplyConfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperClaimNRApplyConfirmedSuccess) AddAccountBalance(
	address string, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return nil
}
func TestClaimNodeRegistration_ApplyConfirmed(t *testing.T) {
	_, txBody, _ := GetFixturesForClaimNoderegistration()
	type fields struct {
		Body                  *model.ClaimNodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
			name: "ApplyConfirmed:fail-{NodePublicKeyNotRegistered}",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				Fee:                   1,
				QueryExecutor:         &mockExecutorApplyConfirmedFailNodeNotFoundClaimNR{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
			errText: "AppErr: NodePublicKeyNotRegistered",
		},
		{
			name: "ApplyConfirmed:success",
			fields: fields{
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				Fee:                   1,
				QueryExecutor:         &mockExecutorApplyConfirmedSuccessClaimNR{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper:  &mockAccountBalanceHelperClaimNRApplyConfirmedSuccess{},
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
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
				AccountBalanceHelper:  tt.fields.AccountBalanceHelper,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
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
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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
		Fee                   int64
		SenderAddress         string
		Height                uint32
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
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
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
		Fee                     int64
		SenderAddress           string
		Height                  uint32
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
