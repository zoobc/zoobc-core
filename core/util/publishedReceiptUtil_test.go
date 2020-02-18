package util

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewPublishedReceiptUtil(t *testing.T) {
	type args struct {
		publishedReceiptQuery query.PublishedReceiptQueryInterface
		queryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *PublishedReceiptUtil
	}{
		{
			name: "NewPublishedReceiptUtil-Success",
			args: args{
				publishedReceiptQuery: nil,
				queryExecutor:         nil,
			},
			want: &PublishedReceiptUtil{
				PublishedReceiptQuery: nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublishedReceiptUtil(tt.args.publishedReceiptQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPublishedReceiptUtil() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// GetPublishedReceiptByLinkedRMR mocks
	mockGetPublishedReceiptByLinkedRMRExecutorSuccess struct {
		query.Executor
	}
	mockGetPublishedReceiptByLinkedRMRPublishedReceiptQueryFail struct {
		query.PublishedReceiptQuery
	}
	mockGetPublishedReceiptByLinkedRMRPublishedReceiptQuerySuccess struct {
		query.PublishedReceiptQuery
	}
	// GetPublishedReceiptByLinkedRMR mocks
)

var (
	// GetPublishedReceiptByLinkedRMR mocks
	mockGetPublishedReceiptByRMRResult = &model.PublishedReceipt{
		BatchReceipt:       &model.BatchReceipt{},
		IntermediateHashes: nil,
		BlockHeight:        1,
		ReceiptIndex:       1,
	}
	// GetPublishedReceiptByLinkedRMR mocks
)

func (*mockGetPublishedReceiptByLinkedRMRExecutorSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockGetPublishedReceiptByLinkedRMRPublishedReceiptQueryFail) Scan(publishedReceipt *model.PublishedReceipt, row *sql.Row) error {
	return errors.New("mockedError")
}

func (*mockGetPublishedReceiptByLinkedRMRPublishedReceiptQuerySuccess) Scan(publishedReceipt *model.PublishedReceipt, row *sql.Row) error {
	*publishedReceipt = *mockGetPublishedReceiptByRMRResult
	return nil
}

func TestPublishedReceiptUtil_GetPublishedReceiptByLinkedRMR(t *testing.T) {
	type fields struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		root []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.PublishedReceipt
		wantErr bool
	}{
		{
			name: "GetPublishedReceiptByLinkedRMR-ScanFail",
			fields: fields{
				PublishedReceiptQuery: &mockGetPublishedReceiptByLinkedRMRPublishedReceiptQueryFail{},
				QueryExecutor:         &mockGetPublishedReceiptByLinkedRMRExecutorSuccess{},
			},
			args: args{
				root: make([]byte, 32),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublishedReceiptByLinkedRMR-Success",
			fields: fields{
				PublishedReceiptQuery: &mockGetPublishedReceiptByLinkedRMRPublishedReceiptQuerySuccess{},
				QueryExecutor:         &mockGetPublishedReceiptByLinkedRMRExecutorSuccess{},
			},
			args: args{
				root: make([]byte, 32),
			},
			want:    mockGetPublishedReceiptByRMRResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psu := &PublishedReceiptUtil{
				PublishedReceiptQuery: tt.fields.PublishedReceiptQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := psu.GetPublishedReceiptByLinkedRMR(tt.args.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublishedReceiptByLinkedRMR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublishedReceiptByLinkedRMR() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// GetPublishedReceiptByLinkedRMR mocks
	mockGetPublishedReceiptsByBlockHeightExecutorFail struct {
		query.Executor
	}

	mockGetPublishedReceiptsByBlockHeightExecutorSuccess struct {
		query.Executor
	}

	mockGetPublishedReceiptsByBlockHeightPublishedReceiptQueryFail struct {
		query.PublishedReceiptQuery
	}

	mockGetPublishedReceiptsByBlockHeightPublishedReceiptQuerySuccess struct {
		query.PublishedReceiptQuery
	}
	// GetPublishedReceiptByLinkedRMR mocks
)

var (
	mockGetPublishedReceiptByBlockHeightResult = []*model.PublishedReceipt{
		{
			BatchReceipt:       nil,
			IntermediateHashes: nil,
			BlockHeight:        1,
			ReceiptIndex:       2,
			PublishedIndex:     3,
		},
	}
)

func (*mockGetPublishedReceiptsByBlockHeightExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPublishedReceiptsByBlockHeightExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetPublishedReceiptsByBlockHeightPublishedReceiptQueryFail) BuildModel(
	prs []*model.PublishedReceipt, rows *sql.Rows,
) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetPublishedReceiptsByBlockHeightPublishedReceiptQuerySuccess) BuildModel(
	prs []*model.PublishedReceipt, rows *sql.Rows,
) ([]*model.PublishedReceipt, error) {
	return mockGetPublishedReceiptByBlockHeightResult, nil
}

func TestPublishedReceiptUtil_GetPublishedReceiptsByBlockHeight(t *testing.T) {
	type fields struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PublishedReceipt
		wantErr bool
	}{
		{
			name: "GetPublishedReceiptsByBlockHeight-ExecuteSelectFail",
			fields: fields{
				PublishedReceiptQuery: &mockGetPublishedReceiptsByBlockHeightPublishedReceiptQuerySuccess{},
				QueryExecutor:         &mockGetPublishedReceiptsByBlockHeightExecutorFail{},
			},
			args: args{
				blockHeight: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublishedReceiptsByBlockHeight-BuildModelFail",
			fields: fields{
				PublishedReceiptQuery: &mockGetPublishedReceiptsByBlockHeightPublishedReceiptQueryFail{},
				QueryExecutor:         &mockGetPublishedReceiptsByBlockHeightExecutorSuccess{},
			},
			args: args{
				blockHeight: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublishedReceiptsByBlockHeight-Success",
			fields: fields{
				PublishedReceiptQuery: &mockGetPublishedReceiptsByBlockHeightPublishedReceiptQuerySuccess{},
				QueryExecutor:         &mockGetPublishedReceiptsByBlockHeightExecutorSuccess{},
			},
			args: args{
				blockHeight: 1,
			},
			want:    mockGetPublishedReceiptByBlockHeightResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psu := &PublishedReceiptUtil{
				PublishedReceiptQuery: tt.fields.PublishedReceiptQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := psu.GetPublishedReceiptsByBlockHeight(tt.args.blockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublishedReceiptsByBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublishedReceiptsByBlockHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// InsertPublishedReceipt mocks
	mockInsertPublishedReceiptExecutorFail struct {
		query.Executor
	}
	mockInsertPublishedReceiptExecutorSuccess struct {
		query.Executor
	}
	// InsertPublishedReceipt mocks
)

func (*mockInsertPublishedReceiptExecutorFail) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("mockedError")
}

func (*mockInsertPublishedReceiptExecutorFail) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("mockedError")
}

func (*mockInsertPublishedReceiptExecutorSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockInsertPublishedReceiptExecutorSuccess) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestPublishedReceiptUtil_InsertPublishedReceipt(t *testing.T) {
	dummyPublishedReceipt := &model.PublishedReceipt{
		BatchReceipt: &model.BatchReceipt{},
	}
	type fields struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		publishedReceipt *model.PublishedReceipt
		tx               bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "InsertPublishedReceipt-txFalse-Fail",
			fields: fields{
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				QueryExecutor:         &mockInsertPublishedReceiptExecutorFail{},
			},
			args: args{
				publishedReceipt: dummyPublishedReceipt,
				tx:               false,
			},
			wantErr: true,
		},
		{
			name: "InsertPublishedReceipt-txFalse-Success",
			fields: fields{
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				QueryExecutor:         &mockInsertPublishedReceiptExecutorSuccess{},
			},
			args: args{
				publishedReceipt: dummyPublishedReceipt,
				tx:               false,
			},
			wantErr: false,
		},
		{
			name: "InsertPublishedReceipt-txTrue-Fail",
			fields: fields{
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				QueryExecutor:         &mockInsertPublishedReceiptExecutorFail{},
			},
			args: args{
				publishedReceipt: dummyPublishedReceipt,
				tx:               true,
			},
			wantErr: true,
		},
		{
			name: "InsertPublishedReceipt-txTrue-Success",
			fields: fields{
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				QueryExecutor:         &mockInsertPublishedReceiptExecutorSuccess{},
			},
			args: args{
				publishedReceipt: dummyPublishedReceipt,
				tx:               true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psu := &PublishedReceiptUtil{
				PublishedReceiptQuery: tt.fields.PublishedReceiptQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := psu.InsertPublishedReceipt(tt.args.publishedReceipt, tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("InsertPublishedReceipt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
