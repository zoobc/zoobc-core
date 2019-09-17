package query

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	// MerkleTreeQueryInterface contain builder func for MerkleTree
	MerkleTreeQueryInterface interface {
		InsertMerkleTree(tree [][]*bytes.Buffer) (qStr string, args []interface{})
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
func (mrQ *MerkleTreeQuery) InsertMerkleTree(tree [][]*bytes.Buffer) (qStr string, args []interface{}) {

	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			mrQ.getTableName(),
			strings.Join(mrQ.Fields, ", "),
			fmt.Sprintf("?%s", strings.Repeat("?, ", len(mrQ.Fields)-1)),
		),
		[]interface{}{tree[:len(tree)-1], tree}
}
