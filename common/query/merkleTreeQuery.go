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
		SelectMerkleTree(
			lowerHeight, upperHeight, limit uint32,
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
SelectMerkleTree represents get merkle tree in range of block_height
and order by block_height ascending
test_expression >= low_expression AND test_expression <= high_expression
*/
func (mrQ *MerkleTreeQuery) SelectMerkleTree(
	lowerHeight, upperHeight, limit uint32,
) string {
	query := fmt.Sprintf("SELECT %s FROM %s AS mt WHERE EXISTS "+
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked) AND "+
		"block_height BETWEEN %d AND %d ORDER BY block_height ASC LIMIT %d",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(), lowerHeight, upperHeight, limit)
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
