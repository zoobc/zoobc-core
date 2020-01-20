package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	FileChunkQueryInterface interface {
		InsertFileChunk(snapshotChunk *model.FileChunk) (str string, args []interface{})
		GetFileChunksByMegablockID(megablockID int64) (str string)
		GetFileChunksByBlockHeight(height uint32, ct chaintype.ChainType) string
		GetFileChunkByChunkHash(chunkHash []byte) (str string, args []interface{})
		GetLastFileChunk(ct chaintype.ChainType) string
		ExtractModel(sc *model.FileChunk) []interface{}
		BuildModel(snapshotChunks []*model.FileChunk, rows *sql.Rows) ([]*model.FileChunk, error)
		Scan(sc *model.FileChunk, row *sql.Row) error
		Rollback(spineBlockHeight uint32) [][]interface{}
	}

	FileChunkQuery struct {
		Fields    []string
		TableName string
	}
)

func NewFileChunkQuery() *FileChunkQuery {
	return &FileChunkQuery{
		Fields: []string{
			"chunk_hash",
			"megablock_id",
			"chunk_index",
			"previous_chunk_hash",
			"spine_block_height",
			"chain_type",
		},
		TableName: "file_chunk",
	}
}

func (scl *FileChunkQuery) getTableName() string {
	return scl.TableName
}

// InsertFileChunk
func (scl *FileChunkQuery) InsertFileChunk(snapshotChunk *model.FileChunk) (str string, args []interface{}) {
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		scl.getTableName(),
		strings.Join(scl.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(scl.Fields)-1)),
	)
	return qryInsert, scl.ExtractModel(snapshotChunk)
}

// GetFileChunksByBlockHeight returns query string to get snapshotChunk at given spine block's height
func (scl *FileChunkQuery) GetFileChunksByBlockHeight(height uint32, ct chaintype.ChainType) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE spine_block_height = %d AND chain_type = %d",
		strings.Join(scl.Fields, ", "), scl.getTableName(), height, ct.GetTypeInt())
	return query
}

// GetFileChunksByMegablockID returns query string to get all snapshotChunks relative to a megablock 
func (scl *FileChunkQuery) GetFileChunksByMegablockID(megablockID int64) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE megablock_id = %d",
		strings.Join(scl.Fields, ", "), scl.getTableName(), megablockID)
	return query
}

// GetFileChunkByChunkHash returns query string to get snapshotChunk with a given chunk hash (pri key)
func (scl *FileChunkQuery) GetFileChunkByChunkHash(chunkHash []byte) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE chunk_hash = ?",
		strings.Join(scl.Fields, ", "), scl.getTableName(),
	), []interface{}{chunkHash}
}

// GetLastFileChunk returns the last snapshotChunk
func (scl *FileChunkQuery) GetLastFileChunk(ct chaintype.ChainType) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chain_type = %d ORDER BY spine_block_height, chunk_index DESC LIMIT 1",
		strings.Join(scl.Fields, ", "), scl.getTableName(), ct.GetTypeInt())
	return query
}

// ExtractModel extract the model struct fields to the order of FileChunkQuery.Fields
func (scl *FileChunkQuery) ExtractModel(sc *model.FileChunk) []interface{} {
	return []interface{}{
		sc.ChunkHash,
		sc.MegablockID,
		sc.ChunkIndex,
		sc.PreviousChunkHash,
		sc.SpineBlockHeight,
		sc.ChainType,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (scl *FileChunkQuery) BuildModel(
	snapshotChunks []*model.FileChunk,
	rows *sql.Rows,
) ([]*model.FileChunk, error) {
	for rows.Next() {
		var (
			sc  model.FileChunk
			err error
		)
		err = rows.Scan(
			&sc.ChunkHash,
			&sc.MegablockID,
			&sc.ChunkIndex,
			&sc.PreviousChunkHash,
			&sc.SpineBlockHeight,
			&sc.ChainType,
		)
		if err != nil {
			return nil, err
		}
		snapshotChunks = append(snapshotChunks, &sc)
	}
	return snapshotChunks, nil
}

// Rollback delete records `WHERE spine_block_height > `height`
func (scl *FileChunkQuery) Rollback(spineBlockHeight uint32) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE spine_block_height > ?", scl.TableName),
			spineBlockHeight,
		},
	}
}

// Scan represents `sql.Scan`
func (scl *FileChunkQuery) Scan(sc *model.FileChunk, row *sql.Row) error {
	err := row.Scan(
		&sc.ChunkHash,
		&sc.MegablockID,
		&sc.ChunkIndex,
		&sc.PreviousChunkHash,
		&sc.SpineBlockHeight,
		&sc.ChainType,
	)
	if err != nil {
		return err
	}
	return nil
}
