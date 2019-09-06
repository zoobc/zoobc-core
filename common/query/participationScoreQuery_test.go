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

func TestParticipationScoreQuery_UpdateParticipationScore(t *testing.T) {
	t.Run("UpdateParticipationScore:success", func(t *testing.T) {

		q, args := mockParticipationScoreQuery.UpdateParticipationScore(mockParticipationScore)
		wantQ0 := "UPDATE participation_score SET latest = 0 WHERE node_id = 1"
		wantQ1 := "INSERT INTO participation_score (node_id,score,latest,height) VALUES(? , ?, ?, ?)"
		wantArg := []interface{}{
			mockParticipationScore.NodeID, mockParticipationScore.Score,
			mockParticipationScore.Latest, mockParticipationScore.Height,
		}
		if q[0] != wantQ0 {
			t.Errorf("update query returned wrong: get: %s\nwant: %s", q, wantQ0)
		}
		if q[1] != wantQ1 {
			t.Errorf("insert query returned wrong: get: %s\nwant: %s", q, wantQ1)
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
