package query

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

var (
	mockBlockQuery = NewBlockQuery(chaintype.GetChainType(0))
	mockBlock      = &model.Block{
		ID:                   1,
		Height:               0,
		BlockSeed:            []byte{1, 2, 3},
		BlockSignature:       []byte{1, 2, 3, 4, 5},
		BlocksmithAddress:    "BCZ",
		CumulativeDifficulty: "0",
		PayloadHash:          []byte{},
		PayloadLength:        1,
		PreviousBlockHash:    []byte{},
		SmithScale:           0,
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
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", tableName, want)
		}
	})
}

func TestBlockQuery_GetBlocks(t *testing.T) {
	t.Run("GetBlocks:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlocks(0, 10)
		wantQ := "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
			"payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height " +
			">= 0 LIMIT 10"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetLastBlock(t *testing.T) {
	t.Run("GetLastBlock:success", func(t *testing.T) {
		q := mockBlockQuery.GetLastBlock()
		wantQ := "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
			"payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version FROM main_block ORDER BY height " +
			"DESC LIMIT 1"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetGenesisBlock(t *testing.T) {
	t.Run("GetGenesisBlock:success", func(t *testing.T) {
		q := mockBlockQuery.GetGenesisBlock()
		wantQ := "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
			"payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height " +
			"= 0 LIMIT 1"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_InsertBlock(t *testing.T) {
	t.Run("InsertBlock:success", func(t *testing.T) {
		q, args := mockBlockQuery.InsertBlock(mockBlock)
		wantQ := "INSERT INTO main_block (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
			"smith_scale, payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version) VALUES(? , ?, ?, ?, " +
			"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		wantArg := mockBlockQuery.ExtractModel(mockBlock)

		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
		if !reflect.DeepEqual(args, wantArg) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", args, wantArg)
		}
	})
}

func TestBlockQuery_GetBlockByID(t *testing.T) {
	t.Run("GetBlockByID:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlockByID(1)
		wantQ := "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
			"payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version FROM main_block WHERE id = 1"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_GetBlockByHeight(t *testing.T) {
	t.Run("GetBlockByHeight:success", func(t *testing.T) {
		q := mockBlockQuery.GetBlockByHeight(0)
		wantQ := "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
			"payload_length, payload_hash, blocksmith_address, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height = 0"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestBlockQuery_ExtractModel(t *testing.T) {
	t.Run("BlockQuery-ExtractModel:success", func(t *testing.T) {
		res := mockBlockQuery.ExtractModel(mockBlock)
		want := []interface{}{
			mockBlock.ID,
			mockBlock.PreviousBlockHash,
			mockBlock.Height,
			mockBlock.Timestamp,
			mockBlock.BlockSeed,
			mockBlock.BlockSignature,
			mockBlock.CumulativeDifficulty,
			mockBlock.SmithScale,
			mockBlock.PayloadLength,
			mockBlock.PayloadHash,
			mockBlock.BlocksmithAddress,
			mockBlock.TotalAmount,
			mockBlock.TotalFee,
			mockBlock.TotalCoinBase,
			mockBlock.Version,
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
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(mockBlock.ID,
				mockBlock.PreviousBlockHash,
				mockBlock.Height,
				mockBlock.Timestamp,
				mockBlock.BlockSeed,
				mockBlock.BlockSignature,
				mockBlock.CumulativeDifficulty,
				mockBlock.SmithScale,
				mockBlock.PayloadLength,
				mockBlock.PayloadHash,
				mockBlock.BlocksmithAddress,
				mockBlock.TotalAmount,
				mockBlock.TotalFee,
				mockBlock.TotalCoinBase,
				mockBlock.Version))

		rows, _ := db.Query("foo")
		var tempBlock []*model.Block
		res := mockBlockQuery.BuildModel(tempBlock, rows)
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
		name     string
		fields   fields
		args     args
		wantStr  []string
		wantArgs uint32
	}{
		{
			name:     "wantSuccess",
			fields:   fields(*mockBlockQuery),
			args:     args{height: uint32(1)},
			wantStr:  []string{"DELETE FROM block WHERE height > 1"},
			wantArgs: uint32(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bq := &BlockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			gotStr, gotArgs := bq.Rollback(tt.args.height)
			if !reflect.DeepEqual(gotStr, tt.wantStr) {
				t.Errorf("Rollback() = %v, want %v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Rollback() = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
