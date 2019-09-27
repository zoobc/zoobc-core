package query

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeRegistrationQueryInterface interface {
		InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{})
		UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{}
		GetNodeRegistrations(registrationHeight, size uint32) (str string)
		GetActiveNodeRegistrations() string
		GetNodeRegistrationByID(id int64) (str string, args []interface{})
		GetNodeRegistrationByNodePublicKey() string
		GetLastVersionedNodeRegistrationByPublicKey(nodePublicKey []byte, height uint32) (str string, args []interface{})
		GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{})
		GetNodeRegistrationsByHighestLockedBalance(limit uint32, queued bool) string
		GetNodeRegistrationsWithZeroScore(queued bool) string
		GetNodeRegistryAtHeight(height uint32) string
		ExtractModel(nr *model.NodeRegistration) []interface{}
		BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration
		BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) []*model.Blocksmith
	}

	NodeRegistrationQuery struct {
		Fields    []string
		TableName string
	}
)

func NewNodeRegistrationQuery() *NodeRegistrationQuery {
	return &NodeRegistrationQuery{
		Fields: []string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"queued",
			"latest",
			"height",
		},
		TableName: "node_registry",
	}
}

func (nr *NodeRegistrationQuery) getTableName() string {
	return nr.TableName
}

func (nr *NodeRegistrationQuery) InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nr.getTableName(),
		strings.Join(nr.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nr.Fields)-1)),
	), nr.ExtractModel(nodeRegistration)
}

// UpdateNodeRegistration returns a slice of two queries.
// 1st update all old noderegistration versions' latest field to 0
// 2nd insert a new version of the noderegisration with updated data
func (nr *NodeRegistrationQuery) UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE ID = ?", nr.getTableName())
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nr.getTableName(),
		strings.Join(nr.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nr.Fields)-1)),
	)

	queries = append(queries,
		append([]interface{}{qryUpdate}, nodeRegistration.NodeID),
		append([]interface{}{qryInsert}, nr.ExtractModel(nodeRegistration)...),
	)

	return queries
}

// GetNodeRegistrations returns query string to get multiple node registrations
func (nr *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND latest=1 LIMIT %d",
		strings.Join(nr.Fields, ", "), nr.getTableName(), registrationHeight, size)
}

// GetActiveNodeRegistrations
func (nr *NodeRegistrationQuery) GetActiveNodeRegistrations() string {
	return fmt.Sprintf("SELECT nr.node_public_key AS node_public_key, ps.score AS participation_score FROM %s AS nr "+
		"INNER JOIN %s AS ps ON nr.id = ps.node_id WHERE nr.latest = 1 AND nr.queued = 0",
		nr.getTableName(), NewParticipationScoreQuery().TableName)
}

// GetNodeRegistrationByID returns query string to get Node Registration by node public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByID(id int64) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName()), []interface{}{id}
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByNodePublicKey() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName())
}

// GetLastVersionedNodeRegistrationByPublicKey returns query string to get Node Registration
// by node public key at a given height (versioned)
func (nr *NodeRegistrationQuery) GetLastVersionedNodeRegistrationByPublicKey(nodePublicKey []byte,
	height uint32) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1",
		strings.Join(nr.Fields, ", "), nr.getTableName()), []interface{}{nodePublicKey, height}
}

// GetNodeRegistrationByAccountID returns query string to get Node Registration by account public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_address = '?' AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName()), []interface{}{accountAddress}
}

// GetNodeRegistrationsByHighestLockedBalance returns query string to get the list of Node Registrations with highest locked balance
// queued or not queued
func (nr *NodeRegistrationQuery) GetNodeRegistrationsByHighestLockedBalance(limit uint32, queued bool) string {
	var (
		queuedInt int
	)
	if queued {
		queuedInt = 1
	} else {
		queuedInt = 0
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE locked_balance > 0 AND queued = %d AND latest=1 ORDER BY locked_balance DESC LIMIT %d",
		strings.Join(nr.Fields, ", "), nr.getTableName(), queuedInt, limit)
}

// GetNodeRegistrationsWithZeroScore returns query string to get the list of Node Registrations with zero participation score
func (nr *NodeRegistrationQuery) GetNodeRegistrationsWithZeroScore(queued bool) string {
	var (
		queuedInt int
	)
	nrTable := nr.getTableName()
	nrTableAlias := "A"
	psTable := NewParticipationScoreQuery().getTableName()
	psTableAlias := "B"
	nrTableFields := make([]string, 0)
	for _, field := range nr.Fields {
		nrTableFields = append(nrTableFields, nrTableAlias+"."+field)
	}
	if queued {
		queuedInt = 1
	} else {
		queuedInt = 0
	}

	return fmt.Sprintf("SELECT %s FROM "+nrTable+" as "+nrTableAlias+" "+
		"INNER JOIN "+psTable+" as "+psTableAlias+" ON "+nrTableAlias+".id = "+psTableAlias+".node_id "+
		"WHERE "+psTableAlias+".score = 0 "+
		"AND "+nrTableAlias+".latest=1 "+
		"AND "+nrTableAlias+".queued=%d "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(nrTableFields, ", "),
		queuedInt)
}

// GetNodeRegistryAtHeight returns unique latest node registry record at specific height
func (nr *NodeRegistrationQuery) GetNodeRegistryAtHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s, max(height) AS max_height FROM %s where height <= %d AND queued == 0 GROUP BY id ORDER BY height DESC",
		strings.Join(nr.Fields, ", "), nr.getTableName(), height)
}

// ExtractModel extract the model struct fields to the order of NodeRegistrationQuery.Fields
func (*NodeRegistrationQuery) ExtractModel(nr *model.NodeRegistration) []interface{} {
	return []interface{}{
		nr.NodeID,
		nr.NodePublicKey,
		nr.AccountAddress,
		nr.RegistrationHeight,
		nr.NodeAddress,
		nr.LockedBalance,
		nr.Queued,
		nr.Latest,
		nr.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (nr *NodeRegistrationQuery) BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration {
	columns, _ := rows.Columns()
	var (
		ignoredAggregateColumns, basicFieldsReceiver []interface{}
		dumpString                                   string
	)
	for i := 0; i < len(columns)-len(nr.Fields); i++ {
		ignoredAggregateColumns = append(ignoredAggregateColumns, &dumpString)
	}

	for rows.Next() {
		var nr model.NodeRegistration
		basicFieldsReceiver = append(
			basicFieldsReceiver,
			&nr.NodeID,
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&nr.NodeAddress,
			&nr.LockedBalance,
			&nr.Queued,
			&nr.Latest,
			&nr.Height,
		)
		basicFieldsReceiver = append(basicFieldsReceiver, ignoredAggregateColumns...)
		_ = rows.Scan(basicFieldsReceiver...)
		nodeRegistrations = append(nodeRegistrations, &nr)
	}
	return nodeRegistrations
}

func (*NodeRegistrationQuery) BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) []*model.Blocksmith {
	for rows.Next() {
		var (
			blocksmith  model.Blocksmith
			scoreString string
		)
		_ = rows.Scan(
			&blocksmith.NodePublicKey,
			&scoreString,
		)
		blocksmith.Score, _ = new(big.Int).SetString(scoreString, 10)
		blocksmiths = append(blocksmiths, &blocksmith)
	}
	return blocksmiths
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (nr *NodeRegistrationQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", nr.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE (height || '_' || id) IN (
				SELECT (MAX(height) || '_' || id) as con
				FROM %s
				WHERE latest = 0
				GROUP BY id
			)`,
				nr.TableName,
				nr.TableName,
			),
			1,
		},
	}
}
