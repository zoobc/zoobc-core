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
		GetMegablocksInTimeInterval(fromTimestamp, toTimestamp int64) string
		GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) string
		ExtractModel(mb *model.Megablock) []interface{}
		BuildModel(megablocks []*model.Megablock, rows *sql.Rows) ([]*model.Megablock, error)
		Scan(mb *model.Megablock, row *sql.Row) error
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
			"file_chunk_hashes",
			"megablock_height",
			"chain_type",
			"megablock_type",
			"expiration_timestamp",
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

// GetLastMegablock returns the last megablock
func (mbl *MegablockQuery) GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chain_type = %d AND megablock_type = %d ORDER BY megablock_height DESC LIMIT 1",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), ct.GetTypeInt(), mbType)
	return query
}

// GetMegablocksInTimeInterval retrieve all megablocks within a time frame
// Note: it is used to get all entities that have expired between spine blocks
func (mbl *MegablockQuery) GetMegablocksInTimeInterval(fromTimestamp, toTimestamp int64) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE expiration_timestamp > %d AND expiration_timestamp <= %d "+
		"ORDER BY megablock_type, chain_type, megablock_height",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), fromTimestamp, toTimestamp)
	return query
}

// ExtractModel extract the model struct fields to the order of MegablockQuery.Fields
func (mbl *MegablockQuery) ExtractModel(mb *model.Megablock) []interface{} {
	return []interface{}{
		mb.ID,
		mb.FullFileHash,
		mb.FileChunkHashes,
		mb.MegablockHeight,
		mb.ChainType,
		mb.MegablockType,
		mb.ExpirationTimestamp,
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
			&mb.FileChunkHashes,
			&mb.MegablockHeight,
			&mb.ChainType,
			&mb.MegablockType,
			&mb.ExpirationTimestamp,
		)
		if err != nil {
			return nil, err
		}
		megablocks = append(megablocks, &mb)
	}
	return megablocks, nil
}

// Scan represents `sql.Scan`
func (mbl *MegablockQuery) Scan(mb *model.Megablock, row *sql.Row) error {
	err := row.Scan(
		&mb.ID,
		&mb.FullFileHash,
		&mb.FileChunkHashes,
		&mb.MegablockHeight,
		&mb.ChainType,
		&mb.MegablockType,
		&mb.ExpirationTimestamp,
	)
	if err != nil {
		return err
	}
	return nil
}
