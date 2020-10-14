package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

var (
	// mock GetPendingTransactionByAddress
	mockGetPendingTransactionsExecutorCountFailParam = &model.GetPendingTransactionsRequest{
		SenderAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		Status: model.PendingTransactionStatus_PendingTransactionPending,
		Pagination: &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      1,
		},
	}
	// mock GetPendingTransactionByAddress
)

type (
	// mock GetPendingTransactionByAddress
	mockGetPendingTransactionsExecutorCountFail struct {
		query.Executor
	}
	mockGetPendingTransactionsExecutorGetPendingTxsFail struct {
		query.Executor
	}
	mockGetPendingTransactionsExecutorGetPendingTxsSuccess struct {
		query.Executor
	}
	mockGetPendingTransactionsPendingTxQueryBuildFail struct {
		query.PendingTransactionQuery
	}
	mockGetPendingTransactionsPendingTxQueryBuildSuccess struct {
		query.PendingTransactionQuery
	}
	// mock GetPendingTransactionByAddress
)

func (*mockGetPendingTransactionsExecutorCountFail) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionsExecutorGetPendingTxsFail) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionsExecutorGetPendingTxsFail) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionsExecutorGetPendingTxsSuccess) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionsExecutorGetPendingTxsSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}).AddRow(1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockGetPendingTransactionsPendingTxQueryBuildFail) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionsPendingTxQueryBuildSuccess) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	return []*model.PendingTransaction{}, nil
}

