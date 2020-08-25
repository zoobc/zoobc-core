package query

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeRegistrationQueryInterface interface {
		InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{})
		InsertNodeRegistrations(nodeRegistrations []*model.NodeRegistration) (str string, args []interface{})
		UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{}
		ClearDeletedNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{}
		GetNodeRegistrations(registrationHeight, size uint32) (str string)
		GetNodeRegistrationsByBlockTimestampInterval(fromTimestamp, toTimestamp int64) string
		GetActiveNodeRegistrations() string
		GetActiveNodeRegistrationsByHeight(height uint32) string
		GetActiveNodeRegistrationsWithNodeAddress() string
		GetNodeRegistrationByID(id int64) (str string, args []interface{})
		GetNodeRegistrationByNodePublicKey() string
		GetLastVersionedNodeRegistrationByPublicKey(nodePublicKey []byte, height uint32) (str string, args []interface{})
		GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress(nodePublicKey []byte, height uint32) (str string, args []interface{})
		GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{})
		GetNodeRegistrationsByHighestLockedBalance(limit uint32, registrationStatus model.NodeRegistrationState) string
		GetNodeRegistrationsWithZeroScore(registrationStatus model.NodeRegistrationState) string
		GetNodeRegistryAtHeight(height uint32) string
		GetNodeRegistryAtHeightWithNodeAddress(height uint32) string
		GetPendingNodeRegistrations(limit uint32) string
		ExtractModel(nr *model.NodeRegistration) []interface{}
		BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) ([]*model.NodeRegistration, error)
		BuildModelWithAddressInfo(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) ([]*model.NodeRegistration, error)
		BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) ([]*model.Blocksmith, error)
		Scan(nr *model.NodeRegistration, row *sql.Row) error
		ScanWithNodeAddress(nr *model.NodeRegistration, row *sql.Row) error
	}

	NodeRegistrationQuery struct {
		Fields                  []string
		JoinedAddressInfoFields []string
		TableName               string
	}
)

func NewNodeRegistrationQuery() *NodeRegistrationQuery {
	return &NodeRegistrationQuery{
		Fields: []string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		JoinedAddressInfoFields: []string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
			"%s.address AS node_address",
			"%s.port AS node_address_port",
			"%s.status AS node_address_status",
		},
		TableName: "node_registry",
	}
}

func (nrq *NodeRegistrationQuery) getTableName() string {
	return nrq.TableName
}

// InsertNodeRegistration inserts a new node registration into DB
func (nrq *NodeRegistrationQuery) InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nrq.getTableName(),
		strings.Join(nrq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(nrq.Fields)-1)),
	), nrq.ExtractModel(nodeRegistration)
}

// InsertNodeRegistrations represents query builder to insert multiple record in single query
func (nrq *NodeRegistrationQuery) InsertNodeRegistrations(nodeRegistrations []*model.NodeRegistration) (str string, args []interface{}) {
	if len(nodeRegistrations) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			nrq.getTableName(),
			strings.Join(nrq.Fields, ", "),
		)
		for k, nodeReg := range nodeRegistrations {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(nrq.Fields)-1),
			)
			if k < len(nodeRegistrations)-1 {
				str += ","
			}
			args = append(args, nrq.ExtractModel(nodeReg)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (nrq *NodeRegistrationQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	nodeRegistrations, ok := payload.([]*model.NodeRegistration)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+nrq.TableName)
	}
	if len(nodeRegistrations) > 0 {

		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(nrq.Fields), len(nodeRegistrations))
		for i := 0; i < rounds; i++ {
			qry, args := nrq.InsertNodeRegistrations(nodeRegistrations[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := nrq.InsertNodeRegistrations(nodeRegistrations[len(nodeRegistrations)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (nrq *NodeRegistrationQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (id, height) NOT IN "+
				"(select t2.id, max(height) from %s t2 group by t2.id)",
			nrq.getTableName(), nrq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (id, height) IN "+
				"(select t2.id, max(height) from %s t2 group by t2.id)",
			nrq.getTableName(), nrq.getTableName()),
	}
}

// UpdateNodeRegistration returns a slice of two queries.
// 1st update all old noderegistration versions' latest field to 0
// 2nd insert a new version of the noderegisration with updated data
func (nrq *NodeRegistrationQuery) UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE ID = ?", nrq.getTableName())
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nrq.getTableName(),
		strings.Join(nrq.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nrq.Fields)-1)),
	)

	queries = append(queries,
		append([]interface{}{qryUpdate}, nodeRegistration.NodeID),
		append([]interface{}{qryInsert}, nrq.ExtractModel(nodeRegistration)...),
	)

	return queries
}

