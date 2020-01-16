package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MegablockQueryInterface interface {
		InsertMegablock(megablock *model.Megablock) (str string, args []interface{})
		GetMegablocksByBlockHeight(height uint32, ct chaintype.ChainType) string
		GetLastMegablock() string
		ExtractModel(mb *model.Megablock) []interface{}
		BuildModel(megablocks []*model.Megablock, rows *sql.Rows) ([]*model.Megablock, error)
		Scan(mb *model.Megablock, row *sql.Row) error
		Rollback(spineBlockHeight uint32) [][]interface{}
	}

	MegablockQuery struct {
		Fields    []string
		TableName string
	}
)

func NewMegablockQuery() *MegablockQuery {
	return &MegablockQuery{
		Fields: []string{
			"full_snapshot_hash",
			"spine_block_height",
			"main_block_height",
		},
		TableName: "megablock",
	}
}

func (mbl *MegablockQuery) getTableName() string {
	return mbl.TableName
}

// InsertMegablock
func (mbl *MegablockQuery) InsertMegablock(megablock *model.Megablock) (str string, args []interface{}) {
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		mbl.getTableName(),
		strings.Join(mbl.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(mbl.Fields)-1)),
	)
	return qryInsert, mbl.ExtractModel(megablock)
}

// GetMegablocksByBlockHeight returns query string to get megablock at given block's height (spine or main)
func (mbl *MegablockQuery) GetMegablocksByBlockHeight(height uint32, ct chaintype.ChainType) (str string) {
	var (
		heightPrefix string
	)
	switch ct.(type) {
	case *chaintype.MainChain:
		heightPrefix = "main"
	case *chaintype.SpineChain:
		heightPrefix = "spine"
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s_block_height = %d",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), heightPrefix, height)
	return query
}

// GetLastMegablock returns the last megablock
func (mbl *MegablockQuery) GetLastMegablock() string {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY spine_block_height DESC LIMIT 1",
		strings.Join(mbl.Fields, ", "), mbl.getTableName())
	return query
}

// ExtractModel extract the model struct fields to the order of MegablockQuery.Fields
func (mbl *MegablockQuery) ExtractModel(mb *model.Megablock) []interface{} {
	return []interface{}{
		mb.FullSnapshotHash,
		mb.SpineBlockHeight,
		mb.MainBlockHeight,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (mbl *MegablockQuery) BuildModel(
	megablocks []*model.Megablock,
	rows *sql.Rows,
) ([]*model.Megablock, error) {
	for rows.Next() {
		var (
			mb  model.Megablock
			err error
		)
		err = rows.Scan(
			&mb.FullSnapshotHash,
			&mb.SpineBlockHeight,
			&mb.MainBlockHeight,
		)
		if err != nil {
			return nil, err
		}
		megablocks = append(megablocks, &mb)
	}
	return megablocks, nil
}

// Rollback delete records `WHERE spine_block_height > `height`
func (mbl *MegablockQuery) Rollback(spineBlockHeight uint32) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE spine_block_height > ?", mbl.TableName),
			spineBlockHeight,
		},
	}
}

// Scan represents `sql.Scan`
func (mbl *MegablockQuery) Scan(mb *model.Megablock, row *sql.Row) error {
	err := row.Scan(
		&mb.FullSnapshotHash,
		&mb.SpineBlockHeight,
		&mb.MainBlockHeight,
	)
	if err != nil {
		return err
	}
	return nil
}
