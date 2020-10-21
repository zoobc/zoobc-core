package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlockchainObjectPropertyQueryInterface interface {
		InsertBlockcahinObjectProperties(properties []*model.BlockchainObjectProperty) (str string, args []interface{})
		ExtractModel(property *model.BlockchainObjectProperty) []interface{}
		BuildModel(property []*model.BlockchainObjectProperty, rows *sql.Rows) ([]*model.BlockchainObjectProperty, error)
	}
	BlockchainObjectPropertyQuery struct {
		Fields    []string
		TableName string
	}
)

// NewBatchReceiptQuery returns BatchReceiptQuery instance
func NewBlockchainObjectPropertyQuery() *BlockchainObjectPropertyQuery {
	return &BlockchainObjectPropertyQuery{
		Fields: []string{
			"blockchain_object_id",
			"key",
			"value",
			"block_height",
		},
		TableName: "blockchain_object_property",
	}
}

// InsertBlockcahinObjectProperties build query for bulk store pas
func (bopq *BlockchainObjectPropertyQuery) InsertBlockcahinObjectProperties(
	properties []*model.BlockchainObjectProperty,
) (str string, args []interface{}) {
	var (
		values []interface{}
		query  = fmt.Sprintf(
			"INSERT INTO %s (%s) ",
			bopq.getTableName(),
			strings.Join(bopq.Fields, ", "),
		)
	)
	for k, property := range properties {
		query += fmt.Sprintf("VALUES(?%s)", strings.Repeat(",? ", len(bopq.Fields)-1))
		if k < len(properties)-1 {
			query += ", "
		}
		values = append(values, bopq.ExtractModel(property)...)
	}
	return query, values
}

func (bopq *BlockchainObjectPropertyQuery) getTableName() string {
	return bopq.TableName
}

func (*BlockchainObjectPropertyQuery) ExtractModel(property *model.BlockchainObjectProperty) []interface{} {
	return []interface{}{
		property.BlockchainObjectID,
		property.Key,
		property.Value,
		property.BlockHeight,
	}
}

func (*BlockchainObjectPropertyQuery) BuildModel(
	properties []*model.BlockchainObjectProperty, rows *sql.Rows,
) ([]*model.BlockchainObjectProperty, error) {
	for rows.Next() {
		var (
			property model.BlockchainObjectProperty
		)
		if err := rows.Scan(
			&property.BlockchainObjectID,
			&property.Key,
			&property.Value,
			&property.BlockHeight,
		); err != nil {
			return nil, err
		}
		properties = append(properties, &property)
	}
	return properties, nil
}

// Rollback delete records `WHERE height > "block_height"
func (bopq *BlockchainObjectPropertyQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", bopq.getTableName()),
			height,
		},
	}
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (bopq *BlockchainObjectPropertyQuery) RecalibrateVersionedTable() []string {
	return []string{} // only table with `latest` column need this
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (bopq *BlockchainObjectPropertyQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	blockchainObjectProperties, ok := payload.([]*model.BlockchainObjectProperty)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+bopq.TableName)
	}
	if len(blockchainObjectProperties) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(bopq.Fields), len(blockchainObjectProperties))
		for i := 0; i < rounds; i++ {
			qry, args := bopq.InsertBlockcahinObjectProperties(blockchainObjectProperties[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := bopq.InsertBlockcahinObjectProperties(blockchainObjectProperties[len(blockchainObjectProperties)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// SelectDataForSnapshot select only the block at snapshot height (fromHeight is unused)
func (bopq *BlockchainObjectPropertyQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		strings.Join(bopq.Fields, ","), bopq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (bopq *BlockchainObjectPropertyQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		bopq.getTableName(), fromHeight, toHeight)
}
