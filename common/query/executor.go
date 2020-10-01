package query

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"
)

type (
	// ExecutorInterface interface
	ExecutorInterface interface {
		BeginTx() error
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
		Db           *sql.DB
		sync.RWMutex // mutex should only lock tx
		Tx           *sql.Tx
	}
)

// NewQueryExecutor create new query executor instance
func NewQueryExecutor(db *sql.DB) *Executor {
	return &Executor{
		Db: db,
	}
}

/*
BeginTx begin database transaction and assign it to the Executor.Tx
lock the struct on begin
*/
func (qe *Executor) BeginTx() error {
	qe.Lock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	tx, err := qe.Db.Begin()

	if err != nil {
		qe.Unlock()
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
	qe.Lock()
	defer qe.Unlock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	result, err := qe.Db.Exec(query)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return result, nil
}

/*
Execute execute a single query string
return error if query not executed successfully
error will be nil otherwise.
*/
func (qe *Executor) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	qe.Lock()
	monitoring.SetDatabaseStats(qe.Db.Stats())
	stmt, err := qe.Db.Prepare(query)

	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer stmt.Close()
	defer qe.Unlock()
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
		qe.Unlock() // either success or not struct access should be unlocked once done

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
		qe.Unlock()
	}()
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}
