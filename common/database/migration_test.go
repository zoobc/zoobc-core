package database

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestMigration_Init(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	type fields struct {
		Versions []string
		Query    *query.Executor
	}
	type args struct {
		query *query.Executor
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
				Versions: []string{
					`CREATE TABLE IF NOT EXISTS "migration" (
						"version" INTEGER DEFAULT 0 NOT NULL,
						"created_date" TIMESTAMP NOT NULL
					);`,
				},
				Query: query.NewQueryExecutor(db),
			},
			args: args{
				query: query.NewQueryExecutor(db),
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
			args: args{
				query: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				Versions: tt.fields.Versions,
				Query:    tt.fields.Query,
			}
			if err := m.Init(tt.args.query); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigration_Apply(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(`
		CREATE TABLE IF NOT EXISTS "accounts" (
			id	INTEGER,
			public_key	BLOB  NOT NULL,
			PRIMARY KEY("public_key")
		);
	`)).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "migration" (
			"version",
			"created_date"
		)
		VALUES (
			0,
			datetime('now')
		);
	`)).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	type fields struct {
		Versions []string
		Query    *query.Executor
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
				},
				Query: query.NewQueryExecutor(db),
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
			if err := m.Apply(); (err != nil) != tt.wantErr {
				t.Errorf("Migration.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Migration.Apply() query: %s, want: %s", tt.fields.Versions[0], err)
			}
		})
	}
}
