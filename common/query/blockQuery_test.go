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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockBlockQuery = NewBlockQuery(chaintype.GetChainType(0))
	bQNodePubKey   = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}

	mockBlock = &model.Block{
		ID:                   1,
		Height:               0,
		BlockSeed:            []byte{1, 2, 3},
		BlockSignature:       []byte{1, 2, 3, 4, 5},
		BlocksmithPublicKey:  bQNodePubKey,
		CumulativeDifficulty: "0",
		PayloadHash:          []byte{},
		PayloadLength:        1,
		BlockHash:            []byte{},
		PreviousBlockHash:    []byte{},
		Timestamp:            1000,
		TotalAmount:          1000,
		TotalCoinBase:        0,
		TotalFee:             1,
		Transactions:         nil,
		Version:              1,
	}
)

func TestNewBlockQuery(t *testing.T) {
	type args struct {
		chaintype chaintype.ChainType
	}
	tests := []struct {
		name string
		args args
		want *BlockQuery
	}{
		{
			name: "NewBlockQuery:success",
			args: args{
				chaintype: &chaintype.MainChain{},
			},
			want: mockBlockQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockQuery(tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockQuery_getTableName(t *testing.T) {
	t.Run("BlockQuery-getTableName:mainchain", func(t *testing.T) {
		tableName := mockBlockQuery.getTableName()
		want := "main_block"
		if tableName != want {
			t.Errorf("arguments returned wrong: get: \n%vwant: \n%v", tableName, want)
		}
	})
}

func TestBlockQuery_GetBlocks(t *testing.T) {
	t.Run("GetBlocks:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlocks(0, 10)
		wantQ := "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
			"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
			"FROM main_block WHERE height >= 0 ORDER BY height ASC LIMIT 10"
		if q != wantQ {
			t.Errorf("query returned wrong: get: \n%s\nwant: \n%s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetLastBlock(t *testing.T) {
	t.Run("GetLastBlock:success", func(t *testing.T) {
		q := mockBlockQuery.GetLastBlock()
		wantQ := "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
			"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block"
		if q != wantQ {
			t.Errorf("query returned wrong: get: \n%swant: \n%s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetGenesisBlock(t *testing.T) {
	t.Run("GetGenesisBlock:success", func(t *testing.T) {
		q := mockBlockQuery.GetGenesisBlock()
		wantQ := "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
			"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE height = 0"
		if q != wantQ {
			t.Errorf("query returned wrong: get: \n%swant: \n%s", q, wantQ)
		}
	})
}

func TestBlockQuery_InsertBlock(t *testing.T) {
	t.Run("InsertBlock:success", func(t *testing.T) {
		q, args := mockBlockQuery.InsertBlock(mockBlock)
		wantQ := "INSERT INTO main_block (height, id, block_hash, previous_block_hash, timestamp, block_seed, " +
			"block_signature, cumulative_difficulty, payload_length, payload_hash, " +
			"blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
			"version, merkle_root, merkle_tree, reference_block_height) VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := mockBlockQuery.ExtractModel(mockBlock)

		if q != wantQ {
			t.Errorf("query returned wrong: get: \n%swant: \n%s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: \n%v\nwant: \n%v", args, wantArg)
		}
	})
}

func TestBlockQuery_GetBlockByID(t *testing.T) {
	t.Run("GetBlockByID:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlockByID(1)
		wantQ := "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, " +
			"total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
			"FROM main_block WHERE id = 1"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetBlockByHeight(t *testing.T) {
	t.Run("GetBlockByHeight:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlockByHeight(0)
		wantQ := "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, " +
			"total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
			"FROM main_block WHERE height = 0"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_ExtractModel(t *testing.T) {
	t.Run("BlockQuery-ExtractModel:success", func(t *testing.T) {
		res := mockBlockQuery.ExtractModel(mockBlock)
		want := []interface{}{
			mockBlock.Height,
			mockBlock.ID,
			mockBlock.BlockHash,
			mockBlock.PreviousBlockHash,
			mockBlock.Timestamp,
			mockBlock.BlockSeed,
			mockBlock.BlockSignature,
			mockBlock.CumulativeDifficulty,
			mockBlock.PayloadLength,
			mockBlock.PayloadHash,
			mockBlock.BlocksmithPublicKey,
			mockBlock.TotalAmount,
			mockBlock.TotalFee,
			mockBlock.TotalCoinBase,
			mockBlock.Version,
			mockBlock.MerkleRoot,
			mockBlock.MerkleTree,
			mockBlock.ReferenceBlockHeight,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}

func TestBlockQuery_BuildModel(t *testing.T) {
	t.Run("BlockQuery-BuildModel:success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(
			sqlmock.NewRows(mockBlockQuery.Fields).
				AddRow(
					mockBlock.Height,
					mockBlock.ID,
					mockBlock.BlockHash,
					mockBlock.PreviousBlockHash,
					mockBlock.Timestamp,
					mockBlock.BlockSeed,
					mockBlock.BlockSignature,
					mockBlock.CumulativeDifficulty,
					mockBlock.PayloadLength,
					mockBlock.PayloadHash,
					mockBlock.BlocksmithPublicKey,
					mockBlock.TotalAmount,
					mockBlock.TotalFee,
					mockBlock.TotalCoinBase,
					mockBlock.Version,
					mockBlock.MerkleRoot,
					mockBlock.MerkleTree,
					mockBlock.ReferenceBlockHeight))

		rows, _ := db.Query("foo")
		var tempBlock []*model.Block
		res, _ := mockBlockQuery.BuildModel(tempBlock, rows)
		if !reflect.DeepEqual(res[0], mockBlock) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, mockBlock)
		}
	})
}

func TestBlockQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
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
			fields: fields(*mockBlockQuery),
			args:   args{height: uint32(1)},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM main_block WHERE height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			multiQueries := bq.Rollback(tt.args.height)
			if !reflect.DeepEqual(multiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", multiQueries, tt.wantMultiQueries)
				return
			}
		})
	}
}

func TestBlockQuery_GetBlockFromHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		startHeight uint32
		limit       uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "GetBlockFromHeight:success",
			fields: fields(*mockBlockQuery),
			args: args{
				limit:       1,
				startHeight: 1,
			},
			want: "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
				"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, " +
				"total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE height >= 1 ORDER BY height LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := bq.GetBlockFromHeight(tt.args.startHeight, tt.args.limit); got != tt.want {
				t.Errorf("BlockQuery.GetBlockFromHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockQuery_GetBlockFromTimestamp(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		startTimestamp int64
		limit          uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "GetBlockFromTimestamp:success",
			fields: fields(*mockBlockQuery),
			args: args{
				limit:          1,
				startTimestamp: 15875392,
			},
			want: "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
				"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, " +
				"total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
				"FROM main_block WHERE timestamp >= 15875392 ORDER BY timestamp LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := bq.GetBlockFromTimestamp(tt.args.startTimestamp, tt.args.limit); got != tt.want {
				t.Errorf("BlockQuery.GetBlockFromTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockQuery_SelectDataForSnapshot(t *testing.T) {
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
			name:   "SelectDataForSnapshot:success",
			fields: fields(*mockBlockQuery),
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "SELECT height,id,block_hash,previous_block_hash,timestamp,block_seed,block_signature," +
				"cumulative_difficulty,payload_length,payload_hash,blocksmith_public_key,total_amount," +
				"total_fee,total_coinbase,version,merkle_root,merkle_tree,reference_block_height " +
				"FROM main_block WHERE height >= 0 AND height <= 10 AND height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := bq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockQuery_TrimDataBeforeSnapshot(t *testing.T) {
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
			name:   "TrimDataBeforeSnapshot:success",
			fields: fields(*mockBlockQuery),
			args: args{
				fromHeight: 1,
				toHeight:   10,
			},
			want: "DELETE FROM main_block WHERE height >= 1 AND height <= 10 AND height != 0",
		},
		{
			name:   "TrimDataBeforeSnapshot:success-{startFromGenesis}",
			fields: fields(*mockBlockQuery),
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM main_block WHERE height >= 0 AND height <= 10 AND height != 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := bq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("BlockQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockQuery_InsertBlocks(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		blocks []*model.Block
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
			fields: fields(*NewBlockQuery(&chaintype.MainChain{})),
			args: args{
				blocks: []*model.Block{
					mockBlock,
				},
			},
			wantStr: "INSERT INTO main_block " +
				"(height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
				"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, " +
				"merkle_root, merkle_tree, reference_block_height) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			wantArgs: NewBlockQuery(&chaintype.MainChain{}).ExtractModel(mockBlock),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := bq.InsertBlocks(tt.args.blocks)
			if gotStr != tt.wantStr {
				t.Errorf("InsertBlocks() gotStr = \n%v want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertBlocks() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestBlockQuery_GetBlockSmithPublicKeyByHeightRange(t *testing.T) {
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
			name:   "GetBlockSmithPublicKeyByHeightRange",
			fields: fields(*mockBlockQuery),
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "SELECT height, blocksmith_public_key FROM main_block WHERE height >= 0 AND height <= 10 ORDER BY height DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := bq.GetBlockSmithPublicKeyByHeightRange(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("GetBlockSmithPublicKeyByHeightRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
