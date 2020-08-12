package fee

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockInsertFeeScaleExecutorFail struct {
		query.ExecutorInterface
	}
	mockInsertFeeScaleExecutorSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockInsertFeeScaleExecutorFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockedError")
}

func (*mockInsertFeeScaleExecutorFail) ExecuteStatement(string, ...interface{}) (sql.Result, error) {
	return nil, errors.New("mockedError")
}

func (*mockInsertFeeScaleExecutorSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*mockInsertFeeScaleExecutorSuccess) ExecuteStatement(string, ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestFeeScaleService_InsertFeeScale(t *testing.T) {
	type fields struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		feeScaleQuery       query.FeeScaleQueryInterface
		mainchainBlockQuery query.BlockQueryInterface
		executor            query.ExecutorInterface
	}
	type args struct {
		feeScale *model.FeeScale
		dbTx     bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "InsertFeeScale-executorFail-txFalse",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockInsertFeeScaleExecutorFail{},
			},
			args: args{
				feeScale: &model.FeeScale{},
				dbTx:     false,
			},
			wantErr: true,
		},
		{
			name: "InsertFeeScale-executorFail-txTrue",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockInsertFeeScaleExecutorFail{},
			},
			args: args{
				feeScale: &model.FeeScale{},
				dbTx:     true,
			},
			wantErr: true,
		},
		{
			name: "InsertFeeScale-executorSuccess-txFalse",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockInsertFeeScaleExecutorSuccess{},
			},
			args: args{
				feeScale: &model.FeeScale{},
				dbTx:     false,
			},
			wantErr: false,
		},
		{
			name: "InsertFeeScale-executorSuccess-txTrue",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockInsertFeeScaleExecutorSuccess{},
			},
			args: args{
				feeScale: &model.FeeScale{},
				dbTx:     true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := &FeeScaleService{
				lastBlockTimestamp:  tt.fields.lastBlockTimestamp,
				lastFeeScale:        tt.fields.lastFeeScale,
				FeeScaleQuery:       tt.fields.feeScaleQuery,
				MainchainBlockQuery: tt.fields.mainchainBlockQuery,
				Executor:            tt.fields.executor,
			}
			if err := fss.InsertFeeScale(tt.args.feeScale); (err != nil) != tt.wantErr {
				t.Errorf("InsertFeeScale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockGetLatestFeeScaleExecutorSelectFail struct {
		query.ExecutorInterface
	}
	mockGetLatestFeeScaleExecutorSelectSuccess struct {
		query.ExecutorInterface
	}
	mockGetLatestFeeScaleQueryScanFail struct {
		query.FeeScaleQuery
	}
	mockGetLatestFeeScaleQueryScanSuccess struct {
		query.FeeScaleQuery
	}
)

func (*mockGetLatestFeeScaleExecutorSelectFail) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetLatestFeeScaleExecutorSelectSuccess) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockGetLatestFeeScaleQueryScanFail) Scan(*model.FeeScale, *sql.Row) error {
	return errors.New("mockedError")
}

func (*mockGetLatestFeeScaleQueryScanSuccess) Scan(*model.FeeScale, *sql.Row) error {
	return nil
}

