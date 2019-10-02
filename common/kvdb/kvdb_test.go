package kvdb

import (
	"os"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/database"

	"github.com/dgraph-io/badger"
)

func getMockedKVDb() *badger.DB {
	badgerDb := &database.BadgerDB{}
	path, specPath := "./testdata", "foo"
	_ = badgerDb.InitializeBadgerDB(path, specPath)
	db, err := (&database.BadgerDB{}).OpenBadgerDB(path, specPath)
	if err != nil {
		panic(err)
	}
	return db
}

func cleanUpTestData() {
	_ = os.RemoveAll("./testdata/foo")
}

func TestKVExecutor_BatchInsert(t *testing.T) {
	t.Run("batch insert success", func(t *testing.T) {
		mockKvDb := getMockedKVDb()
		mockData := map[string][]byte{
			"people_john": []byte("doe"),
		}
		mockExecutor := &KVExecutor{Db: mockKvDb}
		// insert first
		err := mockExecutor.BatchInsert(mockData)
		if err != nil {
			t.Error("should be success")
		}
		result, err := mockExecutor.GetByPrefix("people_")
		if err != nil {
			t.Error("kvdb.GetByPrefix should not return error")
		}
		if !reflect.DeepEqual(mockData, result) {
			t.Error("get by prefix does not return the same data")
		}
		defer cleanUpTestData()

	})
}

func TestKVExecutor_Get(t *testing.T) {
	t.Run("key not found", func(t *testing.T) {
		mockKvDb := getMockedKVDb()
		defer cleanUpTestData()
		mockExecutor := &KVExecutor{Db: mockKvDb}
		_, err := mockExecutor.Get("bar")
		if err == nil {
			t.Error("should return key not found")
		}
	})
	t.Run("success", func(t *testing.T) {
		mockKvDb := getMockedKVDb()
		mockExecutor := &KVExecutor{Db: mockKvDb}
		// insert first
		_ = mockExecutor.Insert("bar", []byte{1, 1, 1, 1}, 60)
		res, err := mockExecutor.Get("bar")
		if err != nil {
			t.Error("should return key not found")
		}
		if !reflect.DeepEqual(res, []byte{1, 1, 1, 1}) {
			t.Error("inserted value and fetched value does not match")
		}
		defer cleanUpTestData()
	})
}

func TestKVExecutor_Insert(t *testing.T) {
	t.Run("success insert", func(t *testing.T) {
		mockKvDb := getMockedKVDb()
		mockExecutor := &KVExecutor{Db: mockKvDb}
		// insert first
		err := mockExecutor.Insert("bar", []byte{1, 1, 1, 1}, 60)
		if err != nil {
			t.Error("should return key not found")
		}
		defer cleanUpTestData()
	})
}

func TestNewKVExecutor(t *testing.T) {
	type args struct {
		db *badger.DB
	}
	tests := []struct {
		name string
		args args
		want *KVExecutor
	}{
		{
			name: "NewKVExecutor:success",
			args: args{db: nil},
			want: &KVExecutor{Db: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKVExecutor(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKVExecutor() = %v, want %v", got, tt.want)
			}
		})
	}
}
