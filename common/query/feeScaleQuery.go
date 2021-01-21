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
	FeeScaleQueryInterface interface {
		GetLatestFeeScale() string
		InsertFeeScale(feeScale *model.FeeScale) [][]interface{}
		InsertFeeScales(feeScales []*model.FeeScale) (qry string, args []interface{})
		ExtractModel(feeScale *model.FeeScale) []interface{}
		BuildModel(feeScales []*model.FeeScale, rows *sql.Rows) ([]*model.FeeScale, error)
		Scan(feeScale *model.FeeScale, row *sql.Row) error
	}

	FeeScaleQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewFeeScaleQuery returns FeeScaleQuery instance
func NewFeeScaleQuery() *FeeScaleQuery {
	return &FeeScaleQuery{
		Fields: []string{
			"fee_scale",
			"block_height",
			"latest",
		},
		TableName: "fee_scale",
	}
}

func (fsq *FeeScaleQuery) getTableName() string {
	return fsq.TableName
}

// GetLatestFeeScale return the latest fee scale
func (fsq *FeeScaleQuery) GetLatestFeeScale() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE latest = true",
		strings.Join(fsq.Fields, ", "), fsq.getTableName())
}

// InsertFeeScale insert new fee scale record
func (fsq *FeeScaleQuery) InsertFeeScale(feeScale *model.FeeScale) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s SET latest = ? WHERE latest = ? AND block_height IN (SELECT MAX(t2.block_height) FROM %s as t2)",
				fsq.getTableName(), fsq.getTableName(),
			),
			0,
			1,
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					fsq.getTableName(),
					strings.Join(fsq.Fields, ", "),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(fsq.Fields)-1)),
				),
			},
			fsq.ExtractModel(feeScale)...,
		),
	}
}

func (fsq *FeeScaleQuery) InsertFeeScales(feeScales []*model.FeeScale) (str string, args []interface{}) {
	if len(feeScales) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			fsq.getTableName(),
			strings.Join(fsq.Fields, ", "),
		)
		for k, feeScale := range feeScales {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(fsq.Fields)-1),
			)
			if k < len(feeScales)-1 {
				str += ","
			}
			args = append(args, fsq.ExtractModel(feeScale)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (fsq *FeeScaleQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	feeScales, ok := payload.([]*model.FeeScale)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+fsq.TableName)
	}
	if len(feeScales) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(fsq.Fields), len(feeScales))
		for i := 0; i < rounds; i++ {
			qry, args := fsq.InsertFeeScales(feeScales[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := fsq.InsertFeeScales(feeScales[len(feeScales)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (fsq *FeeScaleQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND block_height NOT IN "+
				"(select max(t2.block_height) from %s t2)",
			fsq.getTableName(), fsq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND block_height IN "+
				"(select max(t2.block_height) from %s t2)",
			fsq.getTableName(), fsq.getTableName()),
	}
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*FeeScaleQuery) ExtractModel(feeScale *model.FeeScale) []interface{} {
	return []interface{}{
		feeScale.FeeScale,
		feeScale.BlockHeight,
		feeScale.Latest,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*FeeScaleQuery) BuildModel(
	feeScales []*model.FeeScale,
	rows *sql.Rows,
) ([]*model.FeeScale, error) {
	for rows.Next() {
		var (
			feeScale model.FeeScale
			err      error
		)
		err = rows.Scan(
			&feeScale.FeeScale,
			&feeScale.BlockHeight,
			&feeScale.Latest,
		)
		if err != nil {
			return nil, err
		}
		feeScales = append(feeScales, &feeScale)
	}
	return feeScales, nil
}

// Scan similar with `sql.Scan`
func (*FeeScaleQuery) Scan(feeScale *model.FeeScale, row *sql.Row) error {
	err := row.Scan(
		&feeScale.FeeScale,
		&feeScale.BlockHeight,
		&feeScale.Latest,
	)
	return err
}

// Rollback delete records `WHERE height > "block_height"
func (fsq *FeeScaleQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", fsq.getTableName()),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND block_height IN (
				SELECT MAX(t2.block_height)
				FROM %s as t2
			)`,
				fsq.TableName,
				fsq.TableName,
			),
			1,
			0,
		},
	}
}

// SelectDataForSnapshot select only the block at snapshot block_height (fromHeight is unused)
func (fsq *FeeScaleQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height != 0 AND block_height >= %d AND block_height <= %d`,
		strings.Join(fsq.Fields, ","), fsq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fsq *FeeScaleQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		fsq.getTableName(), fromHeight, toHeight)
}