// ClearDeletedNodeRegistration used when registering a new node from and account that has previously deleted another one
// to avoid having multiple node registrations with same account id and latest = true
func (nrq *NodeRegistrationQuery) ClearDeletedNodeRegistration(nodeRegistration *model.NodeRegistration) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE ID = ? AND registration_status = 2", nrq.getTableName())

	queries = append(queries,
		append([]interface{}{qryUpdate}, nodeRegistration.NodeID),
	)

	return queries
}

// GetNodeRegistrations returns query string to get multiple node registrations
func (nrq *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND latest=1 LIMIT %d",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), registrationHeight, size)
}

// GetNodeRegistrationsByBlockTimestampInterval returns query string to get multiple node registrations
// Note: toTimestamp (limit) is excluded from selection to avoid selecting duplicates
func (nrq *NodeRegistrationQuery) GetNodeRegistrationsByBlockTimestampInterval(fromTimestamp, toTimestamp int64) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE "+
		"height >= (SELECT MIN(height) FROM main_block AS mb1 WHERE mb1.timestamp >= %d) AND "+
		"height <= (SELECT MAX(height) FROM main_block AS mb2 WHERE mb2.timestamp < %d) AND "+
		"registration_status != %d AND latest=1 ORDER BY height",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), fromTimestamp, toTimestamp, uint32(model.NodeRegistrationState_NodeQueued))
}

func (nrq *NodeRegistrationQuery) GetActiveNodeRegistrationsByHeight(height uint32) string {
	return fmt.Sprintf("SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score,"+
		" max(nr.height) AS max_height FROM %s AS nr "+
		"INNER JOIN %s AS ps ON nr.id = ps.node_id WHERE nr.height <= %d AND "+
		"nr.registration_status = %d AND nr.latest = 1 AND ps.score > 0 AND ps.latest = 1 GROUP BY nr.id",
		nrq.getTableName(), NewParticipationScoreQuery().TableName, height, uint32(model.NodeRegistrationState_NodeRegistered))
}

// GetNodeRegistrationByID returns query string to get Node Registration by node public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByID(id int64) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND latest=1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{id}
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByNodePublicKey() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName())
}

// GetLastVersionedNodeRegistrationByPublicKey returns query string to get Node Registration
// by node public key at a given height (versioned)
func (nrq *NodeRegistrationQuery) GetLastVersionedNodeRegistrationByPublicKey(nodePublicKey []byte,
	height uint32) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{nodePublicKey, height}
}

// GetLastVersionedNodeRegistrationByPublicKey returns query string to get Node Registration
// by node public key at a given height (versioned)
func (nrq *NodeRegistrationQuery) GetLastVersionedNodeRegistrationByPublicKeyWithNodeAddress(nodePublicKey []byte,
	height uint32) (str string, args []interface{}) {
	joinedFields := strings.Join(nrq.JoinedAddressInfoFields, ", ")
	joinedFieldsStr := fmt.Sprintf(joinedFields, "t2", "t2", "t2")
	return fmt.Sprintf("SELECT %s FROM %s LEFT JOIN %s AS t2 ON id = t2.node_id "+
			"WHERE (node_public_key = ? OR t2.node_id IS NULL) AND height <= ? ORDER BY height DESC LIMIT 1",
			joinedFieldsStr, nrq.getTableName(), NewNodeAddressInfoQuery().TableName),
		[]interface{}{nodePublicKey, height}
}

// GetNodeRegistrationByAccountAddress returns query string to get Node Registration by account public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{accountAddress}
}

// GetNodeRegistrationsByHighestLockedBalance returns query string to get the list of Node Registrations with highest locked balance
// registration_status or not registration_status
func (nrq *NodeRegistrationQuery) GetNodeRegistrationsByHighestLockedBalance(limit uint32,
	registrationStatus model.NodeRegistrationState) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE locked_balance > 0 AND registration_status = %d AND latest=1 "+
		"ORDER BY locked_balance DESC LIMIT %d",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), registrationStatus, limit)
}

