package transaction

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockFeeScaleFeeVoteRevealTXValidateInvalidPhasePeriod struct {
		fee.FeeScaleService
	}

	mockFeeScaleFeeVoteRevealTXValidateSuccess struct {
		fee.FeeScaleService
	}
	mockFeeScaleFeeVoteRevealTXValidateDuplicated struct {
		fee.FeeScaleService
	}

	mockSignatureFeeVoteRevealTXValidateInvalid struct {
		crypto.Signature
	}
	mockSignatureFeeVoteRevealTXValidateSuccess struct {
		crypto.Signature
	}

	mockQueryExecutorFeeVoteRevealTXValidateSuccess struct {
		query.Executor
	}

	// mockBlockQuery
	mockBlockQueryFeeVoteRevealVoteTXValidateFound struct {
		query.BlockQuery
	}
	// mockCommitmentVoteQuery
	mockCommitmentVoteQueryFeeVoteRevealTXValidateNotFound struct {
		query.FeeVoteCommitmentVoteQuery
	}
	mockCommitmentVoteQueryFeeVoteRevealTXValidateFound struct {
		query.FeeVoteCommitmentVoteQuery
	}

	// mockNodeRegistrationQuery
	mockNodeRegistrationQueryFeeVoteRevealTXValidateNotFound struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryFeeVoteRevealTXValidateFound struct {
		query.NodeRegistrationQuery
	}

	// mockVoteRevealQuery
	mockVoteRevealQueryFeeVoteRevealTXValidateFound struct {
		query.FeeVoteRevealVoteQuery
	}
	mockVoteRevealQueryFeeVoteRevealTXValidateNotFound struct {
		query.FeeVoteRevealVoteQuery
	}

	// mockAccountBalanceHelper
	mockAccountBalanceHelperValidateSuccess struct {
		AccountBalanceHelper
	}
)

func (*mockFeeScaleFeeVoteRevealTXValidateInvalidPhasePeriod) GetCurrentPhase(int64, bool) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
}
func (*mockFeeScaleFeeVoteRevealTXValidateSuccess) GetCurrentPhase(int64, bool) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseReveal, false, nil
}
func (*mockFeeScaleFeeVoteRevealTXValidateSuccess) GetLatestFeeScale(feeScale *model.FeeScale) error {
	return nil
}
func (*mockFeeScaleFeeVoteRevealTXValidateDuplicated) GetCurrentPhase(int64, bool) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseReveal, true, nil
}
func (*mockFeeScaleFeeVoteRevealTXValidateDuplicated) GetLatestFeeScale(feeScale *model.FeeScale) error {
	return nil
}
func (*mockSignatureFeeVoteRevealTXValidateInvalid) VerifySignature([]byte, []byte, string) error {
	return errors.New("invalid")
}
func (*mockSignatureFeeVoteRevealTXValidateSuccess) VerifySignature([]byte, []byte, string) error {
	return nil
}

func (*mockBlockQueryFeeVoteRevealVoteTXValidateFound) GetBlockByHeight(uint32) string {
	return ""
}
func (*mockBlockQueryFeeVoteRevealVoteTXValidateFound) Scan(*model.Block, *sql.Row) error {
	return nil
}

func (*mockQueryExecutorFeeVoteRevealTXValidateSuccess) ExecuteSelectRow(qry string, _ bool, _ ...interface{}) (*sql.Row, error) {
	var (
		dbCon, mockDB, _ = sqlmock.New()
		mockedRow        *sqlmock.Rows
	)
	switch {
	case strings.Contains(qry, "FROM main_block"):
		mockedBlock := GetFixturesForBlock(100, 12345678)
		mockedRow = mockDB.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields)
		mockedRow.AddRow(
			mockedBlock.ID,
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			mockedBlock.PreviousBlockHash,
			mockedBlock.Height,
			mockedBlock.Timestamp,
			mockedBlock.BlockSeed,
			mockedBlock.BlockSignature,
			mockedBlock.CumulativeDifficulty,
			mockedBlock.PayloadLength,
			mockedBlock.PayloadHash,
			mockedBlock.BlocksmithPublicKey,
			mockedBlock.TotalAmount,
			mockedBlock.TotalFee,
			mockedBlock.TotalCoinBase,
			mockedBlock.Version,
		)
		mockDB.ExpectQuery(regexp.QuoteMeta(qry)).WillReturnRows(mockedRow)
		return dbCon.QueryRow(qry), nil
	default:
		return nil, nil
	}
}

