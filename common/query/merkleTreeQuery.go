package query

import (
	"database/sql"
	"fmt"
	"strings"
)

type (
	// MerkleTreeQueryInterface contain builder func for MerkleTree
	MerkleTreeQueryInterface interface {
		InsertMerkleTree(root, tree []byte) (qStr string, args []interface{})
		GetMerkleTreeByRoot(root []byte) (qStr string, args []interface{})
		SelectMerkleTree(
			lowerHeight, upperHeight, limit uint32,
		) string
		ScanTree(row *sql.Row) ([]byte, error)
		BuildTree(row *sql.Rows) (map[string][]byte, error)
	}
	// MerkleTreeQuery fields and table name
	MerkleTreeQuery struct {
		Fields    []string
		TableName string
	}
)

// NewMerkleTreeQuery func to create new MerkleTreeInterface
func NewMerkleTreeQuery() MerkleTreeQueryInterface {
	return &MerkleTreeQuery{
		Fields: []string{
			"id",
			"tree",
		},
		TableName: "merkle_tree",
	}

}

func (mrQ *MerkleTreeQuery) getTableName() string {
	return mrQ.TableName

}

// InsertMerkleTree func build insert Query for MerkleTree
func (mrQ *MerkleTreeQuery) InsertMerkleTree(root, tree []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			mrQ.getTableName(),
			strings.Join(mrQ.Fields, ", "),
			fmt.Sprintf("?%s", strings.Repeat(",? ", len(mrQ.Fields)-1)),
		),
		[]interface{}{root, tree}
}

// GetMerkleTreeByRoot is used to retrieve merkle table record, to check if the merkle root specified exist
func (mrQ *MerkleTreeQuery) GetMerkleTreeByRoot(root []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE id = ?",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(),
	), []interface{}{root}
}

func (mrQ *MerkleTreeQuery) SelectMerkleTree(
	lowerHeight, upperHeight, limit uint32,
) string {
	query := fmt.Sprintf("SELECT %s FROM %s AS mt WHERE EXISTS "+
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked AND "+
		"block_height >= %d AND block_height <= %d ) LIMIT %d",
		strings.Join(mrQ.Fields, ", "), mrQ.getTableName(), lowerHeight, upperHeight, limit)
	return query
}

func (mrQ *MerkleTreeQuery) ScanTree(row *sql.Row) ([]byte, error) {
	var (
		root, tree []byte
	)
	err := row.Scan(
		&root,
		&tree,
	)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (mrQ *MerkleTreeQuery) BuildTree(rows *sql.Rows) (map[string][]byte, error) {
	var listTree = make(map[string][]byte)
	for rows.Next() {
		var (
			root, tree []byte
		)
		err := rows.Scan(
			&root,
			&tree,
		)
		if err != nil {
			return nil, err
		}
		listTree[string(root)] = tree
	}
	return listTree, nil
}