func TestMultisigService_GetPendingTransactionByAddress(t *testing.T) {
	type fields struct {
		Executor                query.ExecutorInterface
		BlockService            coreService.BlockServiceInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		Logger                  *logrus.Logger
	}
	type args struct {
		param *model.GetPendingTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionByAddress-fail-countExecuteSelectRow-error-noRow",
			fields: fields{
				Executor:                &mockGetPendingTransactionsExecutorCountFail{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionsExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-fail-GetPendingTxsExecutor-error",
			fields: fields{
				Executor:                &mockGetPendingTransactionsExecutorGetPendingTxsFail{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionsExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-fail-PendingTxQueryBuild-error",
			fields: fields{
				Executor:                &mockGetPendingTransactionsExecutorGetPendingTxsSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: &mockGetPendingTransactionsPendingTxQueryBuildFail{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionsExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-success",
			fields: fields{
				Executor:                &mockGetPendingTransactionsExecutorGetPendingTxsSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: &mockGetPendingTransactionsPendingTxQueryBuildSuccess{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionsExecutorCountFailParam,
			},
			want: &model.GetPendingTransactionsResponse{
				Count:               1,
				Page:                1,
				PendingTransactions: []*model.PendingTransaction{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MultisigService{
				Executor:                tt.fields.Executor,
				BlockService:            tt.fields.BlockService,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:   tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				Logger:                  tt.fields.Logger,
			}
			got, err := ms.GetPendingTransactions(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPendingTransactionByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPendingTransactionByAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// mock GetPendingTransactionByAddress
	mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam = &model.GetPendingTransactionDetailByTransactionHashRequest{
		TransactionHashHex: "1c72a355d480ce3c10b1981a7a22e5c2d7accb0c302dbef47a25119bff1b5e17",
	}
	mockLastBlock = &model.Block{
		Height: 1000,
	}
	mockPendingTransaction = &model.PendingTransaction{
		SenderAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionPending,
		BlockHeight:      800,
		Latest:           true,
	}
	mockMultisigInfo = &model.MultiSignatureInfo{
		MinimumSignatures: 2,
		Nonce:             3,
		MultisigAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BlockHeight: 400,
		Latest:      true,
	}
	mockMultisigInfoWithParticipants = &model.MultiSignatureInfo{
		MinimumSignatures: 2,
		Nonce:             3,
		Addresses: [][]byte{
			{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
				116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
			{0, 0, 0, 0, 8, 32, 68, 38, 181, 138, 127, 184, 190, 125, 84, 174, 13, 162, 122, 62, 183, 130, 70, 18, 103, 47, 177, 161,
				153, 143, 61, 130, 145, 81, 222, 70},
			{0, 0, 0, 0, 160, 121, 129, 83, 225, 164, 195, 123, 8, 181, 41, 251, 17, 3, 93, 37, 182, 109, 32, 174, 168, 68, 193, 212,
				79, 54, 156, 213, 117, 27, 185, 167},
		},
		MultisigAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BlockHeight: 400,
		Latest:      true,
	}
	mockMultisigParticipant1 = &model.MultiSignatureParticipant{
		MultiSignatureAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BlockHeight: 400,
		Latest:      true,
		AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		AccountAddressIndex: 0,
	}
	mockMultisigParticipant2 = &model.MultiSignatureParticipant{
		MultiSignatureAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BlockHeight: 400,
		Latest:      true,
		AccountAddress: []byte{0, 0, 0, 0, 8, 32, 68, 38, 181, 138, 127, 184, 190, 125, 84, 174, 13, 162, 122, 62, 183, 130, 70, 18, 103, 47, 177, 161,
			153, 143, 61, 130, 145, 81, 222, 70},
		AccountAddressIndex: 1,
	}
	mockMultisigParticipant3 = &model.MultiSignatureParticipant{
		MultiSignatureAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BlockHeight: 400,
		Latest:      true,
		AccountAddress: []byte{0, 0, 0, 0, 160, 121, 129, 83, 225, 164, 195, 123, 8, 181, 41, 251, 17, 3, 93, 37, 182, 109, 32, 174, 168, 68, 193, 212,
			79, 54, 156, 213, 117, 27, 185, 167},
		AccountAddressIndex: 2,
	}
	mockMultisigParticipants = []*model.MultiSignatureParticipant{
		mockMultisigParticipant1,
		mockMultisigParticipant2,
		mockMultisigParticipant3,
	}

// mock GetPendingTransactionByAddress
)

type (
	mockGetPendingTransactionByTransactionHashBlockServiceFail struct {
		coreService.BlockService
	}

	mockGetPendingTransactionByTransactionHashBlockServiceSuccess struct {
		coreService.BlockService
	}

	mockGetPendingTransactionByTransactionHashPendingQueryScanNoRow struct {
		query.PendingTransactionQuery
	}
	mockGetPendingTransactionByTransactionHashPendingQueryScanOtherError struct {
		query.PendingTransactionQuery
	}
	mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess struct {
		query.PendingTransactionQuery
	}

	mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess struct {
		query.Executor
	}

	mockGetPendingTransactionByTransactionHashGetPendingSigExecutorFail struct {
		mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess
	}

	mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess struct {
		mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess
	}

	mockGetPendingTransactionByTransactionHashPendingSigQueryBuildFail struct {
		query.PendingSignatureQuery
	}

	mockGetPendingTransactionByTransactionHashPendingSigQueryBuildSuccess struct {
		query.PendingSignatureQuery
	}

	mockGetPendingTransactionByTransactionHashMultisigInfoScanFailOtherError struct {
		query.MultisignatureInfoQuery
	}
	mockGetPendingTransactionByTransactionHashMultisigInfoScanSuccess struct {
		query.MultisignatureInfoQuery
	}
	mockGetPendingTransactionByTransactionHashMultisigParticipantBuildModelSuccess struct {
		query.MultiSignatureParticipantQuery
	}
)

func (*mockGetPendingTransactionByTransactionHashBlockServiceFail) GetLastBlock() (*model.Block, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionByTransactionHashBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return mockLastBlock, nil
}

func (*mockGetPendingTransactionByTransactionHashPendingQueryScanNoRow) Scan(
	pendingTx *model.PendingTransaction, row *sql.Row) error {
	return sql.ErrNoRows
}

func (*mockGetPendingTransactionByTransactionHashPendingQueryScanOtherError) Scan(
	pendingTx *model.PendingTransaction, row *sql.Row) error {
	return errors.New("mockedError")
}

func (*mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess) Scan(
	pendingTx *model.PendingTransaction, row *sql.Row) error {
	*pendingTx = *mockPendingTransaction
	return nil
}

func (*mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess) ExecuteSelectRow(
	qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionByTransactionHashGetPendingSigExecutorFail) ExecuteSelect(
	qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}).AddRow(1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockGetPendingTransactionByTransactionHashPendingSigQueryBuildFail) BuildModel(
	pendingSigs []*model.PendingSignature, rows *sql.Rows,
) ([]*model.PendingSignature, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionByTransactionHashPendingSigQueryBuildSuccess) BuildModel(
	pendingSigs []*model.PendingSignature, rows *sql.Rows,
) ([]*model.PendingSignature, error) {
	return []*model.PendingSignature{}, nil
}

func (*mockGetPendingTransactionByTransactionHashMultisigInfoScanFailOtherError) Scan(
	multisigInfo *model.MultiSignatureInfo, row *sql.Row,
) error {
	return errors.New("mockedError")
}

func (*mockGetPendingTransactionByTransactionHashMultisigInfoScanSuccess) Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error {
	*multisigInfo = *mockMultisigInfo
	return nil
}

func (*mockGetPendingTransactionByTransactionHashMultisigParticipantBuildModelSuccess) BuildModel(rows *sql.Rows) (
	participants []*model.MultiSignatureParticipant,
	err error,
) {
	for _, mockParticipant := range mockMultisigParticipants {
		participants = append(participants, mockParticipant)
	}
	return participants, nil
}

func TestMultisigService_GetPendingTransactionDetailByTransactionHash(t *testing.T) {
	type fields struct {
		Executor                       query.ExecutorInterface
		BlockService                   coreService.BlockServiceInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultisignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		Logger                         *logrus.Logger
	}
	type args struct {
		param *model.GetPendingTransactionDetailByTransactionHashRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionDetailByTransactionHashResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-getlastblock-error",
			fields: fields{
				Executor:                       nil,
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceFail{},
				PendingTransactionQuery:        nil,
				PendingSignatureQuery:          nil,
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-wrongTxHashHex",
			fields: fields{
				Executor:                       nil,
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        nil,
				PendingSignatureQuery:          nil,
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: &model.GetPendingTransactionDetailByTransactionHashRequest{
					TransactionHashHex: "PPPP",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-no-pendingTx",
			fields: fields{
				Executor:                       &mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess{},
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        &mockGetPendingTransactionByTransactionHashPendingQueryScanNoRow{},
				PendingSignatureQuery:          nil,
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},

		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-scanError",
			fields: fields{
				Executor:                       &mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess{},
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        &mockGetPendingTransactionByTransactionHashPendingQueryScanOtherError{},
				PendingSignatureQuery:          nil,
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-executorPendingSigFail",
			fields: fields{
				Executor:                       &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorFail{},
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:          nil,
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-QueryPendingSigFail",
			fields: fields{
				Executor:                       &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess{},
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:          &mockGetPendingTransactionByTransactionHashPendingSigQueryBuildFail{},
				MultisignatureInfoQuery:        nil,
				MultisignatureParticipantQuery: nil,
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-fail-ScanMultsigiInfoFailOtherError",
			fields: fields{
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:   &mockGetPendingTransactionByTransactionHashPendingSigQueryBuildSuccess{},
				MultisignatureInfoQuery: &mockGetPendingTransactionByTransactionHashMultisigInfoScanFailOtherError{},
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash-Success",
			fields: fields{
				Executor:                       &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess{},
				BlockService:                   &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery:        &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:          &mockGetPendingTransactionByTransactionHashPendingSigQueryBuildSuccess{},
				MultisignatureInfoQuery:        &mockGetPendingTransactionByTransactionHashMultisigInfoScanSuccess{},
				MultisignatureParticipantQuery: &mockGetPendingTransactionByTransactionHashMultisigParticipantBuildModelSuccess{},
				Logger:                         logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want: &model.GetPendingTransactionDetailByTransactionHashResponse{
				PendingTransaction: mockPendingTransaction,
				PendingSignatures:  []*model.PendingSignature{},
				MultiSignatureInfo: mockMultisigInfoWithParticipants,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MultisigService{
				Executor:                       tt.fields.Executor,
				BlockService:                   tt.fields.BlockService,
				PendingTransactionQuery:        tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:          tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery:        tt.fields.MultisignatureInfoQuery,
				MultiSignatureParticipantQuery: tt.fields.MultisignatureParticipantQuery,
				Logger:                         tt.fields.Logger,
			}
			got, err := ms.GetPendingTransactionDetailByTransactionHash(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPendingTransactionDetailByTransactionHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPendingTransactionDetailByTransactionHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMultisigService(t *testing.T) {
	type args struct {
		executor                       query.ExecutorInterface
		blockService                   coreService.BlockServiceInterface
		pendingTransactionQuery        query.PendingTransactionQueryInterface
		pendingSignatureQuery          query.PendingSignatureQueryInterface
		multisignatureQuery            query.MultisignatureInfoQueryInterface
		multiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
	}
	tests := []struct {
		name string
		args args
		want *MultisigService
	}{
		{
			name: "NewMultisigService-success",
			args: args{
				executor:                       nil,
				blockService:                   nil,
				pendingTransactionQuery:        nil,
				pendingSignatureQuery:          nil,
				multisignatureQuery:            nil,
				multiSignatureParticipantQuery: nil,
			},
			want: &MultisigService{
				Executor:                       nil,
				BlockService:                   nil,
				PendingTransactionQuery:        nil,
				PendingSignatureQuery:          nil,
				MultiSignatureParticipantQuery: nil,
				Logger:                         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMultisigService(tt.args.executor, tt.args.blockService, tt.args.pendingTransactionQuery,
				tt.args.pendingSignatureQuery, tt.args.multisignatureQuery,
				tt.args.multiSignatureParticipantQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultisigService() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// mock GetMultisigInfo
	mockGetMultisigInfoExecutorParam = &model.GetMultisignatureInfoRequest{
		MultisigAddress: []byte{0, 0, 0, 0, 160, 121, 129, 83, 225, 164, 195, 123, 8, 181, 41, 251, 17, 3, 93, 37, 182, 109, 32, 174, 168,
			68, 193, 212, 79, 54, 156, 213, 117, 27, 185, 167},
		Pagination: &model.Pagination{
			OrderField: "block_height",
			OrderBy:    model.OrderBy_DESC,
			Page:       1,
			Limit:      1,
		},
	}
	// mock GetMultisigInfo
)

type (
	mockGetMultisigInfoExecutorCountFailNoRow struct {
		query.Executor
	}
	mockGetMultisigInfoExecutorCountFailOther struct {
		query.Executor
	}
	mockGetMultisigInfoExecutorExecuteSelectError struct {
		query.Executor
	}
	mockGetMultisigInfoExecutorSuccess struct {
		query.Executor
	}
	mockGetMultisigInfoQueryBuildFail struct {
		query.MultisignatureInfoQuery
	}
	mockGetMultisigInfoQueryBuildSuccess struct {
		query.MultisignatureInfoQuery
	}
)

func (*mockGetMultisigInfoExecutorCountFailNoRow) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigInfoExecutorCountFailOther) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total", "Other", "Other"}).AddRow(1, 1, 1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigInfoExecutorExecuteSelectError) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigInfoExecutorExecuteSelectError) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetMultisigInfoExecutorSuccess) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigInfoExecutorSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}).AddRow(1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockGetMultisigInfoQueryBuildFail) BuildModel(
	multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetMultisigInfoQueryBuildSuccess) BuildModel(
	multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	return []*model.MultiSignatureInfo{}, nil
}

func TestMultisigService_GetMultisignatureInfo(t *testing.T) {
	type fields struct {
		Executor                query.ExecutorInterface
		BlockService            coreService.BlockServiceInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		Logger                  *logrus.Logger
	}
	type args struct {
		param *model.GetMultisignatureInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisignatureInfoResponse
		wantErr bool
	}{
		{
			name: "GetMultisignatureInfo-fail-countRow.Scan()-ErrorNoRow",
			fields: fields{
				Executor:                &mockGetMultisigInfoExecutorCountFailNoRow{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetMultisigInfoExecutorParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo-fail-countRow.Scan()-ErrorOther",
			fields: fields{
				Executor:                &mockGetMultisigInfoExecutorCountFailOther{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetMultisigInfoExecutorParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo-fail-multisigInfoExecuteSelect-Error",
			fields: fields{
				Executor:                &mockGetMultisigInfoExecutorExecuteSelectError{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetMultisigInfoExecutorParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo-fail-multisigInfoQueryBuild-Error",
			fields: fields{
				Executor:                &mockGetMultisigInfoExecutorSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: &mockGetMultisigInfoQueryBuildFail{},
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetMultisigInfoExecutorParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo-success",
			fields: fields{
				Executor:                &mockGetMultisigInfoExecutorSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: &mockGetMultisigInfoQueryBuildSuccess{},
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetMultisigInfoExecutorParam,
			},
			want: &model.GetMultisignatureInfoResponse{
				Count:              1,
				Page:               mockGetMultisigInfoExecutorParam.GetPagination().Page,
				MultisignatureInfo: []*model.MultiSignatureInfo{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MultisigService{
				Executor:                tt.fields.Executor,
				BlockService:            tt.fields.BlockService,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:   tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				Logger:                  tt.fields.Logger,
			}
			got, err := ms.GetMultisignatureInfo(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMultisignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMultisignatureInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetMultisigAddressByParticipantAddressGetTotalNotFound struct {
		query.Executor
	}
	mockGetMultisigAddressByParticipantAddressGetTotalInternalError struct {
		query.Executor
	}
	mockGetMultisigAddressByParticipantAddressExecuteSelectError struct {
		query.Executor
	}
	mockGetMultisigAddressByParticipantAddressSuccess struct {
		query.Executor
	}
)

func (*mockGetMultisigAddressByParticipantAddressGetTotalNotFound) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigAddressByParticipantAddressGetTotalInternalError) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow("NULL"))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetMultisigAddressByParticipantAddressExecuteSelectError) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}
func (*mockGetMultisigAddressByParticipantAddressExecuteSelectError) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetMultisigAddressByParticipantAddressSuccess) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	rows := db.QueryRow(qe)
	return rows, nil
}
func (*mockGetMultisigAddressByParticipantAddressSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

func TestMultisigService_GetMultisigAddressByParticipantAddress(t *testing.T) {
	type fields struct {
		Executor                       query.ExecutorInterface
		BlockService                   coreService.BlockServiceInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		Logger                         *logrus.Logger
	}
	type args struct {
		param *model.GetMultisigAddressByParticipantAddressRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisigAddressByParticipantAddressResponse
		wantErr bool
	}{
		{
			name: "GetTotal:NotFound",
			fields: fields{
				Executor: &mockGetMultisigAddressByParticipantAddressGetTotalNotFound{},
			},
			args: args{
				param: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTotal:InternalError",
			fields: fields{
				Executor: &mockGetMultisigAddressByParticipantAddressGetTotalInternalError{},
			},
			args: args{
				param: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisigAddressByParticipantAddress-ExecuteSelect:Error",
			fields: fields{
				Executor: &mockGetMultisigAddressByParticipantAddressExecuteSelectError{},
			},
			args: args{
				param: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisigAddressByParticipantAddress-Success",
			fields: fields{
				Executor: &mockGetMultisigAddressByParticipantAddressSuccess{},
			},
			args: args{
				param: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want: &model.GetMultisigAddressByParticipantAddressResponse{
				Total:             1,
				MultisigAddresses: [][]byte{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MultisigService{
				Executor:                       tt.fields.Executor,
				BlockService:                   tt.fields.BlockService,
				PendingTransactionQuery:        tt.fields.PendingTransactionQuery,
				PendingSignatureQuery:          tt.fields.PendingSignatureQuery,
				MultisignatureInfoQuery:        tt.fields.MultisignatureInfoQuery,
				MultiSignatureParticipantQuery: tt.fields.MultiSignatureParticipantQuery,
				Logger:                         tt.fields.Logger,
			}
			got, err := ms.GetMultisigAddressByParticipantAddress(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigService.GetMultisigAddressByParticipantAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigService.GetMultisigAddressByParticipantAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
