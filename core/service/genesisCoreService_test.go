package service

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorAddGenesisAccountSuccess struct {
		query.Executor
	}
	mockExecutorAddGenesisAccountFailExecuteTransactions struct {
		query.Executor
	}

	mockExecutorAddGenesisAccountCommitFail struct {
		query.Executor
	}
)

func (*mockExecutorAddGenesisAccountSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockExecutorAddGenesisAccountFailExecuteTransactions) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockError:accountInsertFail")
}

func (*mockExecutorAddGenesisAccountCommitFail) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestAddGenesisAccount(t *testing.T) {
	t.Run("AddGenesisAccount:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin() // we'll skip prepare expectation by mocking the function
		mock.ExpectCommit()
		err := AddGenesisAccount(&mockExecutorAddGenesisAccountSuccess{
			query.Executor{
				Db: db,
			},
		})
		if err != nil {
			t.Error("should be able to add genesis successfully")
		}
	})
	t.Run("AddGenesisAccount:fail-{fail execute tx}", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := AddGenesisAccount(&mockExecutorAddGenesisAccountFailExecuteTransactions{
			query.Executor{
				Db: db,
			},
		})
		if err == nil {
			t.Error("ExecuteTransactionsFailure should causes error")
		}
	})
	t.Run("AddGenesisAccount:fail-{fail commit tx}", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(errors.New("mockError:commitFail"))
		mock.ExpectRollback()
		err := AddGenesisAccount(&mockExecutorAddGenesisAccountCommitFail{
			query.Executor{
				Db: db,
			},
		})
		if err == nil {
			t.Error("ExecuteTransactionsFailure should causes error")
		}
	})
}
