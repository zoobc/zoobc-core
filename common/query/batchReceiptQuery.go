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
		InsertBatchReceipt(receipt *model.BatchReceipt) (qStr string, args []interface{})
		GetBatchReceipts(limit uint32, offset uint64) string
		RemoveBatchReceiptByRoot(merkleRoot []byte) (qStr string, args []interface{})
		RemoveBatchReceipt(datumType uint32, datumHash []byte) (qStr string, args []interface{})
		ExtractModel(receipt *model.BatchReceipt) []interface{}
		BuildModel(receipts []*model.BatchReceipt, rows *sql.Rows) []*model.BatchReceipt
		Scan(receipt *model.BatchReceipt, rows *sql.Row) error
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
			"rmr_linked",
			"recipient_signature",
		},
		TableName: "batch_receipt",
	}
}

func (br *BatchReceiptQuery) getTableName() string {
	return br.TableName
}

// InsertBatchReceipt build insert query for `batch_receipt` table
func (br *BatchReceiptQuery) InsertBatchReceipt(receipt *model.BatchReceipt) (qStr string, args []interface{}) {
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
			"DELETE FROM %s WHERE rmr_linked = ?",
			br.getTableName(),
		),
		[]interface{}{root}
}

// RemoveBatchReceipt query builder to remove `batch_receipt WHERE datum_type = ? AND datum_hash = ?`
func (br *BatchReceiptQuery) RemoveBatchReceipt(datumType uint32, datumHash []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"DELETE FROM %s WHERE datum_type = ? AND datum_hash = ?",
			br.getTableName(),
		),
		[]interface{}{datumType, datumHash}
}

// ExtractModel extract the model struct fields to the order of BatchReceiptQuery.Fields
func (*BatchReceiptQuery) ExtractModel(receipt *model.BatchReceipt) []interface{} {
	return []interface{}{
		&receipt.SenderPublicKey,
		&receipt.RecipientPublicKey,
		&receipt.DatumType,
		&receipt.DatumHash,
		&receipt.ReferenceBlockHeight,
		&receipt.ReferenceBlockHash,
		&receipt.RMRLinked,
		&receipt.RecipientSignature,
	}
}

// BuildModel extract __*sql.Rows__ into []*model.Receipt
func (*BatchReceiptQuery) BuildModel(receipts []*model.BatchReceipt, rows *sql.Rows) []*model.BatchReceipt {
	for rows.Next() {
		var receipt model.BatchReceipt
		_ = rows.Scan(
			&receipt.SenderPublicKey,
			&receipt.RecipientPublicKey,
			&receipt.DatumType,
			&receipt.DatumHash,
			&receipt.ReferenceBlockHeight,
			&receipt.ReferenceBlockHash,
			&receipt.RMRLinked,
			&receipt.RecipientSignature,
		)

		receipts = append(receipts, &receipt)
	}
	return receipts
}

func (*BatchReceiptQuery) Scan(receipt *model.BatchReceipt, row *sql.Row) error {

	err := row.Scan(
		&receipt.SenderPublicKey,
		&receipt.RecipientPublicKey,
		&receipt.DatumType,
		&receipt.DatumHash,
		&receipt.ReferenceBlockHeight,
		&receipt.ReferenceBlockHash,
		&receipt.RMRLinked,
		&receipt.RecipientSignature,
	)
	return err

}
