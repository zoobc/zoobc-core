package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PublishedReceiptQueryInterface interface {
		GetPublishedReceiptByLinkedRMR(root []byte) (str string, args []interface{})
		InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt) (str string, args []interface{})
		Scan(publishedReceipt *model.PublishedReceipt, row *sql.Row) error
		ExtractModel(publishedReceipt *model.PublishedReceipt) []interface{}
	}

	PublishedReceiptQuery struct {
		Fields    []string
		TableName string
	}
)

// NewPublishedReceiptQuery returns PublishedQuery instance
func NewPublishedReceiptQuery() *PublishedReceiptQuery {
	return &PublishedReceiptQuery{
		Fields: []string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"intermediate_hashes",
			"block_height",
			"receipt_index",
			"published_index",
		},
		TableName: "published_receipt",
	}
}

func (prq *PublishedReceiptQuery) getTableName() string {
	return prq.TableName
}

// InsertPublishedReceipt inserts a new receipts into DB
func (prq *PublishedReceiptQuery) InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		prq.getTableName(),
		strings.Join(prq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(prq.Fields)-1)),
	), prq.ExtractModel(publishedReceipt)
}

func (prq *PublishedReceiptQuery) GetPublishedReceiptByLinkedRMR(root []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE rmr_linked = ?", strings.Join(prq.Fields, ", "), prq.getTableName())
	return query, []interface{}{
		root,
	}
}

func (*PublishedReceiptQuery) Scan(receipt *model.PublishedReceipt, row *sql.Row) error {
	err := row.Scan(
		&receipt.BatchReceipt.SenderPublicKey,
		&receipt.BatchReceipt.RecipientPublicKey,
		&receipt.BatchReceipt.DatumType,
		&receipt.BatchReceipt.DatumHash,
		&receipt.BatchReceipt.ReferenceBlockHeight,
		&receipt.BatchReceipt.ReferenceBlockHash,
		&receipt.BatchReceipt.RMRLinked,
		&receipt.BatchReceipt.RecipientSignature,
		&receipt.IntermediateHashes,
		&receipt.BlockHeight,
		&receipt.ReceiptIndex,
		&receipt.Index,
	)
	return err

}

func (*PublishedReceiptQuery) ExtractModel(publishedReceipt *model.PublishedReceipt) []interface{} {
	return []interface{}{
		&publishedReceipt.BatchReceipt.SenderPublicKey,
		&publishedReceipt.BatchReceipt.RecipientPublicKey,
		&publishedReceipt.BatchReceipt.DatumType,
		&publishedReceipt.BatchReceipt.DatumHash,
		&publishedReceipt.BatchReceipt.ReferenceBlockHeight,
		&publishedReceipt.BatchReceipt.ReferenceBlockHash,
		&publishedReceipt.BatchReceipt.RMRLinked,
		&publishedReceipt.BatchReceipt.RecipientSignature,
		&publishedReceipt.IntermediateHashes,
		&publishedReceipt.BlockHeight,
		&publishedReceipt.ReceiptIndex,
		&publishedReceipt.Index,
	}
}
