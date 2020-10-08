package database

import (
	"database/sql"
	"regexp"
	"testing"

	logrus2 "github.com/sirupsen/logrus"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorInitFailRows struct {
		query.Executor
	}
)

func (*mockExecutorInitFailRows) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {

	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(qe).WillReturnRows(
		sqlmock.NewRows([]string{"count"}),
	)
	return db.QueryRow(qe), nil
}

func (*mockExecutorInitFailRows) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Version",
	}).AddRow(1))

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestMigration_Init(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	type fields struct {
		Versions []string
		Query    query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: query.NewQueryExecutor(db, logrus2.New()),
			},
			wantErr: false,
		},
		{
			name: "wantError",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: nil,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:versionNotNil",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: &mockExecutorInitFailRows{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				Versions: tt.fields.Versions,
				Query:    tt.fields.Query,
			}
			if err := m.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var dbMock, mock, _ = sqlmock.New()

type (
	mockQueryExecutorVersionNil struct {
		query.Executor
	}
)

func (*mockQueryExecutorVersionNil) BeginTx() error {
	return nil
}
func (*mockQueryExecutorVersionNil) ExecuteTransaction(qStr string, args ...interface{}) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectPrepare(regexp.QuoteMeta(qStr)).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	stmt, err := db.Prepare(qStr)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(args...)
	return err
}
func (*mockQueryExecutorVersionNil) CommitTx() error {
	return nil
}
func TestMigration_Apply(t *testing.T) {
	currentVersion := 1

	type fields struct {
		CurrentVersion *int
		Versions       []string
		Query          query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "accounts" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
					`CREATE TABLE IF NOT EXISTS "accounts" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
				},
				CurrentVersion: &currentVersion,
				Query: &mockQueryExecutorVersionNil{
					query.Executor{Db: dbMock},
				},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:VersionNil",
			fields: fields{
				CurrentVersion: nil,
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "account" (
						id	INTEGER,
						public_key	BLOB  NOT NULL,
						PRIMARY KEY("public_key")
					);`,
				},
				Query: &mockQueryExecutorVersionNil{
					query.Executor{Db: dbMock},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				CurrentVersion: tt.fields.CurrentVersion,
				Versions:       tt.fields.Versions,
				Query:          tt.fields.Query,
			}
			if err := m.Apply(); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Apply() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Migration.Apply() query: \n%s, want: \n%s", tt.fields.Versions[0], err)
			}
		})
	}
}
