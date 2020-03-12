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
	"github.com/zoobc/zoobc-core/core/service"
)

var (
	// mock GetPendingTransactionByAddress
	mockGetPendingTransactionByAddressExecutorCountFailParam = &model.GetPendingTransactionByAddressRequest{
		SenderAddress: "abc",
		Status:        model.PendingTransactionStatus_PendingTransactionPending,
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
	mockGetPendingTransactionByAddressExecutorCountFail struct {
		query.Executor
	}
	mockGetPendingTransactionByAddressExecutorGetPendingTxsFail struct {
		query.Executor
	}
	mockGetPendingTransactionByAddressExecutorGetPendingTxsSuccess struct {
		query.Executor
	}
	mockGetPendingTransactionByAddressPendingTxQueryBuildFail struct {
		query.PendingTransactionQuery
	}
	mockGetPendingTransactionByAddressPendingTxQueryBuildSuccess struct {
		query.PendingTransactionQuery
	}
	// mock GetPendingTransactionByAddress
)

func (*mockGetPendingTransactionByAddressExecutorCountFail) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionByAddressExecutorGetPendingTxsFail) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionByAddressExecutorGetPendingTxsFail) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionByAddressExecutorGetPendingTxsSuccess) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"Total"}).AddRow(1))
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockGetPendingTransactionByAddressExecutorGetPendingTxsSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"mockedColumn"}).AddRow(1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockGetPendingTransactionByAddressPendingTxQueryBuildFail) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPendingTransactionByAddressPendingTxQueryBuildSuccess) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	return []*model.PendingTransaction{}, nil
}

func TestMultisigService_GetPendingTransactionByAddress(t *testing.T) {
	type fields struct {
		Executor                query.ExecutorInterface
		BlockService            service.BlockServiceInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		Logger                  *logrus.Logger
	}
	type args struct {
		param *model.GetPendingTransactionByAddressRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionByAddressResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionByAddress-fail-countExecuteSelectRow-error-noRow",
			fields: fields{
				Executor:                &mockGetPendingTransactionByAddressExecutorCountFail{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionByAddressExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-fail-GetPendingTxsExecutor-error",
			fields: fields{
				Executor:                &mockGetPendingTransactionByAddressExecutorGetPendingTxsFail{},
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionByAddressExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-fail-PendingTxQueryBuild-error",
			fields: fields{
				Executor:                &mockGetPendingTransactionByAddressExecutorGetPendingTxsSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: &mockGetPendingTransactionByAddressPendingTxQueryBuildFail{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionByAddressExecutorCountFailParam,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionByAddress-success",
			fields: fields{
				Executor:                &mockGetPendingTransactionByAddressExecutorGetPendingTxsSuccess{},
				BlockService:            nil,
				PendingTransactionQuery: &mockGetPendingTransactionByAddressPendingTxQueryBuildSuccess{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
			args: args{
				param: mockGetPendingTransactionByAddressExecutorCountFailParam,
			},
			want: &model.GetPendingTransactionByAddressResponse{
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
			got, err := ms.GetPendingTransactionByAddress(tt.args.param)
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
		SenderAddress:    "ABC",
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 100),
		Status:           model.PendingTransactionStatus_PendingTransactionPending,
		BlockHeight:      800,
		Latest:           true,
	}
	mockMultisigInfo = &model.MultiSignatureInfo{
		MinimumSignatures: 2,
		Nonce:             3,
		Addresses:         []string{"A", "B", "C"},
		MultisigAddress:   "ABC",
		BlockHeight:       400,
		Latest:            true,
	}

// mock GetPendingTransactionByAddress
)

type (
	mockGetPendingTransactionByTransactionHashBlockServiceFail struct {
		service.BlockService
	}

	mockGetPendingTransactionByTransactionHashBlockServiceSuccess struct {
		service.BlockService
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

func TestMultisigService_GetPendingTransactionDetailByTransactionHash(t *testing.T) {
	type fields struct {
		Executor                query.ExecutorInterface
		BlockService            service.BlockServiceInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		Logger                  *logrus.Logger
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
				Executor:                nil,
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceFail{},
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                nil,
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanNoRow{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingTxExecutorSuccess{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanOtherError{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorFail{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:   &mockGetPendingTransactionByTransactionHashPendingSigQueryBuildFail{},
				MultisignatureInfoQuery: nil,
				Logger:                  logrus.New(),
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
				Executor:                &mockGetPendingTransactionByTransactionHashGetPendingSigExecutorSuccess{},
				BlockService:            &mockGetPendingTransactionByTransactionHashBlockServiceSuccess{},
				PendingTransactionQuery: &mockGetPendingTransactionByTransactionHashPendingQueryScanSuccess{},
				PendingSignatureQuery:   &mockGetPendingTransactionByTransactionHashPendingSigQueryBuildSuccess{},
				MultisignatureInfoQuery: &mockGetPendingTransactionByTransactionHashMultisigInfoScanSuccess{},
				Logger:                  logrus.New(),
			},
			args: args{
				param: mockGetPendingTransactionDetailByTransactionHashExecutorCountFailParam,
			},
			want: &model.GetPendingTransactionDetailByTransactionHashResponse{
				PendingTransaction: mockPendingTransaction,
				PendingSignatures:  []*model.PendingSignature{},
				MultiSignatureInfo: mockMultisigInfo,
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
		executor                query.ExecutorInterface
		blockService            service.BlockServiceInterface
		pendingTransactionQuery query.PendingTransactionQueryInterface
		pendingSignatureQuery   query.PendingSignatureQueryInterface
		multisignatureQuery     query.MultisignatureInfoQueryInterface
	}
	tests := []struct {
		name string
		args args
		want *MultisigService
	}{
		{
			name: "NewMultisigService-success",
			args: args{
				executor:                nil,
				blockService:            nil,
				pendingTransactionQuery: nil,
				pendingSignatureQuery:   nil,
				multisignatureQuery:     nil,
			},
			want: &MultisigService{
				Executor:                nil,
				BlockService:            nil,
				PendingTransactionQuery: nil,
				PendingSignatureQuery:   nil,
				MultisignatureInfoQuery: nil,
				Logger:                  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMultisigService(
				tt.args.executor, tt.args.blockService, tt.args.pendingTransactionQuery, tt.args.pendingSignatureQuery,
				tt.args.multisignatureQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultisigService() = %v, want %v", got, tt.want)
			}
		})
	}
}
