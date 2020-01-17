package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SnapshotChunkQueryInterface interface {
		InsertSnapshotChunk(snapshotChunk *model.SnapshotChunk) (str string, args []interface{})
		GetSnapshotChunksByBlockHeight(height uint32) string
		GetSnapshotChunkByChunkHash(chunkHash []byte) (str string, args []interface{})
		GetLastSnapshotChunk() string
		ExtractModel(sc *model.SnapshotChunk) []interface{}
		BuildModel(snapshotChunks []*model.SnapshotChunk, rows *sql.Rows) ([]*model.SnapshotChunk, error)
		Scan(sc *model.SnapshotChunk, row *sql.Row) error
		Rollback(spineBlockHeight uint32) [][]interface{}
	}

	SnapshotChunkQuery struct {
		Fields    []string
		TableName string
	}
)

func NewSnapshotChunkQuery() *SnapshotChunkQuery {
	return &SnapshotChunkQuery{
		Fields: []string{
			"chunk_hash",
			"chunk_index",
			"previous_chunk_hash",
			"spine_block_height",
		},
		TableName: "snapshot_chunk",
	}
}

func (scl *SnapshotChunkQuery) getTableName() string {
	return scl.TableName
}

// InsertSnapshotChunk
func (scl *SnapshotChunkQuery) InsertSnapshotChunk(snapshotChunk *model.SnapshotChunk) (str string, args []interface{}) {
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		scl.getTableName(),
		strings.Join(scl.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(scl.Fields)-1)),
	)
	return qryInsert, scl.ExtractModel(snapshotChunk)
}

// GetSnapshotChunksByBlockHeight returns query string to get snapshotChunk at given spine block's height
func (scl *SnapshotChunkQuery) GetSnapshotChunksByBlockHeight(height uint32) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE spine_block_height = %d",
		strings.Join(scl.Fields, ", "), scl.getTableName(), height)
	return query
}

// GetSnapshotChunkByChunkHash returns query string to get snapshotChunk with a given chunk hash (pri key)
func (scl *SnapshotChunkQuery) GetSnapshotChunkByChunkHash(chunkHash []byte) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE chunk_hash = ?",
		strings.Join(scl.Fields, ", "), scl.getTableName(),
	), []interface{}{chunkHash}
}

// GetLastSnapshotChunk returns the last snapshotChunk
func (scl *SnapshotChunkQuery) GetLastSnapshotChunk() string {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY spine_block_height, chunk_index DESC LIMIT 1",
		strings.Join(scl.Fields, ", "), scl.getTableName())
	return query
}

// ExtractModel extract the model struct fields to the order of SnapshotChunkQuery.Fields
func (scl *SnapshotChunkQuery) ExtractModel(sc *model.SnapshotChunk) []interface{} {
	return []interface{}{
		sc.ChunkHash,
		sc.ChunkIndex,
		sc.PreviousChunkHash,
		sc.SpineBlockHeight,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (scl *SnapshotChunkQuery) BuildModel(
	snapshotChunks []*model.SnapshotChunk,
	rows *sql.Rows,
) ([]*model.SnapshotChunk, error) {
	for rows.Next() {
		var (
			sc  model.SnapshotChunk
			err error
		)
		err = rows.Scan(
			&sc.ChunkHash,
			&sc.ChunkIndex,
			&sc.PreviousChunkHash,
			&sc.SpineBlockHeight,
		)
		if err != nil {
			return nil, err
		}
		snapshotChunks = append(snapshotChunks, &sc)
	}
	return snapshotChunks, nil
}

// Rollback delete records `WHERE spine_block_height > `height`
func (scl *SnapshotChunkQuery) Rollback(spineBlockHeight uint32) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE spine_block_height > ?", scl.TableName),
			spineBlockHeight,
		},
	}
}

// Scan represents `sql.Scan`
func (scl *SnapshotChunkQuery) Scan(sc *model.SnapshotChunk, row *sql.Row) error {
	err := row.Scan(
		&sc.ChunkHash,
		&sc.ChunkIndex,
		&sc.PreviousChunkHash,
		&sc.SpineBlockHeight,
	)
	if err != nil {
		return err
	}
	return nil
}
