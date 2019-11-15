package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	ParticipationScoreQueryInterface interface {
		AddParticipationScore(score int64, causedFields map[string]interface{}) [][]interface{}
		GetParticipationScoreByNodeID(id int64) (str string, args []interface{})
		GetParticipationScoreByAccountAddress(accountAddress string) (str string)
		GetParticipationScoreByNodePublicKey(nodePublicKey []byte) (str string, args []interface{})
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

func (ps *ParticipationScoreQuery) AddParticipationScore(score int64, causedFields map[string]interface{}) [][]interface{} {
	var (
		queries            [][]interface{}
		updateVersionQuery string
	)
	// update or insert new participation_score row
	updateScoreQuery := fmt.Sprintf("INSERT INTO %s AS ps (node_id, score, latest, height) "+
		"VALUES(?, %d, 1, ?) ON CONFLICT(ps.node_id, ps.height) "+
		"DO UPDATE SET (score, height, latest) = (SELECT "+
		"ps1.score + %d, ps1.height, 1 FROM %s AS ps1 WHERE ps1.node_id = ? AND ps1.latest = 1)",
		ps.getTableName(), score, score, ps.getTableName())
	queries = append(queries,
		[]interface{}{
			updateScoreQuery, causedFields["node_id"], causedFields["height"], causedFields["node_id"],
		},
	)

	if causedFields["height"].(uint32) != 0 {
		// set previous version record to latest = false
		updateVersionQuery = fmt.Sprintf("UPDATE %s SET latest = false WHERE node_id = ? AND height != ? AND latest = true",
			ps.getTableName())
		queries = append(queries,
			[]interface{}{
				updateVersionQuery, causedFields["node_id"], causedFields["height"],
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
			WHERE latest = ? AND (height || '_' || node_id) IN (
				SELECT (MAX(height) || '_' || node_id) as con
				FROM %s
				GROUP BY node_id
			)`,
				ps.TableName,
				ps.TableName,
			),
			1, 0,
		},
	}
}
