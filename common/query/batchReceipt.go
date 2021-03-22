// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	BatchReceiptQueryInterface interface {
		InsertReceipt(receipt *model.BatchReceipt) (str string, args []interface{})
		InsertReceipts(receipts []*model.BatchReceipt) (str string, args []interface{})
		GetReceipts(paginate model.Pagination) string
		GetReceiptsByRootInRange(lowerHeight, upperHeight uint32, root []byte) (str string, args []interface{})
		GetReceiptsByRefBlockHeightAndRefBlockHash(refHeight uint32, refHash []byte) (str string, args []interface{})
		GetReceiptsByRootAndDatumHash(root, datumHash []byte, datumType uint32) (str string, args []interface{})
		GetReceiptByRecipientAndDatumHash(datumHash []byte, datumType uint32, recipientPubKey []byte) (str string, args []interface{})
		GetReceiptsWithUniqueRecipient(limit, lowerBlockHeight, upperBlockHeight uint32) string
		SelectReceipt(lowerHeight, upperHeight, limit uint32) (str string)
		PruneData(blockHeight, limit uint32) (string, []interface{})
		ExtractModel(receipt *model.BatchReceipt) []interface{}
		BuildModel(receipts []*model.BatchReceipt, rows *sql.Rows) ([]*model.BatchReceipt, error)
		Scan(receipt *model.BatchReceipt, row *sql.Row) error
	}

	BatchReceiptQuery struct {
		Fields    []string
		TableName string
	}
)

func NewBatchReceipt() *model.BatchReceipt {
	return &model.BatchReceipt{
		Receipt: &model.Receipt{},
	}
}

// NewBatchReceiptQuery returns BatchReceiptQuery instance
func NewBatchReceiptQuery() *BatchReceiptQuery {
	return &BatchReceiptQuery{
		Fields: []string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr",
			"recipient_signature",
			"rmr_batch",
			"rmr_batch_index",
		},
		TableName: "node_receipt",
	}
}

func (rq *BatchReceiptQuery) getTableName() string {
	return rq.TableName
}

// GetReceipts get a set of receipts that satisfies the params from DB
func (rq *BatchReceiptQuery) GetReceipts(paginate model.Pagination) string {

	query := fmt.Sprintf(
		"SELECT %s FROM %s ",
		strings.Join(rq.Fields, ", "),
		rq.getTableName(),
	)

	newLimit := paginate.GetLimit()
	if newLimit == 0 {
		newLimit = constant.ReceiptNodeMaximum
	}

	orderField := paginate.GetOrderField()
	if orderField == "" {
		orderField = "reference_block_height"
	}

	query += fmt.Sprintf(
		"ORDER BY %s %s LIMIT %d OFFSET %d",
		orderField,
		paginate.GetOrderBy(),
		newLimit,
		paginate.GetPage(),
	)
	return query
}

// GetReceiptsWithUniqueRecipient get receipt with unique recipient_public_key
// lowerBlockHeight and upperBlockHeight is passed as window limit of receipt reference_block_height to pick
func (rq *BatchReceiptQuery) GetReceiptsWithUniqueRecipient(
	limit, lowerBlockHeight, upperBlockHeight uint32) string {
	var query string
	if limit == 0 {
		limit = 10
	}
	query = fmt.Sprintf("SELECT %s FROM %s AS rc WHERE "+
		"NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE pr.datum_hash == rc.datum_hash) "+
		"AND reference_block_height BETWEEN %d AND %d "+
		"GROUP BY recipient_public_key ORDER BY reference_block_height ASC LIMIT %d",
		strings.Join(rq.Fields, ", "), rq.getTableName(), lowerBlockHeight, upperBlockHeight, limit)
	return query
}

// GetReceiptsByRootInRange return sql query to fetch pas by its merkle root, the datum_hash should not already exists in
// published_receipt table
func (rq *BatchReceiptQuery) GetReceiptsByRootInRange(
	lowerHeight, upperHeight uint32, root []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s AS rc WHERE rc.rmr_batch = ? AND "+
		"NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE "+
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) AND "+
		"reference_block_height BETWEEN %d AND %d "+
		"GROUP BY recipient_public_key",
		strings.Join(rq.Fields, ", "), rq.getTableName(), lowerHeight, upperHeight)
	return query, []interface{}{
		root,
	}
}

func (rq *BatchReceiptQuery) GetReceiptsByRefBlockHeightAndRefBlockHash(refHeight uint32, refHash []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s AS rc WHERE rc.reference_block_height = ? AND "+
		"rc.reference_block_hash = ? LIMIT 1",
		strings.Join(rq.Fields, ", "), rq.getTableName())
	return query, []interface{}{
		refHeight,
		refHash,
	}
}

