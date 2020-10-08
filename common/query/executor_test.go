package query

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	logrus2 "github.com/sirupsen/logrus"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestExecutor_Execute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	mock.ExpectExec("insert into product_viewers").WillReturnResult(sqlmock.NewResult(1, 1))
	type fields struct {
		Db *sql.DB
	}
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    sql.Result
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields{db},
			args: args{
				query: "insert into product_viewers (user_id, product_id) values (2, 3)",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "wantError",
			fields: fields{db},
			args: args{
				query: "insert wrong query into product_viewers (user_id, product_id) values (2, 3)",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qe := &Executor{
				Db: tt.fields.Db,
			}
			_, err := qe.Execute(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Executor.ExecuteQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExecutor_ExecuteSelect(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			version,
			created_date
		FROM migration limit 1;
	`)).WillReturnRows(sqlmock.NewRows([]string{"version", "created_date"}))

	type fields struct {
		Db *sql.DB
	}
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Db: db,
			},
			args: args{
				query: `SELECT 
					version,
					created_date
				FROM migration limit 1;
				`,
			},
			wantErr: false,
		},
		{
			name: "wantError",
			fields: fields{
				Db: db,
			},
			args: args{
				query: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qe := &Executor{
				Db: tt.fields.Db,
			}
			_, err := qe.ExecuteSelect(tt.args.query, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Executor.ExecuteSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExecutor_ExecuteStatement(t *testing.T) {
	t.Run("PrepareFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").WillReturnError(errors.New("mockError"))

		// test error prepare
		executor := Executor{Db: db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err == nil {
			t.Error("should return error if prepare fail")
		}
	})
	t.Run("ExecFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1, []byte{1, 2, 34}).WillReturnError(errors.New("mockError"))
		executor := Executor{Db: db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err == nil {
			t.Error("should return error if exec fail")
		}
	})
	t.Run("Success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1, []byte{1, 2, 34}).WillReturnResult(sqlmock.NewResult(1, 1))
		executor := Executor{Db: db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err != nil {
			t.Error("should return error if exec fail")
		}
	})

}

func TestExecutor_ExecuteSelectRow(t *testing.T) {
	type (
		fields struct {
			Db *sql.DB
		}
		args struct {
			query string
			args  []interface{}
		}
	)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error("failed while opening database connection")
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
		ELECT () FROM account
		WHERE id = ? AND name = ? limit 1
	`)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"field",
	}).AddRow(1))
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Db: db,
			},
			args: args{
				query: "SELECT () FROM account WHERE id = ? AND name = ? limit 1",
				args: []interface{}{
					1, 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qe := &Executor{
				Db: tt.fields.Db,
			}
			_, _ = qe.ExecuteSelectRow(tt.args.query, false, tt.args.args...)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Executor.ExecuteSelectRow() = %v", err)
			}
		})
	}
}

func TestExecutor_BeginTx(t *testing.T) {
	t.Run("fail:beginFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin().WillReturnError(errors.New("mockError:beginFail"))
		executor := Executor{Db: db}
		err := executor.BeginTx()
		if err == nil {
			t.Errorf("begin tx should fail:begin fail")
		}
	})
	t.Run("success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		executor := Executor{Db: db}
		err := executor.BeginTx()
		if err != nil {
			t.Errorf("begin tx should not fail")
		}
	})
}

func TestExecutor_CommitTx(t *testing.T) {
	t.Run("fail:commitFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(errors.New("mockError:commitFail"))

		executor := Executor{
			Db: db,
		}
		_ = executor.BeginTx()
		err := executor.CommitTx()
		if err == nil {
			t.Errorf("commit tx should fail : commit fail")
		}
	})
	t.Run("success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectCommit()

		executor := Executor{
			Db: db,
		}
		_ = executor.BeginTx()
		err := executor.CommitTx()
		if err != nil {
			t.Errorf("commit tx should not return error")
		}
	})
}

