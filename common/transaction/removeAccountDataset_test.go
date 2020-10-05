package transaction

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	executorRemoveAccountDatasetApplyConfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetApplyConfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetApplyConfirmedSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyConfirmedFail) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorRemoveAccountDatasetApplyConfirmedFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_ApplyConfirmed(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
		Fee                  int64
		SenderAddress        string
		RecipientAddress     string
		Height               uint32
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Body:                mockRemoveAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:UndoUnconfirmedFail",
			fields: fields{
				Body:                &model.RemoveAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              3,
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionsFail",
			fields: fields{
				Body:                &model.RemoveAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              0,
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor: &executorRemoveAccountDatasetApplyConfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetApplyUnconfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetApplyUnconfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetApplyUnconfirmedSuccess) ExecuteSelect(qStr string, _ bool,
	_ ...interface{}) (*sql.Rows, error) {
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

func (*executorRemoveAccountDatasetApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorRemoveAccountDatasetApplyUnconfirmedFail) ExecuteSelect(qStr string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
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

func (*executorRemoveAccountDatasetApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_ApplyUnconfirmed(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()
	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
		Fee                  int64
		SenderAddress        string
		RecipientAddress     string
		Height               uint32
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Body:                mockRemoveAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetApplyUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantErr:ExecuteSpendableBalanceFail",
			fields: fields{
				Body:                mockRemoveAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetApplyUnconfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetUndoUnconfirmedSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetUndoUnconfirmedFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetUndoUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*executorRemoveAccountDatasetUndoUnconfirmedFail) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("MockedError")
}

func TestRemoveAccountDataset_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
		Fee                  int64
		SenderAddress        string
		RecipientAddress     string
		Height               uint32
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				Body:                &model.RemoveAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "",
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetUndoUnconfirmedSuccess{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:fail",
			fields: fields{
				Body:                &model.RemoveAccountDatasetTransactionBody{},
				Fee:                 1,
				SenderAddress:       "",
				AccountDatasetQuery: nil,
				QueryExecutor: &executorRemoveAccountDatasetUndoUnconfirmedFail{
					query.Executor{
						Db: db,
					},
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	executorRemoveAccountDatasetValidateSuccess struct {
		query.Executor
	}
	executorRemoveAccountDatasetValidateFail struct {
		query.Executor
	}
)

func (*executorRemoveAccountDatasetValidateSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
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

func (*executorRemoveAccountDatasetValidateFail) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*executorRemoveAccountDatasetValidateFail) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewAccountDatasetsQuery().Fields),
	)

	return db.QueryRow(qStr), nil
}

func TestRemoveAccountDataset_Validate(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                 *model.RemoveAccountDatasetTransactionBody
		Fee                  int64
		SenderAddress        string
		RecipientAddress     string
		Height               uint32
		AccountDatasetQuery  query.AccountDatasetQueryInterface
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:success",
			fields: fields{
				Body:                 mockRemoveAccountDatasetTransactionBody,
				Fee:                  1,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:     "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorRemoveAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "Validate:BalanceNotEnough",
			fields: fields{
				Body:                 mockRemoveAccountDatasetTransactionBody,
				Fee:                  60,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:     "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				QueryExecutor:        &executorRemoveAccountDatasetValidateSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "Validate:noRow",
			fields: fields{
				Body:                mockRemoveAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &executorRemoveAccountDatasetValidateFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                 tt.fields.Body,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				AccountDatasetQuery:  tt.fields.AccountDatasetQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("RemoveAccountDataset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveAccountDataset_GetSize(t *testing.T) {
	mockRemoveAccountDatasetTransactionBody, _ := GetFixturesForRemoveAccountDataset()

	type fields struct {
		Body                *model.RemoveAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		RecipientAddress    string
		Height              uint32
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
				Body:                mockRemoveAccountDatasetTransactionBody,
				Fee:                 1,
				SenderAddress:       "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				Height:              5,
				AccountDatasetQuery: nil,
				QueryExecutor:       nil,
			},
			want: 21,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &RemoveAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				RecipientAddress:    tt.fields.RecipientAddress,
				Height:              tt.fields.Height,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("RemoveAccountDataset.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveAccountDataset_GetTransactionBody(t *testing.T) {
	mockTxBody, _ := GetFixturesForRemoveAccountDataset()
	type fields struct {
		Body                *model.RemoveAccountDatasetTransactionBody
		Fee                 int64
		SenderAddress       string
		RecipientAddress    string
		Height              uint32
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
			tx := &RemoveAccountDataset{
				Body:                tt.fields.Body,
				Fee:                 tt.fields.Fee,
				SenderAddress:       tt.fields.SenderAddress,
				RecipientAddress:    tt.fields.RecipientAddress,
				Height:              tt.fields.Height,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}
