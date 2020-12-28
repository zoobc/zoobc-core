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
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockSpineBlockManifestServiceQueryExecutor struct {
		testName string
		query.Executor
	}
)

var (
	ssMockSpineBlockManifest = &model.SpineBlockManifest{
		ID:                      1,
		FullFileHash:            ssMockFullHash,
		ManifestReferenceHeight: 720,
		FileChunkHashes:         []byte{},
		ChainType:               0,
		SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
		ExpirationTimestamp:     1000,
	}
)

func (*mockSpineBlockManifestServiceQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) BeginTx() error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) RollbackTx() error {
	return nil
}
func (*mockSpineBlockManifestServiceQueryExecutor) CommitTx() error {
	return nil
}

func TestBlockSpineSnapshotService_CreateSpineBlockManifest(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		SpineBlockManifestQuery   query.SpineBlockManifestQueryInterface
		SpineBlockQuery           query.BlockQueryInterface
		MainBlockQuery            query.BlockQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
	type args struct {
		snapshotHash            []byte
		mainHeight, spineHeight uint32
		megablockTimestamp      int64
		sortedFileChunksHashes  [][]byte
		lastFileChunkHash       []byte
		ct                      chaintype.ChainType
		mbType                  model.SpineBlockManifestType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *model.SpineBlockManifest
	}{
		{
			name: "CreateSpineBlockManifest:success",
			fields: fields{
				QueryExecutor: &mockSpineBlockManifestServiceQueryExecutor{
					testName: "CreateSpineBlockManifest:success",
				},
				SpineBlockManifestQuery: query.NewSpineBlockManifestQuery(),
				SpineBlockQuery:         query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                  log.New(),
			},
			args: args{
				snapshotHash:           make([]byte, 32),
				mainHeight:             ssMockMainBlock.Height,
				megablockTimestamp:     ssMockMainBlock.Timestamp,
				sortedFileChunksHashes: make([][]byte, 0),
				lastFileChunkHash:      make([]byte, 32),
				ct:                     &chaintype.MainChain{},
				mbType:                 model.SpineBlockManifestType_Snapshot,
			},
			wantErr: false,
			want: &model.SpineBlockManifest{
				ID:                      int64(2447379392738367286),
				FullFileHash:            make([]byte, 32),
				ManifestReferenceHeight: ssMockMainBlock.Height,
				ExpirationTimestamp:     int64(1604307615),
				FileChunkHashes:         make([]byte, 0),
				SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
				ChainType:               0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestService{
				QueryExecutor:           tt.fields.QueryExecutor,
				SpineBlockManifestQuery: tt.fields.SpineBlockManifestQuery,
				SpineBlockQuery:         tt.fields.SpineBlockQuery,
				Logger:                  tt.fields.Logger,
			}
			got, err := mbl.CreateSpineBlockManifest(tt.args.snapshotHash, tt.args.mainHeight, tt.args.megablockTimestamp,
				tt.args.sortedFileChunksHashes, tt.args.ct, tt.args.mbType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotService.CreateSpineBlockManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.CreateSpineBlockManifest() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshotService_GetSpineBlockManifestBytes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		SpineBlockManifestQuery query.SpineBlockManifestQueryInterface
		SpineBlockQuery         query.BlockQueryInterface
		Logger                  *log.Logger
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name:   "GetSpineBlockManifestBytes:success",
			fields: fields{},
			args: args{
				spineBlockManifest: ssMockSpineBlockManifest,
			},
			want: []byte{1, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 208, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SpineBlockManifestService{
				QueryExecutor:           tt.fields.QueryExecutor,
				SpineBlockManifestQuery: tt.fields.SpineBlockManifestQuery,
				SpineBlockQuery:         tt.fields.SpineBlockQuery,
				Logger:                  tt.fields.Logger,
			}
			got := ss.GetSpineBlockManifestBytes(tt.args.spineBlockManifest)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.GetSpineBlockManifestBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
