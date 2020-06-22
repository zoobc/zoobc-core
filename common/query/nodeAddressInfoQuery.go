package query

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeAddressInfoQueryInterface interface {
		InsertNodeAddressInfo(peerAddress *model.NodeAddressInfo) (str string, args []interface{})
		UpdateNodeAddressInfo(peerAddress *model.NodeAddressInfo) [][]interface{}
		DeleteNodeAddressInfoByNodeID(nodeID int64) (str string, args []interface{})
		GetNodeAddressInfoByNodeIDs(nodeIDs []int64) string
		GetNodeAddressInfoByAddressPort(address string, port uint32) (str string, args []interface{})
		ExtractModel(pa *model.NodeAddressInfo) []interface{}
		BuildModel(peerAddresss []*model.NodeAddressInfo, rows *sql.Rows) ([]*model.NodeAddressInfo, error)
		Scan(pa *model.NodeAddressInfo, row *sql.Row) error
	}

	NodeAddressInfoQuery struct {
		Fields    []string
		TableName string
	}
)

func NewNodeAddressInfoQuery() *NodeAddressInfoQuery {
	return &NodeAddressInfoQuery{
		Fields: []string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
		},
		TableName: "node_address_info",
	}
}

func (paq *NodeAddressInfoQuery) getTableName() string {
	return paq.TableName
}

// InsertNodeAddressInfo inserts a new peer address into DB. if an old ip/port peer is found with different nodeId,
// replace the old entry with the new one.
func (paq *NodeAddressInfoQuery) InsertNodeAddressInfo(peerAddress *model.NodeAddressInfo) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		paq.getTableName(),
		strings.Join(paq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(paq.Fields)-1)),
	), paq.ExtractModel(peerAddress)
}

// UpdateNodeAddressInfo returns a slice of queries/query parameters containing the update query to be executed.
func (paq *NodeAddressInfoQuery) UpdateNodeAddressInfo(peerAddress *model.NodeAddressInfo) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryUpdate := fmt.Sprintf(
		"UPDATE %s SET"+
			" address = ?,"+
			" port = ?,"+
			" block_height = ?,"+
			" block_hash = ?,"+
			" signature = ?"+
			" WHERE node_id = ?", paq.getTableName())
	// move NodeID at the bottom of the values array
	values := append(paq.ExtractModel(peerAddress)[1:], peerAddress.NodeID)
	queries = append(queries,
		append([]interface{}{qryUpdate}, values...),
	)
	return queries
}

// DeleteNodeAddressInfoByNodeID returns the query string and parameters to be executed to delete a peerAddress record
func (paq *NodeAddressInfoQuery) DeleteNodeAddressInfoByNodeID(nodeID int64) (str string, args []interface{}) {
	return fmt.Sprintf(
		"DELETE FROM %s WHERE node_id = ?",
		paq.getTableName(),
	), []interface{}{nodeID}
}

// GetNodeAddressInfoByID returns query string to get peerAddress by node ID
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByNodeIDs(nodeIDs []int64) string {
	b := make([]string, len(nodeIDs))
	for i, v := range nodeIDs {
		b[i] = strconv.Itoa(int(v))
	}
	nodeIDsStr := strings.Join(b, ", ")
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_id IN (%s)",
		strings.Join(paq.Fields, ", "), paq.getTableName(), nodeIDsStr)
}

// GetNodeAddressInfoByAddressPort returns query string to get peerAddress by node ID
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByAddressPort(address string, port uint32) (str string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_id = ? AND port = ? LIMIT 1",
		strings.Join(paq.Fields, ", "), paq.getTableName()), []interface{}{address, port}
}

// ExtractModel extract the model struct fields to the order of NodeAddressInfoQuery.Fields
func (paq *NodeAddressInfoQuery) ExtractModel(pa *model.NodeAddressInfo) []interface{} {
	return []interface{}{
		pa.NodeID,
		pa.Address,
		pa.Port,
		pa.BlockHeight,
		pa.BlockHash,
		pa.Signature,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (paq *NodeAddressInfoQuery) BuildModel(
	nodeAddressesInfo []*model.NodeAddressInfo,
	rows *sql.Rows,
) ([]*model.NodeAddressInfo, error) {
	for rows.Next() {
		var pa model.NodeAddressInfo
		err := rows.Scan(
			&pa.NodeID,
			&pa.Address,
			&pa.Port,
			&pa.BlockHeight,
			&pa.BlockHash,
			&pa.Signature,
		)
		if err != nil {
			return nil, err
		}
		nodeAddressesInfo = append(nodeAddressesInfo, &pa)
	}
	return nodeAddressesInfo, nil
}

// Scan represents `sql.Scan`
func (paq *NodeAddressInfoQuery) Scan(pa *model.NodeAddressInfo, row *sql.Row) error {
	err := row.Scan(
		&pa.NodeID,
		&pa.Address,
		&pa.Port,
		&pa.BlockHeight,
		&pa.BlockHash,
		&pa.Signature,
	)
	if err != nil {
		return err
	}
	return nil
}
