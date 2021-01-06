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

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TransactionQueryInterface interface {
		InsertTransaction(tx *model.Transaction) (str string, args []interface{})
		InsertTransactions(txs []*model.Transaction) (str string, args []interface{})
		GetTransaction(id int64) string
		GetTransactionsByIds(txIds []int64) (str string, args []interface{})
		GetTransactionsByBlockID(blockID int64) (str string, args []interface{})
		ExtractModel(tx *model.Transaction) []interface{}
		BuildModel(txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error)
		Scan(tx *model.Transaction, row *sql.Row) error
	}

	TransactionQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewTransactionQuery returns TransactionQuery instance
func NewTransactionQuery(chaintype chaintype.ChainType) *TransactionQuery {
	return &TransactionQuery{
		Fields: []string{
			"id",
			"block_id",
			"block_height",
			"sender_account_address",
			"recipient_account_address",
			"transaction_type",
			"fee",
			"timestamp",
			"transaction_hash",
			"transaction_body_length",
			"transaction_body_bytes",
			"signature",
			"version",
			"transaction_index",
			"child_type",
			"message",
		},
		TableName: "\"transaction\"",
		ChainType: chaintype,
	}
}

func (tq *TransactionQuery) getTableName() string {
	return tq.TableName
}

// GetTransaction get a single transaction from DB
func (tq *TransactionQuery) GetTransaction(id int64) string {
	query := fmt.Sprintf(
		"SELECT %s from %s WHERE id = %d",
		strings.Join(tq.Fields, ", "),
		tq.getTableName(),
		id,
	)
	return query
}

// InsertTransaction inserts a new transaction into DB
func (tq *TransactionQuery) InsertTransaction(tx *model.Transaction) (str string, args []interface{}) {
	var value = fmt.Sprintf("?%s", strings.Repeat(", ?", len(tq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		tq.getTableName(), strings.Join(tq.Fields, ", "), value)
	return query, tq.ExtractModel(tx)
}

func (tq *TransactionQuery) InsertTransactions(txs []*model.Transaction) (str string, args []interface{}) {
	if len(txs) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			tq.getTableName(),
			strings.Join(tq.Fields, ", "),
		)
		for k, atomic := range txs {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(tq.Fields)-1),
			)
			if k < len(txs)-1 {
				str += ","
			}

			args = append(args, tq.ExtractModel(atomic)...)
		}
	}
	return str, args

}
func (tq *TransactionQuery) GetTransactionsByBlockID(blockID int64) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_id = ? AND child_type = ? "+
		"ORDER BY transaction_index ASC", strings.Join(tq.Fields, ", "), tq.getTableName())
	return query, []interface{}{blockID, uint32(model.TransactionChildType_NoneChild)}
}

func (tq *TransactionQuery) GetTransactionsByIds(txIds []int64) (str string, args []interface{}) {

	args = append(args, uint32(model.TransactionChildType_NoneChild))
	for _, id := range txIds {
		args = append(args, id)
	}
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE child_type = ? AND id IN(?%s)",
			strings.Join(tq.Fields, ", "),
			tq.getTableName(),
			strings.Repeat(", ?", len(txIds)-1),
		),
		args
}

// ExtractModel extract the model struct fields to the order of TransactionQuery.Fields
func (*TransactionQuery) ExtractModel(tx *model.Transaction) []interface{} {
	return []interface{}{
		&tx.ID,
		&tx.BlockID,
		&tx.Height,
		&tx.SenderAccountAddress,
		&tx.RecipientAccountAddress,
		&tx.TransactionType,
		&tx.Fee,
		&tx.Timestamp,
		&tx.TransactionHash,
		&tx.TransactionBodyLength,
		&tx.TransactionBodyBytes,
		&tx.Signature,
		&tx.Version,
		&tx.TransactionIndex,
		&tx.ChildType,
		&tx.Message,
	}
}

func (*TransactionQuery) BuildModel(txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	for rows.Next() {
		var (
			tx  model.Transaction
			err error
		)
		err = rows.Scan(
			&tx.ID,
			&tx.BlockID,
			&tx.Height,
			&tx.SenderAccountAddress,
			&tx.RecipientAccountAddress,
			&tx.TransactionType,
			&tx.Fee,
			&tx.Timestamp,
			&tx.TransactionHash,
			&tx.TransactionBodyLength,
			&tx.TransactionBodyBytes,
			&tx.Signature,
			&tx.Version,
			&tx.TransactionIndex,
			&tx.ChildType,
			&tx.Message,
		)
		if err != nil {
			return nil, err
		}
		txs = append(txs, &tx)
	}
	return txs, nil
}

func (*TransactionQuery) Scan(tx *model.Transaction, row *sql.Row) error {
	err := row.Scan(
		&tx.ID,
		&tx.BlockID,
		&tx.Height,
		&tx.SenderAccountAddress,
		&tx.RecipientAccountAddress,
		&tx.TransactionType,
		&tx.Fee,
		&tx.Timestamp,
		&tx.TransactionHash,
		&tx.TransactionBodyLength,
		&tx.TransactionBodyBytes,
		&tx.Signature,
		&tx.Version,
		&tx.TransactionIndex,
		&tx.ChildType,
		&tx.Message,
	)
	return err
}

// Rollback delete records `WHERE height > "height"
func (tq *TransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", tq.getTableName()),
			height,
		},
	}
}
