package transaction

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	mockAccountBalanceQuery struct {
		query.AccountBalanceQuery
	}
)

func (m *mockAccountBalanceQuery) GetAccountBalanceByAccountID() string {
	return `SELECT account_id,block_height,spendable_balance,balance,pop_revenue
		FROM account_balance
		WHERE account_id = ?`
}

func TestSendMoney_Validate(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceInt
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errMsg  string
	}{
		{
			name: "wantSuccess:ReturnNil",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height: 0,
			},
			wantErr: false,
		},
		{
			name: "wantError:InvalidAmount",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: -1,
				},
				Height: 1,
			},
			wantErr: true,
			errMsg:  "transaction must have an amount more than 0",
		},
		{
			name: "wantError:InvalidRecipient",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:            1,
				SenderAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				SenderAccountType: 0,
			},
			wantErr: true,
			errMsg:  "transaction must have a valid recipient account id",
		},
		{
			name: "wantError:InvalidSender",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				SenderAccountType:    0,
				RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAccountType: 1,
			},
			wantErr: true,
			errMsg:  "transaction must have a valid sender account id",
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

			if err := tx.Validate(); err != nil {
				if !tt.wantErr {
					t.Errorf("SendMoney.Validate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if ok := strings.Compare(err.Error(), tt.errMsg); ok != 0 {
					t.Errorf("SendMoney.Validate() got errorMessage: %s, wantMessage: %s", err.Error(), tt.errMsg)
					return
				}
			}
		})
	}
	// case account not exists
	db, mock, _ := sqlmock.New()
	txAccountNotExists := SendMoney{
		Body: &model.SendMoneyTransactionBody{
			Amount: 10,
		},
		Height:               1,
		RecipientAccountType: 1,
		RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		SenderAccountType:    1,
		SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		QueryExecutor: query.ExecutorInterface(&query.Executor{
			Db: db,
		}),
		AccountBalanceQuery: (&mockAccountBalanceQuery{}).NewAccountBalanceQuery(),
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			account_id,block_height,spendable_balance,balance,pop_revenue
		FROM account_balance
		WHERE account_id = ?
	`)).WithArgs(util.CreateAccountIDFromAddress(
		txAccountNotExists.RecipientAccountType,
		txAccountNotExists.RecipientAddress,
	)).WillReturnRows(sqlmock.NewRows((&mockAccountBalanceQuery{}).NewAccountBalanceQuery().Fields))

	if err := txAccountNotExists.Validate(); (err != nil) && strings.Compare(err.Error(), "account not exists") != 0 {
		t.Error(err)
	}

	// case balance != amount
	txBalanceNotEnough := SendMoney{
		Body: &model.SendMoneyTransactionBody{
			Amount: 10,
		},
		Height:               1,
		RecipientAccountType: 1,
		RecipientAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		SenderAccountType:    1,
		SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		QueryExecutor: query.ExecutorInterface(&query.Executor{
			Db: db,
		}),
		AccountBalanceQuery: query.AccountBalanceInt(&mockAccountBalanceQuery{}),
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			account_id,block_height,spendable_balance,balance,pop_revenue
		FROM account_balance
		WHERE account_id = ?
	`)).WithArgs(util.CreateAccountIDFromAddress(
		txBalanceNotEnough.RecipientAccountType,
		txBalanceNotEnough.RecipientAddress,
	)).WillReturnRows(
		sqlmock.NewRows((&mockAccountBalanceQuery{}).NewAccountBalanceQuery().Fields).
			AddRow(
				util.CreateAccountIDFromAddress(
					txBalanceNotEnough.RecipientAccountType,
					txBalanceNotEnough.RecipientAddress,
				),
				1,
				5,
				10,
				0,
			),
	)
	if err := txBalanceNotEnough.Validate(); (err != nil) && strings.Compare(err.Error(), "transaction amount not enough") != 0 {
		t.Errorf("fields: %v", mockAccountBalanceQuery{}.Fields)
		t.Error(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("upss: %v", err)
	}
}
