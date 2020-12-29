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

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SpinePublicKeyQueryInterface interface {
		InsertSpinePublicKey(spinePublicKey *model.SpinePublicKey) [][]interface{}
		GetValidSpinePublicKeysByHeightInterval(fromHeight, toHeight uint32) string
		GetSpinePublicKeysByBlockHeight(height uint32) string
		ExtractModel(spk *model.SpinePublicKey) []interface{}
		BuildModel(spinePublicKeys []*model.SpinePublicKey, rows *sql.Rows) ([]*model.SpinePublicKey, error)
		BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) ([]*model.Blocksmith, error)
		Scan(spk *model.SpinePublicKey, row *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}

	SpinePublicKeyQuery struct {
		Fields    []string
		TableName string
	}
)

func NewSpinePublicKeyQuery() *SpinePublicKeyQuery {
	return &SpinePublicKeyQuery{
		Fields: []string{
			"node_public_key",
			"node_id",
			"public_key_action",
			"main_block_height",
			"latest",
			"height",
		},
		TableName: "spine_public_key",
	}
}

func (spkq *SpinePublicKeyQuery) getTableName() string {
	return spkq.TableName
}

// InsertSpinePublicKey
func (spkq *SpinePublicKeyQuery) InsertSpinePublicKey(spinePublicKey *model.SpinePublicKey) [][]interface{} {
	var (
		queries [][]interface{}
	)
	spinePublicKey.Latest = true
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE node_public_key = ?", spkq.getTableName())
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		spkq.getTableName(),
		strings.Join(spkq.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(spkq.Fields)-1)),
	)

	queries = append(queries,
		append([]interface{}{qryUpdate}, spinePublicKey.NodePublicKey),
		append([]interface{}{qryInsert}, spkq.ExtractModel(spinePublicKey)...),
	)

	return queries
}

func (spkq *SpinePublicKeyQuery) GetValidSpinePublicKeysByHeightInterval(fromHeight, toHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND height <= %d AND public_key_action=%d AND latest=1 ORDER BY height",
		strings.Join(spkq.Fields, ", "), spkq.getTableName(), fromHeight, toHeight, uint32(model.SpinePublicKeyAction_AddKey))
}

// GetSpinePublicKeysByBlockHeight returns query string to get Spine public keys for a given block
func (spkq *SpinePublicKeyQuery) GetSpinePublicKeysByBlockHeight(height uint32) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height = %d",
		strings.Join(spkq.Fields, ", "), spkq.getTableName(), height)
	return query
}

// ExtractModel extract the model struct fields to the order of SpinePublicKeyQuery.Fields
func (spkq *SpinePublicKeyQuery) ExtractModel(spk *model.SpinePublicKey) []interface{} {
	return []interface{}{
		spk.NodePublicKey,
		spk.NodeID,
		spk.PublicKeyAction,
		spk.MainBlockHeight,
		spk.Latest,
		spk.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (spkq *SpinePublicKeyQuery) BuildModel(
	spinePublicKeys []*model.SpinePublicKey,
	rows *sql.Rows,
) ([]*model.SpinePublicKey, error) {
	for rows.Next() {
		var (
			spk model.SpinePublicKey
			err error
		)
		err = rows.Scan(
			&spk.NodePublicKey,
			&spk.NodeID,
			&spk.PublicKeyAction,
			&spk.MainBlockHeight,
			&spk.Latest,
			&spk.Height,
		)
		if err != nil {
			return nil, err
		}
		spinePublicKeys = append(spinePublicKeys, &spk)
	}
	return spinePublicKeys, nil
}

func (spkq *SpinePublicKeyQuery) BuildBlocksmith(
	blocksmiths []*model.Blocksmith, rows *sql.Rows,
) ([]*model.Blocksmith, error) {
	for rows.Next() {
		var (
			blocksmith model.Blocksmith
			nodeStatus int64
			height     uint32
			latest     bool
			blockID    int64
		)
		err := rows.Scan(
			&blocksmith.NodePublicKey,
			&blocksmith.NodeID,
			&blockID,
			&nodeStatus,
			&latest,
			&height,
		)
		if err != nil {
			return nil, err
		}
		blocksmiths = append(blocksmiths, &blocksmith)
	}
	return blocksmiths, nil
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (spkq *SpinePublicKeyQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", spkq.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (node_public_key, height) IN (
				SELECT t2.node_public_key, MAX(t2.height)
				FROM %s as t2
				GROUP BY t2.node_public_key
			)`,
				spkq.TableName,
				spkq.TableName,
			),
			1, 0,
		},
	}
}

// Scan represents `sql.Scan`
func (spkq *SpinePublicKeyQuery) Scan(spk *model.SpinePublicKey, row *sql.Row) error {
	err := row.Scan(
		&spk.NodePublicKey,
		&spk.NodeID,
		&spk.PublicKeyAction,
		&spk.MainBlockHeight,
		&spk.Latest,
		&spk.Height,
	)
	if err != nil {
		return err
	}
	return nil
}
