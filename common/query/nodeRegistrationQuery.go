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
		GetNodeRegistrations(registrationHeight, size uint32)
		GetNodeRegistrationNodeByPublicKey(nodePublicKey []byte)
		GetNodeRegistrationByAccountPublicKey(accountPublicKey []byte)
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
		Fields:    []string{"node_public_key", "account_id", "registration_height", "node_address", "locked_balance", "latest", "height"},
		TableName: "node_registry",
	}
}

func (nr *NodeRegistrationQuery) getTableName() string {
	return nr.TableName
}

func (nr *NodeRegistrationQuery) InsertNodeRegistration(nodeRegistration *model.NodeRegistration) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		nr.TableName,
		strings.Join(nr.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(nr.Fields)-1)),
	), nr.ExtractModel(nodeRegistration)
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationNodeByPublicKey(nodePublicKey []byte) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = %d AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName(), nodePublicKey)
}

// GetNodeRegistrationByAccountPublicKey returns query string to get Node Registration by account public key
func (nr *NodeRegistrationQuery) GetNodeRegistrationByAccountID(accountPublicKey []byte) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_id = %d AND latest=1",
		strings.Join(nr.Fields, ", "), nr.getTableName(), accountPublicKey)
}

// GetNodeRegistrations returns query string to get multiple node registrations
func (nr *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND latest=1 LIMIT %d",
		strings.Join(nr.Fields, ", "), nr.getTableName(), registrationHeight, size)
}

// ExtractModel extract the model struct fields to the order of NodeRegistrationQuery.Fields
func (*NodeRegistrationQuery) ExtractModel(nr *model.NodeRegistration) []interface{} {
	return []interface{}{
		nr.NodePublicKey,
		nr.AccountId,
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
			&nr.NodePublicKey,
			&nr.AccountId,
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
