package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-sqlite3"
)

type (
	// SqliteDBInstance as public interface that should implemented
	SqliteDBInstance interface {
		InitializeDB(dbPath, dbName string) error
		OpenDB(
			dbPath, dbName string,
			maximumOpenConnection, maxIdleConnections int,
			maximumLifetimeConnection time.Duration,
		) (*sql.DB, error)
		CloseDB() error
	}
	// SqliteDB must be implemented
	SqliteDB struct{}
)

var (
	conn       *sql.DB
	dbInstance *SqliteDB
)

// NewSqliteDB create new / fetch existing singleton SqliteDB instance.
func NewSqliteDB() *SqliteDB {
	if dbInstance == nil {
		dbInstance = &SqliteDB{}
	}

	return dbInstance
}

/*
InitializeDB initialize sqlite database file from given dbPath and dbName
if dbPath not exist create given dbPath
if dbName / file not exist, create file with given dbName
return nil if dbPath/dbName exist
*/
func (db *SqliteDB) InitializeDB(dbPath, dbName string) error {
	_, err := os.Stat(dbPath)
	if ok := os.IsNotExist(err); ok {
		return err
	}

	_, err = os.OpenFile(fmt.Sprintf("%s/%s", dbPath, dbName), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

/*
OpenDB tries to open the db and if fails logs and exit the application
mutate SqliteDB.Conn to opened connection if success and return nil
return error if error occurred
*/
func (db *SqliteDB) OpenDB(
	dbPath, dbName string,
	maximumOpenConnection, maximumIdleConnections int,
	maximumLifetimeConnection time.Duration,
) (*sql.DB, error) {
	var (
		err     error
		absPath string
	)

	absPath, err = filepath.Abs(filepath.Join(dbPath, dbName))
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return nil, err
	}

	// _mutex parameter for setup threading mode in mattn/go-sqlite3
	// no = SQLITE_OPEN_NOMUTEX, full = SQLITE_OPEN_FULLMUTEX
	conn, err = sql.Open("sqlite3", fmt.Sprintf("%s?_mutex=no", absPath))

	if _, ok := err.(sqlite3.Error); ok {
		return nil, err
	}
	// Higher number of idle connections in the pool will improve performance
	// But it will takes up memory usage
	conn.SetMaxIdleConns(maximumIdleConnections)
	// SetConnMaxLifetime used to controlling the lifecycle of connections,
	// Will be useful when maintaining idle connetions in low traffic
	// todo: andy-shi88 experiment with sql configuration for debugging cpu spike - remove this later
	conn.SetConnMaxLifetime(0)
	// SetMaxOpenConns the maximum number of open connections to the database
	// to prevent unable open database file
	conn.SetMaxOpenConns(maximumOpenConnection)
	return conn, nil
}

/*
CloseDB close database connection and set sqliteD.Conn to nil
return nil if success,
*/
func (db *SqliteDB) CloseDB() error {
	if conn == nil {
		return errors.New("database connection not opened")
	}
	err := conn.Close()
	conn = nil // mutate the sqliteDBInstance : close the connection
	if err != nil {
		return err
	}
	return nil
}
