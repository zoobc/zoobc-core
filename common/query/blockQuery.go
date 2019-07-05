package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"
)

type (
	BlockQueryInterface interface {
		GetAccounts() string
	}

	BlockQuery struct {
		Fields    []string
		TableName string
	}
)

func NewBlockQuery() *BlockQuery {
	return &BlockQuery{
		Fields: []string{
			"id",
			"hash",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"base_target",
			"generator_public_key",
			"total_amount",
			"total_fee",
			"payload_length",
			"payload_hash",
			"version",
		},
		TableName: "blocks",
	}
}

func (bq *BlockQuery) getTableName(ct contract.ChainType) string {
	return ct.GetTablePrefix() + "_" + bq.TableName
}

func (bq *BlockQuery) GetBlocks(ct contract.ChainType, height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(bq.Fields, ", "), bq.getTableName(ct))
}

func (bq *BlockQuery) GetBlockByID(ct contract.ChainType, ID int64) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(ct), ID)
}

func (bq *BlockQuery) GetBlockByHeight(ct contract.ChainType, height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(ct), height)
}
