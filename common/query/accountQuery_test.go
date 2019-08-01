package query

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewAccountQuery(t *testing.T) {
	tests := []struct {
		name string
		want *AccountQuery
	}{
		{
			name: "NewAccountQuery:success",
			want: NewAccountQuery(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

var mockAccountQuery = &AccountQuery{
	Fields:    []string{"id", "account_type", "address"},
	TableName: "account",
}

var mockAccount = &model.Account{
	ID:          []byte{1},
	AccountType: 0,
	Address:     "bar",
}

func TestAccountQuery_GetAccountByID(t *testing.T) {
	t.Run("GetAccountByID:success", func(t *testing.T) {
		q, args := mockAccountQuery.GetAccountByID(mockAccount.ID)
		wantQ := "SELECT id, account_type, address FROM account WHERE id = ?"
		wantArg := []interface{}{
			mockAccount.ID,
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestAccountQuery_GetAccountByIDs(t *testing.T) {
	t.Run("GetAccountByIDs:success", func(t *testing.T) {
		argIn := [][]byte{mockAccount.ID, {2}}
		q, args := mockAccountQuery.GetAccountByIDs(argIn)
		wantQ := "SELECT id,account_type,address FROM account WHERE id in (? ,?)"
		wantArg := []interface{}{
			mockAccount.ID, []byte{2},
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestAccountQuery_InsertAccount(t *testing.T) {
	t.Run("InsertAccount:success", func(t *testing.T) {
		q, args := mockAccountQuery.InsertAccount(mockAccount)
		wantQ := "INSERT OR IGNORE INTO account (id,account_type,address) VALUES(? , ?, ?)"
		wantArg := mockAccountQuery.ExtractModel(mockAccount)

		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestAccountQuery_ExtractModel(t *testing.T) {
	t.Run("AccountQuery-ExtractModel:success", func(t *testing.T) {
		res := mockAccountQuery.ExtractModel(mockAccount)
		want := []interface{}{
			mockAccount.ID, mockAccount.AccountType, mockAccount.Address,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}

func TestAccountQuery_BuildModel(t *testing.T) {
	t.Run("AccountQuery-BuildModel:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows([]string{
			"ID", "AccountType", "Address"}).
			AddRow(mockAccount.ID, mockAccount.AccountType, mockAccount.Address))
		rows, _ := db.Query("foo")
		var tempAccount []*model.Account
		res := mockAccountQuery.BuildModel(tempAccount, rows)
		if !reflect.DeepEqual(res[0], mockAccount) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockAccount)
		}
	})
}

func TestAccountQuery_GetTableName(t *testing.T) {
	t.Run("AccountQuery-GetTableName:success", func(t *testing.T) {
		res := mockAccountQuery.GetTableName()
		if res != mockAccountQuery.TableName {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockAccountQuery.TableName)
		}
	})
}