// GetReceiptsByRootAndDatumHash return sql query to fetch batch receipts by their merkle root
// note: order is important during receipt selection process during block generation
func (rq *BatchReceiptQuery) GetReceiptsByRootAndDatumHash(root, datumHash []byte, datumType uint32) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s AS rc WHERE rc.rmr_batch = ? AND rc.datum_hash = ? AND rc."+
		"datum_type = ? ORDER BY recipient_signature",
		strings.Join(rq.Fields, ", "), rq.getTableName())
	return query, []interface{}{
		root,
		datumHash,
		datumType,
	}
}

func (rq *BatchReceiptQuery) GetReceiptByRecipientAndDatumHash(datumHash []byte, datumType uint32,
	recipientPubKey []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s AS rc WHERE rc.datum_hash = ? AND rc.datum_type = ? AND rc.recipient_public_key = ? LIMIT 1",
		strings.Join(rq.Fields, ", "), rq.getTableName())
	return query, []interface{}{
		datumHash,
		datumType,
		recipientPubKey,
	}
}

// SelectReceipt select list of receipt by some filter
func (rq *BatchReceiptQuery) SelectReceipt(
	lowerHeight, upperHeight, limit uint32,
) (str string) {
	query := fmt.Sprintf("SELECT %s FROM %s AS nr WHERE EXISTS "+
		"(SELECT rmr FROM published_receipt AS pr WHERE nr.rmr_batch = pr.rmr AND "+
		"block_height >= %d AND block_height <= %d ) LIMIT %d",
		strings.Join(rq.Fields, ", "), rq.getTableName(), lowerHeight, upperHeight, limit)

	return query
}

// InsertReceipt inserts a new pas into DB
func (rq *BatchReceiptQuery) InsertReceipt(receipt *model.BatchReceipt) (str string, args []interface{}) {

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		rq.getTableName(),
		strings.Join(rq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(rq.Fields)-1)),
	), rq.ExtractModel(receipt)
}

// InsertReceipts build query for bulk store pas
func (rq *BatchReceiptQuery) InsertReceipts(receipts []*model.BatchReceipt) (str string, args []interface{}) {

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

// PruneData handle query for remove by reference_block_height with limit
func (rq *BatchReceiptQuery) PruneData(blockHeight, limit uint32) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"DELETE FROM %s WHERE reference_block_height IN("+
				"SELECT reference_block_height FROM %s "+
				"WHERE reference_block_height <? "+
				"ORDER BY reference_block_height ASC LIMIT ?)",
			rq.getTableName(),
			rq.getTableName(),
		), []interface{}{
			blockHeight,
			limit,
		}
}

// ExtractModel extract the model struct fields to the order of BatchReceiptQuery.Fields
func (*BatchReceiptQuery) ExtractModel(receipt *model.BatchReceipt) []interface{} {
	return []interface{}{
		&receipt.GetReceipt().SenderPublicKey,
		&receipt.GetReceipt().RecipientPublicKey,
		&receipt.GetReceipt().DatumType,
		&receipt.GetReceipt().DatumHash,
		&receipt.GetReceipt().ReferenceBlockHeight,
		&receipt.GetReceipt().ReferenceBlockHash,
		&receipt.GetReceipt().RMR,
		&receipt.GetReceipt().RecipientSignature,
		&receipt.RMRBatch,
		&receipt.RMRBatchIndex,
	}
}

func (*BatchReceiptQuery) BuildModel(batchReceipts []*model.BatchReceipt, rows *sql.Rows) ([]*model.BatchReceipt, error) {

	for rows.Next() {
		var (
			receipt      model.Receipt
			batchReceipt model.BatchReceipt
			err          error
		)

		err = rows.Scan(
			&receipt.SenderPublicKey,
			&receipt.RecipientPublicKey,
			&receipt.DatumType,
			&receipt.DatumHash,
			&receipt.ReferenceBlockHeight,
			&receipt.ReferenceBlockHash,
			&receipt.RMR,
			&receipt.RecipientSignature,
			&batchReceipt.RMRBatch,
			&batchReceipt.RMRBatchIndex,
		)
		if err != nil {
			return nil, err
		}
		batchReceipt.Receipt = &receipt
		batchReceipts = append(batchReceipts, &batchReceipt)
	}

	return batchReceipts, nil
}

func (*BatchReceiptQuery) Scan(batchReceipt *model.BatchReceipt, row *sql.Row) error {

	err := row.Scan(
		&batchReceipt.Receipt.SenderPublicKey,
		&batchReceipt.Receipt.RecipientPublicKey,
		&batchReceipt.Receipt.DatumType,
		&batchReceipt.Receipt.DatumHash,
		&batchReceipt.Receipt.ReferenceBlockHeight,
		&batchReceipt.Receipt.ReferenceBlockHash,
		&batchReceipt.Receipt.RMR,
		&batchReceipt.Receipt.RecipientSignature,
		&batchReceipt.RMRBatch,
		&batchReceipt.RMRBatchIndex,
	)
	return err

}

func (rq *BatchReceiptQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE reference_block_height > ?", rq.getTableName()),
			height,
		},
	}
}
