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
	mockReceiptQuery = NewBatchReceiptQuery()
	mockBatchReceipt = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey:      []byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
			RecipientPublicKey:   []byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
			DatumType:            uint32(1),
			DatumHash:            []byte{1, 2, 3, 4, 5, 6},
			ReferenceBlockHeight: uint32(1),
			ReferenceBlockHash:   []byte{1, 2, 3, 4, 5, 6},
			RMRLinked:            []byte{1, 2, 3, 4, 5, 6},
			RecipientSignature:   []byte{1, 2, 3, 4, 5, 6},
		},
		RMR:      []byte{1, 2, 3, 4, 5, 6},
		RMRIndex: uint32(4),
	}
)

func TestReceiptQuery_InsertReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipts []*model.BatchReceipt
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
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipts: []*model.BatchReceipt{mockBatchReceipt},
			},
			wantQStr: "INSERT INTO node_receipt " +
				"(sender_public_key, recipient_public_key, " +
				"datum_type, datum_hash, reference_block_height, " +
				"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index) " +
				"VALUES(?,? ,? ,? ,? ,? ,? ,? ,? ,? )",
			wantArgs: []interface{}{
				&mockBatchReceipt.Receipt.SenderPublicKey,
				&mockBatchReceipt.Receipt.RecipientPublicKey,
				&mockBatchReceipt.Receipt.DatumType,
				&mockBatchReceipt.Receipt.DatumHash,
				&mockBatchReceipt.Receipt.ReferenceBlockHeight,
				&mockBatchReceipt.Receipt.ReferenceBlockHash,
				&mockBatchReceipt.Receipt.RMRLinked,
				&mockBatchReceipt.Receipt.RecipientSignature,
				&mockBatchReceipt.RMR,
				&mockBatchReceipt.RMRIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := rq.InsertReceipts(tt.args.batchReceipts)
			if gotQStr != tt.wantQStr {
				t.Errorf("BatchReceiptQuery.InsertReceipts() gotQStr = \n%v, want \n%v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.InsertReceipts() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestReceiptQuery_InsertReceipt(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipt *model.BatchReceipt
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipt: mockBatchReceipt,
			},
			wantStr: "INSERT INTO node_receipt " +
				"(sender_public_key, recipient_public_key, datum_type, datum_hash, " +
				"reference_block_height, reference_block_hash, rmr_linked, " +
				"recipient_signature, rmr, rmr_index) VALUES(? , ? , ? , ? , ? , ? , ? , ? , ? , ? )",
			wantArgs: []interface{}{
				&mockBatchReceipt.Receipt.SenderPublicKey,
				&mockBatchReceipt.Receipt.RecipientPublicKey,
				&mockBatchReceipt.Receipt.DatumType,
				&mockBatchReceipt.Receipt.DatumHash,
				&mockBatchReceipt.Receipt.ReferenceBlockHeight,
				&mockBatchReceipt.Receipt.ReferenceBlockHash,
				&mockBatchReceipt.Receipt.RMRLinked,
				&mockBatchReceipt.Receipt.RecipientSignature,
				&mockBatchReceipt.RMR,
				&mockBatchReceipt.RMRIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := rq.InsertReceipt(tt.args.batchReceipt)
			if gotStr != tt.wantStr {
				t.Errorf("BatchReceiptQuery.InsertReceipt() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BatchReceiptQuery.InsertReceipt() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestReceiptQuery_GetReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		paginate model.Pagination
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{paginate: model.Pagination{
				OrderBy: model.OrderBy_ASC,
			}},
			want: "SELECT sender_public_key, recipient_public_key, datum_type, " +
				"datum_hash, reference_block_height, reference_block_hash, rmr_linked, " +
				"recipient_signature, rmr, rmr_index FROM node_receipt ORDER BY reference_block_height " +
				"ASC LIMIT 256 OFFSET 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := rq.GetReceipts(tt.args.paginate); got != tt.want {
				t.Errorf("BatchReceiptQuery.GetReceipts() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorNodeReceiptScan struct {
		Executor
	}
)

func (*mockQueryExecutorNodeReceiptScan) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockReceiptQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(4),
	))
	return db.QueryRow("")
}

func TestReceiptQuery_Scan(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipt *model.BatchReceipt
		row          *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipt: mockBatchReceipt,
				row:          (&mockQueryExecutorNodeReceiptScan{}).ExecuteSelectRow("", ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := r.Scan(tt.args.batchReceipt, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("BatchReceiptQuery.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorNodeReceiptBuildModel struct {
		Executor
	}
)

func (*mockQueryExecutorNodeReceiptBuildModel) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockReceiptQuery.Fields).AddRow(
		[]byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
		[]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"),
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(1),
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{1, 2, 3, 4, 5, 6},
		uint32(4),
	))
	return db.Query("")
}

func TestReceiptQuery_BuildModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		batchReceipts []*model.BatchReceipt
		rows          *sql.Rows
	}
	rows, err := (&mockQueryExecutorNodeReceiptBuildModel{}).ExecuteSelect("", false, "")
	if err != nil {
		t.Errorf("Rows Failed: %s", err.Error())
		return
	}
	defer rows.Close()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.BatchReceipt
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				batchReceipts: []*model.BatchReceipt{},
				rows:          rows,
			},
			want: []*model.BatchReceipt{mockBatchReceipt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := re.BuildModel(tt.args.batchReceipts, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeReceiptQuery_PruneData(t *testing.T) {
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
			name:   "WantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				blockHeight: 2000,
				limit:       500,
			},
			want: "DELETE FROM node_receipt WHERE reference_block_height IN(" +
				"SELECT reference_block_height FROM node_receipt " +
				"WHERE reference_block_height <? ORDER BY reference_block_height ASC LIMIT ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, args := rq.PruneData(tt.args.blockHeight, tt.args.limit)
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

func TestBatchReceiptQuery_GetReceiptsByRoot(t *testing.T) {
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
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				root: make([]byte, 32),
			},
			wantStr: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash," +
				" rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE rc.rmr = ? ORDER BY datum_hash, " +
				"recipient_public_key, reference_block_height",
			wantArgs: []interface{}{
				make([]byte, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := rq.GetReceiptsByRoot(tt.args.root)
			if gotStr != tt.wantStr {
				t.Errorf("GetReceiptsByRoot() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetReceiptsByRoot() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBatchReceiptQuery_GetReceiptsByRootAndDatumHash(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		root      []byte
		datumHash []byte
		datumType uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockReceiptQuery),
			args: args{
				root:      make([]byte, 32),
				datumType: 1,
				datumHash: make([]byte, 32),
			},
			wantStr: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash," +
				" rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE rc.rmr = ? AND rc.datum_hash = ? AND rc." +
				"datum_type = ? ORDER BY recipient_public_key, reference_block_height",
			wantArgs: []interface{}{
				make([]byte, 32),
				make([]byte, 32),
				uint32(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rq := &BatchReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := rq.GetReceiptsByRootAndDatumHash(tt.args.root, tt.args.datumHash, tt.args.datumType)
			if gotStr != tt.wantStr {
				t.Errorf("GetReceiptsByRootAndDatumHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetReceiptsByRootAndDatumHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
