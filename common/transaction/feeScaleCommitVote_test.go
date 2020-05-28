package transaction

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorApplyFeeScaleVoteSuccess struct {
		query.Executor
	}
)

func (*mockExecutorApplyFeeScaleVoteSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}

func (*mockExecutorApplyFeeScaleVoteSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func TestFeeScaleCommitVote_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                      int64
		Fee                     int64
		SenderAddress           string
		Height                  uint32
		Body                    *model.FeeScaleCommitVoteTransactionsBody
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		BlockQuery              query.BlockQueryInterface
		AccountLedgerQuery      query.AccountLedgerQueryInterface
		FeeScaleVoteCommitQuery query.FeeScaleVoteCommitQueryInterface
		QueryExecutor           query.ExecutorInterface
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
			name: "wantSuccess",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeScaleCommitVoteTransactionsBody{
					VoteHash: []byte{1, 2, 1},
				},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				FeeScaleVoteCommitQuery: query.NewFeeScaleVoteCommitsQuery(),
				QueryExecutor:           &mockExecutorApplyFeeScaleVoteSuccess{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeScaleCommitVote{
				ID:                      tt.fields.ID,
				Fee:                     tt.fields.Fee,
				SenderAddress:           tt.fields.SenderAddress,
				Height:                  tt.fields.Height,
				Body:                    tt.fields.Body,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				BlockQuery:              tt.fields.BlockQuery,
				AccountLedgerQuery:      tt.fields.AccountLedgerQuery,
				FeeScaleVoteCommitQuery: tt.fields.FeeScaleVoteCommitQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("FeeScaleCommitVote.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
