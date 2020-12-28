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
package fee

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/storage"

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
		lastBlockTimestamp    int64
		lastFeeScale          model.FeeScale
		feeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		executor              query.ExecutorInterface
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &storage.BlockStateStorage{},
				executor:              &mockInsertFeeScaleExecutorFail{},
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockInsertFeeScaleExecutorFail{},
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockInsertFeeScaleExecutorSuccess{},
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockInsertFeeScaleExecutorSuccess{},
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
				lastBlockTimestamp:    tt.fields.lastBlockTimestamp,
				lastFeeScale:          tt.fields.lastFeeScale,
				FeeScaleQuery:         tt.fields.feeScaleQuery,
				MainBlockStateStorage: tt.fields.MainBlockStateStorage,
				Executor:              tt.fields.executor,
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
		lastBlockTimestamp    int64
		lastFeeScale          model.FeeScale
		feeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		executor              query.ExecutorInterface
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
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              nil,
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: false,
		},
		{
			name: "GetLatestFeeScale - Executor fail",
			fields: fields{
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         query.NewFeeScaleQuery(),
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetLatestFeeScaleExecutorSelectFail{},
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: true,
		},
		{
			name: "GetLatestFeeScale - scan fail",
			fields: fields{
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         &mockGetLatestFeeScaleQueryScanFail{},
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetLatestFeeScaleExecutorSelectSuccess{},
			},
			args: args{
				feeScale: &model.FeeScale{},
			},
			wantErr: true,
		},
		{
			name: "GetLatestFeeScale - success",
			fields: fields{
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         &mockGetLatestFeeScaleQueryScanSuccess{},
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetLatestFeeScaleExecutorSelectSuccess{},
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
				lastBlockTimestamp:    tt.fields.lastBlockTimestamp,
				lastFeeScale:          tt.fields.lastFeeScale,
				FeeScaleQuery:         tt.fields.feeScaleQuery,
				MainBlockStateStorage: tt.fields.MainBlockStateStorage,
				Executor:              tt.fields.executor,
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
	mockFeeMainBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockFeeMainBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
)

var (
	mockGetCurrentPhaseBlockTimestampNowCommit    = time.Date(2020, 1, 3, 1, 0, 0, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampNowReveal    = time.Date(2020, 1, 16, 1, 0, 0, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampLastAdjust   = time.Date(2019, 12, 31, 23, 59, 59, 0, time.UTC)
	mockGetCurrentPhaseBlockTimestampLastNoAdjust = time.Date(2020, 1, 2, 23, 59, 59, 0, time.UTC)
)

func (*mockFeeMainBlockStateStorageSuccess) GetItem(key, item interface{}) error {
	block, ok := item.(*model.Block)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongType item, expected *model.Block")
	}
	block.Timestamp = mockGetCurrentPhaseBlockTimestampNowCommit.Unix()
	return nil
}

func (*mockFeeMainBlockStateStorageFail) GetItem(key, item interface{}) error {
	return errors.New("mockedError")
}

func (*mockGetCurrentPhaseExecutorGetLastBlockSuccess) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func TestFeeScaleService_GetCurrentPhase(t *testing.T) {
	type fields struct {
		lastBlockTimestamp    int64
		lastFeeScale          model.FeeScale
		feeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		executor              query.ExecutorInterface
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         nil,
				MainBlockStateStorage: &mockFeeMainBlockStateStorageFail{},
				executor:              &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
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
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         nil,
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
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
				lastBlockTimestamp:    mockGetCurrentPhaseBlockTimestampLastAdjust.Unix(),
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         nil,
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
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
				lastBlockTimestamp:    mockGetCurrentPhaseBlockTimestampLastNoAdjust.Unix(),
				lastFeeScale:          model.FeeScale{},
				feeScaleQuery:         nil,
				MainBlockStateStorage: &mockFeeMainBlockStateStorageSuccess{},
				executor:              &mockGetCurrentPhaseExecutorGetLastBlockSuccess{},
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
				lastBlockTimestamp:    tt.fields.lastBlockTimestamp,
				lastFeeScale:          tt.fields.lastFeeScale,
				FeeScaleQuery:         tt.fields.feeScaleQuery,
				MainBlockStateStorage: tt.fields.MainBlockStateStorage,
				Executor:              tt.fields.executor,
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
		feeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		executor              query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *FeeScaleService
	}{
		{
			name: "NewFeeScaleService",
			args: args{
				feeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				executor:              nil,
			},
			want: &FeeScaleService{
				lastBlockTimestamp:    0,
				lastFeeScale:          model.FeeScale{},
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeeScaleService(tt.args.feeScaleQuery, tt.args.MainBlockStateStorage, tt.args.executor); !reflect.DeepEqual(got, tt.want) {
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
		lastBlockTimestamp    int64
		lastFeeScale          model.FeeScale
		FeeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		Executor              query.ExecutorInterface
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
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
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
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
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
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
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
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
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
				FeeScaleQuery:         nil,
				MainBlockStateStorage: nil,
				Executor:              nil,
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
				lastBlockTimestamp:    tt.fields.lastBlockTimestamp,
				lastFeeScale:          tt.fields.lastFeeScale,
				FeeScaleQuery:         tt.fields.FeeScaleQuery,
				MainBlockStateStorage: tt.fields.MainBlockStateStorage,
				Executor:              tt.fields.Executor,
			}
			if got := fss.SelectVote(tt.args.votes, tt.args.originalSendMoneyFee); got != tt.want {
				t.Errorf("SelectVote() = %v, want %v", got, tt.want)
			}
		})
	}
}
