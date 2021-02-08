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
package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAccountBalanceHelperApplyConfirmFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperApplyConfirmSuccess struct {
		AccountBalanceHelper
	}
	mockExecutorFeeVoteCommitApplyConfirmedSuccess struct {
		query.Executor
	}
	mockExecutorFeeVoteCommitApplyConfirmedFail struct {
		query.Executor
	}
)

func (*mockAccountBalanceHelperApplyConfirmFail) AddAccountBalance([]byte, int64, model.EventType, uint32, int64, uint64) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperApplyConfirmSuccess) AddAccountBalance([]byte, int64, model.EventType, uint32, int64, uint64) error {
	return nil
}

func (*mockExecutorFeeVoteCommitApplyConfirmedSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockExecutorFeeVoteCommitApplyConfirmedFail) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("MockedError")
}

func TestFeeVoteCommitTransaction_ApplyConfirmed(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
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
			name: "wantFailed:AddBalance",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmFail{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: true,
		},
		{
			name: "wantFailed:InsertCommitVote",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmSuccess{},
				QueryExecutor:              &mockExecutorFeeVoteCommitApplyConfirmedFail{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmSuccess{},
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
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperApplyUnconfirmedFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperApplyUnconfirmedFail) AddAccountSpendableBalance(address []byte, amount int64) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}

func TestFeeVoteCommitTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail:AddAccountSpendableBalance",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperFeeVoteUndoApplyUnconfirmedFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperUndoApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperFeeVoteUndoApplyUnconfirmedFail) AddAccountSpendableBalance(address []byte, amount int64) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}

func TestFeeVoteCommitTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail:AddAccountSpendableBalance",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperFeeVoteUndoApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:  1,
					Fee: 1,
					SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Height: 1,
				},
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperUndoApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockFeeScaleServiceValidateFail struct {
		fee.FeeScaleService
	}
	mockFeeScaleServiceValidateSuccess struct {
		fee.FeeScaleService
	}

	mockQueryExecutorFeeVoteCommitValidateSuccess struct {
		query.Executor
	}
	mockFeeVoteCommitmentVoteQueryValidateFail struct {
		query.FeeVoteCommitmentVoteQuery
	}

	mockFeeVoteCommitmentVoteQueryValidateDupicated struct {
		query.FeeVoteCommitmentVoteQuery
	}
	mockFeeVoteCommitmentVoteQueryValidateSuccess struct {
		query.FeeVoteCommitmentVoteQuery
	}
	mockBlockQueryFeeVoteCommitGetLastBlockValidateFail struct {
		query.BlockQuery
	}
	mockBlockQueryGetLastBlockFeeVoteCommitValidateWrongPhase struct {
		query.BlockQuery
	}
	mockBlockQueryGetLastBlockFeeVoteCommitValidateSuccess struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightFeeVoteCommitValidateFail struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightFeeVoteCommitValidateDuplicated struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess struct {
		query.BlockQuery
	}
	mockNodeRegistrationQueryFeeVoteCommitValidateFail struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryFeeVoteCommitValidateFailErrNoRow struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryFeeVoteCommitValidateSuccess struct {
		query.NodeRegistrationQuery
	}

	mockAccountBalanceHelperFeeVoteCommitValidateFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperFeeVoteCommitValidateNotEnoughSpendable struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperFeeVoteCommitValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeVoteCommitTxBody, mockFeeVoteCommitTxBodyBytes = GetFixtureForFeeVoteCommitTransaction(&model.FeeVoteInfo{
		RecentBlockHash:   []byte{},
		RecentBlockHeight: 100,
		FeeVote:           10,
	}, "ZOOBC")
	mockTimestampValidateWrongPhase          int64 = 1
	mockTimestampValidateRightPhase          int64 = 2
	mockTimestampValidateRightPhaseExistVote int64 = 3
	mockFeeValidate                          int64 = 10
)

func (*mockFeeScaleServiceValidateFail) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, errors.New("MockedError")
}

func (*mockQueryExecutorFeeVoteCommitValidateSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockFeeVoteCommitmentVoteQueryValidateFail) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockFeeVoteCommitmentVoteQueryValidateSuccess) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	return sql.ErrNoRows
}
func (*mockFeeVoteCommitmentVoteQueryValidateDupicated) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	return nil
}

func (*mockBlockQueryFeeVoteCommitGetLastBlockValidateFail) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryFeeVoteCommitGetLastBlockValidateFail) Scan(block *model.Block, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockBlockQueryGetLastBlockFeeVoteCommitValidateWrongPhase) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetLastBlockFeeVoteCommitValidateWrongPhase) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateWrongPhase
	return nil
}

func (*mockBlockQueryGetLastBlockFeeVoteCommitValidateSuccess) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetLastBlockFeeVoteCommitValidateSuccess) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhase
	return nil
}

