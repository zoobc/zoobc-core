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
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Body:                mockSetupAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       mockSetupAccountDatasetTransactionBody.GetSetterAccountAddress(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetApplyConfirmedSuccess{},
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: false,
		},
		{
			name: "wantErr:UndoUnconfirmFail",
			fields: fields{
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       mockSetupAccountDatasetTransactionBody.GetSetterAccountAddress(),
				Height:              3,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetApplyConfirmedFail{},
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionsFail",
			fields: fields{
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       mockSetupAccountDatasetTransactionBody.GetSetterAccountAddress(),
				Height:              0,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetApplyConfirmedFail{},
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
				AccountLedgerQuery:  tt.fields.AccountLedgerQuery,
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
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
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
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor:       &executorSetupAccountDatasetApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ExecuteSpandableBalanceFail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor:       &executorSetupAccountDatasetApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
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
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor:       &executorSetupAccountDatasetUndoUnconfirmSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:fail",
			fields: fields{
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor:       &executorSetupAccountDatasetUndoUnconfirmFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
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
				"BCZ",
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
				"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
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
				"BCZ",
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
				"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
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
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantErr:BalanceNotEnough",
			fields: fields{
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 60,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetValidateSuccess{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:AlreadyExists",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetValidateAlreadyExists{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:Success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorSetupAccountDatasetValidateSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
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
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
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
				Body:                &model.SetupAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "",
				Height:              5,
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
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
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
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
					SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Property:                "Admin",
					Value:                   "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 116,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("SetupAccountDataset.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccountDataset_GetBodyBytes(t *testing.T) {

	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
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
					SetterAccountAddress:    "Hl891TeTFxGgWOWfOOFKYr_XdhXNxO8JK8WnMJV6g8aL",
					RecipientAccountAddress: "HlZLh3VcnNlvByWoAzXOQ2jAlwFOiyO9_njI3oq5Ygha",
					Property:                "AccountDatasetEscrowApproval",
					Value:                   "Happy birthday",
				},
				Fee:                 1,
				SenderAddress:       "Hl891TeTFxGgWOWfOOFKYr_XdhXNxO8JK8WnMJV6g8aL",
				Height:              5,
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: []byte{
				44, 0, 0, 0, 72, 108, 56, 57, 49, 84, 101, 84, 70, 120, 71, 103, 87, 79, 87, 102, 79, 79,
				70, 75, 89, 114, 95, 88, 100, 104, 88, 78, 120, 79, 56, 74, 75, 56, 87, 110, 77, 74, 86,
				54, 103, 56, 97, 76, 44, 0, 0, 0, 72, 108, 90, 76, 104, 51, 86, 99, 110, 78, 108, 118, 66,
				121, 87, 111, 65, 122, 88, 79, 81, 50, 106, 65, 108, 119, 70, 79, 105, 121, 79, 57, 95, 110,
				106, 73, 51, 111, 113, 53, 89, 103, 104, 97, 28, 0, 0, 0, 65, 99, 99, 111, 117, 110, 116,
				68, 97, 116, 97, 115, 101, 116, 69, 115, 99, 114, 111, 119, 65, 112, 112, 114, 111, 118, 97,
				108, 14, 0, 0, 0, 72, 97, 112, 112, 121, 32, 98, 105, 114, 116, 104, 100, 97, 121,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAccountDataset.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAccountDataset_GetTransactionBody(t *testing.T) {
	mockTxBody, _ := GetFixturesForSetupAccountDataset()
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
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
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}
