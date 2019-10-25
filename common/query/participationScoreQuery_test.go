package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

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

func TestParticipationScoreQuery_GetParticipationScoreByAccountAddress(t *testing.T) {
	testAccountAddress := "TESTACCOUNTADDRESS"
	t.Run("GetParticipationScoreByAccountAddress", func(t *testing.T) {
		res := mockParticipationScoreQuery.GetParticipationScoreByAccountAddress(testAccountAddress)
		want := "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
			"ON A.node_id = B.id WHERE B.account_address='" + testAccountAddress + "' AND B.latest=1 AND B.queued=0 AND A.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestParticipationScoreQuery_GetParticipationScoreByNodePublicKey(t *testing.T) {
	t.Run("GetParticipationScoreByNodePublicKey", func(t *testing.T) {
		res, _ := mockParticipationScoreQuery.GetParticipationScoreByNodePublicKey([]byte{})
		want := "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A " +
			"INNER JOIN node_registry as B ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.queued=0 AND A.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}

func TestParticipationScoreQuery_AddParticipationScore(t *testing.T) {
	t.Run("AddParticipationScore:success", func(t *testing.T) {
		res := mockParticipationScoreQuery.AddParticipationScore(12, 1*constant.OneZBC, 1)
		want := [][]interface{}{
			{
				"INSERT INTO participation_score (node_id, score, height, latest) SELECT node_id, score + " +
					"100000000, 1, latest FROM participation_score WHERE node_id = 12 AND latest = 1 ON " +
					"CONFLICT(node_id, height) DO UPDATE SET (score) = (SELECT score + 100000000 FROM participation_score " +
					"WHERE node_id = 12 AND latest = 1)",
			},
			{
				"UPDATE participation_score SET latest = false WHERE node_id = 12 AND height != 1 AND latest = true",
			},
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}
