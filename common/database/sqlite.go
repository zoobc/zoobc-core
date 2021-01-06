// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
	conn, err = sql.Open("sqlite3", fmt.Sprintf("%s?_mutex=no&cache=private&mode=rwc&journal_mode=WAL", absPath))

	if _, ok := err.(sqlite3.Error); ok {
		return nil, err
	}
	// Higher number of idle connections in the pool will improve performance
	// But it will takes up memory usage
	conn.SetMaxIdleConns(maximumIdleConnections)
	// SetConnMaxLifetime used to controlling the lifecycle of connections,
	// Will be useful when maintaining idle connetions in low traffic
	conn.SetConnMaxLifetime(maximumLifetimeConnection)
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