func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateFail) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateFail) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateFail) Scan(block *model.Block, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateDuplicated) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateDuplicated) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateDuplicated) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhaseExistVote
	return nil
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhase
	return nil
}

func (*mockNodeRegistrationQueryFeeVoteCommitValidateFail) GetNodeRegistrationByAccountAddress(
	accountAddress []byte) (str string, args []interface{},
) {
	return "mock", nil
}

func (*mockNodeRegistrationQueryFeeVoteCommitValidateFail) Scan(nr *model.NodeRegistration, row *sql.Row) error {
	return errors.New("MockedError")
}
func (*mockNodeRegistrationQueryFeeVoteCommitValidateFailErrNoRow) GetNodeRegistrationByAccountAddress(
	accountAddress []byte) (str string, args []interface{},
) {
	return "mockQuery", nil
}

func (*mockNodeRegistrationQueryFeeVoteCommitValidateFailErrNoRow) Scan(
	nr *model.NodeRegistration, row *sql.Row,
) error {
	return sql.ErrNoRows
}
func (*mockNodeRegistrationQueryFeeVoteCommitValidateSuccess) GetNodeRegistrationByAccountAddress(
	accountAddress []byte) (str string, args []interface{},
) {
	return "mockQuery", nil
}

func (*mockNodeRegistrationQueryFeeVoteCommitValidateSuccess) Scan(nr *model.NodeRegistration, row *sql.Row) error {
	return nil
}

func (*mockFeeScaleServiceValidateSuccess) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	switch blockTimestamp {
	case mockTimestampValidateWrongPhase:
		return model.FeeVotePhase_FeeVotePhaseReveal, false, nil
	case mockTimestampValidateRightPhase:
		return model.FeeVotePhase_FeeVotePhaseCommmit, true, nil
	case mockTimestampValidateRightPhaseExistVote:
		return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
	default:
		return model.FeeVotePhase_FeeVotePhaseReveal, false, errors.New("mockErrorInvalidCase")
	}
}

func (*mockFeeScaleServiceValidateSuccess) GetLatestFeeScale(feeScale *model.FeeScale) error {
	feeScale.FeeScale = fee.InitialFeeScale
	return nil
}

func (*mockAccountBalanceHelperFeeVoteCommitValidateFail) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperFeeVoteCommitValidateFail) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return false, sql.ErrNoRows
}
func (*mockAccountBalanceHelperFeeVoteCommitValidateNotEnoughSpendable) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeValidate - 1
	return nil
}
func (*mockAccountBalanceHelperFeeVoteCommitValidateNotEnoughSpendable) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return false, nil
}
func (*mockAccountBalanceHelperFeeVoteCommitValidateSuccess) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return true, nil
}
func (*mockAccountBalanceHelperFeeVoteCommitValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeValidate + 1
	return nil
}

