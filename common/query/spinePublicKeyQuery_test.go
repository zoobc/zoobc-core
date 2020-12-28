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
