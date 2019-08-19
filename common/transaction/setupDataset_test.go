package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

// var db, mock, _ = sqlmock.New()
type (
	executorSetupDatasetApplyConfirmedSuccess struct {
		query.Executor
	}
	executorSetupDatasetApplyConfirmedFail struct {
		query.Executor
	}

	executorSetupDatasetApplyUnconfirmedSuccess struct {
		query.Executor
	}
	executorSetupDatasetApplyUnconfirmedFail struct {
		query.Executor
	}

	executorSetupDatasetUndoUnconfirmSuccess struct {
		query.Executor
	}
	executorSetupDatasetUndoUnconfirmFail struct {
		query.Executor
	}
	executorSetupDatasetValidateSuccess struct {
		query.Executor
	}
)

func (*executorSetupDatasetApplyConfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupDatasetApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSetupDatasetApplyConfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupDatasetApplyConfirmedFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupDatasetApplyUnconfirmedSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
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

func (*executorSetupDatasetApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupDatasetApplyUnconfirmedFail) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
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

func (*executorSetupDatasetApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupDatasetUndoUnconfirmSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupDatasetUndoUnconfirmFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorSetupDatasetValidateSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
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

func TestSetupDataset_ApplyConfirmed(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:      2000,
					Property:      "Admin",
					Value:         "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        query.NewDatasetsQuery(),
				QueryExecutor: &executorSetupDatasetApplyConfirmedSuccess{
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
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              3,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        query.NewDatasetsQuery(),
				QueryExecutor: &executorSetupDatasetApplyConfirmedFail{
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
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              0,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        query.NewDatasetsQuery(),
				QueryExecutor: &executorSetupDatasetApplyConfirmedFail{
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
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupDataset.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupDataset_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:      2000,
					Property:      "Admin",
					Value:         "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetApplyUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ValidateFail",
			fields: fields{
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: nil,
				DatasetQuery:        nil,
				QueryExecutor:       nil,
			},
			wantErr: true,
		},
		{
			name: "wantErr:ExecuteSpandableBalanceFail",
			fields: fields{
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:      2000,
					Property:      "Admin",
					Value:         "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetApplyUnconfirmedFail{
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
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupDataset.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupDataset_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetUndoUnconfirmSuccess{
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
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetUndoUnconfirmFail{
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
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SetupDataset.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupDataset_Validate(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body: &model.SetupDatasetTransactionBody{
					MuchTime: 0,
				},
				Fee:                 1,
				SenderAddress:       "",
				AccountBalanceQuery: nil,
				DatasetQuery:        nil,
				QueryExecutor:       nil,
			},
			wantErr: true,
		},
		{
			name: "wantErr:BalanceNotEnough",
			fields: fields{
				Body: &model.SetupDatasetTransactionBody{
					MuchTime: 2000,
				},
				Fee:                 60,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetValidateSuccess{
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
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					MuchTime:      2000,
					Property:      "Admin",
					Value:         "Welcome",
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				DatasetQuery:        nil,
				QueryExecutor: &executorSetupDatasetValidateSuccess{
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
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := tx.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SetupDataset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetupDataset_GetAmount(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body:                &model.SetupDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "",
				Height:              5,
				AccountBalanceQuery: nil,
				DatasetQuery:        nil,
				QueryExecutor:       nil,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("SetupDataset.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupDataset_GetSize(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					AccountRecipient: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Property:         "Admin",
					Value:            "Welcome",
					MuchTime:         123,
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountBalanceQuery: nil,
				DatasetQuery:        nil,
				QueryExecutor:       nil,
			},
			want: 124,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("SetupDataset.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupDataset_GetBodyBytes(t *testing.T) {
	type fields struct {
		Body                *model.SetupDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		Height              uint32
		AccountBalanceQuery query.AccountBalanceQueryInterface
		DatasetQuery        query.DatasetsQueryInterface
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
				Body: &model.SetupDatasetTransactionBody{
					AccountSetter:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					AccountRecipient: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Property:         "Admin",
					Value:            "Welcome",
					MuchTime:         123,
				},
				Fee:                 1,
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountBalanceQuery: nil,
				DatasetQuery:        nil,
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
			tx := &SetupDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				Height:              tt.fields.Height,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				DatasetQuery:        tt.fields.DatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupDataset.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
