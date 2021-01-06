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
package service

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/crypto"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockNodeRegistrationQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockNodeRegistrationQueryExecutorSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE height >= 1 AND height <= 2 " +
		"AND registration_status != 1 AND latest=1 ORDER BY height":
		mockNodeRegistrationRows := mockSpine.NewRows(query.NewNodeRegistrationQuery().Fields)
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockNodeRegistrationRows)
	default:
		return nil, fmt.Errorf("unmocked query for mockNodeRegistrationQueryExecutorSuccess: %s", qStr)
	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

func TestBlockSpinePublicKeyService_BuildSpinePublicKeysFromNodeRegistry(t *testing.T) {
	type fields struct {
		Signature             crypto.SignatureInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
	}
	type args struct {
		mainFromHeight uint32
		mainToHeight   uint32
		spineHeight    uint32
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantSpinePublicKeys []*model.SpinePublicKey
		wantErr             bool
	}{
		{
			name: "BuildSpinePublicKeysFromNodeRegistry:success",
			fields: fields{
				QueryExecutor:         &mockNodeRegistrationQueryExecutorSuccess{},
				SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				Signature:             nil,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args: args{
				mainFromHeight: 1,
				mainToHeight:   2,
				spineHeight:    1,
			},
			wantSpinePublicKeys: []*model.SpinePublicKey{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bsf := &BlockSpinePublicKeyService{
				Signature:             tt.fields.Signature,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
			}
			gotSpinePublicKeys, err := bsf.BuildSpinePublicKeysFromNodeRegistry(tt.args.mainFromHeight, tt.args.mainToHeight, tt.args.spineHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSpinePublicKeys, tt.wantSpinePublicKeys) {
				t.Errorf("BlockSpinePublicKeyService.BuildSpinePublicKeysFromNodeRegistry() = %v, want %v", gotSpinePublicKeys, tt.wantSpinePublicKeys)
			}
		})
	}
}
