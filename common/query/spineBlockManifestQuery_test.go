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

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestSpineBlockManifestQuery_InsertSpineBlockManifest(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}

	mb1 := &model.SpineBlockManifest{
		FullFileHash:            make([]byte, 64), // sha3-512
		ManifestReferenceHeight: 720,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "InsertSpineBlockManifest:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				spineBlockManifest: mb1,
			},
			want: "INSERT OR REPLACE INTO spine_block_manifest (id,full_file_hash,file_chunk_hashes,manifest_reference_height," +
				"manifest_spine_block_height,chain_type,manifest_type," +
				"expiration_timestamp) VALUES(? , ?, ?, ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := mbl.InsertSpineBlockManifest(tt.args.spineBlockManifest); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpineBlockManifestQuery.InsertSpineBlockManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpineBlockManifestQuery_GetLastSpineBlockManifest(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		ct     chaintype.ChainType
		mbType model.SpineBlockManifestType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetLastSpineBlockManifest:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				ct:     &chaintype.MainChain{},
				mbType: model.SpineBlockManifestType_Snapshot,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, manifest_reference_height, manifest_spine_block_height, " +
				"chain_type, manifest_type, expiration_timestamp FROM spine_block_manifest WHERE chain_type = 0 AND " +
				"manifest_type = 0 ORDER BY manifest_reference_height DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetLastSpineBlockManifest(tt.args.ct, tt.args.mbType); got != tt.want {
				t.Errorf("SpineBlockManifestQuery.GetLastSpineBlockManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpineBlockManifestQuery_GetSpineBlockManifestsInTimeInterval(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromTimestamp int64
		toTimestamp   int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetSpineBlockManifestTimeInterval:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				fromTimestamp: 10,
				toTimestamp:   20,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, manifest_reference_height, manifest_spine_block_height, " +
				"chain_type, manifest_type, expiration_timestamp FROM spine_block_manifest WHERE expiration_timestamp > 10 " +
				"AND expiration_timestamp <= 20 ORDER BY manifest_type, chain_type, manifest_reference_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetSpineBlockManifestTimeInterval(tt.args.fromTimestamp, tt.args.toTimestamp); got != tt.want {
				t.Errorf("SpineBlockManifestQuery.GetSpineBlockManifestTimeInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
