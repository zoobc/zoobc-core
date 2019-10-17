package query

import (
	"database/sql"
	"fmt"
	"strings"
)

type (
	// MerkleTreeQueryInterface contain builder func for MerkleTree
	MerkleTreeQueryInterface interface {
		InsertMerkleTree(root, tree []byte, timestamp int64, blockHeight uint32) (qStr string, args []interface{})
		GetMerkleTreeByRoot(root []byte) (qStr string, args []interface{})
		SelectMerkleTree(
			lowerHeight, upperHeight, limit uint32,
		) string
		GetLastMerkleRoot() (qStr string)
		ScanTree(row *sql.Row) ([]byte, error)
		ScanRoot(row *sql.Row) ([]byte, error)
		BuildTree(row *sql.Rows) (map[string][]byte, error)
		Rollback(height uint32) (multiQueries [][]interface{})
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
			"timestamp",
			"block_height",
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
		[]interface{}{root, tree, timestamp, blockHeight}
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
		root, tree  []byte
		timestamp   int64
		blockHeight uint32
	)
	err := row.Scan(
		&root,
		&tree,
		&timestamp,
		&blockHeight,
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
		&tree,
		&timestamp,
		&blockHeight,
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
			&tree,
			&timestamp,
			&blockHeight,
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
		{ // insert node_receipt to batch_receipt where the root will be deleted
			fmt.Sprintf("INSERT INTO batch_receipt (sender_public_key, recipient_public_key, datum_type, "+
				"datum_hash, reference_block_height, reference_block_hash, rmr_linked, recipient_signature) SELECT ("+
				"sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, "+
				"reference_block_hash, rmr_linked, recipient_signature FROM node_receipt WHERE rmr IN ("+
				"SELECT id FROM merkle_tree WHERE block_height > %d"+
				")", height),
		},
		{ // delete the node receipt related to the merkle root deleted
			fmt.Sprintf("DELETE FROM node_receipt WHERE rmr IN ("+
				"SELECT id FROM merkle_tree WHERE block_height > %d"+
				")", height),
		},
		{ // delete the root and tree
			fmt.Sprintf("DELETE FROM %s WHERE block_height > %d", mrQ.getTableName(), height),
		},
	}
}
