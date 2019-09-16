package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	ParticipationScoreQueryInterface interface {
		InsertParticipationScore(participationScore *model.ParticipationScore) (str string, args []interface{})
		UpdateParticipationScore(participationScore *model.ParticipationScore) (str []string, args []interface{})
		GetParticipationScoreByNodeID(id int64) (str string, args []interface{})
		GetParticipationScoreByAccountAddress(accountAddress string) (str string)
		GetParticipationScoreByNodePublicKey(nodePublicKey []byte) (str string, args []interface{})
		ExtractModel(ps *model.ParticipationScore) []interface{}
		BuildModel(participationScores []*model.ParticipationScore, rows *sql.Rows) []*model.ParticipationScore
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

// UpdateParticipationScore returns a slice of two queries.
// 1st update all old participation scores versions' latest field to 0
// 2nd insert a new version of the participation score with updated data
func (ps *ParticipationScoreQuery) UpdateParticipationScore(
	participationScore *model.ParticipationScore) (str []string, args []interface{}) {
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE node_id = %d", ps.getTableName(), participationScore.NodeID)
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		ps.getTableName(),
		strings.Join(ps.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(ps.Fields)-1)),
	)
	return []string{qryUpdate, qryInsert}, ps.ExtractModel(participationScore)
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
		"AND "+nrTableAlias+".queued=0 "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(psTableFields, ", "),
		accountAddress)
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
		"AND "+nrTableAlias+".queued=0 "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(psTableFields, ", "),
	), []interface{}{nodePublicKey}
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
func (*ParticipationScoreQuery) BuildModel(participationScores []*model.ParticipationScore, rows *sql.Rows) []*model.ParticipationScore {
	for rows.Next() {
		var ps model.ParticipationScore
		_ = rows.Scan(
			&ps.NodeID,
			&ps.Score,
			&ps.Latest,
			&ps.Height,
		)
		participationScores = append(participationScores, &ps)
	}
	return participationScores
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (ps *ParticipationScoreQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", ps.TableName),
			[]interface{}{height},
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE height || '_' || id) IN (
				SELECT (MAX(height) || '_' || id) as con
				FROM %s
				WHERE latest = 0
				GROUP BY id
			)`,
				ps.TableName,
				ps.TableName,
			),
			[]interface{}{1},
		},
	}
}
