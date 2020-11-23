package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BlockchainObjectQueryInterface interface {
		InsertBlockcahinObject(blockchainObject *model.BlockchainObject) (str string, args []interface{})
		InsertBlockcahinObjects(properties []*model.BlockchainObject) (str string, args []interface{})
		GetBlockchainObjectByID(id []byte) (str string, args []interface{})
		ExtractModel(blockchainObject *model.BlockchainObject) []interface{}
		BuildModel(blockchainObject []*model.BlockchainObject, rows *sql.Rows) ([]*model.BlockchainObject, error)
		Scan(blockchainObject *model.BlockchainObject, row *sql.Row) error
	}
	BlockchainObjectQuery struct {
		Fields    []string
		TableName string
	}
)

// NewBlockchainObjectQuery returns BlockchainObjectQuery instance
func NewBlockchainObjectQuery() *BlockchainObjectQuery {
	return &BlockchainObjectQuery{
		Fields: []string{
			"id",
			"owner_account_address",
			"block_height",
			"latest",
		},
		TableName: "blockchain_object",
	}
}

// InsertBlockcahinObjects represents query builder to insert single record blockchain object
func (boq *BlockchainObjectQuery) InsertBlockcahinObject(
	blockchainObject *model.BlockchainObject,
) (str string, args []interface{}) {
	var (
		value = fmt.Sprintf("? %s", strings.Repeat(", ?", len(boq.Fields)-1))
		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", boq.getTableName(), strings.Join(boq.Fields, ", "), value)
	)
	return query, boq.ExtractModel(blockchainObject)
}

// InsertBlockcahinObjects represents query builder to insert multiple record in single query
func (boq *BlockchainObjectQuery) InsertBlockcahinObjects(
	blockchainObjects []*model.BlockchainObject,
) (str string, args []interface{}) {
	var (
		values []interface{}
		query  = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			boq.getTableName(),
			strings.Join(boq.Fields, ", "),
		)
	)
	for k, blockchainObject := range blockchainObjects {
		query += fmt.Sprintf("(?%s)", strings.Repeat(",? ", len(boq.Fields)-1))
		if k < len(blockchainObjects)-1 {
			query += ", "
		}
		values = append(values, boq.ExtractModel(blockchainObject)...)
	}
	return query, values
}

// GetBlockchainObjectByID represents query builder to get latest version blockchain object besed on it's ID
func (boq *BlockchainObjectQuery) GetBlockchainObjectByID(id []byte) (str string, args []interface{}) {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE id = ? AND latest = 1 ORDER BY block_height DESC`,
		strings.Join(boq.Fields, ","), boq.TableName), []interface{}{id}
}

func (boq *BlockchainObjectQuery) getTableName() string {
	return boq.TableName
}

func (*BlockchainObjectQuery) ExtractModel(blockchainObject *model.BlockchainObject) []interface{} {
	return []interface{}{
		blockchainObject.ID,
		blockchainObject.OwnerAccountAddress,
		blockchainObject.BlockHeight,
		blockchainObject.Latest,
	}
}

func (*BlockchainObjectQuery) BuildModel(
	blockchainObjects []*model.BlockchainObject,
	rows *sql.Rows,
) ([]*model.BlockchainObject, error) {
	for rows.Next() {
		var (
			blockchainObject model.BlockchainObject
		)
		if err := rows.Scan(
			&blockchainObject.ID,
			&blockchainObject.OwnerAccountAddress,
			&blockchainObject.BlockHeight,
			&blockchainObject.Latest,
		); err != nil {
			return nil, err
		}
		blockchainObjects = append(blockchainObjects, &blockchainObject)
	}
	return blockchainObjects, nil
}

func (*BlockchainObjectQuery) Scan(blockchainObject *model.BlockchainObject, row *sql.Row) error {
	err := row.Scan(
		&blockchainObject.ID,
		&blockchainObject.OwnerAccountAddress,
		&blockchainObject.BlockHeight,
		&blockchainObject.Latest,
	)
	return err
}

// Rollback delete records `WHERE height > "block_height"
func (boq *BlockchainObjectQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", boq.getTableName()),
			height,
		},
	}
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (boq *BlockchainObjectQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"UPDATE %s SET latest = false WHERE latest = true AND id NOT IN (SELECT MAX(t2.block_height) FROM %s t2",
			boq.getTableName(), boq.getTableName()),
		fmt.Sprintf(
			"UPDATE %s SET latest = true WHERE latest = false AND id IN (SELECT MAX(t2.block_height) FROM %s t2",
			boq.getTableName(), boq.getTableName()),
	}
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (boq *BlockchainObjectQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	blockchainObjects, ok := payload.([]*model.BlockchainObject)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+boq.TableName)
	}
	if len(blockchainObjects) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(boq.Fields), len(blockchainObjects))
		for i := 0; i < rounds; i++ {
			qry, args := boq.InsertBlockcahinObjects(blockchainObjects[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := boq.InsertBlockcahinObjects(blockchainObjects[len(blockchainObjects)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// SelectDataForSnapshot select only the block at snapshot height (fromHeight is unused)
func (boq *BlockchainObjectQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		strings.Join(boq.Fields, ","), boq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (boq *BlockchainObjectQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		boq.getTableName(), fromHeight, toHeight)
}
