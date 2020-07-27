package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	FeeScaleQueryInterface interface {
		GetLatestFeeScale() string
		InsertFeeScale(feeScale *model.FeeScale) [][]interface{}
		ExtractModel(feeScale *model.FeeScale) []interface{}
		BuildModel(feeScales []*model.FeeScale, rows *sql.Rows) ([]*model.FeeScale, error)
		Scan(feeScale *model.FeeScale, row *sql.Row) error
	}

	FeeScaleQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewFeeScaleQuery returns FeeScaleQuery instance
func NewFeeScaleQuery() *FeeScaleQuery {
	return &FeeScaleQuery{
		Fields: []string{
			"fee_scale",
			"block_height",
			"latest",
		},
		TableName: "fee_scale",
	}
}

func (fsq *FeeScaleQuery) getTableName() string {
	return fsq.TableName
}

// GetLatestFeeScale return the latest fee scale
func (fsq *FeeScaleQuery) GetLatestFeeScale() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE latest = true",
		strings.Join(fsq.Fields, ", "), fsq.getTableName())
}

// InsertFeeScale insert new fee scale record
func (fsq *FeeScaleQuery) InsertFeeScale(feeScale *model.FeeScale) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s SET latest = ? WHERE latest = ? AND block_height IN (SELECT MAX(t2.block_height) FROM %s as t2)",
				fsq.getTableName(), fsq.getTableName(),
			),
			0,
			1,
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					fsq.getTableName(),
					strings.Join(fsq.Fields, ", "),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(fsq.Fields)-1)),
				),
			},
			fsq.ExtractModel(feeScale)...,
		),
	}
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*FeeScaleQuery) ExtractModel(feeScale *model.FeeScale) []interface{} {
	return []interface{}{
		feeScale.FeeScale,
		feeScale.BlockHeight,
		feeScale.Latest,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*FeeScaleQuery) BuildModel(
	feeScales []*model.FeeScale,
	rows *sql.Rows,
) ([]*model.FeeScale, error) {
	for rows.Next() {
		var (
			feeScale model.FeeScale
			err      error
		)
		err = rows.Scan(
			&feeScale.FeeScale,
			&feeScale.BlockHeight,
			&feeScale.Latest,
		)
		if err != nil {
			return nil, err
		}
		feeScales = append(feeScales, &feeScale)
	}
	return feeScales, nil
}

// Scan similar with `sql.Scan`
func (*FeeScaleQuery) Scan(feeScale *model.FeeScale, row *sql.Row) error {
	err := row.Scan(
		&feeScale.FeeScale,
		&feeScale.BlockHeight,
		&feeScale.Latest,
	)
	return err
}

// Rollback delete records `WHERE height > "block_height"
func (fsq *FeeScaleQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", fsq.getTableName()),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND block_height IN (
				SELECT MAX(t2.block_height)
				FROM %s as t2
			)`,
				fsq.TableName,
				fsq.TableName,
			),
			1,
			0,
		},
	}
}

// SelectDataForSnapshot select only the block at snapshot block_height (fromHeight is unused)
func (fsq *FeeScaleQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height > 0 AND block_height >= %d AND block_height <= %d`,
		strings.Join(fsq.Fields, ","), fsq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fsq *FeeScaleQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	// do not delete genesis block
	if fromHeight == 0 {
		fromHeight++
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		fsq.getTableName(), fromHeight, toHeight)
}
