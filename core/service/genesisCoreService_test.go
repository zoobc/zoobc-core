package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorAddGenesisAccountSuccess struct {
		query.ExecutorInterface
	}
	mockExecutorAddGenesisAccountFailAccount struct {
		query.ExecutorInterface
	}

	mockExecutorAddGenesisAccountFailAccountBalance struct {
		query.ExecutorInterface
	}
)

func (*mockExecutorAddGenesisAccountSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (*mockExecutorAddGenesisAccountFailAccount) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("mockError:accountInsertFail")
}

func (*mockExecutorAddGenesisAccountFailAccountBalance) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	if qe == "INSERT INTO account (id,account_type,address) VALUES(? , ?, ?)" {
		return nil, nil
	}
	return nil, errors.New("mockError:accountInsertFail")
}

func TestAddGenesisAccount(t *testing.T) {
	type args struct {
		executor query.ExecutorInterface
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "AddGenesisAccount:success",
			args: args{
				executor: &mockExecutorAddGenesisAccountSuccess{},
			},
			wantErr: false,
		},
		{
			name: "AddGenesisAccount:fail-{fail insert account}",
			args: args{
				executor: &mockExecutorAddGenesisAccountFailAccount{},
			},
			wantErr: true,
		},
		{
			name: "AddGenesisAccount:fail-{fail insert account balance}",
			args: args{
				executor: &mockExecutorAddGenesisAccountFailAccountBalance{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddGenesisAccount(tt.args.executor); (err != nil) != tt.wantErr {
				t.Errorf("AddGenesisAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