func (*mockCommitmentVoteQueryFeeVoteRevealTXValidateNotFound) GetVoteCommitByAccountAddress(string) (qStr string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockCommitmentVoteQueryFeeVoteRevealTXValidateNotFound) Scan(*model.FeeVoteCommitmentVote, *sql.Row) error {
	return sql.ErrNoRows
}
func (*mockCommitmentVoteQueryFeeVoteRevealTXValidateFound) GetVoteCommitByAccountAddress(string) (qStr string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockCommitmentVoteQueryFeeVoteRevealTXValidateFound) Scan(vote *model.FeeVoteCommitmentVote, _ *sql.Row) error {
	body, _ := GetFixtureForFeeVoteCommitTransaction(&model.FeeVoteInfo{
		RecentBlockHash:   []byte{},
		RecentBlockHeight: 100,
		FeeVote:           10,
	}, "ZOOBC")
	vote.VoteHash = body.GetVoteHash()
	return nil
}

func (*mockNodeRegistrationQueryFeeVoteRevealTXValidateNotFound) GetNodeRegistrationByAccountAddress(string) (str string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockNodeRegistrationQueryFeeVoteRevealTXValidateNotFound) Scan(_ *model.NodeRegistration, _ *sql.Row) error {
	return sql.ErrNoRows
}
func (*mockNodeRegistrationQueryFeeVoteRevealTXValidateFound) GetNodeRegistrationByAccountAddress(string) (str string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockNodeRegistrationQueryFeeVoteRevealTXValidateFound) Scan(*model.NodeRegistration, *sql.Row) error {
	return nil
}

func (*mockVoteRevealQueryFeeVoteRevealTXValidateFound) GetFeeVoteRevealByAccountAddress(string) (str string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockVoteRevealQueryFeeVoteRevealTXValidateFound) Scan(*model.FeeVoteRevealVote, *sql.Row) error {
	return nil
}
func (*mockVoteRevealQueryFeeVoteRevealTXValidateNotFound) GetFeeVoteRevealByAccountAddress(string) (str string, args []interface{}) {
	return "", []interface{}{}
}
func (*mockVoteRevealQueryFeeVoteRevealTXValidateNotFound) Scan(*model.FeeVoteRevealVote, *sql.Row) error {
	return sql.ErrNoRows
}

