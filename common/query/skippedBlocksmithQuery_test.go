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

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

var (
	mockSkippedBlocksmith = &model.SkippedBlocksmith{
		BlocksmithPublicKey: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		POPChange:           0,
		BlockHeight:         0,
		BlocksmithIndex:     0,
	}
)

func TestSkippedBlocksmithQuery_SelectDataForSnapshot(t *testing.T) {
	qry := NewSkippedBlocksmithQuery()
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "SelectDataForSnapshot",
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith " +
				"WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0 ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbq := &SkippedBlocksmithQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := sbq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("SkippedBlocksmithQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkippedBlocksmithQuery_TrimDataBeforeSnapshot(t *testing.T) {
	qry := NewSkippedBlocksmithQuery()
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "TrimDataBeforeSnapshot",
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM skipped_blocksmith WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbq := &SkippedBlocksmithQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := sbq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("SkippedBlocksmithQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkippedBlocksmithQuery_InsertSkippedBlocksmiths(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		skippedBlocksmiths []*model.SkippedBlocksmith
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewSkippedBlocksmithQuery()),
			args: args{
				skippedBlocksmiths: []*model.SkippedBlocksmith{
					mockSkippedBlocksmith,
				},
			},
			wantStr:  "INSERT INTO skipped_blocksmith (blocksmith_public_key, pop_change, block_height, blocksmith_index) VALUES (?, ?, ?, ?)",
			wantArgs: NewSkippedBlocksmithQuery().ExtractModel(mockSkippedBlocksmith),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbq := &SkippedBlocksmithQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := sbq.InsertSkippedBlocksmiths(tt.args.skippedBlocksmiths)
			if gotStr != tt.wantStr {
				t.Errorf("InsertSkippedBlocksmiths() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertSkippedBlocksmiths() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
