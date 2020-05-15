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
		GetPublishedReceiptByBlockHeight(blockHeight uint32) (str string, args []interface{})
		InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt) (str string, args []interface{})
		Scan(publishedReceipt *model.PublishedReceipt, row *sql.Row) error
		ExtractModel(publishedReceipt *model.PublishedReceipt) []interface{}
		BuildModel(prs []*model.PublishedReceipt, rows *sql.Rows) ([]*model.PublishedReceipt, error)
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

// InsertPublishedReceipt inserts a new pas into DB
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

func (prq *PublishedReceiptQuery) GetPublishedReceiptByBlockHeight(blockHeight uint32) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_height = ? ORDER BY published_index ASC",
		strings.Join(prq.Fields, ", "), prq.getTableName())
	return query, []interface{}{
		blockHeight,
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
		&receipt.PublishedIndex,
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
		&publishedReceipt.PublishedIndex,
	}
}

func (prq *PublishedReceiptQuery) BuildModel(
	prs []*model.PublishedReceipt, rows *sql.Rows,
) ([]*model.PublishedReceipt, error) {
	for rows.Next() {
		var receipt = model.PublishedReceipt{
			BatchReceipt: &model.BatchReceipt{},
		}
		err := rows.Scan(
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
			&receipt.PublishedIndex,
		)
		if err != nil {
			return nil, err
		}
		prs = append(prs, &receipt)
	}
	return prs, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (prq *PublishedReceiptQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", prq.getTableName()),
			height,
		},
	}
}

func (prq *PublishedReceiptQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d ORDER BY block_height",
		strings.Join(prq.Fields, ", "),
		prq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (prq *PublishedReceiptQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		prq.TableName, fromHeight, toHeight)
}
