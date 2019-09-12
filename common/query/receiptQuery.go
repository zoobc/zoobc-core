package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	ReceiptQueryInterface interface {
		InsertReceipt(receipt *model.Receipt) (str string, args []interface{})
		GetReceipts(limit uint32, offset uint64) string
		ExtractModel(receipt *model.Receipt) []interface{}
		BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt
	}

	ReceiptQuery struct {
		Fields    []string
		TableName string
	}
)

// NewTransactionQuery returns TransactionQuery instance
func NewReceiptQuery() *ReceiptQuery {
	return &ReceiptQuery{
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
		TableName: "node_receipt",
	}
}

func (rq *ReceiptQuery) getTableName() string {
	return rq.TableName
}

// GetReceipts get a set of receipts that satisfies the params from DB
func (rq *ReceiptQuery) GetReceipts(limit uint32, offset uint64) string {
	query := fmt.Sprintf("SELECT %s from %s", strings.Join(rq.Fields, ", "), rq.getTableName())

	newLimit := limit
	if limit == 0 {
		newLimit = uint32(10)
	}

	query += fmt.Sprintf(" LIMIT %d,%d", offset, newLimit)
	return query
}

// InsertReceipts inserts a new receipts into DB
func (rq *ReceiptQuery) InsertReceipt(receipt *model.Receipt) (str string, args []interface{}) {
	var value = fmt.Sprintf("? %s", strings.Repeat(", ?", len(rq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		rq.getTableName(), strings.Join(rq.Fields, ", "), value)
	return query, rq.ExtractModel(receipt)
}

// ExtractModel extract the model struct fields to the order of ReceiptQuery.Fields
func (*ReceiptQuery) ExtractModel(receipt *model.Receipt) []interface{} {
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

func (*ReceiptQuery) BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt {
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
