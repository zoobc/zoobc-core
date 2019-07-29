package transaction

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockApplyUnconfirmedQueryExecutor struct {
		query.Executor
	}
)

func (*mockApplyUnconfirmedQueryExecutor) ExecuteStatement(qStr string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("Empty")
}

func TestNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		SenderAccountType     uint32
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		AccountQuery          query.AccountQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.NodeRegistrationTransactionBody{
					NodePublicKey: []byte{},
					AccountType:   0,
				},
				Fee:                 1,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockApplyUnconfirmedQueryExecutor{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				SenderAccountType:     tt.fields.SenderAccountType,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				AccountQuery:          tt.fields.AccountQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		SenderAccountType     uint32
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		AccountQuery          query.AccountQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "NodeRegistration-UndoApplyUnconfirmed:default",
			fields: fields{
				Body: &model.NodeRegistrationTransactionBody{
					NodePublicKey: []byte{},
					AccountType:   0,
				},
				Fee:                 1,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockApplyUnconfirmedQueryExecutor{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				SenderAccountType:     tt.fields.SenderAccountType,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				AccountQuery:          tt.fields.AccountQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