// GetNodeRegistrationsWithZeroScore returns query string to get the list of Node Registrations with zero participation score
func (nrq *NodeRegistrationQuery) GetNodeRegistrationsWithZeroScore(registrationStatus model.NodeRegistrationState) string {
	nrTable := nrq.getTableName()
	nrTableAlias := "A"
	psTable := NewParticipationScoreQuery().getTableName()
	psTableAlias := "B"
	nrTableFields := make([]string, 0)
	for _, field := range nrq.Fields {
		nrTableFields = append(nrTableFields, nrTableAlias+"."+field)
	}

	return fmt.Sprintf("SELECT %s FROM "+nrTable+" as "+nrTableAlias+" "+
		"INNER JOIN "+psTable+" as "+psTableAlias+" ON "+nrTableAlias+".id = "+psTableAlias+".node_id "+
		"WHERE "+psTableAlias+".score <= 0 "+
		"AND "+nrTableAlias+".latest=1 "+
		"AND "+nrTableAlias+".registration_status=%d "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(nrTableFields, ", "),
		registrationStatus)
}

// GetNodeRegistryAtHeight returns unique latest node registry record at specific height
func (nrq *NodeRegistrationQuery) GetNodeRegistryAtHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s where registration_status = 0 AND (id,height) in (SELECT id,MAX(height) "+
		"FROM %s WHERE height <= %d GROUP BY id) ORDER BY height DESC",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), nrq.getTableName(), height)
}

// GetRegisteredNodes the full node registry
func (nrq *NodeRegistrationQuery) GetActiveNodeRegistrations() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE registration_status = 0 AND latest = 1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName())
}

// GetNodeRegistryAtHeightWithNodeAddress returns unique latest node registry record at specific height, with peer addresses too.
// Note: this query is to be used during node scrambling. Only nodes that have a peerAddress will be selected
func (nrq *NodeRegistrationQuery) GetNodeRegistryAtHeightWithNodeAddress(height uint32) string {
	joinedFields := strings.Join(nrq.JoinedAddressInfoFields, ", ")
	joinedFieldsStr := fmt.Sprintf(joinedFields, "t2", "t2", "t2")
	return fmt.Sprintf("SELECT %s FROM %s INNER JOIN %s AS t2 ON id = t2.node_id "+
		"WHERE registration_status = 0 AND (id,height) in (SELECT t1.id,MAX(t1.height) "+
		"FROM %s AS t1 WHERE t1.height <= %d GROUP BY t1.id) "+
		"ORDER BY id, t2.status",
		joinedFieldsStr, nrq.getTableName(), NewNodeAddressInfoQuery().TableName, nrq.getTableName(), height)
}

// GetNodeRegistryAtHeightWithNodeAddress returns unique latest node registry record at specific height, with peer addresses too.
// Note: this query is to be used during node scrambling. Only nodes that have a peerAddress will be selected
func (nrq *NodeRegistrationQuery) GetActiveNodeRegistrationsWithNodeAddress() string {
	joinedFields := strings.Join(nrq.JoinedAddressInfoFields, ", ")
	joinedFieldsStr := fmt.Sprintf(joinedFields, "t2", "t2", "t2")
	return fmt.Sprintf("SELECT %s FROM %s INNER JOIN %s AS t2 ON id = t2.node_id "+
		"WHERE registration_status = 0 "+
		"ORDER BY height DESC",
		joinedFieldsStr, nrq.getTableName(), NewNodeAddressInfoQuery().TableName)
}

// GetPendingNodeRegistrations returns pending node registrations
func (nrq *NodeRegistrationQuery) GetPendingNodeRegistrations(limit uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE registration_status=1 AND latest=1 ORDER BY locked_balance DESC LIMIT %d",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), limit)
}

