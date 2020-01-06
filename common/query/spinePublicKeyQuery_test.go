package query

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockSpinePublicKeyQuery = NewSpinePublicKeyQuery()
	mockSpinePublicKey      = &model.SpinePublicKey{
		NodePublicKey:   []byte{1},
		PublicKeyAction: model.SpinePublicKeyAction_AddKey,
		Latest:          true,
		Height:          0,
	}
)

func TestSpinePublicKeyQuery_InsertSpinePublicKey(t *testing.T) {
	t.Run("InsertSpinePublicKey", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.InsertSpinePublicKey(mockSpinePublicKey)
		wantQry := "UPDATE spine_public_key SET latest = 0 WHERE node_public_key = ?"
		if fmt.Sprintf("%v", res[0][0]) != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[0][0], wantQry)
		}
		wantArg := mockSpinePublicKey.NodePublicKey
		b, _ := res[0][1].([]byte)
		if !bytes.Equal(b, wantArg) {
			t.Errorf("arg does not match:\nget: %v\nwant: %v", res[0][1], wantArg)
		}
		wantQry1 := "INSERT INTO spine_public_key (node_public_key,block_id,public_key_action,latest,height) VALUES(? , ?, ?, ?, ?)"
		if fmt.Sprintf("%v", res[1][0]) != wantQry1 {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[1][0], wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetValidSpinePublicKeysByHeightInterval(t *testing.T) {
	t.Run("GetValidSpinePublicKeysByHeightInterval", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(0, 100)
		wantQry := "SELECT node_public_key, block_id, public_key_action, latest, height FROM spine_public_key " +
			"WHERE height >= 0 AND height <= 100 AND public_key_action=0 AND latest=1 ORDER BY height"
		if res != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetSpinePublicKeysByBlockID(t *testing.T) {
	t.Run("GetValidSpinePublicKeysByHeightInterval", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.GetSpinePublicKeysByBlockID(1)
		wantQry := "SELECT node_public_key, block_id, public_key_action, latest, height FROM spine_public_key WHERE block_id = 1"
		if res != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, wantQry)
		}
	})
}
