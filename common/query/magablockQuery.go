package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MegablockQueryInterface interface {
		InsertMegablock(megablock *model.Megablock) (str string, args []interface{})
		GetMegablocksBySpineBlockHeight(height uint32) string
		GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType(
			height uint32,
			ct chaintype.ChainType,
			mbType model.MegablockType,
		) string
		GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) string
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
			"id",
			"full_file_hash",
			"megablock_payload_length",
			"megablock_payload_hash",
			"spine_block_height",
			"megablock_height",
			"chain_type",
			"megablock_type",
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

// GetMegablocksBySpineBlockHeight returns query string to get all megablocks at given spine block's height
func (mbl *MegablockQuery) GetMegablocksBySpineBlockHeight(height uint32) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE spine_block_height = %d ORDER BY megablock_type, chain_type, id",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), height)
	return query
}

// GetMegablocksBySpineBlockHeight returns query string to get all megablocks at given spine block's height
func (mbl *MegablockQuery) GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType(
	height uint32,
	ct chaintype.ChainType,
	mbType model.MegablockType,
) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE spine_block_height = %d AND chain_type = %d AND megablock_type = %d LIMIT 1",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), height, ct.GetTypeInt(), mbType)
	return query
}

// GetLastMegablock returns the last megablock
func (mbl *MegablockQuery) GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chain_type = %d AND megablock_type = %d ORDER BY spine_block_height DESC LIMIT 1",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), ct.GetTypeInt(), mbType)
	return query
}

// ExtractModel extract the model struct fields to the order of MegablockQuery.Fields
func (mbl *MegablockQuery) ExtractModel(mb *model.Megablock) []interface{} {
	return []interface{}{
		mb.ID,
		mb.FullFileHash,
		mb.MegablockPayloadLength,
		mb.MegablockPayloadHash,
		mb.SpineBlockHeight,
		mb.MegablockHeight,
		mb.ChainType,
		mb.MegablockType,
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
			&mb.ID,
			&mb.FullFileHash,
			&mb.MegablockPayloadLength,
			&mb.MegablockPayloadHash,
			&mb.SpineBlockHeight,
			&mb.MegablockHeight,
			&mb.ChainType,
			&mb.MegablockType,
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
		&mb.ID,
		&mb.FullFileHash,
		&mb.MegablockPayloadLength,
		&mb.MegablockPayloadHash,
		&mb.SpineBlockHeight,
		&mb.MegablockHeight,
		&mb.ChainType,
		&mb.MegablockType,
	)
	if err != nil {
		return err
	}
	return nil
}
