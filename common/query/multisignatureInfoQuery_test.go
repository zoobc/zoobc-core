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

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockMultisigInfoQueryInstance = NewMultisignatureInfoQuery()
)

// mocks fixtures for MultisignatureInfoQuery_BuildModel
func getBuildModelErrorMockRows() *sql.Rows {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

func getBuildModelSuccessMockRows(withParticipant bool) *sql.Rows {
	db, mock, _ := sqlmock.New()
	if withParticipant {
		mockRow := sqlmock.NewRows(append(mockMultisigInfoQueryInstance.Fields, "multisig_address"))
		mockRow.AddRow(
			multisigAccountAddress1,
			uint32(1),
			int64(10),
			uint32(12),
			true,
			multisigAccountAddress2,
		)
		mock.ExpectQuery("").WillReturnRows(mockRow)
		rows, _ := db.Query("")
		return rows
	}
	mockRow := sqlmock.NewRows(mockMultisigInfoQueryInstance.Fields)
	mockRow.AddRow(
		multisigAccountAddress1,
		uint32(1),
		int64(10),
		uint32(12),
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	rows, _ := db.Query("")
	return rows
}

func TestMultisignatureInfoQuery_BuildModelWithParticipant(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		mss  []*model.MultiSignatureInfo
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.MultiSignatureInfo
		wantErr bool
	}{
		{
			name: "BuildModel-RowsError",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelErrorMockRows(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelSuccessMockRows(true),
			},
			want: []*model.MultiSignatureInfo{
				{
					MultisigAddress:   multisigAccountAddress1,
					MinimumSignatures: 1,
					Nonce:             10,
					BlockHeight:       12,
					Latest:            true,
					Addresses: [][]byte{
						multisigAccountAddress2,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := msi.BuildModelWithParticipant(tt.args.mss, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		mss  []*model.MultiSignatureInfo
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.MultiSignatureInfo
		wantErr bool
	}{
		{
			name: "BuildModel-RowsError",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelErrorMockRows(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "BuildModel",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				mss:  []*model.MultiSignatureInfo{},
				rows: getBuildModelSuccessMockRows(false),
			},
			want: []*model.MultiSignatureInfo{
				{
					MultisigAddress:   multisigAccountAddress1,
					MinimumSignatures: 1,
					Nonce:             10,
					BlockHeight:       12,
					Latest:            true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := msi.BuildModel(tt.args.mss, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// Extract mocks
	mockExtractMultisignatureInfoMultisig = &model.MultiSignatureInfo{
		MinimumSignatures: 0,
		Nonce:             0,
		Addresses: [][]byte{
			multisigAccountAddress2,
			multisigAccountAddress3,
		},
		MultisigAddress: nil,
		BlockHeight:     0,
		Latest:          true,
	}
	// Extract mocks
)

func TestMultisignatureInfoQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel-Success",
			fields: fields{
				Fields:    nil,
				TableName: "",
			},
			args: args{
				multisigInfo: mockExtractMultisignatureInfoMultisig,
			},
			want: []interface{}{
				&mockExtractMultisignatureInfoMultisig.MultisigAddress,
				&mockExtractMultisignatureInfoMultisig.MinimumSignatures,
				&mockExtractMultisignatureInfoMultisig.Nonce,
				&mockExtractMultisignatureInfoMultisig.BlockHeight,
				&mockExtractMultisignatureInfoMultisig.Latest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mu.ExtractModel(tt.args.multisigInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_GetMultisignatureInfoByAddressWithParticipants(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigAddress      []byte
		currentHeight, limit uint32
	}

	var (
		multisigAddr = []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	)

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetMultisignatureInfoByAddressWithParticipants-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigAddress: multisigAddr,
				currentHeight:   0,
				limit:           constant.MinRollbackBlocks,
			},
			wantStr: "SELECT t1.multisig_address, t1.minimum_signatures, t1.nonce, t1.block_height, t1.latest, t2.account_address " +
				"FROM multisignature_info t1 LEFT JOIN multisignature_participant t2 ON t1.multisig_address = t2.multisig_address " +
				"WHERE t1.multisig_address = ? AND t1.block_height >= ? AND t1.latest = true AND t2.latest = true " +
				"ORDER BY t2.account_address_index DESC",
			wantArgs: []interface{}{multisigAddr, uint32(0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := msi.GetMultisignatureInfoByAddressWithParticipants(
				tt.args.multisigAddress,
				tt.args.currentHeight,
				tt.args.limit,
			)
			if gotStr != tt.wantStr {
				t.Errorf("GetMultisignatureInfoByAddressWithParticipants() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetMultisignatureInfoByAddressWithParticipants() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

var (
	// InsertMultisignatureInfo mocks
	mockInsertMultisignatureInfoMultisig = &model.MultiSignatureInfo{
		MinimumSignatures: 0,
		Nonce:             0,
		Addresses: [][]byte{
			multisigAccountAddress2,
			multisigAccountAddress3,
			multisigAccountAddress3,
		},
		MultisigAddress: multisigAccountAddress1,
		BlockHeight:     0,
		Latest:          true,
	}
	// InsertMultisignatureInfo mocks
)

func TestMultisignatureInfoQuery_InsertMultisignatureInfo(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name: "InsertMultisigInfo-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: mockInsertMultisignatureInfoMultisig,
			},
			want: [][]interface{}{
				append([]interface{}{
					"INSERT OR REPLACE INTO multisignature_info (multisig_address, minimum_signatures, nonce, block_height, latest) " +
						"VALUES(? , ? , ? , ? , ? )",
				}, mockMultisigInfoQueryInstance.ExtractModel(
					mockInsertMultisignatureInfoMultisig)...),
				{
					"UPDATE multisignature_info SET latest = false WHERE multisig_address = ? AND " +
						"block_height != 0 AND latest = true", mockInsertMultisignatureInfoMultisig.MultisigAddress,
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got := msi.InsertMultisignatureInfo(tt.args.multisigInfo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertMultisignatureInfo() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_Rollback(t *testing.T) {
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
			name: "Rollback-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				height: 10,
			},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM multisignature_info WHERE block_height > ?",
					uint32(10),
				},
				{
					"UPDATE multisignature_info SET latest = ? WHERE latest = ? AND (multisig_address, " +
						"block_height) IN (SELECT t2.multisig_address, MAX(t2.block_height) FROM multisignature_info as t2 GROUP BY t2." +
						"multisig_address)",
					1, 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := msi.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

// mocks fixtures for MultisignatureInfoQuery_Scan
func getNumberScanFailMockRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows([]string{"randomField"})
	mockRow.AddRow(
		"randomMock",
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

func getNumberScanSuccessMockRow() *sql.Row {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(mockMultisigInfoQueryInstance.Fields)
	mockRow.AddRow(
		multisigAccountAddress1,
		uint32(123),
		int64(10),
		uint32(12),
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow("")
}

// mocks fixtures for MultisignatureInfoQuery_Scan

func TestMultisignatureInfoQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
		row          *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "MultisignatureInfoQuery_Scan-Fail-WrongNumberOfColumn",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: &model.MultiSignatureInfo{},
				row:          getNumberScanFailMockRow(),
			},
			wantErr: true,
		},
		{
			name: "MultisignatureInfoQuery_Scan-Success",
			fields: fields{
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				multisigInfo: &model.MultiSignatureInfo{},
				row:          getNumberScanSuccessMockRow(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := mu.Scan(tt.args.multisigInfo, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultisignatureInfoQuery_getTableName(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getTableName",
			fields: fields{
				Fields:    nil,
				TableName: "X",
			},
			want: "X",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// NewMultisignatureInfoQuery mocks
	mockNewMultisignatureInfoQueryResult = &MultisignatureInfoQuery{
		Fields: []string{
			"multisig_address",
			"minimum_signatures",
			"nonce",
			"block_height",
			"latest",
		},
		TableName: "multisignature_info",
	}
	// NewMultisignatureInfoQuery mocks
)

func TestNewMultisignatureInfoQuery(t *testing.T) {
	tests := []struct {
		name string
		want *MultisignatureInfoQuery
	}{
		{
			name: "NewMultisignatureInfoQuery",
			want: mockNewMultisignatureInfoQueryResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMultisignatureInfoQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMultisignatureInfoQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_SelectDataForSnapshot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
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
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "SELECT multisig_address, minimum_signatures, nonce, block_height, latest FROM multisignature_info " +
				"WHERE (multisig_address, block_height) IN (SELECT t2.multisig_address, MAX(t2.block_height) " +
				"FROM multisignature_info t2 WHERE t2.block_height >= 1 AND t2.block_height <= 10 AND t2.block_height != 0 " +
				"GROUP BY t2.multisig_address) ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("MultisignatureInfoQuery.SelectDataForSnapshot() = \n%v, want \n%v", got, tt.want)
			}

		})
	}
}

func TestMultisignatureInfoQuery_TrimDataBeforeSnapshot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
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
				Fields:    mockMultisigInfoQueryInstance.Fields,
				TableName: mockMultisigInfoQueryInstance.TableName,
			},
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "DELETE FROM multisignature_info WHERE block_height >= 1 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("MultisignatureInfoQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisignatureInfoQuery_InsertMultiSignatureInfos(t *testing.T) {
	musigQ := NewMultisignatureInfoQuery()
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		multiSignatureInfos []*model.MultiSignatureInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewMultisignatureInfoQuery()),
			args: args{
				multiSignatureInfos: []*model.MultiSignatureInfo{
					mockInsertMultisignatureInfoMultisig,
				},
			},
			want: [][]interface{}{
				append([]interface{}{
					"INSERT INTO multisignature_info (multisig_address, minimum_signatures, nonce, block_height, latest) VALUES (?, ?, ?, ?, ?)",
				},
					musigQ.ExtractModel(mockInsertMultisignatureInfoMultisig)...,
				),
				{
					"INSERT INTO multisignature_participant (multisig_address, account_address, account_address_index, latest, block_height) " +
						"VALUES(?, ?, ?, ?, ?),(?, ?, ?, ?, ?),(?, ?, ?, ?, ?)",
					multisigAccountAddress1, multisigAccountAddress2, uint32(0), true, uint32(0), multisigAccountAddress1,
					multisigAccountAddress3, uint32(1), true, uint32(0), multisigAccountAddress1, multisigAccountAddress3, uint32(2), true,
					uint32(0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := msi.InsertMultiSignatureInfos(tt.args.multiSignatureInfos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertMultiSignatureInfos() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
