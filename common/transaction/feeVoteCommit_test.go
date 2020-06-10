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
	mockAccountLedgerHelperApplyConfirmFail struct {
		AccountLedgerHelper
	}
	mockAccountLedgerHelperApplyConfirmSuccess struct {
		AccountLedgerHelper
	}
	mockExecutorFeeVoteCommitApplyConfirmedSuccess struct {
		query.Executor
	}
	mockExecutorFeeVoteCommitApplyConfirmedFail struct {
		query.Executor
	}
)

func (*mockAccountBalanceHelperApplyConfirmFail) AddAccountBalance(address string, amount int64, blockHeight uint32) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperApplyConfirmSuccess) AddAccountBalance(address string, amount int64, blockHeight uint32) error {
	return nil
}

func (*mockAccountLedgerHelperApplyConfirmFail) InsertLedgerEntry(accountLedger *model.AccountLedger) error {
	return errors.New("MockedError")
}
func (*mockAccountLedgerHelperApplyConfirmSuccess) InsertLedgerEntry(accountLedger *model.AccountLedger) error {
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
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
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
			name: "wantFailed:InsertLedger",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmSuccess{},
				AccountLedgerHelper:        &mockAccountLedgerHelperApplyConfirmFail{},
			},
			args: args{
				blockTimestamp: 1,
			},
			wantErr: true,
		},
		{
			name: "wantFailed:InsertCommitVote",
			fields: fields{
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmSuccess{},
				AccountLedgerHelper:        &mockAccountLedgerHelperApplyConfirmSuccess{},
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
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperApplyConfirmSuccess{},
				AccountLedgerHelper:        &mockAccountLedgerHelperApplyConfirmSuccess{},
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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

func (*mockAccountBalanceHelperApplyUnconfirmedFail) AddAccountSpendableBalance(address string, amount int64) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperApplyUnconfirmedSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}

func TestFeeVoteCommitTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Timestamp                  int64
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
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
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitTransaction.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperUndoApplyUnconfirmedFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperUndoApplyUnconfirmedSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperUndoApplyUnconfirmedFail) AddAccountSpendableBalance(address string, amount int64) error {
	return errors.New("MockedError")
}
func (*mockAccountBalanceHelperUndoApplyUnconfirmedSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}

func TestFeeVoteCommitTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				ID:            1,
				Fee:           1,
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Height:        1,
				Body: &model.FeeVoteCommitTransactionBody{
					VoteHash: []byte{1, 2, 1},
				},
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				AccountBalanceHelper:       &mockAccountBalanceHelperUndoApplyUnconfirmedFail{},
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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

	mockQueryExecutorValidateSuccess struct {
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
	mockBlockQueryGetLastBlockValidateFail struct {
		query.BlockQuery
	}
	mockBlockQueryGetLastBlockValidateWrongPhase struct {
		query.BlockQuery
	}
	mockBlockQueryGetLastBlockValidateSuccess struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightValidateFail struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightValidateDuplicated struct {
		query.BlockQuery
	}
	mockBlockQueryGetBlockHeightValidateSuccess struct {
		query.BlockQuery
	}
	mockNodeRegistrationQueryValidateFail struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryValidateFailErrNoRow struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryValidateSuccess struct {
		query.NodeRegistrationQuery
	}

	mockAccountBalanceHelperValidateFail struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperValidateNotEnoughSpendable struct {
		AccountBalanceHelper
	}
	mockFeeVoteCommitAccountBalanceHelperValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeVoteCommitTxBody, mockFeeVoteCommitTxBodyBytes = GetFixtureForFeeVoteCommitTransaction(&model.FeeVoteInfo{
		RecentBlockHash:   []byte{},
		RecentBlockHeight: 100,
		FeeVote:           10,
	}, "ZOOBC")
	mockTimestampValidateWrongPhase          int64  = 1
	mockTimestampValidateRightPhase          int64  = 2
	mockTimestampValidateRightPhaseExistVote int64  = 3
	mockBlockHightValidate                   uint32 = 1
	mockFeeValidate                          int64  = 10
)

func (*mockFeeScaleServiceValidateFail) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, errors.New("MockedError")
}

func (*mockQueryExecutorValidateSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
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

func (*mockBlockQueryGetLastBlockValidateFail) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetLastBlockValidateFail) Scan(block *model.Block, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockBlockQueryGetLastBlockValidateWrongPhase) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetLastBlockValidateWrongPhase) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateWrongPhase
	return nil
}

func (*mockBlockQueryGetLastBlockValidateSuccess) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetLastBlockValidateSuccess) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhase
	return nil
}

func (*mockBlockQueryGetBlockHeightValidateFail) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateFail) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateFail) Scan(block *model.Block, row *sql.Row) error {
	return errors.New("MockedError")
}

func (*mockBlockQueryGetBlockHeightValidateDuplicated) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateDuplicated) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateDuplicated) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhaseExistVote
	return nil
}
func (*mockBlockQueryGetBlockHeightValidateSuccess) GetLastBlock() string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateSuccess) GetBlockByHeight(height uint32) string {
	return "mockQuery"
}
func (*mockBlockQueryGetBlockHeightValidateSuccess) Scan(block *model.Block, row *sql.Row) error {
	block.Timestamp = mockTimestampValidateRightPhase
	return nil
}

func (*mockNodeRegistrationQueryValidateFail) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return "mock", nil
}

