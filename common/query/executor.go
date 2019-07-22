package query

import (
	"database/sql"
	"sync"
)

type (

	// ExecutorInterface interface
	ExecutorInterface interface {
		BeginTx() error
		Execute(string) (sql.Result, error)
		ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error)
		ExecuteSelectRow(query string, args ...interface{}) *sql.Row
		ExecuteStatement(query string, args ...interface{}) (sql.Result, error)
		// ExecuteTransactionStatements(queries [][]interface{}) ([]sql.Result, error)
		ExecuteTransaction(query string, args ...interface{}) error
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
	tx, err := qe.Db.Begin()

	if err != nil {
		qe.Unlock()
		return err
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
	result, err := qe.Db.Exec(query)

	if err != nil {
		return nil, err
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
	defer qe.Unlock()
	stmt, err := qe.Db.Prepare(query)

	if err != nil {
		return nil, err
	}
	result, err := stmt.Exec(args...)

	if err != nil {
		return nil, err
	}
	return result, nil
}

/*
ExecuteSelect execute with select method that if you want to get `sql.Rows`.

And ***need to `Close()` the rows***.

This function is necessary if you want to processing the rows,
otherwise you can use `Execute` or `ExecuteTransactions`
*/
func (qe *Executor) ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)

	rows, err = qe.Db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

/*
ExecuteSelectRow execute with select method that if you want to get `sql.Row` (single).
This function is necessary if you want to processing the row,
otherwise you can use `Execute` or `ExecuteTransactions`
*/
func (qe *Executor) ExecuteSelectRow(query string, args ...interface{}) *sql.Row {
	return qe.Db.QueryRow(query, args...)
}

// ExecuteTransaction execute a single transaction without committing it to database
// ExecuteTransaction should only be called after BeginTx and before CommitTx
func (qe *Executor) ExecuteTransaction(qStr string, args ...interface{}) error {
	stmt, err := qe.Tx.Prepare(qStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}
	return nil
}

// ExecuteTransactionCommit commit on every transaction stacked in Executor.Tx
func (qe *Executor) CommitTx() error {
	err := qe.Tx.Commit()
	defer qe.Unlock() // either success or not struct access should be unlocked once done
	if err != nil {
		return err
	}
	return nil
}

// RollbackTx rollback and unlock executor in case any single tx fail
func (qe *Executor) RollbackTx() error {
	err := qe.Tx.Rollback()
	defer qe.Unlock()
	return err
}
