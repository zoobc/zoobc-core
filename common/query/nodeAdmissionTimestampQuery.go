package query

import (
	"database/sql"
	"fmt"
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
		GetFields() []string
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
func (natq *NodeAdmissionTimestampQuery) GetFields() []string {
	return natq.Fields
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
	// do not delete genesis block
	if fromHeight == 0 {
		fromHeight++
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		natq.getTableName(), fromHeight, toHeight)
}
