package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	// GetTransactionsByIds mocks
	mockGetTransactionsByIdsExecutorFail struct {
		query.Executor
	}
	mockGetTransactionsByIdsExecutorSuccess struct {
		query.Executor
	}
	mockGetTransactionsByIdsTransactionQueryBuildFail struct {
		query.TransactionQuery
	}
	mockGetTransactionsByIdsTransactionQueryBuildSuccess struct {
		query.TransactionQuery
	}
	// GetTransactionsByIds mocks
	// GetTransactionsByBlockID mocks
	mockGetTransactionsByBlockIDExecutorFail struct {
		query.Executor
	}
	mockGetTransactionsByBlockIDExecutorSuccess struct {
		query.Executor
	}
	mockGetTransactionsByBlockIDTransactionQueryBuildFail struct {
		query.TransactionQuery
	}
	mockGetTransactionsByBlockIDTransactionQueryBuildSuccess struct {
		query.TransactionQuery
	}
	// GetTransactionsByBlockID mocks
)

var (
	// GetTransactionByIds mocks
	mockGetTransactionByIdsResult = []*model.Transaction{
		{
			TransactionHash: make([]byte, 32),
		},
	}
	mockGetTransactionsByBlockIDResult = []*model.Transaction{
		{
			TransactionHash: make([]byte, 32),
		},
	}
)

func (*mockGetTransactionsByIdsExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByIdsExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetTransactionsByIdsTransactionQueryBuildFail) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByIdsTransactionQueryBuildSuccess) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return mockGetTransactionByIdsResult, nil
}

func (*mockGetTransactionsByBlockIDExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByBlockIDExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetTransactionsByBlockIDTransactionQueryBuildFail) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetTransactionsByBlockIDTransactionQueryBuildSuccess) BuildModel(
	txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	return mockGetTransactionsByBlockIDResult, nil
}

