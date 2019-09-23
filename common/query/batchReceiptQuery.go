package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BatchReceiptQueryInterface interface for BatchReceiptQuery
	BatchReceiptQueryInterface interface {
		InsertBatchReceipt(receipt *model.Receipt) (qStr string, args []interface{})
		GetBatchReceipts(limit uint32, offset uint64) string
		RemoveBatchReceiptByRoot(merkleRoot []byte) (qStr string, args []interface{})
		ExtractModel(receipt *model.Receipt) []interface{}
		BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt
		Scan(receipt *model.Receipt, rows *sql.Row) error
	}
	// BatchReceiptQuery us query for BatchReceipt
	BatchReceiptQuery struct {
		Fields    []string
		TableName string
	}
)

// NewBatchReceiptQuery return new BatchReceiptQuery instance
func NewBatchReceiptQuery() *BatchReceiptQuery {
	return &BatchReceiptQuery{
		Fields: []string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"receipt_merkle_root",
			"recipient_signature",
		},
		TableName: "batch_receipt",
	}
}

func (br *BatchReceiptQuery) getTableName() string {
	return br.TableName
}

// InsertBatchReceipt build insert query for `batch_receipt` table
func (br *BatchReceiptQuery) InsertBatchReceipt(receipt *model.Receipt) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			br.getTableName(),
			strings.Join(br.Fields, ", "),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(br.Fields)-1)),
		),
		br.ExtractModel(receipt)
}

// GetBatchReceipts build select query for `batch_receipt` table
func (br *BatchReceiptQuery) GetBatchReceipts(limit uint32, offset uint64) string {
	query := fmt.Sprintf(
		"SELECT %s FROM %s ",
		strings.Join(br.Fields, ", "),
		br.getTableName(),
	)
	newLimit := limit
	if limit == 0 {
		newLimit = uint32(10)
	}
	query += fmt.Sprintf(
		"ORDER BY reference_block_height LIMIT %d OFFSET %d",
		newLimit,
		offset,
	)
	return query
}

// RemoveBatchReceiptByRoot build delete query  for `batch_receipt` table by `receipt_merkle_root`
func (br *BatchReceiptQuery) RemoveBatchReceiptByRoot(root []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"DELETE FROM %s WHERE receipt_merkle_root = ?",
			br.getTableName(),
		),
		[]interface{}{root}
}

// ExtractModel extract the model struct fields to the order of BatchReceiptQuery.Fields
func (*BatchReceiptQuery) ExtractModel(receipt *model.Receipt) []interface{} {
	return []interface{}{
		&receipt.SenderPublicKey,
		&receipt.RecipientPublicKey,
		&receipt.DatumType,
		&receipt.DatumHash,
		&receipt.ReferenceBlockHeight,
		&receipt.ReferenceBlockHash,
		&receipt.ReceiptMerkleRoot,
		&receipt.RecipientSignature,
	}
}

// BuildModel extract __*sql.Rows__ into []*model.Receipt
func (*BatchReceiptQuery) BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt {

	for rows.Next() {
		var receipt model.Receipt
		_ = rows.Scan(
			&receipt.SenderPublicKey,
			&receipt.RecipientPublicKey,
			&receipt.DatumType,
			&receipt.DatumHash,
			&receipt.ReferenceBlockHeight,
			&receipt.ReferenceBlockHash,
			&receipt.ReceiptMerkleRoot,
			&receipt.RecipientSignature,
		)

		receipts = append(receipts, &receipt)
	}
	return receipts
}
func (*BatchReceiptQuery) Scan(receipt *model.Receipt, row *sql.Row) error {

	err := row.Scan(
		&receipt.SenderPublicKey,
		&receipt.RecipientPublicKey,
		&receipt.DatumType,
		&receipt.DatumHash,
		&receipt.ReferenceBlockHeight,
		&receipt.ReferenceBlockHash,
		&receipt.ReceiptMerkleRoot,
		&receipt.RecipientSignature,
	)
	return err

}
