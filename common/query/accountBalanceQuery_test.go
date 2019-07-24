package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewAccountBalanceQuery(t *testing.T) {
	tests := []struct {
		name string
		want *AccountBalanceQuery
	}{
		{
			name: "NewAccountBalance:success",
			want: NewAccountBalanceQuery(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountBalanceQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountBalanceQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

var mockAccountBalanceQuery = &AccountBalanceQuery{
	Fields: []string{
		"account_id",
		"block_height",
		"spendable_balance",
		"balance",
		"pop_revenue",
		"latest",
	},
	TableName: "account_balance",
}

var causedFields = map[string]interface{}{
	"account_id":   []byte{1},
	"block_height": 1,
}

var mockAccountBalance = &model.AccountBalance{
	AccountID:        []byte{1},
	BlockHeight:      0,
	SpendableBalance: 0,
	Balance:          0,
	PopRevenue:       0,
	Latest:           true,
}

func TestAccountBalanceQuery_GetAccountBalanceByAccountID(t *testing.T) {
	t.Run("GetAccountBalanceByAccountID", func(t *testing.T) {
		res := mockAccountBalanceQuery.GetAccountBalanceByAccountID()
		want := "SELECT account_id,block_height,spendable_balance,balance,pop_revenue,latest " +
			"FROM account_balance WHERE account_id = ? AND latest = 1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestAccountBalanceQuery_AddAccountBalance(t *testing.T) {
	t.Run("AddAccountBalance", func(t *testing.T) {

		res := mockAccountBalanceQuery.AddAccountBalance(100, causedFields)
		var want [][]interface{}
		want = append(want, []interface{}{
			"UPDATE account_balance SET latest = false WHERE account_id = ? AND block_height = 1 - 1 AND latest = true",
			causedFields["account_id"],
		}, []interface{}{
			"INSERT INTO account_balance (account_id, block_height, spendable_balance, balance, pop_revenue, latest) VALUES " +
				"(?, 1, 100, 100, 0, true) ON CONFLICT(account_id, block_height) DO UPDATE SET spendable_balance = spendable_balance " +
				"+ 100, balance = balance + 100", causedFields["account_id"],
		})
		if !reflect.DeepEqual(res, want) {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestAccountBalanceQuery_AddAccountSpendableBalance(t *testing.T) {
	t.Run("AddAccountSpendableBalance:succes", func(t *testing.T) {
		q, args := mockAccountBalanceQuery.AddAccountSpendableBalance(100, causedFields)
		wantQ := "UPDATE account_balance SET spendable_balance = spendable_balance + (100) WHERE account_id = ?"
		wantArg := []interface{}{
			causedFields["account_id"],
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})

}

func TestAccountBalanceQuery_InsertAccountBalance(t *testing.T) {
	t.Run("InsertAccountBalance:success", func(t *testing.T) {

		q, args := mockAccountBalanceQuery.InsertAccountBalance(mockAccountBalance)
		wantQ := "INSERT INTO account_balance (account_id,block_height,spendable_balance,balance,pop_revenue,latest) VALUES(? , ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockAccountBalance.AccountID, mockAccountBalance.BlockHeight, mockAccountBalance.SpendableBalance, mockAccountBalance.Balance,
			mockAccountBalance.PopRevenue, true,
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestAccountBalanceQuery_ExtractModel(t *testing.T) {
	t.Run("ExtractModel:success", func(t *testing.T) {
		res := mockAccountBalanceQuery.ExtractModel(mockAccountBalance)
		want := []interface{}{
			mockAccountBalance.AccountID, mockAccountBalance.BlockHeight, mockAccountBalance.SpendableBalance, mockAccountBalance.Balance,
			mockAccountBalance.PopRevenue, true,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}