func TestFeeVoteCommitTransaction_Validate(t *testing.T) {

	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	type args struct {
		dbTx bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantFail:InvalidHashLength",
			fields: fields{
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1},
				},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		}, {
			name: "wantFail:getLastBlock",
			fields: fields{
				Body:          mockFeeVoteCommitTxBody,
				BlockQuery:    &mockBlockQueryFeeVoteCommitGetLastBlockValidateFail{},
				QueryExecutor: &mockQueryExecutorFeeVoteCommitValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:getCurrentPhaseFirst",
			fields: fields{
				Body:            mockFeeVoteCommitTxBody,
				QueryExecutor:   &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				BlockQuery:      &mockBlockQueryGetLastBlockFeeVoteCommitValidateSuccess{},
				FeeScaleService: &mockFeeScaleServiceValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:notCommitPeriod",
			fields: fields{
				Body:            mockFeeVoteCommitTxBody,
				QueryExecutor:   &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				BlockQuery:      &mockBlockQueryGetLastBlockFeeVoteCommitValidateWrongPhase{},
				FeeScaleService: &mockFeeScaleServiceValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetVoteCommitByAccountAddressAndHeight",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateFail{},
				BlockQuery:                 &mockBlockQueryGetLastBlockFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:DuplicatedVote",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateDupicated{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateDuplicated{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddress",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddressNoRow",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateFailErrNoRow{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddressNoRow",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateFailErrNoRow{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetAccountBalanceByAccountAddress",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateSuccess{},
				AccountBalanceHelper:       &mockAccountBalanceHelperFeeVoteCommitValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetAccountBalanceByAccountAddressNotEnoughSpendable",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateSuccess{},
				AccountBalanceHelper:       &mockAccountBalanceHelperFeeVoteCommitValidateNotEnoughSpendable{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorFeeVoteCommitValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightFeeVoteCommitValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryFeeVoteCommitValidateSuccess{},
				AccountBalanceHelper:       &mockAccountBalanceHelperFeeVoteCommitValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteCommitTransaction_GetAmount(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name:   "wantSuccess",
			fields: fields{},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("FeeVoteCommitTransaction.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteCommitTransaction_GetMinimumFee(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					Escrow: &model.Escrow{},
				},
				Body:            mockFeeVoteCommitTxBody,
				FeeScaleService: &mockFeeScaleServiceValidateSuccess{},
			},
			want:    fee.SendMoneyFeeConstant,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			got, err := f.GetMinimumFee()
			if (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.GetMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FeeVoteCommitTransaction.GetMinimumFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteCommitTransaction_GetSize(t *testing.T) {

	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "wantSucess",
			fields: fields{
				Body: mockFeeVoteCommitTxBody,
			},
			want: uint32(len(mockFeeVoteCommitTxBodyBytes)),
		},
		{
			name: "wantError",
			fields: fields{
				Body: mockFeeVoteCommitTxBody,
			},
			want: uint32(len(mockFeeVoteCommitTxBodyBytes)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("FeeVoteCommitTransaction.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteCommitTransaction_GetTransactionBody(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSucess",
			fields: fields{
				Body: mockFeeVoteCommitTxBody,
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestFeeVoteCommitTransaction_Escrowable(t *testing.T) {
	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   EscrowTypeAction
		want1  bool
	}{
		{
			name: "wantSucess",
			fields: fields{
				TransactionObject: &model.Transaction{},
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			got, got1 := f.Escrowable()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FeeVoteCommitTransaction.Escrowable() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FeeVoteCommitTransaction.Escrowable() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

type (
	mockFeeScaleServiceSkipMempoolFail struct {
		fee.FeeScaleServiceInterface
	}
	mockFeeScaleServiceSkipMempoolWrongPhase struct {
		fee.FeeScaleServiceInterface
	}
	mockFeeScaleServiceSkipMempoolSuccess struct {
		fee.FeeScaleServiceInterface
	}
	mockQueryExecutorSkipMempoolSuccess struct {
		query.Executor
	}
	mockFeeVoteCommitmentVoteQuerySkipMempoolFail struct {
		query.FeeVoteCommitmentVoteQuery
	}
	mockFeeVoteCommitmentVoteQuerySkipMempoolSuccess struct {
		query.FeeVoteCommitmentVoteQuery
	}
)

func (*mockQueryExecutorSkipMempoolSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}
func (*mockFeeVoteCommitmentVoteQuerySkipMempoolFail) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockFeeVoteCommitmentVoteQuerySkipMempoolSuccess) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	return sql.ErrNoRows
}

func (*mockFeeScaleServiceSkipMempoolFail) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, errors.New("MockedError")
}
func (*mockFeeScaleServiceSkipMempoolWrongPhase) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseReveal, false, nil
}
func (*mockFeeScaleServiceSkipMempoolSuccess) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
}

func (*mockFeeScaleServiceSkipMempoolSuccess) GetLatestFeeScale(
	feeScale *model.FeeScale,
) error {
	return nil
}

func TestFeeVoteCommitTransaction_SkipMempoolTransaction(t *testing.T) {

	type fields struct {
		TransactionObject          *model.Transaction
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
		newBlockTimestamp    int64
		newBlockHeight       uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "wantFail:GetCurrentPhase",
			fields: fields{
				FeeScaleService: &mockFeeScaleServiceSkipMempoolFail{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{},
				newBlockTimestamp:    1,
			},
			want:    true,
			wantErr: true,
		},
		{
			name: "wantFail:WrongPhase",
			fields: fields{
				FeeScaleService: &mockFeeScaleServiceSkipMempoolWrongPhase{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{},
				newBlockTimestamp:    1,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "wantDuplicate:onMempool",
			fields: fields{
				TransactionObject: &model.Transaction{},
				FeeScaleService:   &mockFeeScaleServiceSkipMempoolSuccess{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					0: {
						TransactionType: uint32(model.TransactionType_FeeVoteCommitmentVoteTransaction),
					},
				},
				newBlockTimestamp: 1,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "wantFail:checkPreviousVote",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				FeeScaleService:            &mockFeeScaleServiceSkipMempoolSuccess{},
				QueryExecutor:              &mockQueryExecutorSkipMempoolSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQuerySkipMempoolFail{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{},
			},
			want:    true,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject:          &model.Transaction{},
				FeeScaleService:            &mockFeeScaleServiceSkipMempoolSuccess{},
				QueryExecutor:              &mockQueryExecutorSkipMempoolSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQuerySkipMempoolSuccess{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				TransactionObject:          tt.fields.TransactionObject,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.newBlockTimestamp, tt.args.newBlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FeeVoteCommitTransaction.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
