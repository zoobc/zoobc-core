// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package query

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	feeVoterAccountAddress1 = []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
)

func TestFeeVoteRevealVoteQuery_GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountAddress []byte
		blockHeight    uint32
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
				accountAddress: feeVoterAccountAddress1,
				blockHeight:    100,
			},
			want: "SELECT recent_block_hash, recent_block_height, fee_vote, voter_address, voter_signature, block_height " +
				"FROM fee_vote_reveal_vote WHERE voter_address = ? AND recent_block_height = ? ORDER BY block_height DESC LIMIT 1",
			want1: []interface{}{feeVoterAccountAddress1, uint32(100)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fvr := &FeeVoteRevealVoteQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, got1 := fvr.GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(tt.args.accountAddress, tt.args.blockHeight)
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
					VoterAddress:   feeVoterAccountAddress1,
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
				feeVoterAccountAddress1,
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
			feeVoterAccountAddress1,
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
					VoterAddress:   feeVoterAccountAddress1,
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
				"FROM fee_vote_reveal_vote WHERE block_height >= 100 AND block_height <= 170 AND block_height != 0",
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
			want: "DELETE FROM fee_vote_reveal_vote WHERE block_height >= 100 AND block_height <= 170 AND block_height != 0",
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

var (
	mockFeeVoteRevealVoteQuery                    = NewFeeVoteRevealVoteQuery()
	mockFeeVoteRevealVoteQueryBuildModelRowResult = []*model.FeeVoteRevealVote{
		{
			VoteInfo: &model.FeeVoteInfo{
				RecentBlockHash:   make([]byte, 32),
				RecentBlockHeight: 100,
				FeeVote:           constant.OneZBC,
			},
			VoterSignature: make([]byte, 68),
			VoterAddress:   feeVoterAccountAddress1,
			BlockHeight:    120,
		},
		{
			VoteInfo: &model.FeeVoteInfo{
				RecentBlockHash:   make([]byte, 32),
				RecentBlockHeight: 105,
				FeeVote:           constant.OneZBC,
			},
			VoterSignature: make([]byte, 72),
			VoterAddress:   feeVoterAccountAddress1,
			BlockHeight:    130,
		},
	}
)

func TestFeeVoteRevealVoteQuery_BuildModel(t *testing.T) {
	t.Run("FeeVoteRevealVote:BuildModel", func(t *testing.T) {
		var err error
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockFeeVoteRevealVoteQuery.Fields).
			AddRow(
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetVoteInfo().GetRecentBlockHash(),
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetVoteInfo().GetRecentBlockHeight(),
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetVoteInfo().GetFeeVote(),
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetVoterAddress(),
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetVoterSignature(),
				mockFeeVoteRevealVoteQueryBuildModelRowResult[0].GetBlockHeight(),
			).AddRow(
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetVoteInfo().GetRecentBlockHash(),
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetVoteInfo().GetRecentBlockHeight(),
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetVoteInfo().GetFeeVote(),
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetVoterAddress(),
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetVoterSignature(),
			mockFeeVoteRevealVoteQueryBuildModelRowResult[1].GetBlockHeight(),
		))
		rows, _ := db.Query("")
		var result []*model.FeeVoteRevealVote
		result, err = mockFeeVoteRevealVoteQuery.BuildModel(result, rows)
		if err != nil {
			t.Errorf("error calling FeeVoteRevealVoteQuery.BuildModel - %v", err)
		}
		if !reflect.DeepEqual(result, mockFeeVoteRevealVoteQueryBuildModelRowResult) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", result, mockAccountBalance)
		}
	})
}

func TestFeeVoteRevealVoteQuery_GetFeeVoteRevealsInPeriod(t *testing.T) {
	t.Run("FeeVoteRevealVoteQuery:success", func(t *testing.T) {
		qry, args := mockFeeVoteRevealVoteQuery.GetFeeVoteRevealsInPeriod(0, 720)
		expectQry := "SELECT recent_block_hash, recent_block_height, fee_vote, voter_address, voter_signature, block_height " +
			"FROM fee_vote_reveal_vote WHERE block_height between ? AND ? ORDER BY fee_vote ASC"
		expectArgs := []interface{}{
			uint32(0),
			uint32(720),
		}
		if qry != expectQry {
			t.Errorf("expected: %s\tgot: %s\n", expectQry, qry)
		}
		if !reflect.DeepEqual(args, expectArgs) {
			t.Errorf("expeact-args: %v\tgot:%v\n", expectArgs, args)
		}

	})
}
