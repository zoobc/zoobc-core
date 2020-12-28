// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
		ConfirmNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) [][]interface{}
		DeleteNodeAddressInfoByNodeID(nodeID int64, addressStatuses []model.NodeAddressStatus) (str string, args []interface{})
		GetNodeAddressInfoByNodeIDs(nodeIDs []int64, addressStatuses []model.NodeAddressStatus) string
		GetNodeAddressInfoByNodeID(nodeID int64, addressStatuses []model.NodeAddressStatus) string
		GetNodeAddressInfo() string
		GetNodeAddressInfoByStatus(addressStatuses []model.NodeAddressStatus) string
		GetNodeAddressInfoByAddressPort(
			address string, port uint32,
			addressStatuses []model.NodeAddressStatus,
		) (str string, args []interface{})
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
			"status",
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
			" signature = ?,"+
			" status = ?"+
			" WHERE node_id = ? AND status = ?", paq.getTableName())
	// move NodeID at the bottom of the values array
	values := append(paq.ExtractModel(peerAddress)[1:], peerAddress.NodeID, peerAddress.Status)
	queries = append(queries,
		append([]interface{}{qryUpdate}, values...),
	)
	return queries
}

// ConfirmNodeAddressInfo returns a slice of queries/query parameters containing the insert/delete queries to be executed.
func (paq *NodeAddressInfoQuery) ConfirmNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryDeleteDuplicateAddress := fmt.Sprintf(
		"DELETE FROM %s WHERE address = ? AND port = ? AND node_id != ?",
		paq.getTableName(),
	)
	qryDeleteOld := fmt.Sprintf(
		"DELETE FROM %s WHERE node_id = ? AND status != ?",
		paq.getTableName(),
	)

	nodeAddressInfo.Status = model.NodeAddressStatus_NodeAddressConfirmed
	qryInsertReplace := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		paq.getTableName(),
		strings.Join(paq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(paq.Fields)-1)),
	)
	queries = append(queries,
		append([]interface{}{qryDeleteDuplicateAddress}, nodeAddressInfo.Address, nodeAddressInfo.Port, nodeAddressInfo.NodeID),
		append([]interface{}{qryDeleteOld}, nodeAddressInfo.GetNodeID(), uint32(model.NodeAddressStatus_NodeAddressPending)),
		append([]interface{}{qryInsertReplace}, paq.ExtractModel(nodeAddressInfo)...),
	)
	return queries
}

// DeleteNodeAddressInfoByNodeID returns the query string and parameters to be executed to delete a peerAddress record
func (paq *NodeAddressInfoQuery) DeleteNodeAddressInfoByNodeID(
	nodeID int64,
	addressStatuses []model.NodeAddressStatus,
) (str string, args []interface{}) {
	c := make([]string, len(addressStatuses))
	for i, v := range addressStatuses {
		c[i] = strconv.Itoa(int(v))
	}
	addrStatusesStr := strings.Join(c, ", ")
	return fmt.Sprintf(
		"DELETE FROM %s WHERE node_id = ? AND status IN (%s)",
		paq.getTableName(),
		addrStatusesStr,
	), []interface{}{nodeID}
}

// GetNodeAddressInfoByIDs returns query string to get peerAddress by node IDs and address statuses
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByNodeIDs(nodeIDs []int64, addressStatuses []model.NodeAddressStatus) string {
	b := make([]string, len(nodeIDs))
	for i, v := range nodeIDs {
		b[i] = strconv.Itoa(int(v))
	}
	nodeIDsStr := strings.Join(b, ", ")
	c := make([]string, len(addressStatuses))
	for i, v := range addressStatuses {
		c[i] = strconv.Itoa(int(v))
	}
	addrStatusesStr := strings.Join(c, ", ")
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_id IN (%s) AND status IN (%s) ORDER BY node_id, status ASC",
		strings.Join(paq.Fields, ", "), paq.getTableName(), nodeIDsStr, addrStatusesStr)
}

// GetNodeAddressInfoByID returns query string to get peerAddress by node ID and address statuses
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByNodeID(nodeID int64, addressStatuses []model.NodeAddressStatus) string {
	c := make([]string, len(addressStatuses))
	for i, v := range addressStatuses {
		c[i] = strconv.Itoa(int(v))
	}
	addrStatusesStr := strings.Join(c, ", ")
	return fmt.Sprintf("SELECT %s FROM %s WHERE node_id = %d AND status IN (%s) ORDER BY status ASC",
		strings.Join(paq.Fields, ", "), paq.getTableName(), nodeID, addrStatusesStr)
}

// GetNodeAddressInfo returns query string to get contents of node_address_info table
func (paq *NodeAddressInfoQuery) GetNodeAddressInfo() string {
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY node_id, status ASC",
		strings.Join(paq.Fields, ", "), paq.getTableName())
}

// GetNodeAddressInfoByStatus returns query string to get contents of node_address_info table
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByStatus(addressStatuses []model.NodeAddressStatus) string {
	c := make([]string, len(addressStatuses))
	for i, v := range addressStatuses {
		c[i] = strconv.Itoa(int(v))
	}
	addrStatusesStr := strings.Join(c, ", ")
	return fmt.Sprintf("SELECT %s FROM %s WHERE status IN (%s) ORDER BY node_id, status ASC",
		strings.Join(paq.Fields, ", "), paq.getTableName(), addrStatusesStr)
}

// GetNodeAddressInfoByAddressPort returns query string to get peerAddress by node ID
func (paq *NodeAddressInfoQuery) GetNodeAddressInfoByAddressPort(
	address string, port uint32,
	addressStatuses []model.NodeAddressStatus,
) (str string, args []interface{}) {
	c := make([]string, len(addressStatuses))
	for i, v := range addressStatuses {
		c[i] = strconv.Itoa(int(v))
	}
	addrStatusesStr := strings.Join(c, ", ")
	return fmt.Sprintf("SELECT %s FROM %s WHERE address = ? AND port = ? AND status IN (%s) ORDER BY status ASC",
		strings.Join(paq.Fields, ", "), paq.getTableName(), addrStatusesStr), []interface{}{address, port}
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
		pa.Status,
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
			&pa.Status,
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
		&pa.Status,
	)
	if err != nil {
		return err
	}
	return nil
}
