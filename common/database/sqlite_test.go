package database

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSqliteDB(t *testing.T) {
	tests := []struct {
		name string
		want *SqliteDB
	}{
		{
			name: "wantSuccess",
			want: &SqliteDB{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSqliteDB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSqliteDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqliteDB_InitializeDB(t *testing.T) {

	db, _, err := sqlmock.New()
	if err != nil {
		t.Errorf("error while opening database connection")
	}
	defer db.Close()

	type args struct {
		dbPath string
		dbName string
	}
	tests := []struct {
		name    string
		db      *SqliteDB
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			db:   &SqliteDB{},
			args: args{
				dbPath: "./testdata/",
				dbName: "zoobc_test.db",
			},
			wantErr: false,
		},
		{
			name: "wantError",
			db:   &SqliteDB{},
			args: args{
				dbPath: "",
				dbName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &SqliteDB{}
			if err := db.InitializeDB(tt.args.dbPath, tt.args.dbName); (err != nil) != tt.wantErr {
				t.Errorf("SqliteDB.InitializeDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSqliteDB_OpenDB(t *testing.T) {

	db, _, err := sqlmock.New()
	if err != nil {
		t.Errorf("error while opening database connection")
	}
	defer db.Close()

	type args struct {
		dbPath                 string
		dbName                 string
		maximumIdleConnections int
		maximumOpenConnection  int
		maximumOpenConnections time.Duration
	}
	tests := []struct {
		name    string
		db      *SqliteDB
		args    args
		want    *sql.DB
		wantErr bool
	}{
		{
			name: "wantSuccess",
			db:   &SqliteDB{},
			args: args{
				dbPath:                 "./testdata/",
				dbName:                 "zoobc_test.db",
				maximumOpenConnection:  10,
				maximumIdleConnections: 10,
				maximumOpenConnections: 10 * time.Second,
			},
			want:    db,
			wantErr: false,
		},
		{
			name: "wantError",
			db:   &SqliteDB{},
			args: args{
				dbPath:                 "_",
				dbName:                 "_",
				maximumOpenConnection:  10,
				maximumIdleConnections: 10,
				maximumOpenConnections: 10 * time.Second,
			},
			want:    db,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &SqliteDB{}
			got, err := db.OpenDB(
				tt.args.dbPath,
				tt.args.dbName,
				tt.args.maximumOpenConnection,
				tt.args.maximumIdleConnections,
				tt.args.maximumOpenConnections)

			if (err != nil) != tt.wantErr {
				t.Errorf("SqliteDB.OpenDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == tt.want {
				t.Errorf("SqliteDB.OpenDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqliteDB_CloseDB(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("error while opening database connection")
	}
	defer db.Close()

	mock.ExpectClose()

	conn = db

	tests := []struct {
		name    string
		db      *SqliteDB
		wantErr bool
	}{
		{
			name:    "wantSuccess",
			db:      &SqliteDB{},
			wantErr: false,
		},
		{
			name:    "wantError",
			db:      nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &SqliteDB{}
			if err := db.CloseDB(); (err != nil) != tt.wantErr {
				t.Errorf("SqliteDB.CloseDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
