package database

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
)

var (
	badgerConn       *badger.DB
	badgerDbInstance *BadgerDB
)

type (
	// BadgerDBInstance as public interface that should implemented
	BadgerDBInstance interface {
		InitializeBadgerDB(dbPath, dbName string) error
		OpenBadgerDB(dbPath, dbName string) (*badger.DB, error)
		CloseBadgerDB() error
	}
	BadgerDB struct{}
)

func InstrumentBadgerMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badgerMetrics := make(map[string]float64)

		expvar.Do(func(kv expvar.KeyValue) {
			isBadger := strings.Contains(kv.Key, "badger")
			if isBadger {
				parsedValue, err := strconv.ParseFloat(kv.Value.String(), 64)
				if err == nil {
					badgerMetrics[kv.Key] = parsedValue
				}
			}
		})
		monitoring.SetBadgerMetrics(badgerMetrics)
		next.ServeHTTP(w, r)
	})
}

// NewBadgerDB create new / fetch existing singleton BadgerDB instance.
func NewBadgerDB() *BadgerDB {
	if badgerDbInstance == nil {
		badgerDbInstance = &BadgerDB{}
	}
	return badgerDbInstance
}

/*
InitializeBadgerDB initialize badger database file from given dbPath and dbName
if dbPath not exist create given dbPath
if dbName / file not exist, create file with given dbName
return nil if dbPath/dbName exist
*/
func (bdb *BadgerDB) InitializeBadgerDB(dbPath, dbName string) error {
	_, err := os.Stat(dbPath)
	if ok := os.IsNotExist(err); ok {
		return err
	}
	_, err = os.Stat(fmt.Sprintf("%s/%s", dbPath, dbName))
	if ok := os.IsNotExist(err); ok {
		err = os.Mkdir(fmt.Sprintf("%s/%s", dbPath, dbName), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// OpenBadgerDB will open badger db connection
func (bdb *BadgerDB) OpenBadgerDB(dbPath, dbName string) (*badger.DB, error) {
	logFile, err := os.OpenFile(filepath.Join(dbPath, dbName, "badger.log"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logrus.SetOutput(logFile)

	opts := badger.DefaultOptions(filepath.Join(dbPath, dbName))
	// avoid memory-mapping log files
	opts.TableLoadingMode = options.FileIO
	// limit the in-memory log filesize
	opts.ValueLogFileSize = 1<<28 - 1
	// limit in-memory db entries
	opts.ValueLogMaxEntries = 250000
	// direct badger's log to file stored in ./badgerFolder/badger.log
	opts.Logger = logger
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
