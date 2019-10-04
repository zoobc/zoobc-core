package database

import (
	"os"
	"path/filepath"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/dgraph-io/badger"
)

var (
	badgerConn       *badger.DB
	badgerDbInstance *BadgerDB
)

type (
	// SqliteDBInstance as public interface that should implemented
	BadgerDBInstance interface {
		InitializeBadgerDB(dbPath, dbName string) error
		OpenBadgerDB(dbPath, dbName string) (*badger.DB, error)
		CloseBadgerDB() error
	}
	BadgerDB struct{}
)

// NewBadgerDB create new / fetch existing singleton BadgerDB instance.
func NewBadgerDB() *BadgerDB {
	if badgerDbInstance == nil {
		badgerDbInstance = &BadgerDB{}
	}
	return badgerDbInstance
}

/*
InitializeDB initialize badger database file from given dbPath and dbName
if dbPath not exist create given dbPath
if dbName / file not exist, create file with given dbName
return nil if dbPath/dbName exist
*/
func (bdb *BadgerDB) InitializeBadgerDB(dbPath, dbName string) error {
	_, err := os.Stat(dbPath)
	if ok := os.IsNotExist(err); ok {
		return err
	}
	_, err = os.Stat(filepath.Join(dbPath, dbName))
	if ok := os.IsNotExist(err); ok {
		err = os.Mkdir(filepath.Join(dbPath, dbName), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bdb *BadgerDB) OpenBadgerDB(dbPath, dbName string) (*badger.DB, error) {
	opts := badger.DefaultOptions(filepath.Join(dbPath, dbName)).WithValueLogFileSize(1<<28 - 1).WithValueLogMaxEntries(250000)
	badgerConn, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return badgerConn, nil
}

func (bdb *BadgerDB) CloseBadgerDB() error {
	if badgerConn == nil {
		return blocker.NewBlocker(
			blocker.DBErr,
			"Badger DB failed to close")
	}
	err := badgerConn.Close()
	conn = nil // mutate the sqliteDBInstance : close the connection
	if err != nil {
		return err
	}
	return nil
}