func TestTransactionCoreService_GetTransactionsByIds(t *testing.T) {
	type fields struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
	type args struct {
		transactionIds []int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransactionByIds-ExecuteSelect-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByIdsTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByIdsExecutorFail{},
			},
			args: args{
				transactionIds: []int64{1, 2, 3},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionByIds-BuildModel-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByIdsTransactionQueryBuildFail{},
				QueryExecutor:    &mockGetTransactionsByIdsExecutorSuccess{},
			},
			args: args{
				transactionIds: []int64{1, 2, 3},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionByIds-BuildModel-Success",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByIdsTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByIdsExecutorSuccess{},
			},
			args: args{
				transactionIds: []int64{1},
			},
			want:    mockGetTransactionByIdsResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery: tt.fields.TransactionQuery,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			got, err := tg.GetTransactionsByIds(tt.args.transactionIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionsByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionsByIds() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionCoreService_GetTransactionsByBlockID(t *testing.T) {
	type fields struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
	type args struct {
		blockID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransactionsByBlockID-ExecuteSelect-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByBlockIDTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByBlockIDExecutorFail{},
			},
			args: args{
				blockID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionsByBlockID-BuildModel-Fail",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByBlockIDTransactionQueryBuildFail{},
				QueryExecutor:    &mockGetTransactionsByBlockIDExecutorSuccess{},
			},
			args: args{
				blockID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionsByBlockID-BuildModel-Success",
			fields: fields{
				TransactionQuery: &mockGetTransactionsByBlockIDTransactionQueryBuildSuccess{},
				QueryExecutor:    &mockGetTransactionsByBlockIDExecutorSuccess{},
			},
			args: args{
				blockID: 1,
			},
			want:    mockGetTransactionsByBlockIDResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery: tt.fields.TransactionQuery,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			got, err := tg.GetTransactionsByBlockID(tt.args.blockID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionsByBlockID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionsByBlockID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorExpiringEscrowSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockQueryExecutorExpiringEscrowSuccess) BeginTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) CommitTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) RollbackTx() error {
	return nil
}
func (*mockQueryExecutorExpiringEscrowSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mockRows := sqlmock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mockRows.AddRow(
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRows)
	return db.Query(qStr)
}
func (*mockQueryExecutorExpiringEscrowSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestTransactionCoreService_ExpiringEscrowTransactions(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		blockHeight uint32
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
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				EscrowTransactionQuery: query.NewEscrowTransactionQuery(),
				QueryExecutor:          &mockQueryExecutorExpiringEscrowSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ExpiringEscrowTransactions(tt.args.blockHeight, false); (err != nil) != tt.wantErr {
				t.Errorf("ExpiringEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockUndoApplyUnconfirmedTransaction_EscrowFalse struct {
	transaction.TXEmpty
}

func (*mockUndoApplyUnconfirmedTransaction_EscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockUndoApplyUnconfirmedTransaction_EscrowFalse) UndoApplyUnconfirmed() error {
	return nil
}

type mockUndoApplyUnconfirmedTransaction_EscrowUndoApplyUnconfirmed struct {
	transaction.NodeRegistration
}

func (*mockUndoApplyUnconfirmedTransaction_EscrowUndoApplyUnconfirmed) EscrowUndoApplyUnconfirmed() error {
	return nil
}

type mockUndoApplyUnconfirmedTransaction_EscrowTrue struct {
	transaction.TXEmpty
}

func (*mockUndoApplyUnconfirmedTransaction_EscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockUndoApplyUnconfirmedTransaction_EscrowUndoApplyUnconfirmed{}, true
}

func TestTransactionCoreService_UndoApplyUnconfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "UndoApplyUnconfirmedTransaction:EscrowFalse",
			args: args{
				txAction: &mockUndoApplyUnconfirmedTransaction_EscrowFalse{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockUndoApplyUnconfirmedTransaction_EscrowTrue{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.UndoApplyUnconfirmedTransaction(tt.args.txAction); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.UndoApplyUnconfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockApplyConfirmedTransaction_EscrowFalse struct {
	transaction.TXEmpty
}

func (*mockApplyConfirmedTransaction_EscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockApplyConfirmedTransaction_EscrowFalse) ApplyConfirmed(blockTimestamp int64) error {
	return nil
}

type mockApplyConfirmedTransaction_EscrowApplyConfirmed struct {
	transaction.EscrowTypeAction
}

func (*mockApplyConfirmedTransaction_EscrowApplyConfirmed) EscrowApplyConfirmed(blockTimestamp int64) error {
	return nil
}

type mockApplyConfirmedTransaction_EscrowTrue struct {
	transaction.TXEmpty
}

func (*mockApplyConfirmedTransaction_EscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockApplyConfirmedTransaction_EscrowApplyConfirmed{}, true
}

func TestTransactionCoreService_ApplyConfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction       transaction.TypeAction
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ApplyConfirmedTransaction:EscrowFalse",
			args: args{
				txAction:       &mockApplyConfirmedTransaction_EscrowFalse{},
				blockTimestamp: 0,
			},
			wantErr: false,
		},
		{
			name: "ApplyConfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockApplyConfirmedTransaction_EscrowTrue{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ApplyConfirmedTransaction(tt.args.txAction, tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ApplyConfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockApplyUnconfirmedTransaction_EscrowApplyUnconfirmed struct {
	transaction.EscrowTypeAction
}

func (*mockApplyUnconfirmedTransaction_EscrowApplyUnconfirmed) EscrowApplyUnconfirmed() error {
	return nil
}

type mockApplyUnconfirmedTransaction_EscrowTrue struct {
	transaction.TypeAction
}

func (*mockApplyUnconfirmedTransaction_EscrowTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockApplyUnconfirmedTransaction_EscrowApplyUnconfirmed{}, true
}

type mockApplyUnconfirmedTransaction_EscrowFalse struct {
	transaction.TypeAction
}

func (*mockApplyUnconfirmedTransaction_EscrowFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}
func (*mockApplyUnconfirmedTransaction_EscrowFalse) ApplyUnconfirmed() error {
	return nil
}

func TestTransactionCoreService_ApplyUnconfirmedTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ApplyUnconfirmedTransaction:EscrowTrue",
			args: args{
				txAction: &mockApplyUnconfirmedTransaction_EscrowTrue{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmedTransaction:EscrowFalse",
			args: args{
				txAction: &mockApplyUnconfirmedTransaction_EscrowFalse{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ApplyUnconfirmedTransaction(tt.args.txAction); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ApplyUnconfirmedTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockValidateTransaction_EscrowValidate struct {
	transaction.EscrowTypeAction
}

func (*mockValidateTransaction_EscrowValidate) EscrowValidate(dbTx bool) error {
	return nil
}

type mockValidateTransaction_EscrowableTrue struct {
	transaction.TypeAction
}

func (*mockValidateTransaction_EscrowableTrue) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockValidateTransaction_EscrowValidate{}, true
}

type mockValidateTransaction_EscrowableFalse struct {
	transaction.TypeAction
}

func (*mockValidateTransaction_EscrowableFalse) Escrowable() (transaction.EscrowTypeAction, bool) {
	return nil, false
}

func (*mockValidateTransaction_EscrowableFalse) Validate(dbTx bool) error {
	return nil
}

func TestTransactionCoreService_ValidateTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery       query.TransactionQueryInterface
		EscrowTransactionQuery query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
	}
	type args struct {
		txAction transaction.TypeAction
		useTX    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ValidateTransaction:EscrowableTrue",
			args: args{
				txAction: &mockValidateTransaction_EscrowableTrue{},
			},
			wantErr: false,
		},
		{
			name: "ValidateTransaction:EscrowableFalse",
			args: args{
				txAction: &mockValidateTransaction_EscrowableFalse{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &TransactionCoreService{
				TransactionQuery:       tt.fields.TransactionQuery,
				EscrowTransactionQuery: tt.fields.EscrowTransactionQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
			}
			if err := tg.ValidateTransaction(tt.args.txAction, tt.args.useTX); (err != nil) != tt.wantErr {
				t.Errorf("TransactionCoreService.ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
