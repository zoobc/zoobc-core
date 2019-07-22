package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeRegistrationQueryInterface interface {
		InsertNodeRegistration() string
		ExtractModel(nr *model.NodeRegistration) []interface{}
		GetNodeRegistrations(registrationHeight, size uint32)
		GetNodeRegistrationNodeByPublicKey(nodePublicKey []byte)
		GetNodeRegistrationByAccountPublicKey(accountPublicKey []byte)
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

func (nr *NodeRegistrationQuery) InsertNodeRegistration() string {
	var value = ":" + nr.Fields[0]
	for _, field := range nr.Fields[1:] {
		value += (", :" + field)

	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s); ",
		nr.getTableName(), strings.Join(nr.Fields, ", "), value)
	return query
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
		nr.Latest,
		nr.Height,
	}
}