// ExtractModel extract the model struct fields to the order of NodeRegistrationQuery.Fields
func (nrq *NodeRegistrationQuery) ExtractModel(tx *model.NodeRegistration) []interface{} {
	return []interface{}{
		tx.NodeID,
		tx.NodePublicKey,
		tx.AccountAddress,
		tx.RegistrationHeight,
		tx.LockedBalance,
		tx.RegistrationStatus,
		tx.Latest,
		tx.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (nrq *NodeRegistrationQuery) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {

	var (
		ignoredAggregateColumns []interface{}
		dumpString              string
	)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(columns)-len(nrq.Fields); i++ {
		ignoredAggregateColumns = append(ignoredAggregateColumns, &dumpString)
	}

	for rows.Next() {
		var (
			nr                  model.NodeRegistration
			basicFieldsReceiver []interface{}
		)
		basicFieldsReceiver = append(
			basicFieldsReceiver,
			&nr.NodeID,
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&nr.LockedBalance,
			&nr.RegistrationStatus,
			&nr.Latest,
			&nr.Height,
		)
		basicFieldsReceiver = append(basicFieldsReceiver, ignoredAggregateColumns...)
		err := rows.Scan(basicFieldsReceiver...)
		if err != nil {
			return nil, err
		}
		nodeRegistrations = append(nodeRegistrations, &nr)
	}
	return nodeRegistrations, nil
}

// BuildModelWithAddressInfo will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
// note: this is to be used with queries that join node_address_info table
func (nrq *NodeRegistrationQuery) BuildModelWithAddressInfo(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	for rows.Next() {
		var (
			nr                  model.NodeRegistration
			nrAddress           string
			nrAddressPort       uint32
			nrAddressInfoStatus model.NodeAddressStatus
		)
		err := rows.Scan(
			&nr.NodeID,
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&nr.LockedBalance,
			&nr.RegistrationStatus,
			&nr.Latest,
			&nr.Height,
			&nrAddress,
			&nrAddressPort,
			&nrAddressInfoStatus,
		)
		if err != nil {
			return nil, err
		}
		nr.NodeAddressInfo = &model.NodeAddressInfo{
			Address: nrAddress,
			Port:    nrAddressPort,
			Status:  nrAddressInfoStatus,
		}
		nodeRegistrations = append(nodeRegistrations, &nr)
	}
	return nodeRegistrations, nil
}

func (*NodeRegistrationQuery) BuildBlocksmith(
	blocksmiths []*model.Blocksmith, rows *sql.Rows,
) ([]*model.Blocksmith, error) {
	for rows.Next() {
		var (
			blocksmith  model.Blocksmith
			scoreString string
			maxHeight   uint32
		)
		err := rows.Scan(
			&blocksmith.NodeID,
			&blocksmith.NodePublicKey,
			&scoreString,
			&maxHeight,
		)
		if err != nil {
			return nil, err
		}
		blocksmith.Score, _ = new(big.Int).SetString(scoreString, 10)
		blocksmiths = append(blocksmiths, &blocksmith)
	}
	return blocksmiths, nil
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (nrq *NodeRegistrationQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", nrq.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (id, height) IN (
				SELECT t2.id, MAX(t2.height)
				FROM %s as t2
				GROUP BY t2.id
			)`,
				nrq.TableName,
				nrq.TableName,
			),
			1, 0,
		},
	}
}

// Scan represents `sql.Scan`
func (nrq *NodeRegistrationQuery) Scan(nr *model.NodeRegistration, row *sql.Row) error {

	err := row.Scan(
		&nr.NodeID,
		&nr.NodePublicKey,
		&nr.AccountAddress,
		&nr.RegistrationHeight,
		&nr.LockedBalance,
		&nr.RegistrationStatus,
		&nr.Latest,
		&nr.Height,
	)
	if err != nil {
		return err
	}
	return nil
}

// ScanWithNodeAddress represents `sql.Scan` and includes address info
func (nrq *NodeRegistrationQuery) ScanWithNodeAddress(nr *model.NodeRegistration, row *sql.Row) error {

	var (
		err       error
		nrAddress string
		nrPort    uint32
		nrStatus  model.NodeAddressStatus
	)
	err = row.Scan(
		&nr.NodeID,
		&nr.NodePublicKey,
		&nr.AccountAddress,
		&nr.RegistrationHeight,
		&nr.LockedBalance,
		&nr.RegistrationStatus,
		&nr.Latest,
		&nr.Height,
		&nrAddress,
		&nrPort,
		&nrStatus,
	)
	nrAddressInfo := &model.NodeAddressInfo{
		Address: nrAddress,
		Port:    nrPort,
		Status:  nrStatus,
	}
	nr.NodeAddressInfo = nrAddressInfo
	if err != nil {
		return err
	}
	return nil
}

// SelectDataForSnapshot this query selects only node registry latest state from height 0 to 'fromHeight' (
// excluded) and all records from 'fromHeight' to 'toHeight',
// removing from first selection records that have duplicate ids with second second selection.
// This way we make sure only one version of every id has 'latest' field set to true
func (nrq *NodeRegistrationQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	if fromHeight > 0 {
		return fmt.Sprintf("SELECT %s FROM %s WHERE (id, height) IN (SELECT t2.id, "+
			"MAX(t2.height) FROM %s as t2 WHERE t2.height > 0 AND t2.height < %d GROUP BY t2.id) "+
			"UNION ALL SELECT %s FROM %s WHERE height >= %d AND height <= %d "+
			"ORDER BY height, id",
			strings.Join(nrq.Fields, ","), nrq.getTableName(), nrq.getTableName(), fromHeight,
			strings.Join(nrq.Fields, ","), nrq.getTableName(), fromHeight, toHeight)
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND height <= %d AND height != 0 ORDER BY height, id",
		strings.Join(nrq.Fields, ","), nrq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (nrq *NodeRegistrationQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		nrq.TableName, fromHeight, toHeight)
}
