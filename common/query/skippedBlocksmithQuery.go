package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SkippedBlocksmithQueryInterface interface {
		GetSkippedBlocksmithsByBlockHeight(blockHeight uint32) (qStr string)
		InsertSkippedBlocksmith(skippedBlocksmith *model.SkippedBlocksmith) (qStr string, args []interface{})
		ExtractModel(skippedBlocksmith *model.SkippedBlocksmith) []interface{}
		BuildModel(skippedBlocksmiths []*model.SkippedBlocksmith, rows *sql.Rows) ([]*model.SkippedBlocksmith, error)
		Scan(skippedBlocksmith *model.SkippedBlocksmith, rows *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}

	SkippedBlocksmithQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewSkippedBlocksmithQuery will create a new SkippedBlocksmithQuery instance
func NewSkippedBlocksmithQuery() *SkippedBlocksmithQuery {
	return &SkippedBlocksmithQuery{
		Fields: []string{
			"blocksmith_public_key",
			"pop_change",
			"block_height",
			"blocksmith_index",
		},
		TableName: "skipped_blocksmith",
	}
}

func (sbq *SkippedBlocksmithQuery) getTableName() string {
	return sbq.TableName
}

func (sbq *SkippedBlocksmithQuery) GetSkippedBlocksmithsByBlockHeight(blockHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height = %d",
		strings.Join(sbq.Fields, ", "),
		sbq.getTableName(),
		blockHeight,
	)
}

func (sbq *SkippedBlocksmithQuery) InsertSkippedBlocksmith(
	skippedBlocksmith *model.SkippedBlocksmith,
) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			sbq.getTableName(),
			strings.Join(sbq.Fields, ", "),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(sbq.Fields)-1)),
		),
		sbq.ExtractModel(skippedBlocksmith)
}

func (*SkippedBlocksmithQuery) ExtractModel(skippedModel *model.SkippedBlocksmith) []interface{} {
	return []interface{}{
		&skippedModel.BlocksmithPublicKey,
		&skippedModel.POPChange,
		&skippedModel.BlockHeight,
		&skippedModel.BlocksmithIndex,
	}
}

func (*SkippedBlocksmithQuery) BuildModel(
	skippedBlocksmiths []*model.SkippedBlocksmith,
	rows *sql.Rows,
) ([]*model.SkippedBlocksmith, error) {
	for rows.Next() {
		var (
			skippedBlocksmith model.SkippedBlocksmith
			err               error
		)
		err = rows.Scan(
			&skippedBlocksmith.BlocksmithPublicKey,
			&skippedBlocksmith.POPChange,
			&skippedBlocksmith.BlockHeight,
			&skippedBlocksmith.BlocksmithIndex,
		)
		if err != nil {
			return nil, err
		}
		skippedBlocksmiths = append(skippedBlocksmiths, &skippedBlocksmith)
	}
	return skippedBlocksmiths, nil
}

func (*SkippedBlocksmithQuery) Scan(skippedBlocksmith *model.SkippedBlocksmith, row *sql.Row) error {
	err := row.Scan(
		&skippedBlocksmith.BlocksmithPublicKey,
		&skippedBlocksmith.POPChange,
		&skippedBlocksmith.BlockHeight,
		&skippedBlocksmith.BlocksmithIndex,
	)
	return err
}
func (sbq *SkippedBlocksmithQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", sbq.getTableName()),
			height,
		},
	}
}
