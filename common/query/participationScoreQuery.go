package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	ParticipationScoreQueryInterface interface {
		InsertParticipationScore(participationScore *model.ParticipationScore) (str string, args []interface{})
		InsertParticipationScores(scores []*model.ParticipationScore) (str string, args []interface{})
		UpdateParticipationScore(
			nodeID, score int64,
			blockHeight uint32,
		) [][]interface{}
		GetParticipationScoreByNodeID(id int64) (str string, args []interface{})
		GetParticipationScoreByAccountAddress(accountAddress string) (str string)
		GetParticipationScoreByNodePublicKey(nodePublicKey []byte) (str string, args []interface{})
		GetParticipationScoresByBlockHeightRange(
			fromBlockHeight, toBlockHeight uint32) (str string, args []interface{})
		Scan(participationScore *model.ParticipationScore, row *sql.Row) error
		ExtractModel(ps *model.ParticipationScore) []interface{}
		BuildModel(participationScores []*model.ParticipationScore, rows *sql.Rows) ([]*model.ParticipationScore, error)
	}

	ParticipationScoreQuery struct {
		Fields    []string
		TableName string
	}
)

func NewParticipationScoreQuery() *ParticipationScoreQuery {
	return &ParticipationScoreQuery{
		Fields: []string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		TableName: "participation_score",
	}
}

func (ps *ParticipationScoreQuery) getTableName() string {
	return ps.TableName
}

func (ps *ParticipationScoreQuery) InsertParticipationScore(participationScore *model.ParticipationScore) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		ps.getTableName(),
		strings.Join(ps.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(ps.Fields)-1)),
	), ps.ExtractModel(participationScore)
}

// InsertParticipationScores represents query builder to insert multiple record in single query
func (ps *ParticipationScoreQuery) InsertParticipationScores(scores []*model.ParticipationScore) (str string, args []interface{}) {
	if len(scores) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			ps.getTableName(),
			strings.Join(ps.Fields, ", "),
		)
		for k, score := range scores {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(ps.Fields)-1),
			)
			if k < len(scores)-1 {
				str += ","
			}
			args = append(args, ps.ExtractModel(score)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (ps *ParticipationScoreQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	participationScores, ok := payload.([]*model.ParticipationScore)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+ps.TableName)
	}
	if len(participationScores) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(ps.Fields), len(participationScores))
		for i := 0; i < rounds; i++ {
			qry, args := ps.InsertParticipationScores(participationScores[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := ps.InsertParticipationScores(participationScores[len(participationScores)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (ps *ParticipationScoreQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (node_id, height) NOT IN "+
				"(select t2.node_id, max(t2.height) from %s t2 group by t2.node_id)",
			ps.getTableName(), ps.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (node_id, height) IN "+
				"(select t2.node_id, max(t2.height) from %s t2 group by t2.node_id)",
			ps.getTableName(), ps.getTableName()),
	}
}

func (ps *ParticipationScoreQuery) UpdateParticipationScore(
	nodeID, score int64,
	blockHeight uint32,
) [][]interface{} {
	var (
		queries            [][]interface{}
		updateVersionQuery string
	)
	// update or insert new participation_score row
	// note: the participation score passed to this functions must already be = last recorded part score + increment (or decrement)
	updateScoreQuery := fmt.Sprintf("INSERT INTO %s (node_id, score, height, latest) "+
		"VALUES(%d, %d, %d, 1) "+
		"ON CONFLICT(node_id, height) DO UPDATE SET (score) = %d",
		ps.getTableName(), nodeID, score, blockHeight, score,
	)
	queries = append(queries,
		[]interface{}{
			updateScoreQuery,
		},
	)
	if blockHeight != 0 {
		// set previous version record to latest = false
		updateVersionQuery = fmt.Sprintf("UPDATE %s SET latest = false WHERE node_id = %d AND height != %d AND latest = true",
			ps.getTableName(), nodeID, blockHeight)
		queries = append(queries,
			[]interface{}{
				updateVersionQuery,
			},
		)
	}
	return queries
}

// GetParticipationScoreByNodeID returns query string to get participation score by node id
func (ps *ParticipationScoreQuery) GetParticipationScoreByNodeID(id int64) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_id = ? AND latest=1",
		strings.Join(ps.Fields, ", "), ps.getTableName()), []interface{}{id}
}

func (ps *ParticipationScoreQuery) GetParticipationScoreByAccountAddress(accountAddress string) (str string) {
	psTable := ps.getTableName()
	psTableAlias := "A"
	nrTable := NewNodeRegistrationQuery().getTableName()
	nrTableAlias := "B"
	psTableFields := make([]string, 0)
	for _, field := range ps.Fields {
		psTableFields = append(psTableFields, psTableAlias+"."+field)
	}

	return fmt.Sprintf("SELECT %s FROM "+psTable+" as "+psTableAlias+" "+
		"INNER JOIN "+nrTable+" as "+nrTableAlias+" ON "+psTableAlias+".node_id = "+nrTableAlias+".id "+
		"WHERE "+nrTableAlias+".account_address='%s' "+
		"AND "+nrTableAlias+".latest=1 "+
		"AND "+nrTableAlias+".registration_status=%d "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(psTableFields, ", "),
		accountAddress, uint32(model.NodeRegistrationState_NodeRegistered))
}

func (ps *ParticipationScoreQuery) GetParticipationScoreByNodePublicKey(nodePublicKey []byte) (str string, args []interface{}) {
	psTable := ps.getTableName()
	psTableAlias := "A"
	nrTable := NewNodeRegistrationQuery().getTableName()
	nrTableAlias := "B"
	psTableFields := make([]string, 0)
	for _, field := range ps.Fields {
		psTableFields = append(psTableFields, psTableAlias+"."+field)
	}

	return fmt.Sprintf("SELECT %s FROM "+psTable+" as "+psTableAlias+" "+
		"INNER JOIN "+nrTable+" as "+nrTableAlias+" ON "+psTableAlias+".node_id = "+nrTableAlias+".id "+
		"WHERE "+nrTableAlias+".node_public_key=? "+
		"AND "+nrTableAlias+".latest=1 "+
		"AND "+nrTableAlias+".registration_status=%d "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(psTableFields, ", "),
		uint32(model.NodeRegistrationState_NodeRegistered),
	), []interface{}{nodePublicKey}
}

func (ps *ParticipationScoreQuery) GetParticipationScoresByBlockHeightRange(
	fromBlockHeight, toBlockHeight uint32,
) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height BETWEEN ? AND ? ORDER BY height ASC",
		strings.Join(ps.Fields, ", "), ps.getTableName())
	return query, []interface{}{
		fromBlockHeight, toBlockHeight,
	}
}

