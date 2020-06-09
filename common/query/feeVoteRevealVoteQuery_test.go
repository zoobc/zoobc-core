package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestFeeVoteRevealVoteQuery_GetFeeVoteRevealByAccountAddress(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountAddress string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewFeeVoteRevealVoteQuery()),
			args: args{
				accountAddress: "ABSCasjkdahsdasd",
			},
			want: "SELECT recent_block_hash, recent_block_height, fee_vote, voter_address, voter_signature, block_height " +
				"FROM fee_vote_reveal_vote WHERE voter_address = ? ORDER BY block_height DESC LIMIT 1",
			want1: []interface{}{"ABSCasjkdahsdasd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, got1 := fvr.GetFeeVoteRevealByAccountAddress(tt.args.accountAddress)
			if got != tt.want {
				t.Errorf("GetFeeVoteRevealByAccountAddress() got = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetFeeVoteRevealByAccountAddress() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestFeeVoteRevealVoteQuery_InsertRevealVote(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		revealVote *model.FeeVoteRevealVote
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQry  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewFeeVoteRevealVoteQuery()),
			args: args{
				revealVote: &model.FeeVoteRevealVote{
					VoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterAddress:   "ABSCasjkdahsdasd",
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
					BlockHeight:    230,
				},
			},
			wantQry: "INSERT INTO fee_vote_reveal_vote (recent_block_hash, recent_block_height, fee_vote, voter_address," +
				" voter_signature, block_height) VALUES(?, ?, ?, ?, ?, ?)",
			wantArgs: []interface{}{
				[]byte{1, 2, 3, 4, 5, 6, 7, 8},
				uint32(100),
				int64(10),
				"ABSCasjkdahsdasd",
				[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				uint32(230),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQry, gotArgs := fvr.InsertRevealVote(tt.args.revealVote)
			if gotQry != tt.wantQry {
				t.Errorf("InsertRevealVote() gotQry = %v, want %v", gotQry, tt.wantQry)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertRevealVote() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFeeVoteRevealVoteQuery_Scan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(mock.NewRows(NewFeeVoteRevealVoteQuery().Fields).
		AddRow(
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			100,
			10,
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			"ABSCasjkdahsdasd",
			230,
		))
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		vote *model.FeeVoteRevealVote
		row  *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewFeeVoteRevealVoteQuery()),
			args: args{
				vote: &model.FeeVoteRevealVote{
					VoteInfo: &model.FeeVoteInfo{
						RecentBlockHash:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
						RecentBlockHeight: 100,
						FeeVote:           10,
					},
					VoterAddress:   "ABSCasjkdahsdasd",
					VoterSignature: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
					BlockHeight:    230,
				},
				row: db.QueryRow(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := fvr.Scan(tt.args.vote, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeeVoteRevealVoteQuery_SelectDataForSnapshot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewFeeVoteRevealVoteQuery()),
			args: args{
				fromHeight: 100,
				toHeight:   170,
			},
			want: "SELECT recent_block_hash, recent_block_height, fee_vote, voter_address, voter_signature, block_height " +
				"FROM fee_vote_reveal_vote WHERE block_height >= 100 AND block_height <= 170",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := fvr.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeVoteRevealVoteQuery_TrimDataBeforeSnapshot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewFeeVoteRevealVoteQuery()),
			args: args{
				fromHeight: 100,
				toHeight:   170,
			},
			want: "DELETE FROM fee_vote_reveal_vote WHERE block_height >= 100 AND block_height <= 170",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := fvr.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
