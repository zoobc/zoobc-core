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
	"github.com/zoobc/zoobc-core/common/blocker"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SkippedBlocksmithQueryInterface interface {
		GetNumberOfSkippedBlocksmithsByBlockHeight(blockHeight uint32) (qStr string)
		GetSkippedBlocksmithsByBlockHeight(blockHeight uint32) (qStr string)
		InsertSkippedBlocksmith(skippedBlocksmith *model.SkippedBlocksmith) (qStr string, args []interface{})
		InsertSkippedBlocksmiths(skippedBlockSmiths []*model.SkippedBlocksmith) (str string, args []interface{})
		ExtractModel(skippedBlocksmith *model.SkippedBlocksmith) []interface{}
		BuildModel(skippedBlocksmiths []*model.SkippedBlocksmith, rows *sql.Rows) ([]*model.SkippedBlocksmith, error)
		Scan(skippedBlocksmith *model.SkippedBlocksmith, rows *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}

	SkippedBlocksmithQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewSkippedBlocksmithQuery will create a new SkippedBlocksmithQuery instance
func NewSkippedBlocksmithQuery(ct chaintype.ChainType) *SkippedBlocksmithQuery {
	var tableName = "skipped_blocksmith"
	if chaintype.IsSpineChain(ct) {
		tableName = "spine_skipped_blocksmith"
	}
	return &SkippedBlocksmithQuery{
		Fields: []string{
			"blocksmith_public_key",
			"pop_change",
			"block_height",
			"blocksmith_index",
		},
		TableName: tableName,
	}
}

func (sbq *SkippedBlocksmithQuery) getTableName() string {
	return sbq.TableName
}

func (sbq *SkippedBlocksmithQuery) GetSkippedBlocksmithsByBlockHeight(blockHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height = %d",
		strings.Join(sbq.Fields, ", "),
		sbq.getTableName(),
		blockHeight,
	)
}

func (sbq *SkippedBlocksmithQuery) GetNumberOfSkippedBlocksmithsByBlockHeight(blockHeight uint32) string {
	return fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE block_height = %d",
		sbq.getTableName(),
		blockHeight,
	)
}

func (sbq *SkippedBlocksmithQuery) InsertSkippedBlocksmith(
	skippedBlocksmith *model.SkippedBlocksmith,
) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			sbq.getTableName(),
			strings.Join(sbq.Fields, ", "),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(sbq.Fields)-1)),
		),
		sbq.ExtractModel(skippedBlocksmith)
}

// InsertSkippedBlocksmiths represents query builder to insert multiple record in single query
func (sbq *SkippedBlocksmithQuery) InsertSkippedBlocksmiths(skippedBlocksmiths []*model.SkippedBlocksmith) (str string, args []interface{}) {
	if len(skippedBlocksmiths) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			sbq.getTableName(),
			strings.Join(sbq.Fields, ", "),
		)
		for k, skippedBlocksmith := range skippedBlocksmiths {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(sbq.Fields)-1),
			)
			if k < len(skippedBlocksmiths)-1 {
				str += ","
			}
			args = append(args, sbq.ExtractModel(skippedBlocksmith)...)
		}
	}
	return str, args

}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (sbq *SkippedBlocksmithQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	skippedBlocksmiths, ok := payload.([]*model.SkippedBlocksmith)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+sbq.TableName)
	}
	if len(skippedBlocksmiths) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(sbq.Fields), len(skippedBlocksmiths))
		for i := 0; i < rounds; i++ {
			qry, args := sbq.InsertSkippedBlocksmiths(skippedBlocksmiths[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := sbq.InsertSkippedBlocksmiths(skippedBlocksmiths[len(skippedBlocksmiths)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (sbq *SkippedBlocksmithQuery) RecalibrateVersionedTable() []string {
	return []string{}
}

func (*SkippedBlocksmithQuery) ExtractModel(skippedModel *model.SkippedBlocksmith) []interface{} {
	return []interface{}{
		&skippedModel.BlocksmithPublicKey,
		&skippedModel.POPChange,
		&skippedModel.BlockHeight,
		&skippedModel.BlocksmithIndex,
	}
}

func (*SkippedBlocksmithQuery) BuildModel(
	skippedBlocksmiths []*model.SkippedBlocksmith,
	rows *sql.Rows,
) ([]*model.SkippedBlocksmith, error) {
	for rows.Next() {
		var (
			skippedBlocksmith model.SkippedBlocksmith
			err               error
		)
		err = rows.Scan(
			&skippedBlocksmith.BlocksmithPublicKey,
			&skippedBlocksmith.POPChange,
			&skippedBlocksmith.BlockHeight,
			&skippedBlocksmith.BlocksmithIndex,
		)
		if err != nil {
			return nil, err
		}
		skippedBlocksmiths = append(skippedBlocksmiths, &skippedBlocksmith)
	}
	return skippedBlocksmiths, nil
}

func (*SkippedBlocksmithQuery) Scan(skippedBlocksmith *model.SkippedBlocksmith, row *sql.Row) error {
	err := row.Scan(
		&skippedBlocksmith.BlocksmithPublicKey,
		&skippedBlocksmith.POPChange,
		&skippedBlocksmith.BlockHeight,
		&skippedBlocksmith.BlocksmithIndex,
	)
	return err
}

func (sbq *SkippedBlocksmithQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", sbq.getTableName()),
			height,
		},
	}
}

func (sbq *SkippedBlocksmithQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0 ORDER BY block_height",
		strings.Join(sbq.Fields, ", "),
		sbq.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (sbq *SkippedBlocksmithQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		sbq.getTableName(), fromHeight, toHeight)
}
