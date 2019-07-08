package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlockQueryInterface interface {
		GetBlocks() string
		GetLastBlock() string
		GetGenesisBlock() string
		InsertBlock() string
		ExtractModel(block model.Block) []interface{}
	}

	BlockQuery struct {
		Fields    []string
		Chaintype contract.ChainType
	}
)

func NewBlockQuery(chaintype contract.ChainType) *BlockQuery {
	return &BlockQuery{
		Fields: []string{"id", "previous_block_hash", "height", "timestamp", "block_seed", "block_signature", "cumulative_difficulty",
			"smith_scale", "payload_length", "payload_hash", "blocksmith_id", "total_amount", "total_fee", "total_coinbase", "version",
		},
		Chaintype: chaintype,
	}
}

func (bq *BlockQuery) getTableName() string {
	return bq.Chaintype.GetTablePrefix() + "_block"
}

func (bq *BlockQuery) GetBlocks() string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) GetLastBlock() string {
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY HEIGHT DESC LIMIT 1", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) GetGenesisBlock() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE HEIGHT = 0 ORDER BY HEIGHT DESC LIMIT 1", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *BlockQuery) InsertBlock() string {
	var value = ":" + bq.Fields[0]
	for _, field := range bq.Fields[1:] {
		value += (", :" + field)

	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		bq.getTableName(), strings.Join(bq.Fields, ", "), value)
	return query
}

// ExtractModel extract the model struct fields to the order of BlockQuery.Fields
func (*BlockQuery) ExtractModel(block model.Block) []interface{} {
	return []interface{}{block.ID, block.PreviousBlockHash, block.Height, block.Timestamp, block.BlockSeed, block.BlockSignature, block.CumulativeDifficulty,
		block.SmithScale, block.PayloadLength, block.PayloadHash, block.BlocksmithID, block.TotalAmount, block.TotalFee, block.TotalCoinBase, block.Version}
}
