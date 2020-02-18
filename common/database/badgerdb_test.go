package database

import (
	"fmt"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/model"
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

func TestBytesWrite(t *testing.T) {
	ExampleEncoder()
}

func ExampleEncoder() {
	var v2 []*model.Block
	v1 := []*model.Block{
		{
			Height:               10,
			Timestamp:            123464884,
			ID:                   -1,
			Version:              1,
			TotalCoinBase:        100,
			TotalFee:             1000,
			TotalAmount:          100000000,
			BlocksmithPublicKey:  make([]byte, 32),
			PayloadHash:          make([]byte, 64),
			PayloadLength:        1,
			CumulativeDifficulty: "23423897509472358325780",
		},
		{
			Height:               11,
			Timestamp:            46456466435645,
			ID:                   -1000,
			Version:              1,
			TotalCoinBase:        100,
			TotalFee:             1000,
			TotalAmount:          100000000,
			BlocksmithPublicKey:  make([]byte, 32),
			PayloadHash:          make([]byte, 64),
			PayloadLength:        1,
			CumulativeDifficulty: "539405843078458937593857",
		},
	}
	var b []byte = make([]byte, 0, 64)
	var h codec.Handle = new(codec.JsonHandle)
	var enc *codec.Encoder = codec.NewEncoderBytes(&b, h)
	var err error = enc.Encode(v1) // any of v1 ... v8
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%v\n", b)

	// ... assume b contains the bytes to decode from
	var dec *codec.Decoder = codec.NewDecoderBytes(b, h)
	err = dec.Decode(&v2) // v2 or v8, or a pointer to v1, v3, v4, v5, v6, v7
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%v\n", v2)
}
