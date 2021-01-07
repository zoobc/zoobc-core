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
package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/queue"
)

type (
	// ExecutorInterface interface
	ExecutorInterface interface {
		BeginTx(params ...int) error //STEF to test only, change with proper function param when done
		Execute(string) (sql.Result, error)
		ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error)
		ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error)
		ExecuteStatement(query string, args ...interface{}) (sql.Result, error)
		ExecuteTransaction(query string, args ...interface{}) error
		ExecuteTransactions(queries [][]interface{}) error
		// CommitTx commit on every transaction stacked in Executor.Tx
		// note: rollback is called in this function if commit fail, to avoid locking complication
		CommitTx() error
		RollbackTx() error
	}

	// Executor struct
	Executor struct {
		Db   *sql.DB
		Lock queue.PriorityLock // mutex should only lock tx
		Tx   *sql.Tx
	}
)

// NewQueryExecutor create new query executor instance
func NewQueryExecutor(db *sql.DB, lock queue.PriorityLock) *Executor {
	return &Executor{
		Db:   db,
		Lock: lock,
	}
}

/*
BeginTx begin database transaction and assign it to the Executor.Tx
lock the struct on begin
*/
func (qe *Executor) BeginTx(params ...int) error {
	qe.Lock.Lock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	tx, err := qe.Db.Begin()

	if err != nil {
		qe.Lock.Unlock()
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	qe.Tx = tx
	return nil
}

/*
Execute execute a single query string
return error if query not executed successfully
error will be nil otherwise.
*/
func (qe *Executor) Execute(query string) (sql.Result, error) {
	qe.Lock.Lock()
	defer qe.Lock.Unlock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	result, err := qe.Db.Exec(query)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return result, nil
}

/*
ExecuteStatement execute a single query string
return error if query not executed successfully
error will be nil otherwise.
*/
func (qe *Executor) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	qe.Lock.Lock()
	defer qe.Lock.Unlock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	stmt, err := qe.Db.Prepare(query)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer stmt.Close()
	result, err := stmt.Exec(args...)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return result, nil
}

/*
ExecuteSelect execute with select method that if you want to get `sql.Rows`.

And ***need to `Close()` the rows***.

This function is necessary if you want to processing the rows,
otherwise you can use `Execute` or `ExecuteTransactions`
*/
func (qe *Executor) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)
	monitoring.SetDatabaseStats(qe.Db.Stats())
	if tx {
		if qe.Tx != nil {
			rows, err = qe.Tx.Query(query, args...)
		} else {
			return nil, blocker.NewBlocker(
				blocker.DBErr,
				"transaction need to be begun before read the transaction state",
			)
		}
	} else {
		rows, err = qe.Db.Query(query, args...)
	}
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return rows, nil
}

/*
ExecuteSelectRow execute with select method that if you want to get `sql.Row` (single).
This function is necessary if you want to processing the row,
otherwise you can use `Execute` or `ExecuteTransactions`
*/
func (qe *Executor) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	var (
		row *sql.Row
	)
	monitoring.SetDatabaseStats(qe.Db.Stats())
	if tx {
		if qe.Tx != nil {
			row = qe.Tx.QueryRow(query, args...)
		} else {
			return nil, blocker.NewBlocker(
				blocker.DBErr,
				"ExecuteSelectRow, transaction need to be begun before read the transaction state",
			)
		}
	} else {
		row = qe.Db.QueryRow(query, args...)
	}
	return row, nil
}

// ExecuteTransaction execute a single transaction without committing it to database
// ExecuteTransaction should only be called after BeginTx and before CommitTx
func (qe *Executor) ExecuteTransaction(qStr string, args ...interface{}) error {
	if qe.Tx == nil {
		return blocker.NewBlocker(
			blocker.DBErr,
			"ExecuteTransaction, transaction need to be begun before read the transaction state",
		)
	}
	var stmt, err = qe.Tx.Prepare(qStr)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer stmt.Close()

	monitoring.SetDatabaseStats(qe.Db.Stats())
	_, err = stmt.Exec(args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

// ExecuteTransactions execute multiple transactions without committing it to database
// ExecuteTransactions should only be called after BeginTx and before CommitTx
func (qe *Executor) ExecuteTransactions(queries [][]interface{}) error {
	for _, query := range queries {
		stmt, err := qe.Tx.Prepare(fmt.Sprintf("%v", query[0]))
		if err != nil {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		monitoring.SetDatabaseStats(qe.Db.Stats())
		_, err = stmt.Exec(query[1:]...)
		if err != nil {
			_ = stmt.Close()
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		_ = stmt.Close()
	}
	return nil
}

// CommitTx commit on every transaction stacked in Executor.Tx
// note: rollback is called in this function if commit fail, to avoid locking complication
func (qe *Executor) CommitTx() error {
	monitoring.SetDatabaseStats(qe.Db.Stats())
	err := qe.Tx.Commit()
	defer func() {
		qe.Tx = nil
		qe.Lock.Unlock() // either success or not struct access should be unlocked once done

	}()
	if err != nil {
		var errRollback = qe.Tx.Rollback()
		if errRollback != nil {
			return blocker.NewBlocker(
				blocker.DBErr,
				fmt.Sprintf("error Commit: %s; err Rollback: %s", err.Error(), errRollback.Error()),
			)
		}
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

// RollbackTx rollback and unlock executor in case any single tx fail
func (qe *Executor) RollbackTx() error {
	monitoring.SetDatabaseStats(qe.Db.Stats())
	var err = qe.Tx.Rollback()
	defer func() {
		qe.Tx = nil
		qe.Lock.Unlock()
	}()
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}
