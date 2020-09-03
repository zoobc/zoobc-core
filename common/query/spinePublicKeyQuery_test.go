package query

import (
	"bytes"
	"fmt"
	"reflect"
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
		wantQry1 := "INSERT INTO spine_public_key (node_public_key,node_id,public_key_action,main_block_height,latest,height) VALUES(" +
			"? , ?, ?, ?, ?, ?)"
		if fmt.Sprintf("%v", res[1][0]) != wantQry1 {
			t.Errorf("string not match:\nget: %s\nwant: %s", res[1][0], wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetValidSpinePublicKeysByHeightInterval(t *testing.T) {
	t.Run("GetValidSpinePublicKeysByHeightInterval", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.GetValidSpinePublicKeysByHeightInterval(0, 100)
		wantQry := "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key " +
			"WHERE height >= 0 AND height <= 100 AND public_key_action=0 AND latest=1 ORDER BY height"
		if res != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_GetSpinePublicKeysByBlockHeight(t *testing.T) {
	t.Run("GetValidSpinePublicKeysByHeightInterval", func(t *testing.T) {
		res := mockSpinePublicKeyQuery.GetSpinePublicKeysByBlockHeight(1)
		wantQry := "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height " +
			"FROM spine_public_key WHERE height = 1"
		if res != wantQry {
			t.Errorf("string not match:\nget: %s\nwant: %s", res, wantQry)
		}
	})
}

func TestSpinePublicKeyQuery_Rollback(t *testing.T) {
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
			name:   "WantSuccess",
			fields: fields(*mockSpinePublicKeyQuery),
			args:   args{height: 1},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM spine_public_key WHERE height > ?",
					uint32(1),
				},
				{
					`
			UPDATE spine_public_key SET latest = ?
			WHERE latest = ? AND (node_public_key, height) IN (
				SELECT t2.node_public_key, MAX(t2.height)
				FROM spine_public_key as t2
				GROUP BY t2.node_public_key
			)`,
					1,
					0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spk := &SpinePublicKeyQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := spk.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = \n%v, want \n%v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}
