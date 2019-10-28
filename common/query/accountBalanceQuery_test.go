package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

var (
	mockAccountBalanceQuery = NewAccountBalanceQuery()

	causedFields = map[string]interface{}{
		"account_address": "BCZ",
		"block_height":    uint32(1),
	}
	mockAccountBalance = &model.AccountBalance{
		AccountAddress:   "BCZ",
		BlockHeight:      0,
		SpendableBalance: 0,
		Balance:          0,
		PopRevenue:       0,
		Latest:           true,
	}
	mockAccountBalanceRow = []interface{}{
		"BCZ",
		1,
		100,
		10,
		0,
		true,
	}
)

var _ = mockAccountBalanceRow

func TestAccountBalanceQuery_GetAccountBalanceByAccountID(t *testing.T) {
	t.Run("GetAccountBalanceByAccountID", func(t *testing.T) {
		res, args := mockAccountBalanceQuery.GetAccountBalanceByAccountAddress("BCZ")
		want := "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest " +
			"FROM account_balance WHERE account_address = ? AND latest = 1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		wantArg := []interface{}{
			mockAccountBalance.AccountAddress,
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}

	})
}

func TestAccountBalanceQuery_AddAccountBalance(t *testing.T) {
	t.Run("AddAccountBalance", func(t *testing.T) {

		res := mockAccountBalanceQuery.AddAccountBalance(100, causedFields)
		var want [][]interface{}
		want = append(want, []interface{}{
			"INSERT INTO account_balance (account_address, block_height, spendable_balance, balance, pop_revenue, latest) SELECT ?, " +
				"1, 0, 0, 0, 1 WHERE NOT EXISTS (SELECT account_address FROM account_balance WHERE account_address = ?)",
			causedFields["account_address"], causedFields["account_address"],
		}, []interface{}{
			"INSERT INTO account_balance (account_address, block_height, spendable_balance, balance, pop_revenue, latest) SELECT account_address, " +
				"1, spendable_balance + 100, balance + 100, pop_revenue, latest FROM account_balance WHERE account_address = ? AND latest = 1 " +
				"ON CONFLICT(account_address, block_height) DO UPDATE SET (spendable_balance, balance) = (SELECT spendable_balance + 100, balance " +
				"+ 100 FROM account_balance WHERE account_address = ? AND latest = 1)",
			causedFields["account_address"], causedFields["account_address"],
		}, []interface{}{
			"UPDATE account_balance SET latest = false WHERE account_address = ? AND block_height != 1 AND latest = true",
			causedFields["account_address"],
		})
		if !reflect.DeepEqual(res, want) {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestAccountBalanceQuery_AddAccountSpendableBalance(t *testing.T) {
	t.Run("AddAccountSpendableBalance:success", func(t *testing.T) {
		q, args := mockAccountBalanceQuery.AddAccountSpendableBalance(100, causedFields)
		wantQ := "UPDATE account_balance SET spendable_balance = spendable_balance + (100) WHERE account_address = ?" +
			" AND latest = 1"
		wantArg := []interface{}{"BCZ"}
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
		wantQ := "INSERT INTO account_balance (account_address,block_height,spendable_balance,balance,pop_revenue,latest) " +
			"VALUES(? , ?, ?, ?, ?, ?)"
		wantArg := []interface{}{
			mockAccountBalance.AccountAddress, mockAccountBalance.BlockHeight, mockAccountBalance.SpendableBalance, mockAccountBalance.Balance,
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
			mockAccountBalance.AccountAddress, mockAccountBalance.BlockHeight, mockAccountBalance.SpendableBalance, mockAccountBalance.Balance,
			mockAccountBalance.PopRevenue, true,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}

func TestAccountBalanceQuery_BuildModel(t *testing.T) {
	t.Run("AccountBalanceQuery-BuildModel:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress", "BlockHeight", "SpendableBalance", "Balance", "PopRevenue", "Latest"}).
			AddRow(mockAccountBalance.AccountAddress, mockAccountBalance.BlockHeight, mockAccountBalance.SpendableBalance,
				mockAccountBalance.Balance, mockAccountBalance.PopRevenue, mockAccountBalance.Latest))
		rows, _ := db.Query("foo")
		var tempAccount []*model.AccountBalance
		res := mockAccountBalanceQuery.BuildModel(tempAccount, rows)
		if !reflect.DeepEqual(res[0], mockAccountBalance) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockAccountBalance)
		}
	})
}

func TestAccountBalanceQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountBalanceQuery),
			args:   args{height: uint32(1)},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM account_balance WHERE block_height > ?",
					uint32(1),
				},
				{`
			UPDATE account_balance SET latest = ?
			WHERE (block_height || '_' || account_address) IN (
				SELECT (MAX(block_height) || '_' || account_address) as con
				FROM account_balance
				WHERE latest = 0
				GROUP BY account_address
			)`,
					1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AccountBalanceQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := q.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollbacks() = \n%v, want \n%v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

type (
	mockRowBalanceQueryScan struct {
		Executor
	}
)

func (*mockRowBalanceQueryScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockAccountBalanceQuery.Fields).AddRow(
			"BCZ",
			1,
			100,
			10,
			0,
			true,
		),
	)
	return db.QueryRow("")
}

func TestAccountBalanceQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountBalance *model.AccountBalance
		row            *sql.Row
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountBalanceQuery),
			args: args{
				accountBalance: mockAccountBalance,
				row:            (&mockRowBalanceQueryScan{}).ExecuteSelectRow("", nil),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountBalanceQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := a.Scan(tt.args.accountBalance, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountBalanceQuery_GetAccountBalances(t *testing.T) {
	t.Run("GetAccountBalances", func(t *testing.T) {
		q := mockAccountBalanceQuery.GetAccountBalances()
		wantQ := "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE latest = 1"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}
