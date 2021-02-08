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
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	multisigAccountAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
		239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	multisigAccountAddress2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79,
		137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	multisigAccountAddress3 = []byte{0, 0, 0, 0, 131, 252, 92, 188, 219, 93, 20, 95, 223, 162, 209, 53, 10, 27, 14, 67, 202, 149, 108,
		229, 12, 146, 136, 6, 143, 228, 45, 178, 0, 80, 142, 52}
)

func TestMultiSignatureParticipantQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participant *model.MultiSignatureParticipant
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participant: &model.MultiSignatureParticipant{
					AccountAddressIndex:   0,
					MultiSignatureAddress: multisigAccountAddress1,
					AccountAddress:        multisigAccountAddress2,
					BlockHeight:           100,
				},
			},
			want: []interface{}{
				multisigAccountAddress1,
				multisigAccountAddress2,
				uint32(0),
				false,
				uint32(100),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msq.ExtractModel(tt.args.participant); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_BuildModel(t *testing.T) {
	dbMock, sqlMock, _ := sqlmock.New()
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(NewMultiSignatureParticipantQuery().Fields).
		AddRow(
			multisigAccountAddress1,
			multisigAccountAddress2,
			0,
			true,
			100,
		))
	rows, _ := dbMock.Query("")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantParticipants []*model.MultiSignatureParticipant
		wantErr          bool
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				rows: rows,
			},
			wantParticipants: []*model.MultiSignatureParticipant{
				{
					MultiSignatureAddress: multisigAccountAddress1,
					AccountAddress:        multisigAccountAddress2,
					AccountAddressIndex:   0,
					Latest:                true,
					BlockHeight:           100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotParticipants, err := msq.BuildModel(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParticipants, tt.wantParticipants) {
				t.Errorf("BuildModel() gotParticipants = %v, want %v", gotParticipants, tt.wantParticipants)
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_Scan(t *testing.T) {
	dbMock, sqlMock, _ := sqlmock.New()
	sqlMock.ExpectQuery("").WillReturnRows(sqlMock.NewRows(NewMultiSignatureParticipantQuery().Fields).
		AddRow(
			multisigAccountAddress1,
			multisigAccountAddress2,
			0,
			true,
			100,
		))
	row := dbMock.QueryRow("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participant *model.MultiSignatureParticipant
		row         *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participant: &model.MultiSignatureParticipant{},
				row:         row,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := msq.Scan(tt.args.participant, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ok := reflect.DeepEqual(tt.args.participant, new(model.MultiSignatureParticipant)); ok {
				t.Errorf("Scan() did not update reference")
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_InsertMultisignatureParticipants(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		participants []*model.MultiSignatureParticipant
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantQueries [][]interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultiSignatureParticipantQuery()),
			args: args{
				participants: []*model.MultiSignatureParticipant{
					{
						MultiSignatureAddress: multisigAccountAddress1,
						AccountAddress:        multisigAccountAddress2,
						AccountAddressIndex:   0,
						BlockHeight:           100,
						Latest:                true,
					},
					{
						MultiSignatureAddress: multisigAccountAddress2,
						AccountAddress:        multisigAccountAddress1,
						AccountAddressIndex:   1,
						BlockHeight:           100,
						Latest:                true,
					},
				},
			},
			wantQueries: [][]interface{}{
				{
					"INSERT OR REPLACE INTO multisignature_participant (multisig_address, account_address, account_address_index, latest, " +
						"block_height) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
					multisigAccountAddress1,
					multisigAccountAddress2,
					uint32(0),
					true,
					uint32(100),
					multisigAccountAddress2,
					multisigAccountAddress1,
					uint32(1),
					true,
					uint32(100),
				},
				{
					"UPDATE multisignature_participant SET latest = ? WHERE multisig_address = ? AND block_height != ? AND latest = ?",
					false, multisigAccountAddress1, uint32(100), true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotQueries := msq.InsertMultisignatureParticipants(tt.args.participants); !reflect.DeepEqual(gotQueries, tt.wantQueries) {
				t.Errorf("InsertMultisignatureParticipants() = \n%v, want \n%v", gotQueries, tt.wantQueries)
			}
		})
	}
}

func TestMultiSignatureParticipantQuery_GetMultiSignatureParticipantsByMultisigAddressAndHeightRange(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigAddress []byte
		fromHeight      uint32
		toHeight        uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetMultiSignatureParticipantsByMultisigAddressAndHeightRange-wantSuccess",
			args: args{
				fromHeight:      1,
				toHeight:        10,
				multisigAddress: multisigAccountAddress1,
			},
			fields: fields(*NewMultiSignatureParticipantQuery()),
			wantStr: "SELECT multisig_address,account_address,account_address_index,latest,block_height " +
				"FROM multisignature_participant WHERE multisig_address = ? AND block_height >= ? AND block_height <= ? " +
				"ORDER BY account_address_index",
			wantArgs: []interface{}{
				multisigAccountAddress1,
				uint32(1),
				uint32(10),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msq := &MultiSignatureParticipantQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := msq.GetMultiSignatureParticipantsByMultisigAddressAndHeightRange(tt.args.multisigAddress, tt.args.fromHeight, tt.args.toHeight)
			if gotStr != tt.wantStr {
				t.Errorf("GetMultiSignatureParticipantsByMultisigAddressAndHeightRange() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetMultiSignatureParticipantsByMultisigAddressAndHeightRange() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
