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

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// NodeAdmissionTimestampQueryInterface methods must have
	NodeAdmissionTimestampQueryInterface interface {
		GetNextNodeAdmision() string
		InsertNextNodeAdmission(nodeAdmissionTimestamp *model.NodeAdmissionTimestamp) [][]interface{}
		InsertNextNodeAdmissions(
			nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp,
		) (str string, args []interface{})
		ExtractModel(nextNodeAdmission *model.NodeAdmissionTimestamp) []interface{}
		BuildModel(
			nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp,
			rows *sql.Rows,
		) ([]*model.NodeAdmissionTimestamp, error)
		Scan(nextNodeAdmission *model.NodeAdmissionTimestamp, row *sql.Row) error
	}
	// NodeAdmissionTimestampQuery fields must have
	NodeAdmissionTimestampQuery struct {
		Fields    []string
		TableName string
	}
)

// NewNodeAdmissionTimestampQuery returns NewNodeAdmissionTimestampQuery instance
func NewNodeAdmissionTimestampQuery() *NodeAdmissionTimestampQuery {
	return &NodeAdmissionTimestampQuery{
		Fields: []string{
			"timestamp",
			"block_height",
			"latest",
		},
		TableName: "node_admission_timestamp",
	}
}

func (natq *NodeAdmissionTimestampQuery) getTableName() string {
	return natq.TableName
}

// GetNextNodeAdmision return the next node admission timestamp
func (natq *NodeAdmissionTimestampQuery) GetNextNodeAdmision() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE latest = true  ORDER BY block_height DESC",
		strings.Join(natq.Fields, ", "), natq.getTableName())
}

// InsertNextNodeAdmission insert next timestamp node admission
func (natq *NodeAdmissionTimestampQuery) InsertNextNodeAdmission(
	nodeAdmissionTimestamp *model.NodeAdmissionTimestamp,
) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(`
				UPDATE %s SET latest = ? 
				WHERE latest = ? AND block_height IN (
					SELECT MAX(t2.block_height) FROM %s as t2
				)`,
				natq.getTableName(), natq.getTableName(),
			),
			0,
			1,
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					natq.getTableName(),
					strings.Join(natq.Fields, ", "),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(natq.Fields)-1)),
				),
			},
			natq.ExtractModel(nodeAdmissionTimestamp)...,
		),
	}
}

// InsertNextNodeAdmissions represents query builder to insert multiple record in single query
// note: this query only use for inserting snapshot (applaying some lastest version of this table).
func (natq *NodeAdmissionTimestampQuery) InsertNextNodeAdmissions(
	nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp,
) (str string, args []interface{}) {
	if len(nodeAdmissionTimestamps) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			natq.getTableName(),
			strings.Join(natq.Fields, ", "),
		)
		for k, nodeReg := range nodeAdmissionTimestamps {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(natq.Fields)-1),
			)
			if k < len(nodeAdmissionTimestamps)-1 {
				str += ","
			}
			args = append(args, natq.ExtractModel(nodeReg)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (natq *NodeAdmissionTimestampQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	timestamps, ok := payload.([]*model.NodeAdmissionTimestamp)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+natq.TableName)
	}
	if len(timestamps) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(natq.Fields), len(timestamps))
		for i := 0; i < rounds; i++ {
			qry, args := natq.InsertNextNodeAdmissions(timestamps[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := natq.InsertNextNodeAdmissions(timestamps[len(timestamps)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (natq *NodeAdmissionTimestampQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND block_height NOT IN "+
				"(select max(t2.block_height) from %s t2)",
			natq.getTableName(), natq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND block_height IN "+
				"(select max(t2.block_height) from %s t2)",
			natq.getTableName(), natq.getTableName()),
	}
}

// ExtractModel extract the model struct fields to the order of NodeAdmissionTimestampQuery.Fields
func (*NodeAdmissionTimestampQuery) ExtractModel(
	nextNodeAdmission *model.NodeAdmissionTimestamp,
) []interface{} {
	return []interface{}{
		nextNodeAdmission.Timestamp,
		nextNodeAdmission.BlockHeight,
		nextNodeAdmission.Latest,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*NodeAdmissionTimestampQuery) BuildModel(
	nodeAdmissionTimestamps []*model.NodeAdmissionTimestamp,
	rows *sql.Rows,
) ([]*model.NodeAdmissionTimestamp, error) {
	for rows.Next() {
		var (
			nodeAdmissionTimestamp model.NodeAdmissionTimestamp
			err                    error
		)
		err = rows.Scan(
			&nodeAdmissionTimestamp.Timestamp,
			&nodeAdmissionTimestamp.BlockHeight,
			&nodeAdmissionTimestamp.Latest,
		)
		if err != nil {
			return nil, err
		}
		nodeAdmissionTimestamps = append(nodeAdmissionTimestamps, &nodeAdmissionTimestamp)
	}
	return nodeAdmissionTimestamps, nil
}

// Scan similar with `sql.Scan`
func (natq *NodeAdmissionTimestampQuery) Scan(
	nextNodeAdmission *model.NodeAdmissionTimestamp,
	row *sql.Row,
) error {
	err := row.Scan(
		&nextNodeAdmission.Timestamp,
		&nextNodeAdmission.BlockHeight,
		&nextNodeAdmission.Latest,
	)
	return err
}

// Rollback delete records `WHERE height > "block_height"
func (natq *NodeAdmissionTimestampQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", natq.getTableName()),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND block_height IN (
				SELECT MAX(t2.block_height)
				FROM %s as t2
			)`,
				natq.getTableName(),
				natq.getTableName(),
			),
			1,
			0,
		},
	}
}

// SelectDataForSnapshot select only the block at snapshot block_height (fromHeight is unused)
func (natq *NodeAdmissionTimestampQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		strings.Join(natq.Fields, ","), natq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (natq *NodeAdmissionTimestampQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {

	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		natq.getTableName(), fromHeight, toHeight)
}
