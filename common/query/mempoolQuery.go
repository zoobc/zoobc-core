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
	MempoolQueryInterface interface {
		GetMempoolTransactions() string
		GetMempoolTransaction() string
		InsertMempoolTransaction(mempoolTx *model.MempoolTransaction) (qStr string, args []interface{})
		DeleteMempoolTransaction() string
		DeleteMempoolTransactions([]string) string
		DeleteExpiredMempoolTransactions(expiration int64) string
		GetMempoolTransactionsWantToByHeight(height uint32) (qStr string)
		ExtractModel(block *model.MempoolTransaction) []interface{}
		BuildModel(mempools []*model.MempoolTransaction, rows *sql.Rows) ([]*model.MempoolTransaction, error)
		Scan(mempool *model.MempoolTransaction, row *sql.Row) error
	}

	MempoolQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewMempoolQuery returns MempoolQuery instance
func NewMempoolQuery(chaintype chaintype.ChainType) *MempoolQuery {
	return &MempoolQuery{
		Fields: []string{
			"id",
			"block_height",
			"fee_per_byte",
			"arrival_timestamp",
			"transaction_bytes",
			"sender_account_address",
			"recipient_account_address",
		},
		TableName: "mempool",
		ChainType: chaintype,
	}
}

func (mpq *MempoolQuery) getTableName() string {
	return mpq.TableName
}

// GetMempoolTransactions returns query string to get multiple mempool transactions
func (mpq *MempoolQuery) GetMempoolTransactions() string {
	return fmt.Sprintf("SELECT %s FROM %s ORDER BY fee_per_byte DESC", strings.Join(mpq.Fields, ", "), mpq.getTableName())
}

// GetMempoolTransaction returns query string to get multiple mempool transactions
func (mpq *MempoolQuery) GetMempoolTransaction() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = :id", strings.Join(mpq.Fields, ", "), mpq.getTableName())
}

func (mpq *MempoolQuery) InsertMempoolTransaction(mempoolTx *model.MempoolTransaction) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		mpq.getTableName(),
		strings.Join(mpq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(mpq.Fields)-1)),
	), mpq.ExtractModel(mempoolTx)
}

// DeleteMempoolTransaction delete one mempool transaction by id
func (mpq *MempoolQuery) DeleteMempoolTransaction() string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = :id", mpq.getTableName())
}

// DeleteMempoolTransactions delete one mempool transaction by id
func (mpq *MempoolQuery) DeleteMempoolTransactions(idsStr []string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", mpq.getTableName(), strings.Join(idsStr, ","))
}

// DeleteExpiredMempoolTransactions delete expired mempool transactions
func (mpq *MempoolQuery) DeleteExpiredMempoolTransactions(expiration int64) string {
	return fmt.Sprintf(
		"DELETE FROM %s WHERE arrival_timestamp <= %d",
		mpq.getTableName(),
		expiration,
	)
}

func (mpq *MempoolQuery) GetMempoolTransactionsWantToByHeight(height uint32) (qStr string) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height > %d",
		strings.Join(mpq.Fields, ", "),
		mpq.getTableName(),
		height,
	)
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*MempoolQuery) ExtractModel(mempool *model.MempoolTransaction) []interface{} {
	return []interface{}{
		mempool.ID,
		mempool.BlockHeight,
		mempool.FeePerByte,
		mempool.ArrivalTimestamp,
		mempool.TransactionBytes,
		mempool.SenderAccountAddress,
		mempool.RecipientAccountAddress,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*MempoolQuery) BuildModel(
	mempools []*model.MempoolTransaction,
	rows *sql.Rows,
) ([]*model.MempoolTransaction, error) {
	for rows.Next() {
		var (
			mempool model.MempoolTransaction
			err     error
		)
		err = rows.Scan(
			&mempool.ID,
			&mempool.BlockHeight,
			&mempool.FeePerByte,
			&mempool.ArrivalTimestamp,
			&mempool.TransactionBytes,
			&mempool.SenderAccountAddress,
			&mempool.RecipientAccountAddress,
		)
		if err != nil {
			return nil, err
		}
		mempools = append(mempools, &mempool)
	}
	return mempools, nil
}

// Scan similar with `sql.Scan`
func (*MempoolQuery) Scan(mempool *model.MempoolTransaction, row *sql.Row) error {
	err := row.Scan(
		&mempool.ID,
		&mempool.BlockHeight,
		&mempool.FeePerByte,
		&mempool.ArrivalTimestamp,
		&mempool.TransactionBytes,
		&mempool.SenderAccountAddress,
		&mempool.RecipientAccountAddress,
	)
	return err
}

// Rollback delete records `WHERE height > "block_height"
func (mpq *MempoolQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", mpq.getTableName()),
			height,
		},
	}
}
