package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-sqlite3"
)

type (
	// SqliteDBInstance as public interface that should implemented
	SqliteDBInstance interface {
		InitializeDB(dbPath, dbName string) error
		OpenDB(dbPath, dbName string, maxIdleConnections, maxOpenConnections int) (*sql.DB, error)
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
func (db *SqliteDB) OpenDB(dbPath, dbName string, maximumIdleConnections, maximumOpenConnections int) (*sql.DB, error) {
	var (
		err error
	)
	_, err = os.Stat(dbPath + dbName)
	if os.IsNotExist(err) {
		return nil, err
	}

	conn, err = sql.Open("sqlite3", dbPath+dbName)

	if _, ok := err.(sqlite3.Error); ok {
		return nil, err
	}

	conn.SetMaxIdleConns(maximumIdleConnections)
	conn.SetMaxOpenConns(maximumOpenConnections)

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
