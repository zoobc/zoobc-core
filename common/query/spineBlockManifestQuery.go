// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
		GetSpineBlockManifestTimeInterval(fromTimestamp, toTimestamp int64) string
		GetManifestBySpineBlockHeight(spineBlockHeight uint32) string
		GetManifestsFromSpineBlockHeight(spineBlockHeight uint32) string
		GetLastSpineBlockManifest(ct chaintype.ChainType, mbType model.SpineBlockManifestType) string
		GetManifestsFromManifestReferenceHeightRange(fromHeight, toHeight uint32) (qry string, args []interface{})
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
			"manifest_spine_block_height",
			"chain_type",
			"manifest_type",
			"expiration_timestamp",
		},
		TableName: "spine_block_manifest",
	}
}

func (mbl *SpineBlockManifestQuery) getTableName() string {
	return mbl.TableName
}

// InsertSpineBlockManifest insert new spine block manifest
// Note: a new one with same id will replace a previous one, if present.
// this is to allow blocks downloaded from peers to override spine block manifests created locally and insure that the correct
// snapshot is downloaded by the node when first joins the network
func (mbl *SpineBlockManifestQuery) InsertSpineBlockManifest(
	spineBlockManifest *model.SpineBlockManifest,
) (str string, args []interface{}) {
	qryInsert := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
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

// GetSpineBlockManifestTimeInterval retrieve all spineBlockManifests within a time frame
// Note: it is used to get all entities that have expired between spine blocks
func (mbl *SpineBlockManifestQuery) GetSpineBlockManifestTimeInterval(fromTimestamp, toTimestamp int64) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE expiration_timestamp > %d AND expiration_timestamp <= %d "+
		"ORDER BY manifest_type, chain_type, manifest_reference_height",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), fromTimestamp, toTimestamp)
	return query
}

// GetManifestBySpineBlockHeight retrieve manifests of binded to a spineblock height
func (mbl *SpineBlockManifestQuery) GetManifestBySpineBlockHeight(spineBlockHeight uint32) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE manifest_spine_block_height = %d "+
		"ORDER BY manifest_type, chain_type, manifest_reference_height",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), spineBlockHeight)
	return query
}

func (mbl *SpineBlockManifestQuery) GetManifestsFromSpineBlockHeight(spineBlockHeight uint32) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE manifest_spine_block_height > %d ",
		strings.Join(mbl.Fields, ", "), mbl.getTableName(), spineBlockHeight)
	return query
}

func (mbl *SpineBlockManifestQuery) GetManifestsFromManifestReferenceHeightRange(fromHeight, toHeight uint32) (qry string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE manifest_reference_height >= ? AND manifest_reference_height <= ? ORDER BY manifest_reference_height",
			strings.Join(mbl.Fields, ", "),
			mbl.getTableName(),
		),
		[]interface{}{
			fromHeight,
			toHeight,
		}
}

// ExtractModel extract the model struct fields to the order of SpineBlockManifestQuery.Fields
func (mbl *SpineBlockManifestQuery) ExtractModel(mb *model.SpineBlockManifest) []interface{} {
	return []interface{}{
		mb.ID,
		mb.FullFileHash,
		mb.FileChunkHashes,
		mb.ManifestReferenceHeight,
		mb.ManifestSpineBlockHeight,
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
			&mb.ManifestReferenceHeight,
			&mb.ManifestSpineBlockHeight,
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
		&mb.ManifestReferenceHeight,
		&mb.ManifestSpineBlockHeight,
		&mb.ChainType,
		&mb.SpineBlockManifestType,
		&mb.ExpirationTimestamp,
	)
	if err != nil {
		return err
	}
	return nil
}

// Rollback delete records `WHERE block_height > "height - constant.MinRollbackBlocks"`
// Note: we subtract constant.MinRollbackBlocks from height because that's the block height the snapshot is taken in respect of current
// block height
func (mbl *SpineBlockManifestQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE manifest_spine_block_height > ?", mbl.getTableName()),
			height,
		},
	}
}
