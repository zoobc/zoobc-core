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

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PublishedReceiptQueryInterface interface {
		GetPublishedReceiptByLinkedRMR(root []byte) (str string, args []interface{})
		GetPublishedReceiptByBlockHeight(blockHeight uint32) (str string, args []interface{})
		GetUnlinkedPublishedReceiptByBlockHeightAndReceiver(blockHeight uint32, recipientPubKey []byte) (str string, args []interface{})
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

func NewPublishedReceipt() *model.PublishedReceipt {
	return &model.PublishedReceipt{
		Receipt: &model.Receipt{},
	}
}

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
			"rmr",
			"recipient_signature",
			"intermediate_hashes",
			"block_height",
			"rmr_linked",
			"rmr_linked_index",
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

func (prq *PublishedReceiptQuery) GetUnlinkedPublishedReceipt(root []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE rmr_linked != ?", strings.Join(prq.Fields, ", "), prq.getTableName())
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

func (prq *PublishedReceiptQuery) GetUnlinkedPublishedReceiptByBlockHeightAndReceiver(blockHeight uint32,
	recipientPubKey []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_height = ? AND recipient_public_key = ? AND rmr_linked IS NULL LIMIT 1",
		strings.Join(prq.Fields, ", "), prq.getTableName())
	return query, []interface{}{
		blockHeight,
		recipientPubKey,
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
		&receipt.Receipt.RMR,
		&receipt.Receipt.RecipientSignature,
		&receipt.IntermediateHashes,
		&receipt.BlockHeight,
		&receipt.RMRLinked,
		&receipt.RMRLinkedIndex,
		&receipt.PublishedIndex,
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
		&publishedReceipt.Receipt.RMR,
		&publishedReceipt.Receipt.RecipientSignature,
		&publishedReceipt.IntermediateHashes,
		&publishedReceipt.BlockHeight,
		&publishedReceipt.RMRLinked,
		&publishedReceipt.RMRLinkedIndex,
		&publishedReceipt.PublishedIndex,
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
			&receipt.Receipt.RMR,
			&receipt.Receipt.RecipientSignature,
			&receipt.IntermediateHashes,
			&receipt.BlockHeight,
			&receipt.RMRLinked,
			&receipt.RMRLinkedIndex,
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
