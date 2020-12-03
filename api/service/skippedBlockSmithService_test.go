package service

import (
	"database/sql"
	"errors"
	chaintype2 "github.com/zoobc/zoobc-core/common/chaintype"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewSkippedBlockSmithService(t *testing.T) {
	type args struct {
		skippedBlocksmithQuery *query.SkippedBlocksmithQuery
		queryExecutor          query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want SkippedBlockSmithServiceInterface
	}{
		{
			name: "wantSuccess",
			args: args{
				skippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(&chaintype2.MainChain{}),
				queryExecutor:          &query.Executor{},
			},
			want: &SkippedBlockSmithService{
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(&chaintype2.MainChain{}),
				QueryExecutor:          &query.Executor{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSkippedBlockSmithService(tt.args.skippedBlocksmithQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSkippedBlockSmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetSkippedBlockSmithsSelectFail struct {
		query.Executor
	}
	mockGetSkippedBlockSmithsSelectSuccess struct {
		query.Executor
	}
)

func (*mockGetSkippedBlockSmithsSelectFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}
func (*mockGetSkippedBlockSmithsSelectFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}
func (*mockGetSkippedBlockSmithsSelectSuccess) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockRows := mock.NewRows(query.NewSkippedBlocksmithQuery(&chaintype2.MainChain{}).Fields)
	mockRows.AddRow(
		[]byte{1},
		1,
		1,
		1,
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}
func (*mockGetSkippedBlockSmithsSelectSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}

func TestSkippedBlockSmithService_GetSkippedBlockSmiths(t *testing.T) {
	type fields struct {
		QueryExecutor          query.ExecutorInterface
		SkippedBlocksmithQuery *query.SkippedBlocksmithQuery
	}
	type args struct {
		req *model.GetSkippedBlocksmithsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetSkippedBlocksmithsResponse
		wantErr bool
	}{
		{
			name: "wantFail:",
			fields: fields{
				QueryExecutor:          &mockGetSkippedBlockSmithsSelectFail{},
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(&chaintype2.MainChain{}),
			},
			args: args{
				req: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 1,
					BlockHeightEnd:   2,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "WantSuccess",
			fields: fields{
				QueryExecutor:          &mockGetSkippedBlockSmithsSelectSuccess{},
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(&chaintype2.MainChain{}),
			},
			args: args{
				req: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 1,
					BlockHeightEnd:   2,
				},
			},
			want: &model.GetSkippedBlocksmithsResponse{
				Total: 1,
				SkippedBlocksmiths: []*model.SkippedBlocksmith{
					{
						BlocksmithPublicKey: []byte{1},
						POPChange:           1,
						BlockHeight:         1,
						BlocksmithIndex:     1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbs := &SkippedBlockSmithService{
				QueryExecutor:          tt.fields.QueryExecutor,
				SkippedBlocksmithQuery: tt.fields.SkippedBlocksmithQuery,
			}
			got, err := sbs.GetSkippedBlockSmiths(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkippedBlockSmithService.GetSkippedBlockSmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SkippedBlockSmithService.GetSkippedBlockSmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}
