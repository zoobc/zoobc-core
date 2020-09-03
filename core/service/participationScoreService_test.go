package service

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

func TestNewParticipationScoreService(t *testing.T) {
	type args struct {
		participationScoreQuery query.ParticipationScoreQueryInterface
		queryExecutor           query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *ParticipationScoreService
	}{
		{
			name: "NewParticipationScore",
			args: args{
				participationScoreQuery: nil,
				queryExecutor:           nil,
			},
			want: &ParticipationScoreService{
				ParticipationScoreQuery: nil,
				QueryExecutor:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewParticipationScoreService(tt.args.participationScoreQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewParticipationScoreService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// GetParticipationScore mocks
	mockGetParticipationScoreExecutorFail struct {
		query.Executor
	}
	mockGetParticipationScoreExecutorSuccess struct {
		query.Executor
	}
	mockGetParticipationScoreParticipationScoreQuerySuccess struct {
		query.ParticipationScoreQuery
	}
	mockGetParticipationScoreParticipationScoreQueryFail struct {
		query.ParticipationScoreQuery
	}
	// GetParticipationScore mocks

	// GetParticipationScoreByBlockHeightRange mocks
	mockGetParticipationScoreByBlockHeightRangeExecutorFail struct {
		query.Executor
	}
	mockGetParticipationScoreByBlockHeightRangeExecutorSuccess struct {
		query.Executor
	}
	mockGetParticipationScoreByBlockHeightRangeParticipationScoreQuerySuccess struct {
		query.ParticipationScoreQuery
	}
	mockGetParticipationScoreByBlockHeightRangeParticipationScoreQueryFail struct {
		query.ParticipationScoreQuery
	}
	// GetParticipationScoreByBlockHeightRange mocks
)

var (
	// GetParticipationScore mocks
	mockGetParticipationScoreResult = &model.ParticipationScore{
		Score: 1000,
	}
	// GetParticipationScore mocks
)

func (*mockGetParticipationScoreExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockGetParticipationScoreExecutorSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetParticipationScoreParticipationScoreQuerySuccess) BuildModel(
	participationScores []*model.ParticipationScore, rows *sql.Rows,
) ([]*model.ParticipationScore, error) {
	return []*model.ParticipationScore{
		mockGetParticipationScoreResult,
	}, nil
}

func (*mockGetParticipationScoreParticipationScoreQueryFail) BuildModel(
	participationScores []*model.ParticipationScore, rows *sql.Rows,
) ([]*model.ParticipationScore, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetParticipationScoreByBlockHeightRangeExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockGetParticipationScoreByBlockHeightRangeExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetParticipationScoreByBlockHeightRangeParticipationScoreQuerySuccess) BuildModel(
	participationScores []*model.ParticipationScore, rows *sql.Rows) ([]*model.ParticipationScore, error) {
	return []*model.ParticipationScore{
		mockGetParticipationScoreResult,
	}, nil
}

func (*mockGetParticipationScoreByBlockHeightRangeParticipationScoreQueryFail) BuildModel(
	participationScores []*model.ParticipationScore, rows *sql.Rows) ([]*model.ParticipationScore, error) {
	return nil, errors.New("mockError")
}

func TestParticipationScoreService_GetParticipationScore(t *testing.T) {
	type fields struct {
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		nodePublicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GetParticipationScore-ExecuteSelectFail",
			fields: fields{
				ParticipationScoreQuery: &mockGetParticipationScoreParticipationScoreQuerySuccess{},
				QueryExecutor:           &mockGetParticipationScoreExecutorFail{},
			},
			args: args{
				nodePublicKey: make([]byte, 32),
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "GetParticipationScore-BuildModelFail-OR-ReturnEmptySlice",
			fields: fields{
				ParticipationScoreQuery: &mockGetParticipationScoreParticipationScoreQueryFail{},
				QueryExecutor:           &mockGetParticipationScoreExecutorSuccess{},
			},
			args: args{
				nodePublicKey: make([]byte, 32),
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "GetParticipationScore-Success",
			fields: fields{
				ParticipationScoreQuery: &mockGetParticipationScoreParticipationScoreQuerySuccess{},
				QueryExecutor:           &mockGetParticipationScoreExecutorSuccess{},
			},
			args: args{
				nodePublicKey: make([]byte, 32),
			},
			want:    mockGetParticipationScoreResult.Score,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pss := &ParticipationScoreService{
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			got, err := pss.GetParticipationScore(tt.args.nodePublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetParticipationScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetParticipationScore() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParticipationScoreService_GetParticipationScoreByBlockHeightRange(t *testing.T) {
	type fields struct {
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		fromBlockHeight uint32
		toBlockHeight   uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.ParticipationScore
		wantErr bool
	}{
		{
			name: "wantError:ExecuteSelectError",
			fields: fields{
				QueryExecutor:           &mockGetParticipationScoreByBlockHeightRangeExecutorFail{},
				ParticipationScoreQuery: &mockGetParticipationScoreByBlockHeightRangeParticipationScoreQuerySuccess{},
			},
			args: args{
				fromBlockHeight: 1,
				toBlockHeight:   2,
			},
			wantErr: true,
		},
		{
			name: "wantError:BuildModelError",
			fields: fields{
				QueryExecutor:           &mockGetParticipationScoreByBlockHeightRangeExecutorSuccess{},
				ParticipationScoreQuery: &mockGetParticipationScoreByBlockHeightRangeParticipationScoreQueryFail{},
			},
			args: args{
				fromBlockHeight: 1,
				toBlockHeight:   2,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor:           &mockGetParticipationScoreByBlockHeightRangeExecutorSuccess{},
				ParticipationScoreQuery: &mockGetParticipationScoreByBlockHeightRangeParticipationScoreQuerySuccess{},
			},
			args: args{
				fromBlockHeight: 1,
				toBlockHeight:   2,
			},
			want: []*model.ParticipationScore{
				mockGetParticipationScoreResult,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pss := &ParticipationScoreService{
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			got, err := pss.GetParticipationScoreByBlockHeightRange(tt.args.fromBlockHeight, tt.args.toBlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParticipationScoreService.GetParticipationScoreByBlockHeightRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParticipationScoreService.GetParticipationScoreByBlockHeightRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