func (*mockAccountBalanceHelperValidateSuccess) GetBalanceByAccountAddress(accountBalance *model.AccountBalance, _ string, _ bool) error {
	accountBalance.SpendableBalance = 100
	return nil
}
func TestFeeVoteRevealTransaction_Validate(t *testing.T) {
	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
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
			name: "WantErr:InvalidPhasePeriod",
			fields: fields{
				Fee:             1,
				Timestamp:       12345678,
				FeeScaleService: &mockFeeScaleFeeVoteRevealTXValidateInvalidPhasePeriod{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:InvalidSignature",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateSuccess{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateInvalid{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:InvalidRecentBlock",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateSuccess{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateSuccess{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				BlockQuery:             &mockBlockQueryFeeVoteRevealVoteTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateFound{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:CommitVoteNotFound",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateSuccess{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateSuccess{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				BlockQuery:             &mockBlockQueryFeeVoteRevealVoteTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateNotFound{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:NotNodeOwner",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateSuccess{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateSuccess{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				BlockQuery:             &mockBlockQueryFeeVoteRevealVoteTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateFound{},
				NodeRegistrationQuery:  &mockNodeRegistrationQueryFeeVoteRevealTXValidateNotFound{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:RevealVoteDuplicated",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateDuplicated{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateSuccess{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				BlockQuery:             &mockBlockQueryFeeVoteRevealVoteTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateFound{},
				FeeVoteRevealVoteQuery: &mockVoteRevealQueryFeeVoteRevealTXValidateFound{},
				NodeRegistrationQuery:  &mockNodeRegistrationQueryFeeVoteRevealTXValidateFound{},
			},
			wantErr: true,
		},
		{
			name: "Success",
			fields: fields{
				Fee:                1,
				Timestamp:          12345678,
				FeeScaleService:    &mockFeeScaleFeeVoteRevealTXValidateSuccess{},
				SignatureInterface: &mockSignatureFeeVoteRevealTXValidateSuccess{},
				Body: &model.FeeVoteRevealTransactionBody{
					FeeVoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				BlockQuery:             &mockBlockQueryFeeVoteRevealVoteTXValidateFound{},
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealTXValidateSuccess{},
				FeeVoteCommitVoteQuery: &mockCommitmentVoteQueryFeeVoteRevealTXValidateFound{},
				FeeVoteRevealVoteQuery: &mockVoteRevealQueryFeeVoteRevealTXValidateNotFound{},
				NodeRegistrationQuery:  &mockNodeRegistrationQueryFeeVoteRevealTXValidateFound{},
				AccountBalanceHelper:   &mockAccountBalanceHelperValidateSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteRevealTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Timestamp:              tt.fields.Timestamp,
				Body:                   tt.fields.Body,
				FeeScaleService:        tt.fields.FeeScaleService,
				SignatureInterface:     tt.fields.SignatureInterface,
				BlockQuery:             tt.fields.BlockQuery,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				FeeVoteCommitVoteQuery: tt.fields.FeeVoteCommitVoteQuery,
				FeeVoteRevealVoteQuery: tt.fields.FeeVoteRevealVoteQuery,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:    tt.fields.AccountLedgerHelper,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperFeeVoteRevealSuccess struct {
		AccountBalanceHelper
	}
	mockAccountLedgerHelperFeeVoteRevealSuccess struct {
		AccountLedgerHelper
	}
	mockQueryExecutorFeeVoteRevealApplyConfirmedSuccess struct {
		query.Executor
	}
)

func (*mockAccountBalanceHelperFeeVoteRevealSuccess) AddAccountSpendableBalance(address string, amount int64) error {
	return nil
}
func (*mockAccountBalanceHelperFeeVoteRevealSuccess) AddAccountBalance(address string, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64) error {
	return nil
}

func (*mockAccountLedgerHelperFeeVoteRevealSuccess) InsertLedgerEntry(*model.AccountLedger) error {
	return nil
}

func (*mockQueryExecutorFeeVoteRevealApplyConfirmedSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}

func TestFeeVoteRevealTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperFeeVoteRevealSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteRevealTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Timestamp:              tt.fields.Timestamp,
				Body:                   tt.fields.Body,
				FeeScaleService:        tt.fields.FeeScaleService,
				SignatureInterface:     tt.fields.SignatureInterface,
				BlockQuery:             tt.fields.BlockQuery,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				FeeVoteCommitVoteQuery: tt.fields.FeeVoteCommitVoteQuery,
				FeeVoteRevealVoteQuery: tt.fields.FeeVoteRevealVoteQuery,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:    tt.fields.AccountLedgerHelper,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteRevealTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperFeeVoteRevealSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteRevealTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Timestamp:              tt.fields.Timestamp,
				Body:                   tt.fields.Body,
				FeeScaleService:        tt.fields.FeeScaleService,
				SignatureInterface:     tt.fields.SignatureInterface,
				BlockQuery:             tt.fields.BlockQuery,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				FeeVoteCommitVoteQuery: tt.fields.FeeVoteCommitVoteQuery,
				FeeVoteRevealVoteQuery: tt.fields.FeeVoteRevealVoteQuery,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:    tt.fields.AccountLedgerHelper,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteRevealTransaction_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
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
			name: "WantSuccess",
			fields: fields{
				AccountBalanceHelper:   &mockAccountBalanceHelperFeeVoteRevealSuccess{},
				AccountLedgerHelper:    &mockAccountLedgerHelperFeeVoteRevealSuccess{},
				FeeVoteRevealVoteQuery: query.NewFeeVoteRevealVoteQuery(),
				QueryExecutor:          &mockQueryExecutorFeeVoteRevealApplyConfirmedSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteRevealTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Timestamp:              tt.fields.Timestamp,
				Body:                   tt.fields.Body,
				FeeScaleService:        tt.fields.FeeScaleService,
				SignatureInterface:     tt.fields.SignatureInterface,
				BlockQuery:             tt.fields.BlockQuery,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				FeeVoteCommitVoteQuery: tt.fields.FeeVoteCommitVoteQuery,
				FeeVoteRevealVoteQuery: tt.fields.FeeVoteRevealVoteQuery,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:    tt.fields.AccountLedgerHelper,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteRevealTransaction_ParseBodyBytes(t *testing.T) {
	txBody := &model.FeeVoteRevealTransactionBody{
		FeeVoteInfo: &model.FeeVoteInfo{
			RecentBlockHeight: 100,
			RecentBlockHash:   make([]byte, sha256.Size),
			FeeVote:           100,
		},
		VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}

	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "WantSuccess",
			args: args{
				txBodyBytes: (&FeeVoteRevealTransaction{
					Body: txBody,
				}).GetBodyBytes(),
			},
			want: txBody,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &FeeVoteRevealTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Timestamp:              tt.fields.Timestamp,
				Body:                   tt.fields.Body,
				FeeScaleService:        tt.fields.FeeScaleService,
				SignatureInterface:     tt.fields.SignatureInterface,
				BlockQuery:             tt.fields.BlockQuery,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				FeeVoteCommitVoteQuery: tt.fields.FeeVoteCommitVoteQuery,
				FeeVoteRevealVoteQuery: tt.fields.FeeVoteRevealVoteQuery,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				AccountLedgerHelper:    tt.fields.AccountLedgerHelper,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBodyBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
