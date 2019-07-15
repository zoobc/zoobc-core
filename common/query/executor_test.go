package query

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestExecutor_ExecuteTransactions(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(`
		CREATE TABLE IF NOT EXISTS "test" (
			"version" INTEGER DEFAULT 0 NOT NULL,
			"created_date" TIMESTAMP NOT NULL
		);
	`)).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	type fields struct {
		Db *sql.DB
	}
	type args struct {
		queries []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields{db},
			args: args{
				queries: []string{
					`CREATE TABLE IF NOT EXISTS "test" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qe := &Executor{
				Db: tt.fields.Db,
			}
			got, err := qe.ExecuteTransactions(tt.args.queries)
			if (err != nil) != tt.wantErr {
				t.Fatal(got, err)
				t.Errorf("Executor.ExecuteTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

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
			_, err := qe.ExecuteSelect(tt.args.query)
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
		executor := Executor{db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err == nil {
			t.Error("should return error if prepare fail")
		}
	})
	t.Run("ExecFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1, []byte{1, 2, 34}).WillReturnError(errors.New("mockError"))
		executor := Executor{db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err == nil {
			t.Error("should return error if exec fail")
		}
	})
	t.Run("Success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").ExpectExec().WithArgs(1, []byte{1, 2, 34}).WillReturnResult(sqlmock.NewResult(1, 1))
		executor := Executor{db}
		_, err := executor.ExecuteStatement("insert into blocks(id, blocksmith_id) values(?, ?)", 1, []byte{1, 2, 34})
		if err != nil {
			t.Error("should return error if exec fail")
		}
	})

}

func TestExecutor_ExecuteTransactionStatements(t *testing.T) {
	const insertBlockQuery = "insert into blocks(id, blocksmith_id) values(?, ?)"
	t.Run("PrepareFail", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectPrepare("insert into").WillReturnError(errors.New("mockError"))
		queries := make(map[*string][]interface{})
		insertBlock := "insert into blocks(id, blocksmith_id) values(?, ?)"
		queries[&insertBlock] = []interface{}{1, []byte{1, 2, 34}}
		// test error prepare
		executor := Executor{db}
		_, err := executor.ExecuteTransactionStatements(queries)
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

		queries := make(map[*string][]interface{})
		insertBlock := insertBlockQuery
		insertBlockAgain := insertBlockQuery
		queries[&insertBlock] = []interface{}{1, []byte{1, 2, 34}}
		queries[&insertBlockAgain] = []interface{}{1, []byte{1, 2, 14}}
		// test error prepare
		executor := Executor{db}
		_, err := executor.ExecuteTransactionStatements(queries)
		if err != nil {
			t.Error("transaction should have been committed without error")
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

		queries := make(map[*string][]interface{})
		insertBlock := insertBlockQuery
		insertBlockAgain := insertBlockQuery
		queries[&insertBlock] = []interface{}{1, []byte{1, 2, 34}}
		queries[&insertBlockAgain] = []interface{}{1, []byte{1, 2, 14}}
		// test error prepare
		executor := Executor{db}
		_, err := executor.ExecuteTransactionStatements(queries)
		if err == nil {
			t.Error("should return error if exec fail")
		}
	})
}