func TestFeeScaleService_GetLatestFeeScale(t *testing.T) {
	type fields struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		feeScaleQuery       query.FeeScaleQueryInterface
		mainchainBlockQuery query.BlockQueryInterface
		executor            query.ExecutorInterface
	}
	type args struct {
		feeScale *model.FeeScale
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetLatestFeeScale - cached",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: constant.OneZBC,
				},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            nil,
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: false,
		},
		{
			name: "GetLatestFeeScale - Executor fail",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       query.NewFeeScaleQuery(),
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockGetLatestFeeScaleExecutorSelectFail{},
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: true,
		},
		{
			name: "GetLatestFeeScale - scan fail",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       &mockGetLatestFeeScaleQueryScanFail{},
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockGetLatestFeeScaleExecutorSelectSuccess{},
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: true,
		},
		{
			name: "GetLatestFeeScale - success",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       &mockGetLatestFeeScaleQueryScanSuccess{},
				mainchainBlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				executor:            &mockGetLatestFeeScaleExecutorSelectSuccess{},
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := &FeeScaleService{
				lastBlockTimestamp:  tt.fields.lastBlockTimestamp,
				lastFeeScale:        tt.fields.lastFeeScale,
				FeeScaleQuery:       tt.fields.feeScaleQuery,
				MainchainBlockQuery: tt.fields.mainchainBlockQuery,
				Executor:            tt.fields.executor,
			}
			if err := fss.GetLatestFeeScale(tt.args.feeScale); (err != nil) != tt.wantErr {
				t.Errorf("GetLatestFeeScale() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockGetCurrentPhaseExecutorGetLastBlockSuccess struct {
		query.ExecutorInterface
	}
	mockGetCurrentPhaseBlockQueryGetLastBlockSuccess struct {
		query.BlockQueryInterface
	}
	mockGetCurrentPhaseBlockQueryGetLastBlockFail struct {
		query.BlockQueryInterface
	}
)

var (
	mockGetCurrentPhaseBlockTimestampNowCommit    = time.Date(2020, 1, 3, 1, 0, 0, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampNowReveal    = time.Date(2020, 1, 16, 1, 0, 0, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampLastAdjust   = time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampLastNoAdjust = time.Date(2020, 1, 2, 23, 59, 59, 0, time.UTC)
)

func (*mockGetCurrentPhaseExecutorGetLastBlockSuccess) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockGetCurrentPhaseBlockQueryGetLastBlockFail) Scan(*model.Block, *sql.Row) error {
	return errors.New("mockedError")
}

func (*mockGetCurrentPhaseBlockQueryGetLastBlockFail) GetLastBlock() string {
	return "mockQuery"
}

func (*mockGetCurrentPhaseBlockQueryGetLastBlockSuccess) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockGetCurrentPhaseBlockTimestampNowCommit.Unix()
	return nil
}

func (*mockGetCurrentPhaseBlockQueryGetLastBlockSuccess) GetLastBlock() string {
	return "mockQuery"
}

func TestFeeScaleService_GetCurrentPhase(t *testing.T) {
	type fields struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		feeScaleQuery       query.FeeScaleQueryInterface
		mainchainBlockQuery query.BlockQueryInterface
		executor            query.ExecutorInterface
	}
	type args struct {
		blockTimestamp    int64
		isPostTransaction bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.FeeVotePhase
		want1   bool
		wantErr bool
	}{
		{
			name: "GetCurrentPhase - not cached - getlastblock fail",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       nil,
				mainchainBlockQuery: &mockGetCurrentPhaseBlockQueryGetLastBlockFail{},
				executor:            &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
			},
			args: args{
				blockTimestamp:    0,
				isPostTransaction: false,
			},
			want:    0,
			want1:   false,
			wantErr: true,
		},
		{
			name: "GetCurrentPhase - not cached - getlastblock success - not PostTransaction",
			fields: fields{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       nil,
				mainchainBlockQuery: &mockGetCurrentPhaseBlockQueryGetLastBlockSuccess{},
				executor:            &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
			},
			args: args{
				blockTimestamp:    mockGetCurrentPhaseBlockTimestampNowCommit.Unix(),
				isPostTransaction: false,
			},
			want:    model.FeeVotePhase_FeeVotePhaseCommmit,
			want1:   false,
			wantErr: false,
		},
		{
			name: "GetCurrentPhase - cached - adjust",
			fields: fields{
				lastBlockTimestamp:  mockGetCurrentPhaseBlockTimestampLastAdjust.Unix(),
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       nil,
				mainchainBlockQuery: &mockGetCurrentPhaseBlockQueryGetLastBlockSuccess{},
				executor:            &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
			},
			args: args{
				blockTimestamp:    mockGetCurrentPhaseBlockTimestampNowCommit.Unix(),
				isPostTransaction: false,
			},
			want:    model.FeeVotePhase_FeeVotePhaseCommmit,
			want1:   true,
			wantErr: false,
		},
		{
			name: "GetCurrentPhase - cached - reveal",
			fields: fields{
				lastBlockTimestamp:  mockGetCurrentPhaseBlockTimestampLastNoAdjust.Unix(),
				lastFeeScale:        model.FeeScale{},
				feeScaleQuery:       nil,
				mainchainBlockQuery: &mockGetCurrentPhaseBlockQueryGetLastBlockSuccess{},
				executor:            &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
			},
			args: args{
				blockTimestamp:    mockGetCurrentPhaseBlockTimestampNowReveal.Unix(),
				isPostTransaction: false,
			},
			want:    model.FeeVotePhase_FeeVotePhaseReveal,
			want1:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := &FeeScaleService{
				lastBlockTimestamp:  tt.fields.lastBlockTimestamp,
				lastFeeScale:        tt.fields.lastFeeScale,
				FeeScaleQuery:       tt.fields.feeScaleQuery,
				MainchainBlockQuery: tt.fields.mainchainBlockQuery,
				Executor:            tt.fields.executor,
			}
			got, got1, err := fss.GetCurrentPhase(tt.args.blockTimestamp, tt.args.isPostTransaction)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentPhase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCurrentPhase() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetCurrentPhase() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewFeeScaleService(t *testing.T) {
	type args struct {
		feeScaleQuery       query.FeeScaleQueryInterface
		mainchainBlockQuery query.BlockQueryInterface
		executor            query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *FeeScaleService
	}{
		{
			name: "NewFeeScaleService",
			args: args{
				feeScaleQuery:       nil,
				mainchainBlockQuery: nil,
				executor:            nil,
			},
			want: &FeeScaleService{
				lastBlockTimestamp:  0,
				lastFeeScale:        model.FeeScale{},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeeScaleService(tt.args.feeScaleQuery, tt.args.mainchainBlockQuery, tt.args.executor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeeScaleService() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockMedianLowerConstraintsPassed = []*model.FeeVoteInfo{
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant - (SendMoneyFeeConstant - 3), // i:4
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant - (SendMoneyFeeConstant - 1), // i:2 less than 0.5 than previous
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant - (SendMoneyFeeConstant - 2), // i:3
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant - SendMoneyFeeConstant/10, // i:1
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant - SendMoneyFeeConstant/30, // i:0
		},
	}
	mockMedianHigherConstraintsPassed = []*model.FeeVoteInfo{
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant + 3), // i:4
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant + 1), // i:2 more than 2.0 than previous
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant + 2), // i:3
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/10, // i:1
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/30, // i:0
		},
	}
	mockMedianWithinConstraintsEven = []*model.FeeVoteInfo{
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 4), // i:5
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 3), // i:4
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 1), // i:2
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 2), // i:3
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/10, // i:1
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/30, // i:0
		},
	}
	mockMedianWithinConstraintsOdd = []*model.FeeVoteInfo{
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 4), // i:5
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 1), // i:2
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + (SendMoneyFeeConstant - 2), // i:3
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/10, // i:1
		},
		{
			RecentBlockHash:   nil,
			RecentBlockHeight: 0,
			FeeVote:           SendMoneyFeeConstant + SendMoneyFeeConstant/30, // i:0
		},
	}
)

