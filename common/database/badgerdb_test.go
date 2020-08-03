package database

import (
	"os"
	"reflect"
	"testing"
)

func cleanUpTestData() {
	_ = os.RemoveAll("./testdata/foo")
}

func TestBadgerDB_CloseBadgerDB(t *testing.T) {
	t.Run("badgerConn-nil", func(t *testing.T) {
		badgerConn = nil
		badgerDb := &BadgerDB{}
		err := badgerDb.CloseBadgerDB()
		if err == nil {
			t.Error("badgerConn nil should cause close badger db to return error")
		}
		cleanUpTestData()
	})
	t.Run("badgerConn-close success", func(t *testing.T) {
		badgerConn = nil
		badgerDb := &BadgerDB{}
		_ = badgerDb.InitializeBadgerDB("./testdata", "foo")
		badgerConn, _ = badgerDb.OpenBadgerDB("./testdata", "foo")
		err := badgerDb.CloseBadgerDB()
		if err != nil {
			t.Error("badgerConn nil should cause close badger db to return error")
		}
		cleanUpTestData()
	})
}

func TestBadgerDB_InitializeBadgerDB(t *testing.T) {
	type args struct {
		dbPath string
		dbName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "IntializeBadgerDB:fail - path not exist",
			args: args{
				dbPath: "",
				dbName: "",
			},
			wantErr: true,
		},
		{
			name: "InitializeBadgerDB:success - create new directory",
			args: args{
				dbPath: "./testdata",
				dbName: "foo",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdb := &BadgerDB{}
			if err := bdb.InitializeBadgerDB(tt.args.dbPath, tt.args.dbName); (err != nil) != tt.wantErr {
				t.Errorf("InitializeBadgerDB() error = %v, wantErr %v", err, tt.wantErr)
			}

			cleanUpTestData()
		})
	}
}

func TestBadgerDB_OpenBadgerDB(t *testing.T) {
	type args struct {
		dbPath string
		dbName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OpenBadgerDB:fail - provide file path instead of path",
			args: args{
				dbPath: "./testdata/random/path",
				dbName: "zoobc.badger",
			},
			wantErr: true,
		},
		{
			name: "OpenBadgerDB:success",
			args: args{
				dbPath: "./testdata",
				dbName: "foo",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdb := &BadgerDB{}
			_ = bdb.InitializeBadgerDB(tt.args.dbPath, tt.args.dbName)
			conn, err := bdb.OpenBadgerDB(tt.args.dbPath, tt.args.dbName)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenBadgerDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if conn != nil {
				conn.Close()
			}
			defer cleanUpTestData()
		})
	}
}

func TestNewBadgerDB(t *testing.T) {
	tests := []struct {
		name string
		want *BadgerDB
	}{
		{
			name: "NewBadgerDB:success",
			want: &BadgerDB{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBadgerDB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBadgerDB() = %v, want %v", got, tt.want)
			}
		})
	}
}
