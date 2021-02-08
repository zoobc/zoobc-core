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
)

var (
	mockMerkleTreeQuery = NewMerkleTreeQuery()
	mockRoot            = make([]byte, 32)
	mockBlockHeight     = uint32(0)
	mockTree            = make([]byte, 14*32)
)

func TestMerkleTreeQuery_SelectMerkleTree(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		lowerHeight uint32
		upperHeight uint32
		limit       uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "SelectMerkleTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				lowerHeight: 0,
				upperHeight: 10,
				limit:       20,
			},
			want: "SELECT id, block_height, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS (SELECT rmr_linked FROM " +
				"published_receipt AS pr WHERE mt.id = pr.rmr_linked) AND " +
				"block_height BETWEEN 0 AND 10 ORDER BY block_height ASC LIMIT 20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mrQ.SelectMerkleTree(tt.args.lowerHeight, tt.args.upperHeight, tt.args.limit); got != tt.want {
				t.Errorf("SelectMerkleTree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMerkleTreeQuery(t *testing.T) {
	tests := []struct {
		name string
		want MerkleTreeQueryInterface
	}{
		{
			name: "NewMerkleTreeQuery:success",
			want: mockMerkleTreeQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMerkleTreeQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMerkleTreeQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_InsertMerkleTree(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root        []byte
		blockHeight uint32
		tree        []byte
		timestamp   int64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name: "InsertMerkleTreeQuery:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				root:        mockRoot,
				tree:        mockTree,
				timestamp:   0,
				blockHeight: 0,
			},
			wantQStr: "INSERT INTO merkle_tree (id, block_height, tree, timestamp) VALUES(?,? ,? ,? )",
			wantArgs: []interface{}{
				mockRoot,
				uint32(0),
				mockTree,
				int64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := mrQ.InsertMerkleTree(tt.args.root, tt.args.tree, tt.args.timestamp, tt.args.blockHeight)
			if gotQStr != tt.wantQStr {
				t.Errorf("InsertMerkleTree() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertMerkleTree() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestMerkleTreeQuery_GetMerkleTreeByRoot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name: "GetMerkleTreeByRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				root: mockRoot,
			},
			wantQStr: "SELECT id, block_height, tree, timestamp FROM merkle_tree WHERE id = ?",
			wantArgs: []interface{}{mockRoot},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := mrQ.GetMerkleTreeByRoot(tt.args.root)
			if gotQStr != tt.wantQStr {
				t.Errorf("GetMerkleTreeByRoot() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetMerkleTreeByRoot() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestMerkleTreeQuery_GetLastMerkleRoot(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name     string
		fields   fields
		wantQStr string
	}{
		{
			name: "GetLastMerkleRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			wantQStr: "SELECT id, block_height, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotQStr := mrQ.GetLastMerkleRoot(); gotQStr != tt.wantQStr {
				t.Errorf("GetLastMerkleRoot() = %v, want %v", gotQStr, tt.wantQStr)
			}
		})
	}
}

func TestMerkleTreeQuery_ScanTree(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockBlockHeight,
		mockTree,
		int64(0),
	))
	mock.ExpectQuery("wrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))
	row := db.QueryRow("")
	wrongRow := db.QueryRow("wrongQuery")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ScanTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: row,
			},
			want:    mockTree,
			wantErr: false,
		},
		{
			name: "ScanTree:wrongColumn",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: wrongRow,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.ScanTree(tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanTree() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_ScanRoot(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockBlockHeight,
		mockTree,
		int64(0),
	))
	mock.ExpectQuery("wrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))
	row := db.QueryRow("")
	wrongRow := db.QueryRow("wrongQuery")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		row *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ScanRoot:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: row,
			},
			want:    mockRoot,
			wantErr: false,
		},
		{
			name: "ScanRoot:wrongColumn",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				row: wrongRow,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.ScanRoot(tt.args.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanRoot() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_BuildTree(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockMerkleTreeQuery.Fields).AddRow(
		mockRoot,
		mockBlockHeight,
		mockTree,
		int64(0),
	))

	rows, _ := db.Query("")
	defer rows.Close()
	mock.ExpectQuery("WrongQuery").WillReturnRows(sqlmock.NewRows([]string{"foo"}).AddRow(
		mockRoot,
	))

	rowsWrong, _ := db.Query("WrongQuery")
	defer rowsWrong.Close()
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		rows *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "BuildTree:success",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				rows: rows,
			},
			want: map[string][]byte{
				string(mockRoot): mockTree,
			},
			wantErr: false,
		},
		{
			name: "BuildTree:fail - wrong number field",
			fields: fields{
				Fields:    mockMerkleTreeQuery.Fields,
				TableName: mockMerkleTreeQuery.TableName,
			},
			args: args{
				rows: rowsWrong,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := mrQ.BuildTree(tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildTree() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTreeQuery_PruneData(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockHeight uint32
		limit       uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockMerkleTreeQuery),
			args:   args{blockHeight: 2000, limit: 500},
			want: "DELETE FROM merkle_tree WHERE block_height IN(" +
				"SELECT block_height FROM merkle_tree WHERE block_height < ? " +
				"ORDER BY block_height ASC LIMIT ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, args := mrQ.PruneData(tt.args.blockHeight, tt.args.limit)
			if got != tt.want {
				t.Errorf("PruneData() = \n%v, want \n%v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(args, []interface{}{tt.args.blockHeight, tt.args.limit}) {
				t.Errorf("PruneData() = \n%v, want \n%v", args, []interface{}{tt.args.blockHeight, tt.args.limit})
			}
		})
	}
}

func TestMerkleTreeQuery_Rollback(t *testing.T) {
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
			fields: fields(*mockMerkleTreeQuery),
			args:   args{height: 1},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM merkle_tree WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrQ := &MerkleTreeQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := mrQ.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("MerkleTreeQuery.Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}