// ExtractModel extract the model struct fields to the order of ParticipationScoreQuery.Fields
func (*ParticipationScoreQuery) ExtractModel(ps *model.ParticipationScore) []interface{} {
	return []interface{}{
		ps.NodeID,
		ps.Score,
		ps.Latest,
		ps.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*ParticipationScoreQuery) BuildModel(
	participationScores []*model.ParticipationScore,
	rows *sql.Rows,
) ([]*model.ParticipationScore, error) {
	for rows.Next() {
		var (
			ps  model.ParticipationScore
			err error
		)
		err = rows.Scan(
			&ps.NodeID,
			&ps.Score,
			&ps.Latest,
			&ps.Height,
		)
		if err != nil {
			return nil, err
		}
		participationScores = append(participationScores, &ps)
	}
	return participationScores, nil
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (ps *ParticipationScoreQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", ps.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (node_id, height) IN (
				SELECT t2.node_id, MAX(t2.height)
				FROM %s as t2
				GROUP BY t2.node_id
			)`,
				ps.TableName,
				ps.TableName,
			),
			1, 0,
		},
	}
}

func (*ParticipationScoreQuery) Scan(ps *model.ParticipationScore, row *sql.Row) error {
	err := row.Scan(
		&ps.NodeID,
		&ps.Score,
		&ps.Latest,
		&ps.Height,
	)
	return err
}

func (ps *ParticipationScoreQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(""+
		"SELECT %s FROM %s WHERE (node_id, height) IN (SELECT t2.node_id, MAX(t2.height) FROM %s as t2 "+
		"WHERE t2.height >= %d AND t2.height <= %d AND t2.height != 0 GROUP BY t2.node_id ) ORDER by height",
		strings.Join(ps.Fields, ","),
		ps.getTableName(),
		ps.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (ps *ParticipationScoreQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		ps.TableName, fromHeight, toHeight)
}
