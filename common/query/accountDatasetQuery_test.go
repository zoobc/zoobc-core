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
	mockDatasetQuery = NewAccountDatasetsQuery()
	mockDataset      = &model.AccountDataset{
		SetterAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 174, 8, 69, 186, 181, 103, 207, 111, 16, 204, 183, 18, 162, 64, 217, 82, 41, 208, 14,
			252, 193, 14, 191, 200, 158, 211, 172, 37, 0, 58, 107, 64},
		Property: "Admin",
		Value:    "You're Welcome",
		IsActive: true,
		Latest:   true,
		Height:   5,
	}
)

func TestNewAccountDatasetsQuery(t *testing.T) {
	tests := []struct {
		name string
		want *AccountDatasetQuery
	}{
		{
			name: "success",
			want: mockDatasetQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountDatasetsQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountDatasetsQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetsQuery_GetLastAccountDataset(t *testing.T) {
	type args struct {
		SetterAccountAddress    []byte
		RecipientAccountAddress []byte
		property                string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "wantSuccess",
			args: args{
				SetterAccountAddress:    mockDataset.GetSetterAccountAddress(),
				RecipientAccountAddress: mockDataset.GetRecipientAccountAddress(),
				property:                mockDataset.GetProperty(),
			},
			wantQuery: "SELECT setter_account_address, recipient_account_address, property, value, is_active, latest, height " +
				"FROM account_dataset WHERE setter_account_address = ? AND recipient_account_address = ? AND property = ? AND latest = ?",
			wantArgs: []interface{}{
				mockDataset.GetSetterAccountAddress(),
				mockDataset.GetRecipientAccountAddress(),
				mockDataset.GetProperty(),
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := mockDatasetQuery.GetLatestAccountDataset(tt.args.SetterAccountAddress, tt.args.RecipientAccountAddress, tt.args.property)
			if gotQuery != tt.wantQuery {
				t.Errorf("AccountDatasetQuery.GetLastDataset() gotQuery = \n%v want \n%v", gotQuery, tt.wantQuery)
				return
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("AccountDatasetQuery.GetLastDataset() gotArgs = \n%v want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountDatasetQuery_InsertAccountDataset(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		dataset *model.AccountDataset
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantStr [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockDatasetQuery),
			args:   args{dataset: mockDataset},
			wantStr: [][]interface{}{
				{
					"UPDATE account_dataset SET latest = ? WHERE setter_account_address = ? AND recipient_account_address = ? " +
						"AND property = ? AND height < ? AND latest = ?",
					false,
					mockDataset.GetSetterAccountAddress(),
					mockDataset.GetRecipientAccountAddress(),
					mockDataset.GetProperty(),
					mockDataset.GetHeight(),
					true,
				},
				{
					"INSERT INTO account_dataset (setter_account_address, recipient_account_address, property, value, is_active, latest, height) " +
						"VALUES(?, ?, ?, ?, ?, ?, ?) " +
						"ON CONFLICT(setter_account_address, recipient_account_address, property, height) " +
						"DO UPDATE SET value = ?, is_active = ?, latest = ?",
					mockDataset.GetSetterAccountAddress(),
					mockDataset.GetRecipientAccountAddress(),
					mockDataset.GetProperty(),
					mockDataset.GetValue(),
					true,
					true,
					mockDataset.GetHeight(),
					mockDataset.GetValue(),
					mockDataset.GetIsActive(),
					mockDataset.GetLatest(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := adq.InsertAccountDataset(tt.args.dataset); !reflect.DeepEqual(gotStr, tt.wantStr) {
				t.Errorf("InsertAccountDataset() = \n%v, want \n%v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestAccountDatasetsQuery_GetAccountDatasetEscrowApproval(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountAddress []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockDatasetQuery),
			args: args{accountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
				239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}},
			wantQStr: "SELECT setter_account_address, recipient_account_address, property, value, is_active, latest, height FROM account_dataset " +
				"WHERE setter_account_address = ? AND recipient_account_address = ? AND property = ? AND latest = ?",
			wantArgs: []interface{}{
				[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
					239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72,
					239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				"AccountDatasetEscrowApproval",
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := adq.GetAccountDatasetEscrowApproval(tt.args.accountAddress)
			if gotQStr != tt.wantQStr {
				t.Errorf("GetAccountDatasetEscrowApproval() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetAccountDatasetEscrowApproval() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountDatasetsQuery_ExtractModel(t *testing.T) {
	type args struct {
		dataset *model.AccountDataset
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "success",
			args: args{
				dataset: mockDataset,
			},
			want: []interface{}{
				mockDataset.GetSetterAccountAddress(),
				mockDataset.GetRecipientAccountAddress(),
				mockDataset.GetProperty(),
				mockDataset.GetValue(),
				mockDataset.GetIsActive(),
				mockDataset.GetLatest(),
				mockDataset.GetHeight(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mockDatasetQuery.ExtractModel(tt.args.dataset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockExecutorAccountDatasetBuildModel struct {
		Executor
	}
)

func (*mockExecutorAccountDatasetBuildModel) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(mockDatasetQuery.Fields)
	mockRows.AddRow(
		mockDataset.GetSetterAccountAddress(),
		mockDataset.GetRecipientAccountAddress(),
		mockDataset.GetProperty(),
		mockDataset.GetValue(),
		mockDataset.GetIsActive(),
		mockDataset.GetLatest(),
		mockDataset.GetHeight(),
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}
func TestAccountDatasetQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		datasets []*model.AccountDataset
		rows     *sql.Rows
	}
	rows, _ := (&mockExecutorAccountDatasetBuildModel{}).ExecuteSelect("", false, nil)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.AccountDataset
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockDatasetQuery),
			args: args{
				datasets: []*model.AccountDataset{},
				rows:     rows,
			},
			want: []*model.AccountDataset{
				mockDataset,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := adq.BuildModel(tt.args.datasets, tt.args.rows)
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

type (
	mockRowAccountDatasetQueryScan struct {
		Executor
	}
)

func (*mockRowAccountDatasetQueryScan) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockDatasetQuery.Fields).AddRow(
			mockDataset.GetSetterAccountAddress(),
			mockDataset.GetRecipientAccountAddress(),
			mockDataset.GetProperty(),
			mockDataset.GetValue(),
			mockDataset.GetIsActive(),
			mockDataset.GetLatest(),
			mockDataset.GetHeight(),
		),
	)
	return db.QueryRow(""), nil
}

func TestAccountDatasetsQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		dataset *model.AccountDataset
		row     *sql.Row
	}

	row, _ := (&mockRowAccountDatasetQueryScan{}).ExecuteSelectRow("", false, nil)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockDatasetQuery),
			args: args{
				dataset: mockDataset,
				row:     row,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := a.Scan(tt.args.dataset, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("AccountDatasetQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountDatasetQuery_Rollback(t *testing.T) {
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
			name:   "wantSuccess",
			fields: fields(*mockDatasetQuery),
			args:   args{height: 5},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM account_dataset WHERE height > ?",
					uint32(5),
				},
				{
					`
				UPDATE account_dataset SET latest = ?
				WHERE latest = ? AND (setter_account_address, recipient_account_address, property, height) IN (
					SELECT setter_account_address, recipient_account_address, property, MAX(height)
					FROM account_dataset
					GROUP BY setter_account_address, recipient_account_address, property
				)`,
					1, 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := adq.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = \n%v, want \n%v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestAccountDatasetsQuery_SelectDataForSnapshot(t *testing.T) {
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
			name:   "SelectDataForSnapshot",
			fields: fields(*mockDatasetQuery),
			args: args{
				fromHeight: 0,
				toHeight:   1,
			},
			want: `
			SELECT setter_account_address, recipient_account_address, property, value, is_active, latest, height FROM account_dataset
			WHERE (setter_account_address, recipient_account_address, property, height) IN (
				SELECT setter_account_address, recipient_account_address, property, MAX(height) FROM account_dataset
				WHERE height >= 0 AND height <= 1 AND height != 0
				GROUP BY setter_account_address, recipient_account_address, property
			) ORDER BY height`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := adq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("AccountDatasetQuery.SelectDataForSnapshot() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetsQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			name:   "TrimDataBeforeSnapshot",
			fields: fields(*mockDatasetQuery),
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM account_dataset WHERE height >= 0 AND height <= 10 AND height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := adq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("AccountDatasetQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetQuery_InsertAccountDatasets(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		datasets []*model.AccountDataset
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
			fields: fields(*NewAccountDatasetsQuery()),
			args: args{
				datasets: []*model.AccountDataset{
					mockDataset,
				},
			},
			wantStr: "INSERT INTO account_dataset (setter_account_address, recipient_account_address, property, value, is_active, latest, height) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?)",
			wantArgs: NewAccountDatasetsQuery().ExtractModel(mockDataset),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adq := &AccountDatasetQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := adq.InsertAccountDatasets(tt.args.datasets)
			if gotStr != tt.wantStr {
				t.Errorf("InsertAccountDatasets() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertAccountDatasets() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
