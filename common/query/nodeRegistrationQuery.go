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

func (bq *NodeRegistrationQuery) getTableName() string {
	return bq.TableName
}

func (bq *NodeRegistrationQuery) InsertNodeRegistration() string {
	var value = ":" + bq.Fields[0]
	for _, field := range bq.Fields[1:] {
		value += (", :" + field)

	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		bq.getTableName(), strings.Join(bq.Fields, ", "), value)
	return query
}

// GetNodeRegistrationByNodePublicKey returns query string to get Node Registration by node public key
func (bq *NodeRegistrationQuery) GetNodeRegistrationNodeByPublicKey(nodePublicKey []byte) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), nodePublicKey)
}

// GetNodeRegistrationByAccountPublicKey returns query string to get Node Registration by account public key
func (bq *NodeRegistrationQuery) GetNodeRegistrationByAccountPublicKey(accountPublicKey []byte) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", strings.Join(bq.Fields, ", "), bq.getTableName(), accountPublicKey)
}

// GetNodeRegistrations returns query string to get multiple node registrations
func (bq *NodeRegistrationQuery) GetNodeRegistrations(registrationHeight, size uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d LIMIT %d", strings.Join(bq.Fields, ", "), bq.getTableName(), registrationHeight, size)
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
