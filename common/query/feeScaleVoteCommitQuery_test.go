package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockFeeScaleVoteCommitsQuery = NewFeeScaleVoteCommitsQuery()
	mockFeeScaleVoteCommit       = model.FeeScaleVoteCommit{
		VoteHash:     []byte{1, 2, 1},
		VoterAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		BlockHeight:  1,
	}
)

func TestNewFeeScaleVoteCommitsQuery(t *testing.T) {
	tests := []struct {
		name string
		want *FeeScaleVoteCommitQuery
	}{
		{
			name: "wantSuccess",
			want: mockFeeScaleVoteCommitsQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeeScaleVoteCommitsQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeeScaleVoteCommitsQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeScaleVoteCommitQuery_InsertCommitVote(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		voteCommit *model.FeeScaleVoteCommit
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockFeeScaleVoteCommitsQuery),
			args: args{
				voteCommit: &mockFeeScaleVoteCommit,
			},
			wantStr: "INSERT INTO fee_scale_vote_commits (vote_hash,voter_address,block_height) VALUES(? , ?, ?)",
			wantArgs: []interface{}{
				mockFeeScaleVoteCommit.GetVoteHash(),
				mockFeeScaleVoteCommit.GetVoterAddress(),
				mockFeeScaleVoteCommit.GetBlockHeight(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeScaleVoteCommitQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := fsvc.InsertCommitVote(tt.args.voteCommit)
			if gotStr != tt.wantStr {
				t.Errorf("FeeScaleVoteCommitQuery.InsertCommitVote() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("FeeScaleVoteCommitQuery.InsertCommitVote() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFeeScaleVoteCommitQuery_GetVoteCommitByAccountAddress(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountAddress string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockFeeScaleVoteCommitsQuery),
			args: args{
				accountAddress: mockFeeScaleVoteCommit.GetVoterAddress(),
			},
			wantQStr: "SELECT vote_hash,voter_address,block_height FROM fee_scale_vote_commits WHERE voter_address = ? ORDER BY block_height DESC LIMIT 1",
			wantArgs: []interface{}{
				mockFeeScaleVoteCommit.GetVoterAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeScaleVoteCommitQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := fsvc.GetVoteCommitByAccountAddress(tt.args.accountAddress)
			if gotQStr != tt.wantQStr {
				t.Errorf("FeeScaleVoteCommitQuery.GetVoteCommitByAccountAddress() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("FeeScaleVoteCommitQuery.GetVoteCommitByAccountAddress() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

type (
	mockRowFeeScaleVoteCommitQueryScan struct {
		Executor
	}
)

func (*mockRowFeeScaleVoteCommitQueryScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(NewFeeScaleVoteCommitsQuery().Fields).AddRow(
			mockFeeScaleVoteCommit.GetVoteHash(),
			mockFeeScaleVoteCommit.GetVoterAddress(),
			mockFeeScaleVoteCommit.GetBlockHeight(),
		),
	)
	return db.QueryRow("")
}

func TestFeeScaleVoteCommitQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		voteCommit *model.FeeScaleVoteCommit
		row        *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockFeeScaleVoteCommitsQuery),
			args: args{
				voteCommit: &model.FeeScaleVoteCommit{},
				row:        (&mockRowFeeScaleVoteCommitQueryScan{}).ExecuteSelectRow("", ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeScaleVoteCommitQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := f.Scan(tt.args.voteCommit, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("FeeScaleVoteCommitQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeScaleVoteCommitQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockFeeScaleVoteCommitsQuery),
			args:   args{height: uint32(1)},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM fee_scale_vote_commits WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeScaleVoteCommitQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := fsvc.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("FeeScaleVoteCommitQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}
