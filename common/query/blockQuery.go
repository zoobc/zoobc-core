package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlockQueryInterface interface {
		Rollback(height uint32) (multiQueries [][]interface{})
		GetBlocks(height, size uint32) string
		GetLastBlock() string
		GetGenesisBlock() string
		GetBlockByID(int64) string
		GetBlockByHeight(uint32) string
		GetBlockFromHeight(uint32, uint32) string
		InsertBlock(block *model.Block) (str string, args []interface{})
		ExtractModel(block *model.Block) []interface{}
		BuildModel(blocks []*model.Block, rows *sql.Rows) []*model.Block
	}

	BlockQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewBlockQuery returns BlockQuery instance
func NewBlockQuery(chaintype chaintype.ChainType) *BlockQuery {
	return &BlockQuery{
		Fields: []string{
			"id",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"smith_scale",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		},
		TableName: "block",
		ChainType: chaintype,
	}
}

func (bq *BlockQuery) getTableName() string {
	return bq.ChainType.GetTablePrefix() + "_" + bq.TableName
}

// GetBlocks returns query string to get multiple blocks
func (bq *BlockQuery) GetBlocks(height, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d ORDER BY height ASC LIMIT %d",
		strings.Join(bq.Fields, ", "), bq.getTableName(), height, size)
}

func (bq *BlockQuery) GetLastBlock() string {
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY height DESC LIMIT 1", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) GetGenesisBlock() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height = 0 LIMIT 1", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) InsertBlock(block *model.Block) (str string, args []interface{}) {
	var value = fmt.Sprintf("? %s", strings.Repeat(", ?", len(bq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		bq.getTableName(), strings.Join(bq.Fields, ", "), value)
	return query, bq.ExtractModel(block)
}

// GetBlockByID returns query string to get block by ID
func (bq *BlockQuery) GetBlockByID(id int64) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), id)
}

// GetBlockByHeight returns query string to get block by height
func (bq *BlockQuery) GetBlockByHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), height)
}

// GetBlockFromHeight returns query string to get blocks from a certain height
func (bq *BlockQuery) GetBlockFromHeight(startHeight, limit uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE HEIGHT >= %d ORDER BY HEIGHT LIMIT %d",
		strings.Join(bq.Fields, ", "), bq.getTableName(), startHeight, limit)
}

// ExtractModel extract the model struct fields to the order of BlockQuery.Fields
func (*BlockQuery) ExtractModel(block *model.Block) []interface{} {
	return []interface{}{
		block.ID,
		block.PreviousBlockHash,
		block.Height,
		block.Timestamp,
		block.BlockSeed,
		block.BlockSignature,
		block.CumulativeDifficulty,
		block.SmithScale,
		block.PayloadLength,
		block.PayloadHash,
		block.BlocksmithPublicKey,
		block.TotalAmount,
		block.TotalFee,
		block.TotalCoinBase,
		block.Version,
	}
}

func (*BlockQuery) BuildModel(blocks []*model.Block, rows *sql.Rows) []*model.Block {
	for rows.Next() {
		var block model.Block
		_ = rows.Scan(
			&block.ID,
			&block.PreviousBlockHash,
			&block.Height,
			&block.Timestamp,
			&block.BlockSeed,
			&block.BlockSignature,
			&block.CumulativeDifficulty,
			&block.SmithScale,
			&block.PayloadLength,
			&block.PayloadHash,
			&block.BlocksmithPublicKey,
			&block.TotalAmount,
			&block.TotalFee,
			&block.TotalCoinBase,
			&block.Version,
		)
		blocks = append(blocks, &block)
	}
	return blocks
}

// Rollback delete records `WHERE height > "height"`
func (bq *BlockQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", bq.getTableName()),
			height,
		},
	}
}
