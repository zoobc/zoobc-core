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
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockParticipationScoreQuery = NewParticipationScoreQuery()
	mockParticipationScore      = &model.ParticipationScore{
		NodeID: 1,
		Score:  100000000,
		Latest: true,
		Height: 0,
	}
)

func TestParticipationScoreQuery_InsertParticipationScore(t *testing.T) {
	t.Run("InsertParticipationScore:success", func(t *testing.T) {

		q, args := mockParticipationScoreQuery.InsertParticipationScore(mockParticipationScore)
		wantQ := "INSERT INTO participation_score (node_id,score,latest,height) VALUES(? , ?, ?, ?)"
		wantArg := []interface{}{
			mockParticipationScore.NodeID, mockParticipationScore.Score,
			mockParticipationScore.Latest, mockParticipationScore.Height,
		}
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestParticipationScoreQuery_GetParticipationScoreByNodeID(t *testing.T) {
	t.Run("GetParticipationScoreByNodeID", func(t *testing.T) {
		res, arg := mockParticipationScoreQuery.GetParticipationScoreByNodeID(1)
		want := "SELECT node_id, score, latest, height FROM participation_score WHERE node_id = ? AND latest=1"
		wantArg := []interface{}{int64(1)}
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
		if !reflect.DeepEqual(arg, wantArg) {
			t.Errorf("argument not match:\nget: %v\nwant: %v", arg, wantArg)
		}
	})
}

func TestParticipationScoreQuery_GetParticipationScoreByNodePublicKey(t *testing.T) {
	t.Run("GetParticipationScoreByNodePublicKey", func(t *testing.T) {
		res, _ := mockParticipationScoreQuery.GetParticipationScoreByNodePublicKey([]byte{})
		want := "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A " +
			"INNER JOIN node_registry as B ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.registration_status=0 AND A.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestParticipationScoreQuery_UpdateParticipationScore(t *testing.T) {
	t.Run("UpdateParticipationScore", func(t *testing.T) {
		res := mockParticipationScoreQuery.UpdateParticipationScore(int64(1111), int64(10), uint32(1))
		want0 := "INSERT INTO participation_score (node_id, score, height, latest) VALUES(1111, 10, 1, 1) " +
			"ON CONFLICT(node_id, height) DO UPDATE SET (score) = 10"
		want1 := "UPDATE participation_score SET latest = false WHERE node_id = 1111 AND height != 1 AND latest = true"
		if res[0][0] != want0 {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[0][0], want0)
		}
		if res[1][0] != want1 {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[1][0], want1)
		}
	})
}

func TestParticipationScoreQuery_SelectDataForSnapshot(t *testing.T) {
	t.Run("SelectDataForSnapshot", func(t *testing.T) {
		res := mockParticipationScoreQuery.SelectDataForSnapshot(0, 1)
		want := "SELECT node_id,score,latest,height FROM participation_score WHERE (node_id, height) IN (SELECT t2.node_id, " +
			"MAX(t2.height) FROM participation_score as t2 WHERE t2.height >= 0 AND t2.height <= 1 AND t2.height != 0 GROUP BY t2.node_id ) ORDER by height"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestParticipationScoreQuery_TrimDataBeforeSnapshot(t *testing.T) {
	t.Run("TrimDataBeforeSnapshot", func(t *testing.T) {
		res := mockParticipationScoreQuery.TrimDataBeforeSnapshot(0, 10)
		want := "DELETE FROM participation_score WHERE height >= 0 AND height <= 10 AND height != 0"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestParticipationScoreQuery_InsertParticipationScores(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		scores []*model.ParticipationScore
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewParticipationScoreQuery()),
			args: args{
				scores: []*model.ParticipationScore{
					mockParticipationScore,
				},
			},
			wantStr:  "INSERT INTO participation_score (node_id, score, latest, height) VALUES (?, ?, ?, ?)",
			wantArgs: NewParticipationScoreQuery().ExtractModel(mockParticipationScore),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &ParticipationScoreQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ps.InsertParticipationScores(tt.args.scores)
			if gotStr != tt.wantStr {
				t.Errorf("InsertParticipationScores() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertParticipationScores() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestParticipationScoreQuery_GetParticipationScoresByBlockHeightRange(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromBlockHeight uint32
		toBlockHeight   uint32
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
			fields: fields(*NewParticipationScoreQuery()),
			args: args{
				fromBlockHeight: 20,
				toBlockHeight:   123,
			},
			wantStr:  "SELECT node_id, score, latest, height FROM participation_score WHERE height BETWEEN ? AND ? ORDER BY height ASC",
			wantArgs: []interface{}{uint32(20), uint32(123)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &ParticipationScoreQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := ps.GetParticipationScoresByBlockHeightRange(tt.args.fromBlockHeight, tt.args.toBlockHeight)
			if gotStr != tt.wantStr {
				t.Errorf("ParticipationScoreQuery.GetParticipationScoresByBlockHeightRange() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ParticipationScoreQuery.GetParticipationScoresByBlockHeightRange() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
