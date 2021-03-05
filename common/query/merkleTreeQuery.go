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
	"strings"
)

type (
	// MerkleTreeQueryInterface contain builder func for MerkleTree
	MerkleTreeQueryInterface interface {
		InsertMerkleTree(
			root, tree []byte, timestamp int64, blockHeight uint32) (qStr string, args []interface{})
		GetMerkleTreeByRoot(root []byte) (qStr string, args []interface{})
		SelectMerkleTreeForPublishedReceipts(
			height uint32,
		) string
		SelectMerkleTreeAtHeight(
			height uint32,
		) string
		GetLastMerkleRoot() (qStr string)
		PruneData(blockHeight, limit uint32) (string, []interface{})
		ScanTree(row *sql.Row) ([]byte, error)
		ScanRoot(row *sql.Row) ([]byte, error)
		BuildTree(row *sql.Rows) (map[string][]byte, error)
	}
	// MerkleTreeQuery fields and table name
	MerkleTreeQuery struct {
		Fields    []string
		TableName string
	}
)

// NewMerkleTreeQuery func to create new MerkleTreeInterface
func NewMerkleTreeQuery() *MerkleTreeQuery {
	return &MerkleTreeQuery{
		Fields: []string{
			"id",
			"block_height",
			"tree",
			"timestamp",
		},
		TableName: "merkle_tree",
	}

}

func (mrQ *MerkleTreeQuery) getTableName() string {
	return mrQ.TableName
}

// InsertMerkleTree func build insert Query for MerkleTree
func (mrQ *MerkleTreeQuery) InsertMerkleTree(
	root, tree []byte, timestamp int64, blockHeight uint32,
) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			mrQ.getTableName(),
			strings.Join(mrQ.Fields, ", "),
			fmt.Sprintf("?%s", strings.Repeat(",? ", len(mrQ.Fields)-1)),
		),
		[]interface{}{root, blockHeight, tree, timestamp}
}

// GetMerkleTreeByRoot is used to retrieve merkle table record, to check if the merkle root specified exist
func (mrQ *MerkleTreeQuery) GetMerkleTreeByRoot(root []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = ?",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(),
	), []interface{}{root}
}

func (mrQ *MerkleTreeQuery) GetLastMerkleRoot() (qStr string) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY timestamp DESC LIMIT 1",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName())
	return query
}

/*
SelectMerkleTreeAtHeight represents get merkle tree of block_height
*/
func (mrQ *MerkleTreeQuery) SelectMerkleTreeAtHeight(
	height uint32,
) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_height = %d",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(), height)
	return query
}

/*
SelectMerkleTreeForPublishedReceipts represents get merkle tree in range of block_height
and order by block_height ascending
test_expression >= low_expression AND test_expression <= high_expression
*/
func (mrQ *MerkleTreeQuery) SelectMerkleTreeForPublishedReceipts(
	height uint32,
) string {
	query := fmt.Sprintf("SELECT %s FROM %s AS mt WHERE EXISTS "+
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked) AND "+
		"block_height = %d",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(), height)
	return query
}

// PruneData represents query remove in range block_height with limit
func (mrQ *MerkleTreeQuery) PruneData(blockHeight, limit uint32) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"DELETE FROM %s WHERE block_height IN("+
				"SELECT block_height FROM %s WHERE block_height < ? "+
				"ORDER BY block_height ASC LIMIT ?)",
			mrQ.getTableName(),
			mrQ.getTableName(),
		), []interface{}{
			blockHeight,
			limit,
		}
}

func (mrQ *MerkleTreeQuery) ScanTree(row *sql.Row) ([]byte, error) {
	var (
		root, tree  []byte
		timestamp   int64
		blockHeight uint32
	)
	err := row.Scan(
		&root,
		&blockHeight,
		&tree,
		&timestamp,
	)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (mrQ *MerkleTreeQuery) ScanRoot(row *sql.Row) ([]byte, error) {
	var (
		root, tree  []byte
		timestamp   int64
		blockHeight uint32
	)
	err := row.Scan(
		&root,
		&blockHeight,
		&tree,
		&timestamp,
	)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (mrQ *MerkleTreeQuery) BuildTree(rows *sql.Rows) (map[string][]byte, error) {
	var listTree = make(map[string][]byte)
	for rows.Next() {
		var (
			root, tree  []byte
			timestamp   int64
			blockHeight uint32
		)
		err := rows.Scan(
			&root,
			&blockHeight,
			&tree,
			&timestamp,
		)
		if err != nil {
			return nil, err
		}
		listTree[string(root)] = tree
	}
	return listTree, nil
}

func (mrQ *MerkleTreeQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", mrQ.getTableName()),
			height,
		},
	}
}