func TestFeeScaleService_SelectVote(t *testing.T) {
	previousScale := constant.OneZBC
	type fields struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		FeeScaleQuery       query.FeeScaleQueryInterface
		MainchainBlockQuery query.BlockQueryInterface
		Executor            query.ExecutorInterface
	}
	type args struct {
		votes                []*model.FeeVoteInfo
		originalSendMoneyFee int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "EmptyVotes",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: previousScale,
					Latest:   true,
				},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
			args: args{
				votes:                []*model.FeeVoteInfo{},
				originalSendMoneyFee: SendMoneyFeeConstant,
			},
			want: previousScale,
		},
		{
			name: "MedianPassLowerConstraints",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: previousScale,
					Latest:   true,
				},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
			args: args{
				votes:                mockMedianLowerConstraintsPassed,
				originalSendMoneyFee: SendMoneyFeeConstant,
			},
			want: previousScale / 2,
		},
		{
			name: "MedianPassHigherConstraints",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: previousScale,
					Latest:   true,
				},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
			args: args{
				votes:                mockMedianHigherConstraintsPassed,
				originalSendMoneyFee: SendMoneyFeeConstant,
			},
			want: previousScale * 2,
		},
		{
			name: "WithinConstraints - Even number of votes",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: previousScale,
					Latest:   true,
				},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
			args: args{
				votes:                mockMedianWithinConstraintsEven,
				originalSendMoneyFee: SendMoneyFeeConstant,
			},
			want: 199999650,
		},
		{
			name: "WithinConstraints - Odd number of votes",
			fields: fields{
				lastBlockTimestamp: 0,
				lastFeeScale: model.FeeScale{
					FeeScale: previousScale,
					Latest:   true,
				},
				FeeScaleQuery:       nil,
				MainchainBlockQuery: nil,
				Executor:            nil,
			},
			args: args{
				votes:                mockMedianWithinConstraintsOdd,
				originalSendMoneyFee: SendMoneyFeeConstant,
			},
			want: 199999600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fss := &FeeScaleService{
				lastBlockTimestamp:  tt.fields.lastBlockTimestamp,
				lastFeeScale:        tt.fields.lastFeeScale,
				FeeScaleQuery:       tt.fields.FeeScaleQuery,
				MainchainBlockQuery: tt.fields.MainchainBlockQuery,
				Executor:            tt.fields.Executor,
			}
			if got := fss.SelectVote(tt.args.votes, tt.args.originalSendMoneyFee); got != tt.want {
				t.Errorf("SelectVote() = %v, want %v", got, tt.want)
			}
		})
	}
}
