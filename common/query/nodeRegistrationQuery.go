package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeRegistrationQueryInterface interface {
		InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{})
		UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) (str []string, args []interface{})
		GetNodeRegistrations(registrationHeight, size uint32) (str string)
		GetNodeRegistrationByID(id int64) (str string, args []interface{})
		GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (str string, args []interface{})
		GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{})
		ExtractModel(nr *model.NodeRegistration) []interface{}
		BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration
	}

	NodeRegistrationQuery struct {
		Fields    []string
		TableName string
	}
)

func NewNodeRegistrationQuery() *NodeRegistrationQuery {
	return &NodeRegistrationQuery{
		Fields: []string{"id", "node_public_key", "account_address", "registration_height", "node_address", "locked_balance", "queued",
			"latest", "height"},
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
func (nr *NodeRegistrationQuery) UpdateNodeRegistration(nodeRegistration *model.NodeRegistration) (str []string, args []interface{}) {
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE ID = %d", nr.getTableName(), nodeRegistration.NodeID)
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nr.getTableName(),
		strings.Join(nr.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nr.Fields)-1)),
	)
	return []string{qryUpdate, qryInsert}, nr.ExtractModel(nodeRegistration)
}

// GetNodeRegistrations returns query string to get multiple node registrations
func (nr *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND latest=1 LIMIT %d",
		strings.Join(nr.Fields, ", "), nr.getTableName(), registrationHeight, size)
}

// GetNodeRegistrationByID returns query string to get Node Registration by node public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByID(id int64) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName()), []interface{}{id}
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName()), []interface{}{nodePublicKey}
}

// GetNodeRegistrationByAccountID returns query string to get Node Registration by account public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByAccountAddress(accountAddress string) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_address = %s AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName(), accountAddress), []interface{}{accountAddress}
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
func (*NodeRegistrationQuery) BuildModel(nodeRegistrations []*model.NodeRegistration, rows *sql.Rows) []*model.NodeRegistration {
	for rows.Next() {
		var nr model.NodeRegistration
		_ = rows.Scan(
			&nr.NodeID,
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&nr.NodeAddress,
			&nr.LockedBalance,
			&nr.Queued,
			&nr.Latest,
			&nr.Height)
		nodeRegistrations = append(nodeRegistrations, &nr)
	}
	return nodeRegistrations
}
