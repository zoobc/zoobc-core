package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockFeeVoteCommitmentVoteQuery = NewFeeVoteCommitmentVoteQuery()
	mockFeeVoteCommitmentVote      = model.FeeVoteCommitmentVote{
		VoteHash:     []byte{1, 2, 1},
		VoterAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		BlockHeight:  1,
	}
)

func TestNewFeeVoteCommitmentVoteQuery(t *testing.T) {
	tests := []struct {
		name string
		want *FeeVoteCommitmentVoteQuery
	}{
		{
			name: "wantSuccess",
			want: mockFeeVoteCommitmentVoteQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeeVoteCommitmentVoteQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeeVoteCommitmentVoteQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteCommitmentVoteQuery_InsertCommitVote(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		voteCommit *model.FeeVoteCommitmentVote
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
			fields: fields(*mockFeeVoteCommitmentVoteQuery),
			args: args{
				voteCommit: &mockFeeVoteCommitmentVote,
			},
			wantStr: "INSERT INTO fee_vote_commitment_vote (vote_hash,voter_address,block_height) VALUES(? , ?, ?)",
			wantArgs: []interface{}{
				mockFeeVoteCommitmentVote.GetVoteHash(),
				mockFeeVoteCommitmentVote.GetVoterAddress(),
				mockFeeVoteCommitmentVote.GetBlockHeight(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeVoteCommitmentVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := fsvc.InsertCommitVote(tt.args.voteCommit)
			if gotStr != tt.wantStr {
				t.Errorf("FeeVoteCommitmentVoteQuery.InsertCommitVote() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("FeeVoteCommitmentVoteQuery.InsertCommitVote() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFeeVoteCommitmentVoteQuery_GetVoteCommitByAccountAddressAndHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountAddress string
		height         uint32
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
			fields: fields(*mockFeeVoteCommitmentVoteQuery),
			args: args{
				accountAddress: mockFeeVoteCommitmentVote.GetVoterAddress(),
				height:         mockFeeVoteCommitmentVote.GetBlockHeight(),
			},
			wantQStr: "SELECT vote_hash,voter_address,block_height FROM fee_vote_commitment_vote WHERE voter_address = ? AND block_height>= ?",
			wantArgs: []interface{}{
				mockFeeVoteCommitmentVote.GetVoterAddress(),
				mockFeeVoteCommitmentVote.GetBlockHeight(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeVoteCommitmentVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := fsvc.GetVoteCommitByAccountAddressAndHeight(tt.args.accountAddress, tt.args.height)
			if gotQStr != tt.wantQStr {
				t.Errorf("FeeVoteCommitmentVoteQuery.GetVoteCommitByAccountAddressAndHeight() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("FeeVoteCommitmentVoteQuery.GetVoteCommitByAccountAddressAndHeight() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

type (
	mockRowFeeVoteCommitmentVoteQueryScan struct {
		Executor
	}
)

func (*mockRowFeeVoteCommitmentVoteQueryScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(NewFeeVoteCommitmentVoteQuery().Fields).AddRow(
			mockFeeVoteCommitmentVote.GetVoteHash(),
			mockFeeVoteCommitmentVote.GetVoterAddress(),
			mockFeeVoteCommitmentVote.GetBlockHeight(),
		),
	)
	return db.QueryRow("")
}

func TestFeeVoteCommitmentVoteQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		voteCommit *model.FeeVoteCommitmentVote
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
			fields: fields(*mockFeeVoteCommitmentVoteQuery),
			args: args{
				voteCommit: &model.FeeVoteCommitmentVote{},
				row:        (&mockRowFeeVoteCommitmentVoteQueryScan{}).ExecuteSelectRow("", ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FeeVoteCommitmentVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := f.Scan(tt.args.voteCommit, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("FeeVoteCommitmentVoteQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteCommitmentVoteQuery_Rollback(t *testing.T) {
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
			fields: fields(*mockFeeVoteCommitmentVoteQuery),
			args:   args{height: uint32(1)},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM fee_vote_commitment_vote WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsvc := &FeeVoteCommitmentVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := fsvc.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("FeeVoteCommitmentVoteQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}
