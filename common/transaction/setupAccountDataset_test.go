package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
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

	executorSetupAccountDatasetApplyUnconfirmedSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetApplyUnconfirmedFail struct {
		query.Executor
	}

	executorSetupAccountDatasetUndoUnconfirmSuccess struct {
		query.Executor
	}
	executorSetupAccountDatasetUndoUnconfirmFail struct {
		query.Executor
	}
	executorSetupAccountDatasetValidateSuccess struct {
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

func (*executorSetupAccountDatasetUndoUnconfirmSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupAccountDatasetUndoUnconfirmFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupAccountDatasetValidateSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
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

	return db.QueryRow(qStr), nil
}

func TestSetupAccountDataset_ApplyConfirmed(t *testing.T) {
	mockSetupAccountDatasetTransactionBody, _ := GetFixturesForSetupAccountDataset()

	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
				Body:                mockSetupAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       mockSetupAccountDatasetTransactionBody.GetSetterAccountAddress(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorSetupAccountDatasetApplyConfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
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
				QueryExecutor: &executorSetupAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
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
				QueryExecutor: &executorSetupAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
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
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupAccountDataset.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupAccountDataset_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
					MuchTime:             2000,
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor: &executorSetupAccountDatasetApplyUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ExecuteSpandableBalanceFail",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:             2000,
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor: &executorSetupAccountDatasetApplyUnconfirmedFail{
					query.Executor{
						Db: db,
					},
				},
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

func TestSetupAccountDataset_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
				QueryExecutor: &executorSetupAccountDatasetUndoUnconfirmSuccess{
					query.Executor{
						Db: db,
					},
				},
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
				QueryExecutor: &executorSetupAccountDatasetUndoUnconfirmFail{
					query.Executor{
						Db: db,
					},
				},
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

func TestSetupAccountDataset_Validate(t *testing.T) {
	type fields struct {
		Body                *model.SetupAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountDatasetQuery query.AccountDatasetsQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantErr:MuchTimeZero",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					MuchTime: 0,
				},
				Fee:                 1,
				SenderAddress:       "",
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			wantErr: true,
		},
		{
			name: "wantErr:BalanceNotEnough",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					MuchTime: 2000,
				},
				Fee:                 60,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor: &executorSetupAccountDatasetValidateSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Validate:success",
			fields: fields{
				Body: &model.SetupAccountDatasetTransactionBody{
					SetterAccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:             2000,
					Property:             "Admin",
					Value:                "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: nil,
				QueryExecutor: &executorSetupAccountDatasetValidateSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: false,
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
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
					MuchTime:                123,
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 124,
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
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
					SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Property:                "Admin",
					Value:                   "Welcome",
					MuchTime:                123,
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountBalanceQuery: nil,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: []byte{
				44, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86,
				70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 44, 0, 0, 0, 66, 67,
				90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118,
				66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 5, 0, 0, 0, 65, 100, 109, 105, 110, 7, 0,
				0, 0, 87, 101, 108, 99, 111, 109, 101, 123, 0, 0, 0, 0, 0, 0, 0,
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
		AccountDatasetQuery query.AccountDatasetsQueryInterface
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
