package query

import (
	"database/sql"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	"github.com/zoobc/zoobc-core/common/constant"
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
		GetNodeRegistrationsByHighestLockedBalance(limit uint32, registrationStatus uint32) string
		GetNodeRegistrationsWithZeroScore(registrationStatus uint32) string
		GetNodeRegistryAtHeight(height uint32) string
		ExtractModel(nr *model.NodeRegistration) []interface{}
		BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration
		BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) []*model.Blocksmith
		BuildNodeAddress(fullNodeAddress string) *model.NodeAddress
		ExtractNodeAddress(nodeAddress *model.NodeAddress) string
		Scan(nr *model.NodeRegistration, row *sql.Row) error
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
			"registration_status",
			"latest",
			"height",
		},
		TableName: "node_registry",
	}
}

func (nrq *NodeRegistrationQuery) getTableName() string {
	return nrq.TableName
}

func (nrq *NodeRegistrationQuery) InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nrq.getTableName(),
		strings.Join(nrq.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nrq.Fields)-1)),
	), nrq.ExtractModel(nodeRegistration)
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

// GetNodeRegistrations returns query string to get multiple node registrations
func (nrq *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND latest=1 LIMIT %d",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), registrationHeight, size)
}

// GetActiveNodeRegistrations
func (nrq *NodeRegistrationQuery) GetActiveNodeRegistrations() string {
	return fmt.Sprintf("SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score FROM %s AS nr "+
		"INNER JOIN %s AS ps ON nr.id = ps.node_id WHERE "+
		"account_address = %s AND nr.latest = 1 AND nr.registration_status = 0 AND ps.score > 0 AND ps.latest = 1",
		nrq.getTableName(), NewParticipationScoreQuery().TableName, constant.DeletedNodeAccountAddress)
}

// GetNodeRegistrationByID returns query string to get Node Registration by node public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByID(id int64) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND latest=1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{id}
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByNodePublicKey() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND latest=1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName())
}

// GetLastVersionedNodeRegistrationByPublicKey returns query string to get Node Registration
// by node public key at a given height (versioned)
func (nrq *NodeRegistrationQuery) GetLastVersionedNodeRegistrationByPublicKey(nodePublicKey []byte,
	height uint32) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{nodePublicKey, height}
}

// GetNodeRegistrationByAccountID returns query string to get Node Registration by account public key
func (nrq *NodeRegistrationQuery) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_address = ? AND latest=1",
		strings.Join(nrq.Fields, ", "), nrq.getTableName()), []interface{}{accountAddress}
}

// GetNodeRegistrationsByHighestLockedBalance returns query string to get the list of Node Registrations with highest locked balance
// registration_status or not registration_status
func (nrq *NodeRegistrationQuery) GetNodeRegistrationsByHighestLockedBalance(limit uint32, registrationStatus uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE locked_balance > 0 AND registration_status = %d AND latest=1 ORDER BY locked_balance DESC LIMIT %d",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), registrationStatus, limit)
}

// GetNodeRegistrationsWithZeroScore returns query string to get the list of Node Registrations with zero participation score
func (nrq *NodeRegistrationQuery) GetNodeRegistrationsWithZeroScore(registrationStatus uint32) string {
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
		"WHERE "+psTableAlias+".score = 0 "+
		"AND "+nrTableAlias+".latest=1 "+
		"AND "+nrTableAlias+".registration_status=%d "+
		"AND "+psTableAlias+".latest=1",
		strings.Join(nrTableFields, ", "),
		registrationStatus)
}

// GetNodeRegistryAtHeight returns unique latest node registry record at specific height
func (nrq *NodeRegistrationQuery) GetNodeRegistryAtHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s, max(height) AS max_height FROM %s where height <= %d AND registration_status = 0 GROUP BY id ORDER BY height DESC",
		strings.Join(nrq.Fields, ", "), nrq.getTableName(), height)
}

// ExtractModel extract the model struct fields to the order of NodeRegistrationQuery.Fields
func (nrq *NodeRegistrationQuery) ExtractModel(tx *model.NodeRegistration) []interface{} {

	return []interface{}{
		tx.NodeID,
		tx.NodePublicKey,
		tx.AccountAddress,
		tx.RegistrationHeight,
		nrq.ExtractNodeAddress(tx.GetNodeAddress()),
		tx.LockedBalance,
		tx.RegistrationStatus,
		tx.Latest,
		tx.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (nrq *NodeRegistrationQuery) BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration {

	var (
		ignoredAggregateColumns []interface{}
		dumpString              string
	)

	columns, _ := rows.Columns()
	for i := 0; i < len(columns)-len(nrq.Fields); i++ {
		ignoredAggregateColumns = append(ignoredAggregateColumns, &dumpString)
	}

	for rows.Next() {
		var (
			fullNodeAddress     string
			nr                  model.NodeRegistration
			basicFieldsReceiver []interface{}
		)
		basicFieldsReceiver = append(
			basicFieldsReceiver,
			&nr.NodeID,
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&fullNodeAddress,
			&nr.LockedBalance,
			&nr.RegistrationStatus,
			&nr.Latest,
			&nr.Height,
		)
		basicFieldsReceiver = append(basicFieldsReceiver, ignoredAggregateColumns...)
		_ = rows.Scan(basicFieldsReceiver...)

		nr.NodeAddress = nrq.BuildNodeAddress(fullNodeAddress)
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
			&blocksmith.NodeID,
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
func (nrq *NodeRegistrationQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", nrq.TableName),
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
				nrq.TableName,
				nrq.TableName,
			),
			1,
		},
	}
}

// BuildNodeAddress to build joining the NodeAddress.Address and NodeAddress.Port
func (*NodeRegistrationQuery) BuildNodeAddress(fullNodeAddress string) *model.NodeAddress {
	var (
		host, port string
		err        error
	)

	host, port, err = net.SplitHostPort(fullNodeAddress)
	if err != nil {
		host = fullNodeAddress
	}

	uintPort, _ := strconv.ParseUint(port, 0, 32)
	return &model.NodeAddress{
		Address: host,
		Port:    uint32(uintPort),
	}
}

// NodeAddressToString to build fully node address include port to NodeAddress struct
func (*NodeRegistrationQuery) ExtractNodeAddress(nodeAddress *model.NodeAddress) string {

	if nodeAddress == nil {
		return ""
	}

	if nodeAddress.GetPort() != 0 {
		return fmt.Sprintf("%s:%d", nodeAddress.GetAddress(), nodeAddress.GetPort())
	}

	return nodeAddress.GetAddress()
}

// Scan represents `sql.Scan`
func (nrq *NodeRegistrationQuery) Scan(nr *model.NodeRegistration, row *sql.Row) error {

	var (
		stringAddress string
		err           error
	)
	err = row.Scan(
		&nr.NodeID,
		&nr.NodePublicKey,
		&nr.AccountAddress,
		&nr.RegistrationHeight,
		&stringAddress,
		&nr.LockedBalance,
		&nr.RegistrationStatus,
		&nr.Latest,
		&nr.Height,
	)
	if err != nil {
		return err
	}
	nodeAddress := nrq.BuildNodeAddress(stringAddress)
	nr.NodeAddress = nodeAddress
	return nil
}
