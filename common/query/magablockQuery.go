package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SpineBlockManifestQueryInterface interface {
		InsertSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) (str string, args []interface{})
		GetSpineBlockManifestsInTimeInterval(fromTimestamp, toTimestamp int64) string
		GetLastSpineBlockManifest(ct chaintype.ChainType, mbType model.SpineBlockManifestType) string
		ExtractModel(mb *model.SpineBlockManifest) []interface{}
		BuildModel(spineBlockManifests []*model.SpineBlockManifest, rows *sql.Rows) ([]*model.SpineBlockManifest, error)
		Scan(mb *model.SpineBlockManifest, row *sql.Row) error
	}

	SpineBlockManifestQuery struct {
		Fields    []string
		TableName string
	}
)

func NewSpineBlockManifestQuery() *SpineBlockManifestQuery {
	return &SpineBlockManifestQuery{
		Fields: []string{
			"id",
			"full_file_hash",
			"file_chunk_hashes",
			"manifest_reference_height",
			"chain_type",
			"manifest_type",
			"manifest_timestamp",
		},
		TableName: "spine_block_manifest",
	}
}

func (mbl *SpineBlockManifestQuery) getTableName() string {
	return mbl.TableName
}

// InsertSpineBlockManifest
func (mbl *SpineBlockManifestQuery) InsertSpineBlockManifest(
	spineBlockManifest *model.SpineBlockManifest,
) (str string, args []interface{}) {
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		mbl.getTableName(),
		strings.Join(mbl.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(mbl.Fields)-1)),
	)
	return qryInsert, mbl.ExtractModel(spineBlockManifest)
}

// GetLastSpineBlockManifest returns the last spineBlockManifest
func (mbl *SpineBlockManifestQuery) GetLastSpineBlockManifest(ct chaintype.ChainType, mbType model.SpineBlockManifestType) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chain_type = %d AND manifest_type = %d ORDER BY manifest_reference_height "+
		"DESC LIMIT 1", strings.Join(mbl.Fields, ", "), mbl.getTableName(), ct.GetTypeInt(), mbType)
	return query
}

// GetSpineBlockManifestsInTimeInterval retrieve all spineBlockManifests within a time frame
// Note: it is used to get all entities that have expired between spine blocks
func (mbl *SpineBlockManifestQuery) GetSpineBlockManifestsInTimeInterval(fromTimestamp, toTimestamp int64) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE manifest_timestamp > %d AND manifest_timestamp <= %d "+
		"ORDER BY manifest_type, chain_type, manifest_reference_height",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), fromTimestamp, toTimestamp)
	return query
}

// ExtractModel extract the model struct fields to the order of SpineBlockManifestQuery.Fields
func (mbl *SpineBlockManifestQuery) ExtractModel(mb *model.SpineBlockManifest) []interface{} {
	return []interface{}{
		mb.ID,
		mb.FullFileHash,
		mb.FileChunkHashes,
		mb.SpineBlockManifestHeight,
		mb.ChainType,
		mb.SpineBlockManifestType,
		mb.ExpirationTimestamp,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (mbl *SpineBlockManifestQuery) BuildModel(
	spineBlockManifests []*model.SpineBlockManifest,
	rows *sql.Rows,
) ([]*model.SpineBlockManifest, error) {
	for rows.Next() {
		var (
			mb  model.SpineBlockManifest
			err error
		)
		err = rows.Scan(
			&mb.ID,
			&mb.FullFileHash,
			&mb.FileChunkHashes,
			&mb.SpineBlockManifestHeight,
			&mb.ChainType,
			&mb.SpineBlockManifestType,
			&mb.ExpirationTimestamp,
		)
		if err != nil {
			return nil, err
		}
		spineBlockManifests = append(spineBlockManifests, &mb)
	}
	return spineBlockManifests, nil
}

// Scan represents `sql.Scan`
func (mbl *SpineBlockManifestQuery) Scan(mb *model.SpineBlockManifest, row *sql.Row) error {
	err := row.Scan(
		&mb.ID,
		&mb.FullFileHash,
		&mb.FileChunkHashes,
		&mb.SpineBlockManifestHeight,
		&mb.ChainType,
		&mb.SpineBlockManifestType,
		&mb.ExpirationTimestamp,
	)
	if err != nil {
		return err
	}
	return nil
}
