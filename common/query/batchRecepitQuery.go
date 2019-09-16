package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BatchReceiptQueryInterface interface {
		InsertBatchReceipt(receipt *model.Receipt) (qStr string, args []interface{})
		GetBatchReceipts() string
		RemoveBatchReceipts() string
		ExtractModel(receipt *model.Receipt) []interface{}
	}
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
func (br *BatchReceiptQuery) GetBatchReceipts() string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(br.Fields, ", "), br.getTableName())
}

// RemoveBatchReceipts build delete query for `batch_receipt` table
func (br *BatchReceiptQuery) RemoveBatchReceipts() string {
	return fmt.Sprintf("DELETE FROM %s", br.TableName)
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