func (*mockNodeRegistrationQueryValidateFail) Scan(nr *model.NodeRegistration, row *sql.Row) error {
	return errors.New("MockedError")
}
func (*mockNodeRegistrationQueryValidateFailErrNoRow) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return "mockQuery", nil
}

func (*mockNodeRegistrationQueryValidateFailErrNoRow) Scan(nr *model.NodeRegistration, row *sql.Row) error {
	return sql.ErrNoRows
}
func (*mockNodeRegistrationQueryValidateSuccess) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return "mockQuery", nil
}

func (*mockNodeRegistrationQueryValidateSuccess) Scan(nr *model.NodeRegistration, row *sql.Row) error {
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
	return nil
}

func (*mockAccountBalanceHelperValidateFail) GetBalanceByAccountID(accountBalance *model.AccountBalance, address string, dbTx bool) error {
	return errors.New("MockedError")
}

func (*mockAccountBalanceHelperValidateNotEnoughSpendable) GetBalanceByAccountID(
	accountBalance *model.AccountBalance, address string, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeValidate - 1
	return nil
}
func (*mockFeeVoteCommitAccountBalanceHelperValidateSuccess) GetBalanceByAccountID(
	accountBalance *model.AccountBalance, address string, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeValidate + 1
	return nil
}

func TestFeeVoteCommitTransaction_Validate(t *testing.T) {

	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				BlockQuery:    &mockBlockQueryGetLastBlockValidateFail{},
				QueryExecutor: &mockQueryExecutorValidateSuccess{},
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
				QueryExecutor:   &mockQueryExecutorValidateSuccess{},
				BlockQuery:      &mockBlockQueryGetLastBlockValidateSuccess{},
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
				QueryExecutor:   &mockQueryExecutorValidateSuccess{},
				BlockQuery:      &mockBlockQueryGetLastBlockValidateWrongPhase{},
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
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateFail{},
				BlockQuery:                 &mockBlockQueryGetLastBlockValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:DuplicatedVote",
			fields: fields{
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateDupicated{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateDuplicated{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddress",
			fields: fields{
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddressNoRow",
			fields: fields{
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateFailErrNoRow{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetNodeRegistrationByAccountAddressNoRow",
			fields: fields{
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateFailErrNoRow{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetAccountBalanceByAccountAddress",
			fields: fields{
				Body:                       mockFeeVoteCommitTxBody,
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateSuccess{},
				AccountBalanceHelper:       &mockAccountBalanceHelperValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantFail:scan_GetAccountBalanceByAccountAddressNotEnoughSpandable",
			fields: fields{
				Fee:                        mockFeeValidate,
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateSuccess{},
				AccountBalanceHelper:       &mockAccountBalanceHelperValidateNotEnoughSpendable{},
				FeeScaleService:            &mockFeeScaleServiceValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "wantSucess",
			fields: fields{
				Fee:                        mockFeeValidate,
				Body:                       mockFeeVoteCommitTxBody,
				QueryExecutor:              &mockQueryExecutorValidateSuccess{},
				FeeVoteCommitmentVoteQuery: &mockFeeVoteCommitmentVoteQueryValidateSuccess{},
				BlockQuery:                 &mockBlockQueryGetBlockHeightValidateSuccess{},
				NodeRegistrationQuery:      &mockNodeRegistrationQueryValidateSuccess{},
				AccountBalanceHelper:       &mockFeeVoteCommitAccountBalanceHelperValidateSuccess{},
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name:    "wantSuccess",
			fields:  fields{},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeVoteCommitTransaction{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteCommitTransaction{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("FeeVoteCommitTransaction.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteCommitTransaction_GetTransactionBody(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestFeeVoteCommitTransaction_Escrowable(t *testing.T) {
	type fields struct {
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   EscrowTypeAction
		want1  bool
	}{
		{
			name:   "wantSucess",
			fields: fields{},
			want:   nil,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeVoteCommitTransaction{
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
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
		ID                         int64
		Fee                        int64
		SenderAddress              string
		Height                     uint32
		Body                       *model.FeeVoteCommitTransactionBody
		FeeScaleService            fee.FeeScaleServiceInterface
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		BlockQuery                 query.BlockQueryInterface
		FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper       AccountBalanceHelperInterface
		AccountLedgerHelper        AccountLedgerHelperInterface
		QueryExecutor              query.ExecutorInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
		blockTimestamp       int64
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
				blockTimestamp:       1,
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
				blockTimestamp:       1,
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "wantDuplicate:onMempool",
			fields: fields{
				FeeScaleService: &mockFeeScaleServiceSkipMempoolSuccess{},
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					0: {
						TransactionType: uint32(model.TransactionType_FeeVoteCommitmentVoteTransaction),
					},
				},
				blockTimestamp: 1,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "wantFail:checkPreviousVote",
			fields: fields{
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
				ID:                         tt.fields.ID,
				Fee:                        tt.fields.Fee,
				SenderAddress:              tt.fields.SenderAddress,
				Height:                     tt.fields.Height,
				Body:                       tt.fields.Body,
				FeeScaleService:            tt.fields.FeeScaleService,
				NodeRegistrationQuery:      tt.fields.NodeRegistrationQuery,
				BlockQuery:                 tt.fields.BlockQuery,
				FeeVoteCommitmentVoteQuery: tt.fields.FeeVoteCommitmentVoteQuery,
				AccountBalanceHelper:       tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:        tt.fields.AccountLedgerHelper,
				QueryExecutor:              tt.fields.QueryExecutor,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.blockTimestamp)
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
