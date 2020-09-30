package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlockQueryInterface interface {
		Rollback(height uint32) (multiQueries [][]interface{})
		GetBlocks(height, size uint32) string
		GetLastBlock() string
		GetGenesisBlock() string
		GetBlockByID(id int64) string
		GetBlockByHeight(height uint32) string
		GetBlockFromHeight(startHeight, limit uint32) string
		GetBlockFromTimestamp(startTimestamp int64, limit uint32) string
		InsertBlock(block *model.Block) (str string, args []interface{})
		InsertBlocks(blocks []*model.Block) (str string, args []interface{})
		ExtractModel(block *model.Block) []interface{}
		BuildModel(blocks []*model.Block, rows *sql.Rows) ([]*model.Block, error)
		Scan(block *model.Block, row *sql.Row) error
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
			"height",
			"id",
			"block_hash",
			"previous_block_hash",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
			"merkle_root",
			"merkle_tree",
			"reference_block_height",
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
	return fmt.Sprintf("SELECT MAX(height), %s FROM %s", strings.Join(bq.Fields[1:], ", "), bq.getTableName())
}

func (bq *BlockQuery) GetGenesisBlock() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height = 0", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) InsertBlock(block *model.Block) (str string, args []interface{}) {
	var value = fmt.Sprintf("? %s", strings.Repeat(", ?", len(bq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		bq.getTableName(), strings.Join(bq.Fields, ", "), value)
	return query, bq.ExtractModel(block)
}

// InsertBlocks represents query builder to insert multiple record in single query
func (bq *BlockQuery) InsertBlocks(blocks []*model.Block) (str string, args []interface{}) {
	if len(blocks) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			bq.getTableName(),
			strings.Join(bq.Fields, ", "),
		)
		for k, block := range blocks {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(bq.Fields)-1),
			)
			if k < len(blocks)-1 {
				str += ","
			}
			args = append(args, bq.ExtractModel(block)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (bq *BlockQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	blocks, ok := payload.([]*model.Block)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+bq.TableName)
	}
	if len(blocks) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(bq.Fields), len(blocks))
		for i := 0; i < rounds; i++ {
			qry, args := bq.InsertBlocks(blocks[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := bq.InsertBlocks(blocks[len(blocks)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (bq *BlockQuery) RecalibrateVersionedTable() []string {
	return []string{} // only table with `latest` column need this
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
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d ORDER BY height LIMIT %d",
		strings.Join(bq.Fields, ", "), bq.getTableName(), startHeight, limit)
}

// GetBlockFromTimestamp returns query string to get blocks from a certain block timestamp
func (bq *BlockQuery) GetBlockFromTimestamp(startTimestamp int64, limit uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE timestamp >= %d ORDER BY timestamp LIMIT %d",
		strings.Join(bq.Fields, ", "), bq.getTableName(), startTimestamp, limit)
}

// ExtractModel extract the model struct fields to the order of BlockQuery.Fields
func (*BlockQuery) ExtractModel(block *model.Block) []interface{} {
	return []interface{}{
		block.Height,
		block.ID,
		block.BlockHash,
		block.PreviousBlockHash,
		block.Timestamp,
		block.BlockSeed,
		block.BlockSignature,
		block.CumulativeDifficulty,
		block.PayloadLength,
		block.PayloadHash,
		block.BlocksmithPublicKey,
		block.TotalAmount,
		block.TotalFee,
		block.TotalCoinBase,
		block.Version,
		block.MerkleRoot,
		block.MerkleTree,
		block.ReferenceBlockHeight,
	}
}

func (*BlockQuery) BuildModel(blocks []*model.Block, rows *sql.Rows) ([]*model.Block, error) {
	for rows.Next() {
		var (
			block model.Block
			err   error
		)

		err = rows.Scan(
			&block.Height,
			&block.ID,
			&block.BlockHash,
			&block.PreviousBlockHash,
			&block.Timestamp,
			&block.BlockSeed,
			&block.BlockSignature,
			&block.CumulativeDifficulty,
			&block.PayloadLength,
			&block.PayloadHash,
			&block.BlocksmithPublicKey,
			&block.TotalAmount,
			&block.TotalFee,
			&block.TotalCoinBase,
			&block.Version,
			&block.MerkleRoot,
			&block.MerkleTree,
			&block.ReferenceBlockHeight,
		)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}
	return blocks, nil
}

func (*BlockQuery) Scan(block *model.Block, row *sql.Row) error {
	err := row.Scan(
		&block.Height,
		&block.ID,
		&block.BlockHash,
		&block.PreviousBlockHash,
		&block.Timestamp,
		&block.BlockSeed,
		&block.BlockSignature,
		&block.CumulativeDifficulty,
		&block.PayloadLength,
		&block.PayloadHash,
		&block.BlocksmithPublicKey,
		&block.TotalAmount,
		&block.TotalFee,
		&block.TotalCoinBase,
		&block.Version,
		&block.MerkleRoot,
		&block.MerkleTree,
		&block.ReferenceBlockHeight,
	)
	if err != nil {
		return err

	}
	return nil
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

// SelectDataForSnapshot select only the block at snapshot height (fromHeight is unused)
func (bq *BlockQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		strings.Join(bq.Fields, ","), bq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (bq *BlockQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		bq.getTableName(), fromHeight, toHeight)
}
