package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

var (
	mockParticipationScoreQuery = NewParticipationScoreQuery()
)

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
			"ON A.node_id = B.id WHERE B.account_address='" + testAccountAddress + "' AND B.latest=1 AND B.registration_status=0 AND A.latest=1"
		if res != want {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
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

func TestParticipationScoreQuery_AddParticipationScore(t *testing.T) {
	t.Run("AddParticipationScore:success", func(t *testing.T) {
		causedFields = map[string]interface{}{
			"node_id": int64(12),
			"height":  uint32(1),
		}
		res := mockParticipationScoreQuery.AddParticipationScore(1*constant.OneZBC, causedFields)
		var want [][]interface{}
		want = append(want, []interface{}{
			"INSERT INTO participation_score AS ps (node_id, score, latest, height) " +
				"SELECT ?, 0, 1, ? WHERE NOT EXISTS (SELECT ps1.node_id FROM participation_score AS ps1 " +
				"WHERE ps1.node_id = ?)",
			causedFields["node_id"], causedFields["height"], causedFields["node_id"],
		}, []interface{}{
			"INSERT INTO participation_score AS ps (node_id, score, latest, height) " +
				"SELECT ps1.node_id, ps1.score + 100000000, 1, ? FROM participation_score AS ps1 " +
				"WHERE ps1.node_id = ? AND ps1.latest = 1 ON CONFLICT(ps.node_id, ps.height) " +
				"DO UPDATE SET (score, height, latest) = (SELECT ps2.score + 100000000, ps2.height, 1 " +
				"FROM participation_score AS ps2 WHERE ps2.node_id = ? AND ps2.latest = 1)",
			causedFields["height"], causedFields["node_id"], causedFields["node_id"],
		}, []interface{}{
			"UPDATE participation_score SET latest = false WHERE node_id = ? AND height != ? AND latest = true",
			causedFields["node_id"], causedFields["height"],
		},
		)
		if !reflect.DeepEqual(res, want) {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, want)
		}
	})
}
