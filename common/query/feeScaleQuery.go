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
		InsertFeeScale(feeScale *model.FeeScale) (qStr string, args []interface{})
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

// GetFeeScale return the latest fee scale
func (fsq *FeeScaleQuery) GetLatestFeeScale() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE latest = true",
		strings.Join(fsq.Fields, ", "), fsq.getTableName())
}

// InsertFeeScale insert new fee scale record
func (fsq *FeeScaleQuery) InsertFeeScale(feeScale *model.FeeScale) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fsq.getTableName(),
		strings.Join(fsq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(fsq.Fields)-1)),
	), fsq.ExtractModel(feeScale)
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*FeeScaleQuery) ExtractModel(feeScale *model.FeeScale) []interface{} {
	return []interface{}{
		&feeScale.FeeScale,
		&feeScale.BlockHeight,
		&feeScale.Latest,
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
