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
	mockPublishedReceiptQuery = &PublishedReceiptQuery{
		Fields: []string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"intermediate_hashes",
			"block_height",
			"receipt_index",
			"published_index",
		},
		TableName: "published_receipt",
	}

	mockPublishedReceipt = &model.PublishedReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey:      make([]byte, 32),
			RecipientPublicKey:   make([]byte, 32),
			DatumType:            1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHeight: 0,
			ReferenceBlockHash:   make([]byte, 32),
			RMRLinked:            make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		IntermediateHashes: nil,
		BlockHeight:        0,
		ReceiptIndex:       0,
		PublishedIndex:     0,
	}
)

func TestNewPublishedReceiptQuery(t *testing.T) {
	tests := []struct {
		name string
		want *PublishedReceiptQuery
	}{
		{
			name: "NewPublishedReceipt:Success",
			want: mockPublishedReceiptQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublishedReceiptQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPublishedReceiptQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedReceiptQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		publishedReceipt *model.PublishedReceipt
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name: "ExtractModel:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			args: args{
				publishedReceipt: mockPublishedReceipt,
			},
			want: []interface{}{
				&mockPublishedReceipt.Receipt.SenderPublicKey,
				&mockPublishedReceipt.Receipt.RecipientPublicKey,
				&mockPublishedReceipt.Receipt.DatumType,
				&mockPublishedReceipt.Receipt.DatumHash,
				&mockPublishedReceipt.Receipt.ReferenceBlockHeight,
				&mockPublishedReceipt.Receipt.ReferenceBlockHash,
				&mockPublishedReceipt.Receipt.RMRLinked,
				&mockPublishedReceipt.Receipt.RecipientSignature,
				&mockPublishedReceipt.IntermediateHashes,
				&mockPublishedReceipt.BlockHeight,
				&mockPublishedReceipt.ReceiptIndex,
				&mockPublishedReceipt.PublishedIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pu := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := pu.ExtractModel(tt.args.publishedReceipt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedReceiptQuery_GetPublishedReceiptByLinkedRMR(t *testing.T) {
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
			name: "GetPublishedReceiptByLinkedRMR:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			args: args{
				root: make([]byte, 32),
			},
			wantStr: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
				"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, " +
				"receipt_index, published_index FROM published_receipt WHERE rmr_linked = ?",
			wantArgs: []interface{}{
				make([]byte, 32),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := prq.GetPublishedReceiptByLinkedRMR(tt.args.root)
			if gotStr != tt.wantStr {
				t.Errorf("GetPublishedReceiptByLinkedRMR() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetPublishedReceiptByLinkedRMR() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPublishedReceiptQuery_InsertPublishedReceipt(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		publishedReceipt *model.PublishedReceipt
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertPublishedReceipt:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			args: args{publishedReceipt: mockPublishedReceipt},
			wantStr: "INSERT INTO published_receipt (sender_public_key, recipient_public_key, datum_type, datum_hash, " +
				"reference_block_height, reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, " +
				"block_height, receipt_index, published_index) VALUES(? , ? , ? , ? , ? , ? , ? , ? , ? , ? , ? , ? )",
			wantArgs: []interface{}{
				&mockPublishedReceipt.Receipt.SenderPublicKey,
				&mockPublishedReceipt.Receipt.RecipientPublicKey,
				&mockPublishedReceipt.Receipt.DatumType,
				&mockPublishedReceipt.Receipt.DatumHash,
				&mockPublishedReceipt.Receipt.ReferenceBlockHeight,
				&mockPublishedReceipt.Receipt.ReferenceBlockHash,
				&mockPublishedReceipt.Receipt.RMRLinked,
				&mockPublishedReceipt.Receipt.RecipientSignature,
				&mockPublishedReceipt.IntermediateHashes,
				&mockPublishedReceipt.BlockHeight,
				&mockPublishedReceipt.ReceiptIndex,
				&mockPublishedReceipt.PublishedIndex,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := prq.InsertPublishedReceipt(tt.args.publishedReceipt)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPublishedReceipt() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPublishedReceipt() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPublishedReceiptQuery_Scan(t *testing.T) {
	var mockTempReceipt = model.PublishedReceipt{
		Receipt: &model.Receipt{},
	}

	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(mockPublishedReceiptQuery.Fields).AddRow(
		&mockPublishedReceipt.Receipt.SenderPublicKey,
		&mockPublishedReceipt.Receipt.RecipientPublicKey,
		&mockPublishedReceipt.Receipt.DatumType,
		&mockPublishedReceipt.Receipt.DatumHash,
		&mockPublishedReceipt.Receipt.ReferenceBlockHeight,
		&mockPublishedReceipt.Receipt.ReferenceBlockHash,
		&mockPublishedReceipt.Receipt.RMRLinked,
		&mockPublishedReceipt.Receipt.RecipientSignature,
		&mockPublishedReceipt.IntermediateHashes,
		&mockPublishedReceipt.BlockHeight,
		&mockPublishedReceipt.ReceiptIndex,
		&mockPublishedReceipt.PublishedIndex,
	))
	row := db.QueryRow("")
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipt *model.PublishedReceipt
		row     *sql.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Scan:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			args: args{
				receipt: &mockTempReceipt,
				row:     row,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pu := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if err := pu.Scan(tt.args.receipt, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublishedReceiptQuery_getTableName(t *testing.T) {
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
			name: "getTableName:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			want: mockPublishedReceiptQuery.TableName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := prq.getTableName(); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedReceiptQuery_SelectDataForSnapshot(t *testing.T) {
	prQry := NewPublishedReceiptQuery()
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
			args: args{
				toHeight:   1,
				fromHeight: 0,
			},
			fields: fields{
				TableName: prQry.TableName,
				Fields:    prQry.Fields,
			},
			want: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
				"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
				"published_index FROM published_receipt WHERE block_height >= 0 AND block_height <= 1 AND block_height != 0 ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := prq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PublishedReceiptQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedReceiptQuery_TrimDataBeforeSnapshot(t *testing.T) {
	prQry := NewPublishedReceiptQuery()
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
			args: args{
				toHeight:   10,
				fromHeight: 0,
			},
			fields: fields{
				TableName: prQry.TableName,
				Fields:    prQry.Fields,
			},
			want: "DELETE FROM published_receipt WHERE block_height >= 0 AND block_height <= 10 AND block_height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := prq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("PublishedReceiptQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedReceiptQuery_InsertPublishedReceipts(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		receipts []*model.PublishedReceipt
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
			fields: fields(*NewPublishedReceiptQuery()),
			args: args{
				receipts: []*model.PublishedReceipt{
					mockPublishedReceipt,
				},
			},
			wantStr: "INSERT INTO published_receipt (sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
				"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, published_index) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			wantArgs: NewPublishedReceiptQuery().ExtractModel(mockPublishedReceipt),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := prq.InsertPublishedReceipts(tt.args.receipts)
			if gotStr != tt.wantStr {
				t.Errorf("InsertPublishedReceipts() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertPublishedReceipts() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestPublishedReceiptQuery_GetPublishedReceiptByBlockHeightRange(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		qry := NewPublishedReceiptQuery()
		qStr, args := qry.GetPublishedReceiptByBlockHeightRange(0, 100)
		result := "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
			"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
			"published_index FROM published_receipt WHERE block_height BETWEEN ? AND ? ORDER BY block_height, published_index ASC"
		if qStr != result {
			t.Fatalf("expect: %s\ngot: %s", result, qStr)
		}
		if args[0] != uint32(0) && args[1] != uint32(100) {
			t.Fatalf("expect arguments: %s\ngot: %s", []interface{}{
				uint32(0), uint32(100),
			}, args)
		}
	})
}

func TestPublishedReceiptQuery_GetUnlinkedPublishedReceiptByBlockHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetUnlinkedPublishedReceiptByBlockHeight:success",
			fields: fields{
				Fields:    mockPublishedReceiptQuery.Fields,
				TableName: mockPublishedReceiptQuery.TableName,
			},
			args: args{
				blockHeight: 1,
			},
			wantStr: "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash," +
				" rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
				"published_index FROM published_receipt WHERE block_height = ? AND rmr_linked IS NULL ORDER BY published_index ASC",
			wantArgs: []interface{}{
				uint32(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prq := &PublishedReceiptQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := prq.GetUnlinkedPublishedReceiptByBlockHeight(tt.args.blockHeight)
			if gotStr != tt.wantStr {
				t.Errorf("GetUnlinkedPublishedReceiptByBlockHeight() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("GetUnlinkedPublishedReceiptByBlockHeight() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
