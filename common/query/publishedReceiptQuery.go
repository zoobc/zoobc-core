package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PublishedReceiptQueryInterface interface {
		GetPublishedReceiptByLinkedRMR(root []byte) (str string, args []interface{})
		GetPublishedReceiptByBlockHeight(blockHeight uint32) (str string, args []interface{})
		GetPublishedReceiptByBlockHeightRange(
			fromBlockHeight, toBlockHeight uint32,
		) (str string, args []interface{})
		InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt) (str string, args []interface{})
		InsertPublishedReceipts(receipts []*model.PublishedReceipt) (str string, args []interface{})
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
			"batch_reference_block_height",
			"receipt_index",
			"published_index",
			"published_receipt_type",
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

// InsertPublishedReceipts represents query builder to insert multiple record in single query
func (prq *PublishedReceiptQuery) InsertPublishedReceipts(receipts []*model.PublishedReceipt) (str string, args []interface{}) {
	if len(receipts) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			prq.getTableName(),
			strings.Join(prq.Fields, ", "),
		)
		for k, receipt := range receipts {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(prq.Fields)-1),
			)
			if k < len(receipts)-1 {
				str += ","
			}
			args = append(args, prq.ExtractModel(receipt)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (prq *PublishedReceiptQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	publishedReceipts, ok := payload.([]*model.PublishedReceipt)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+prq.TableName)
	}
	if len(publishedReceipts) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(prq.Fields), len(publishedReceipts))
		for i := 0; i < rounds; i++ {
			qry, args := prq.InsertPublishedReceipts(publishedReceipts[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := prq.InsertPublishedReceipts(publishedReceipts[len(publishedReceipts)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (prq *PublishedReceiptQuery) RecalibrateVersionedTable() []string {
	return []string{} // only table with `latest` column need this
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

func (prq *PublishedReceiptQuery) GetPublishedReceiptByBlockHeightRange(
	fromBlockHeight, toBlockHeight uint32,
) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_height BETWEEN ? AND ? ORDER BY block_height, published_index ASC",
		strings.Join(prq.Fields, ", "), prq.getTableName())
	return query, []interface{}{
		fromBlockHeight, toBlockHeight,
	}
}

func (*PublishedReceiptQuery) Scan(receipt *model.PublishedReceipt, row *sql.Row) error {
	err := row.Scan(
		&receipt.Receipt.SenderPublicKey,
		&receipt.Receipt.RecipientPublicKey,
		&receipt.Receipt.DatumType,
		&receipt.Receipt.DatumHash,
		&receipt.Receipt.ReferenceBlockHeight,
		&receipt.Receipt.ReferenceBlockHash,
		&receipt.Receipt.RMRLinked,
		&receipt.Receipt.RecipientSignature,
		&receipt.IntermediateHashes,
		&receipt.BlockHeight,
		&receipt.BatchReferenceBlockHeight,
		&receipt.ReceiptIndex,
		&receipt.PublishedIndex,
		&receipt.PublishedReceiptType,
	)
	return err

}

func (*PublishedReceiptQuery) ExtractModel(publishedReceipt *model.PublishedReceipt) []interface{} {
	return []interface{}{
		&publishedReceipt.Receipt.SenderPublicKey,
		&publishedReceipt.Receipt.RecipientPublicKey,
		&publishedReceipt.Receipt.DatumType,
		&publishedReceipt.Receipt.DatumHash,
		&publishedReceipt.Receipt.ReferenceBlockHeight,
		&publishedReceipt.Receipt.ReferenceBlockHash,
		&publishedReceipt.Receipt.RMRLinked,
		&publishedReceipt.Receipt.RecipientSignature,
		&publishedReceipt.IntermediateHashes,
		&publishedReceipt.BlockHeight,
		&publishedReceipt.BatchReferenceBlockHeight,
		&publishedReceipt.ReceiptIndex,
		&publishedReceipt.PublishedIndex,
		&publishedReceipt.PublishedReceiptType,
	}
}

func (prq *PublishedReceiptQuery) BuildModel(
	prs []*model.PublishedReceipt, rows *sql.Rows,
) ([]*model.PublishedReceipt, error) {
	for rows.Next() {
		var receipt = model.PublishedReceipt{
			Receipt: &model.Receipt{},
		}
		err := rows.Scan(
			&receipt.Receipt.SenderPublicKey,
			&receipt.Receipt.RecipientPublicKey,
			&receipt.Receipt.DatumType,
			&receipt.Receipt.DatumHash,
			&receipt.Receipt.ReferenceBlockHeight,
			&receipt.Receipt.ReferenceBlockHash,
			&receipt.Receipt.RMRLinked,
			&receipt.Receipt.RecipientSignature,
			&receipt.IntermediateHashes,
			&receipt.BlockHeight,
			&receipt.BatchReferenceBlockHeight,
			&receipt.ReceiptIndex,
			&receipt.PublishedIndex,
			&receipt.PublishedReceiptType,
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
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0 ORDER BY block_height",
		strings.Join(prq.Fields, ", "),
		prq.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (prq *PublishedReceiptQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		prq.TableName, fromHeight, toHeight)
}
