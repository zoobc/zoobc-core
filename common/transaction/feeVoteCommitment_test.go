package transaction

import (
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorFeeVoteCommitApplyConfirmedSuccess struct {
		query.Executor
	}
	mockExecutorFeeVoteCommitmentVoteApplyConfirmedFail struct {
		query.Executor
	}
)

func (*mockExecutorFeeVoteCommitApplyConfirmedSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*mockExecutorFeeVoteCommitmentVoteApplyConfirmedFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockedError")
}

func TestFeeVoteCommit_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitmentTransactionBody
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		BlockQuery                 query.BlockQueryInterface
		AccountLedgerQuery         query.AccountLedgerQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		QueryExecutor              query.ExecutorInterface
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantFailed",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitmentTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:        query.NewAccountBalanceQuery(),
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:         query.NewAccountLedgerQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				QueryExecutor:              &mockExecutorFeeVoteCommitmentVoteApplyConfirmedFail{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: true,
		},

		{
			name: "wantSuccess",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitmentTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:        query.NewAccountBalanceQuery(),
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:         query.NewAccountLedgerQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				QueryExecutor:              &mockExecutorFeeVoteCommitApplyConfirmedSuccess{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitment{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				AccountBalanceQuery:        tt.fields.AccountBalanceQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				AccountLedgerQuery:         tt.fields.AccountLedgerQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitment.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockExecutorFeeVoteCommitApplyUnconfirmedSuccess struct {
		query.Executor
	}
	mockExecutorFeeVoteCommitApplyUnconfirmedFail struct {
		query.Executor
	}
)

func (*mockExecutorFeeVoteCommitApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorFeeVoteCommitApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}
func TestFeeVoteCommit_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitmentTransactionBody
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		BlockQuery                 query.BlockQueryInterface
		AccountLedgerQuery         query.AccountLedgerQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitmentTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:        query.NewAccountBalanceQuery(),
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:         query.NewAccountLedgerQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				QueryExecutor:              &mockExecutorFeeVoteCommitApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitmentTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:        query.NewAccountBalanceQuery(),
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:         query.NewAccountLedgerQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				QueryExecutor:              &mockExecutorFeeVoteCommitApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitment{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				AccountBalanceQuery:        tt.fields.AccountBalanceQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				AccountLedgerQuery:         tt.fields.AccountLedgerQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitment.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockExecutorFeeVoteCommitUndoApplyUnconfirmedSuccess struct {
		query.Executor
	}
	mockExecutorFeeVoteCommitUndoApplyUnconfirmedFail struct {
		query.Executor
	}
)

func (*mockExecutorFeeVoteCommitUndoApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorFeeVoteCommitUndoApplyUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestFeeVoteCommit_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitmentTransactionBody
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		BlockQuery                 query.BlockQueryInterface
		AccountLedgerQuery         query.AccountLedgerQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitmentTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:        query.NewAccountBalanceQuery(),
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:         query.NewAccountLedgerQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				QueryExecutor:              &mockExecutorFeeVoteCommitUndoApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitment{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				AccountBalanceQuery:        tt.fields.AccountBalanceQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				AccountLedgerQuery:         tt.fields.AccountLedgerQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitment.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
