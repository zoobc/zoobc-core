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
		NodePublicKey:      []byte{1},
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
		Latest:             true,
		Height:             0,
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
		wantQry1 := "INSERT INTO spine_public_key (node_public_key,registration_status,latest,height) VALUES(? , ?, ?, ?)"
		if fmt.Sprintf("%v", res[1][0]) != wantQry1 {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[1][0], wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetValidSpinePublicKeysByHeight(t *testing.T) {
	t.Run("GetValidSpinePublicKeysByHeight", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.GetValidSpinePublicKeysByHeight(100)
		wantQry := "SELECT node_public_key, registration_status, latest, height FROM spine_public_key " +
			"WHERE height <= 100 AND registration_status=0 AND latest=1"
		if res != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetSpinePublicKeyByNodePublicKey(t *testing.T) {
	t.Run("GetSpinePublicKeyByNodePublicKey", func(t *testing.T) {
		query, args := mockSpinePublicKeyQuery.GetSpinePublicKeyByNodePublicKey(mockSpinePublicKey.NodePublicKey)
		wantQry := "SELECT node_public_key, registration_status, latest, height FROM spine_public_key " +
			"WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1"
		wantArg := mockSpinePublicKey.NodePublicKey
		if query != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", query, wantQry)
		}
		b, _ := args[0].([]byte)
		if !bytes.Equal(b, wantArg) {
			t.Errorf("arg does not match:\nget: %v\nwant: %v", b, wantArg)
		}
	})
}
