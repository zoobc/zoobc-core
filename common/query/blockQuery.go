package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"
)

type (
	// BlockQueryInterface interface of BlockQuery
	BlockQueryInterface interface {
		GetBlocks(contract.ChainType, uint32) string
		GetBlockByID(contract.ChainType, int64) string
		GetBlockByHeight(contract.ChainType, uint32) string
	}

	// BlockQuery holds needed in querying a block
	BlockQuery struct {
		ChainType contract.ChainType
		Fields    []string
		TableName string
	}
)

// NewBlockQuery returns BlockQuery instance
func NewBlockQuery(chainType contract.ChainType) *BlockQuery {
	blockQuery := BlockQuery{
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
			"blocksmith_id",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		},
		TableName: "block",
		ChainType: chainType,
	}

	return &blockQuery
}

func (bq *BlockQuery) getTableName() string {
	return bq.ChainType.GetTablePrefix() + "_" + bq.TableName
}

// GetBlocks returns query string to get multiple blocks
func (bq *BlockQuery) GetBlocks(height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(bq.Fields, ", "), bq.getTableName())
}

// GetBlockByID returns query string to get block by ID
func (bq *BlockQuery) GetBlockByID(ID int64) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), ID)
}

// GetBlockByHeight returns query string to get block by height
func (bq *BlockQuery) GetBlockByHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), height)
}
