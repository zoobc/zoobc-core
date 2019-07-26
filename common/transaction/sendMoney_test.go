package transaction

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var db, mock, _ = sqlmock.New()

type (
	executorAccountCreateSuccess struct {
		query.Executor
	}
	executorAccountCountSuccess struct {
		query.Executor
	}
	executorAccountCountFail struct {
		query.Executor
	}

	executorValidateSuccess struct {
		query.Executor
	}
	executorApplySuccess struct {
		executorValidateSuccess
	}

	executorFailUpdateAccount struct {
		executorAccountCountSuccess
	}

	executorSuccessUpdateAccount struct {
		query.Executor
	}
)

func (*executorApplySuccess) ExecuteTransactionStatements(queries [][]interface{}) ([]sql.Result, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("")).ExpectExec().
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("")
	result, _ := stmt.Exec("")
	err = tx.Commit()
	return []sql.Result{result}, err
}

func (*executorValidateSuccess) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"total_record",
	}).AddRow(2))

	return db.QueryRow(qStr, 1, 2)
}
func (*executorValidateSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
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
func (*executorValidateSuccess) ExecuteTransactionStatements(queries [][]interface{}) ([]sql.Result, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("")).ExpectExec().
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("")
	result, _ := stmt.Exec("")
	err = tx.Commit()
	return []sql.Result{result}, err
}

func (*executorAccountCreateSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorAccountCreateSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorAccountCreateSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewAccountQuery().Fields,
	))
	return db.Query(qStr, 1)
}
func (*executorAccountCreateSuccess) ExecuteTransactionStatements(queries [][]interface{}) ([]sql.Result, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("")).ExpectExec().
		WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("")
	result, _ := stmt.Exec("")
	err = tx.Commit()
	return []sql.Result{result}, err
}
func (*executorAccountCreateSuccess) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"total_record",
	}).AddRow(2))

	return db.QueryRow(qStr, 1, 2)
}

func (*executorAccountCountFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:accountInsertFail")
}

func (*executorAccountCountFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockError:accountInsertFail")
}

func (*executorFailUpdateAccount) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:accountbalanceFail")
}

func (*executorFailUpdateAccount) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockError:senderFail")
}

func (*executorSuccessUpdateAccount) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSuccessUpdateAccount) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorAccountCountSuccess) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"total_record",
	}).AddRow(2))

	return db.QueryRow(qStr, 1, 2)
}
func (*executorAccountCountSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewAccountBalanceQuery().Fields,
	).AddRow(1, 2, 3, 4, 5, 6))
	return db.Query(qStr, 1)
}

func (*executorAccountCountSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func TestSendMoney_Validate(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceQueryInterface
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:AmountNotEnough",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:RecipientInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				RecipientAddress:     "",
				RecipientAccountType: 0,
			},
			wantErr: true,
		},
		{
			name: "wantError:SenderInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			wantErr: true,
		},
		{
			name: "wantError:SenderNotExists",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCreateSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:CountAccountGot0",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCountFail{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:AmountNotEnough",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCountSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorValidateSuccess{
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
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if err := tx.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSendMoney_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceQueryInterface
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:ValidateInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCountSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:ApplySuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorApplySuccess{
					executorValidateSuccess{
						query.Executor{
							Db: db,
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSendMoney_ApplyConfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceQueryInterface
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:ValidateInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCountSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:UndoUnconfirmedInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor: &executorAccountCountFail{
					query.Executor{
						Db: db,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantFail:deductAddMoneyFail",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        &executorFailUpdateAccount{},
			},
			wantErr: true,
		},
		{
			name: "wantsuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        &executorSuccessUpdateAccount{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMoney_GetAmount(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceQueryInterface
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 100,
				},
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        &executorSuccessUpdateAccount{},
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("SendMoney.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMoney_GetSize(t *testing.T) {
	t.Run("SendMoney:GetSize", func(t *testing.T) {
		tx := &SendMoney{}
		size := tx.GetSize()
		if size != 8 {
			t.Errorf("SendMoney size should be 8\nget: %d instead", size)
		}
	})
}

func TestSendMoney_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceQueryInterface
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        &executorAccountCountSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:executeTransactionFail/",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        &executorAccountCountFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