func TestExecutor_RollbackTx(t *testing.T) {
	t.Run("fail:rollbackError", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(errors.New("mockError:rollbackFail"))

		executor := Executor{
			Db: db,
		}
		_ = executor.BeginTx()
		err := executor.RollbackTx()
		if err == nil {
			t.Errorf("rollback should return error")
		}
	})
	t.Run("success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectRollback()

		executor := Executor{
			Db: db,
		}
		_ = executor.BeginTx()
		err := executor.RollbackTx()
		if err != nil {
			t.Errorf("rollback should not return error")
		}
	})
}

func TestNewQueryExecutor(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want *Executor
	}{
		{
			name: "NewQueryExecutor:success",
			args: args{
				db: nil,
			},
			want: &Executor{
				Db: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryExecutor(tt.args.db, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_ExecuteTransaction(t *testing.T) {
	t.Run("ExecuteTransaction:fail-{prepareFail}", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		executor := NewQueryExecutor(db, logrus2.New())
		mock.ExpectBegin()
		mock.ExpectPrepare("fail prepare").WillReturnError(errors.New("mockError:prepareFail"))
		_ = executor.BeginTx()
		err := executor.ExecuteTransaction("fail prepare")
		if err == nil {
			t.Error("prepare should have failed the whole function")
		}
	})
	t.Run("ExecuteTransaction:fail-{execFail}", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		executor := NewQueryExecutor(db, logrus2.New())
		mock.ExpectBegin()
		mock.ExpectPrepare("fail exec")
		mock.ExpectExec("fail exec").WillReturnError(errors.New("mockError:execFail"))
		_ = executor.BeginTx()
		err := executor.ExecuteTransaction("fail exec")
		if err == nil {
			t.Error("exec should have failed the whole function")
		}
	})
	t.Run("ExecuteTransaction:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		executor := NewQueryExecutor(db, logrus2.New())
		mock.ExpectBegin()
		mock.ExpectPrepare("success")
		mock.ExpectExec("success").WillReturnResult(sqlmock.NewResult(1, 1))
		_ = executor.BeginTx()
		err := executor.ExecuteTransaction("success")
		if err != nil {
			t.Errorf("function should return nil if prepare and exec success\nreturned: %v instead", err)
		}
	})
}

func TestExecutor_ExecuteTransactions(t *testing.T) {
	const insertBlockQuery = "insert into blocks(id, blocksmith_id) values(?, ?)"
	t.Run("PrepareFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectPrepare("insert into").WillReturnError(errors.New("mockError"))
		queries := [][]interface{}{
			{
				insertBlockQuery, 1, []byte{1, 2, 34},
			},
		}
		// test error prepare
		executor := Executor{Db: db}
		_ = executor.BeginTx()
		err := executor.ExecuteTransactions(queries)
		if err == nil {
			t.Error("should return error if prepare fail")
		}
	})
	t.Run("MultipleIdenticalQuery:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1,
			[]byte{1, 2, 34}).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1,
			[]byte{1, 2, 14}).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectClose()
		var queries [][]interface{}
		queries = append(queries, []interface{}{
			insertBlockQuery, 1, []byte{1, 2, 34},
		}, []interface{}{
			insertBlockQuery, 1, []byte{1, 2, 14},
		})
		// test error prepare
		executor := Executor{Db: db}
		_ = executor.BeginTx()
		err := executor.ExecuteTransactions(queries)
		if err != nil {
			t.Errorf("transaction should have been committed without error: %v", err)
		}
	})
	t.Run("MultipleIdenticalQuery:execFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1,
			[]byte{1, 2, 34}).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1,
			[]byte{1, 2, 14}).WillReturnError(errors.New("mockError"))

		var queries [][]interface{}
		queries = append(queries, []interface{}{
			insertBlockQuery, 1, []byte{1, 2, 34},
		}, []interface{}{
			insertBlockQuery, 1, []byte{1, 2, 14},
		})
		// test error prepare
		executor := Executor{Db: db}
		_ = executor.BeginTx()
		err := executor.ExecuteTransactions(queries)
		if err == nil {
			t.Error("should return error if exec fail")
		}
	})
}
