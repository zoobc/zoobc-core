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
		InsertReceipts(receipts []*model.Receipt) (str string, args []interface{})
		GetReceipts(limit uint32, offset uint64) string
		ExtractModel(receipt *model.Receipt) []interface{}
		BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt
		Scan(receipt *model.Receipt, row *sql.Row) error
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
			"rmr_linked",
			"recipient_signature",
			"rmr",
			"rmr_index",
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

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		rq.getTableName(),
		strings.Join(rq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(rq.Fields)-1)),
	), rq.ExtractModel(receipt)
}

// InsertReceipts build query for bulk store receipts
func (rq *ReceiptQuery) InsertReceipts(receipts []*model.Receipt) (qStr string, args []interface{}) {

	var (
		query  string
		values []interface{}
	)

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) ",
		rq.getTableName(),
		strings.Join(rq.Fields, ", "),
	)

	for k, receipt := range receipts {
		query += fmt.Sprintf("VALUES(?%s)", strings.Repeat(",? ", len(rq.Fields)-1))
		if k < len(receipts)-1 {
			query += ", "
		}
		values = append(values, rq.ExtractModel(receipt)...)
	}
	return query, values
}

// ExtractModel extract the model struct fields to the order of ReceiptQuery.Fields
func (*ReceiptQuery) ExtractModel(receipt *model.Receipt) []interface{} {
	return []interface{}{
		&receipt.BatchReceipt.SenderPublicKey,
		&receipt.BatchReceipt.RecipientPublicKey,
		&receipt.BatchReceipt.DatumType,
		&receipt.BatchReceipt.DatumHash,
		&receipt.BatchReceipt.ReferenceBlockHeight,
		&receipt.BatchReceipt.ReferenceBlockHash,
		&receipt.BatchReceipt.RMRLinked,
		&receipt.BatchReceipt.RecipientSignature,
		&receipt.RMR,
		&receipt.RMRIndex,
	}
}

func (*ReceiptQuery) BuildModel(receipts []*model.Receipt, rows *sql.Rows) []*model.Receipt {

	for rows.Next() {
		var (
			receipt      model.Receipt
			batchReceipt model.BatchReceipt
		)

		_ = rows.Scan(
			&batchReceipt.SenderPublicKey,
			&batchReceipt.RecipientPublicKey,
			&batchReceipt.DatumType,
			&batchReceipt.DatumHash,
			&batchReceipt.ReferenceBlockHeight,
			&batchReceipt.ReferenceBlockHash,
			&batchReceipt.RMRLinked,
			&batchReceipt.RecipientSignature,
			&receipt.RMR,
			&receipt.RMRIndex,
		)
		receipt.BatchReceipt = &batchReceipt
		receipts = append(receipts, &receipt)
	}

	return receipts
}

func (*ReceiptQuery) Scan(receipt *model.Receipt, row *sql.Row) error {

	err := row.Scan(
		&receipt.BatchReceipt.SenderPublicKey,
		&receipt.BatchReceipt.RecipientPublicKey,
		&receipt.BatchReceipt.DatumType,
		&receipt.BatchReceipt.DatumHash,
		&receipt.BatchReceipt.ReferenceBlockHeight,
		&receipt.BatchReceipt.ReferenceBlockHash,
		&receipt.BatchReceipt.RMRLinked,
		&receipt.BatchReceipt.RecipientSignature,
		&receipt.RMR,
		&receipt.RMRIndex,
	)
	return err

}
