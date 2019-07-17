package query

import (
	"database/sql"
)

type (

	// ExecutorInterface interface
	ExecutorInterface interface {
		Execute(string) (sql.Result, error)
		ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error)
		ExecuteSelectRow(query string, args ...interface{}) *sql.Row
		ExecuteTransactions(queries []string) ([]sql.Result, error)
		ExecuteStatement(query string, args ...interface{}) (sql.Result, error)
		ExecuteTransactionStatements(queries map[*string][]interface{}) ([]sql.Result, error)
	}

	// Executor struct
	Executor struct {
		Db *sql.DB
	}
)

// NewQueryExecutor create new query executor instance
func NewQueryExecutor(db *sql.DB) *Executor {
	return &Executor{
		Db: db,
	}
}

/*
Execute execute a single query string
return error if query not executed successfully
error will be nil otherwise.
*/
func (qe *Executor) Execute(query string) (sql.Result, error) {

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

// ExecuteTransactions execute list of queries in transaction
// will return error in case one or more of the query fail
func (qe *Executor) ExecuteTransactions(queries []string) ([]sql.Result, error) {

	var (
		stmt    *sql.Stmt
		tx      *sql.Tx
		err     error
		results []sql.Result
	)

	tx, err = qe.Db.Begin()
	if err != nil {
		return nil, err
	}

	for _, query := range queries {
		stmt, err = tx.Prepare(query)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		result, err := stmt.Exec()
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		results = append(results, result)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteTransactionStatements execute list of statement in transaction
// will return error in case one or more of the query fail
func (qe *Executor) ExecuteTransactionStatements(queries map[*string][]interface{}) ([]sql.Result, error) {
	var (
		stmt    *sql.Stmt
		tx      *sql.Tx
		err     error
		results []sql.Result
	)

	tx, err = qe.Db.Begin()
	if err != nil {
		return nil, err
	}

	for query, value := range queries {
		stmt, err = tx.Prepare(*query)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		result, err := stmt.Exec(value...)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		results = append(results, result)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}
