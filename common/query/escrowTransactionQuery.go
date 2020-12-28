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
	// EscrowTransactionQuery fields must have
	EscrowTransactionQuery struct {
		Fields    []string
		TableName string
	}

	// EscrowTransactionQueryInterface methods must have
	EscrowTransactionQueryInterface interface {
		InsertEscrowTransaction(escrow *model.Escrow) [][]interface{}
		InsertEscrowTransactions(escrows []*model.Escrow) (str string, args []interface{})
		GetLatestEscrowTransactionByID(int64) (string, []interface{})
		GetEscrowTransactions(fields map[string]interface{}) (string, []interface{})
		GetExpiredEscrowTransactionsAtCurrentBlock(blockHeight uint32) string
		GetEscrowTransactionsByTransactionIdsAndStatus(
			transactionIds []string, status model.EscrowStatus,
		) string
		ExpiringEscrowTransactions(blockHeight uint32) (string, []interface{})
		ExtractModel(*model.Escrow) []interface{}
		BuildModels(*sql.Rows) ([]*model.Escrow, error)
		Scan(escrow *model.Escrow, row *sql.Row) error
	}
)

// NewEscrowTransactionQuery build an EscrowTransactionQuery
func NewEscrowTransactionQuery() *EscrowTransactionQuery {
	return &EscrowTransactionQuery{
		Fields: []string{
			"id",
			"sender_address",
			"recipient_address",
			"approver_address",
			"amount",
			"commission",
			"timeout",
			"status",
			"block_height",
			"latest",
			"instruction",
		},
		TableName: "escrow_transaction",
	}
}

func (et *EscrowTransactionQuery) getTableName() string {
	return et.TableName
}

/*
InsertEscrowTransaction represents insert query for escrow_transaction table.
There 2 queries result:
		1. Update the previous record to latest = false
		2. Insert new record which is the newest
*/
func (et *EscrowTransactionQuery) InsertEscrowTransaction(escrow *model.Escrow) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE id = ?",
				et.getTableName(),
			),
			false,
			escrow.GetID(),
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					et.getTableName(),
					strings.Join(et.Fields, ","),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(et.Fields)-1))),
			},
			et.ExtractModel(escrow)...,
		),
	}
}

// InsertEscrowTransactions represents query builder to insert multiple record in single query
func (et *EscrowTransactionQuery) InsertEscrowTransactions(escrows []*model.Escrow) (str string, args []interface{}) {
	if len(escrows) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			et.getTableName(),
			strings.Join(et.Fields, ","),
		)
		for k, escrow := range escrows {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(et.Fields)-1),
			)

			if k < len(escrows)-1 {
				str += ","
			}
			args = append(args, et.ExtractModel(escrow)...)
		}
	}

	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (et *EscrowTransactionQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	escrows, ok := payload.([]*model.Escrow)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+et.TableName)
	}
	if len(escrows) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(et.Fields), len(escrows))
		for i := 0; i < rounds; i++ {
			qry, args := et.InsertEscrowTransactions(escrows[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := et.InsertEscrowTransactions(escrows[len(escrows)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (et *EscrowTransactionQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (id, block_height) NOT IN "+
				"(select t2.id, max(t2.block_height) from %s t2 group by t2.id)",
			et.getTableName(), et.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (id, block_height) IN "+
				"(select t2.id, max(t2.block_height) from %s t2 group by t2.id)",
			et.getTableName(), et.getTableName()),
	}
}

// GetLatestEscrowTransactionByID represents getting latest escrow by id
func (et *EscrowTransactionQuery) GetLatestEscrowTransactionByID(id int64) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE id = ? AND latest = ?",
			strings.Join(et.Fields, ", "),
			et.getTableName(),
		),
		[]interface{}{id, true}
}

// GetEscrowTransactions represents SELECT with multiple clauses connected via AND operand
func (et *EscrowTransactionQuery) GetEscrowTransactions(fields map[string]interface{}) (qStr string, args []interface{}) {
	qStr = fmt.Sprintf("SELECT %s FROM %s ", strings.Join(et.Fields, ", "), et.getTableName())

	if len(fields) > 0 {
		qStr += "WHERE "
		i := 1
		for k, v := range fields {
			qStr += fmt.Sprintf("%s = ? ", k)
			if i < len(fields) {
				qStr += "AND "
			}
			args = append(args, v)
			i++
		}
	}

	return qStr, args
}

// GetExpiredEscrowTransactionsAtCurrentBlock fetch provided block height expired escrow transaction
func (et *EscrowTransactionQuery) GetExpiredEscrowTransactionsAtCurrentBlock(blockHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE timeout + block_height = %d AND latest = true AND status = %d",
		strings.Join(et.Fields, ", "), et.getTableName(), blockHeight, model.EscrowStatus_Pending)
}

// ExpiringEscrowTransactions represents update escrows status to expired where that has been expired by blockHeight
func (et *EscrowTransactionQuery) ExpiringEscrowTransactions(blockHeight uint32) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"UPDATE %s SET latest = ?, status = ? WHERE timeout < ? AND status = 0",
			et.getTableName(),
		),
		[]interface{}{
			1,
			model.EscrowStatus_Expired,
			blockHeight,
		}
}

func (et *EscrowTransactionQuery) GetEscrowTransactionsByTransactionIdsAndStatus(
	transactionIds []string, status model.EscrowStatus,
) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE id IN (%s) AND status = %d",
		strings.Join(et.Fields, ", "),
		et.getTableName(),
		strings.Join(transactionIds, ", "),
		status,
	)
}

// ExtractModel will extract values of escrow as []interface{}
func (et *EscrowTransactionQuery) ExtractModel(escrow *model.Escrow) []interface{} {
	return []interface{}{
		escrow.GetID(),
		escrow.GetSenderAddress(),
		escrow.GetRecipientAddress(),
		escrow.GetApproverAddress(),
		escrow.GetAmount(),
		escrow.GetCommission(),
		escrow.GetTimeout(),
		escrow.GetStatus(),
		escrow.GetBlockHeight(),
		escrow.GetLatest(),
		escrow.GetInstruction(),
	}
}

// BuildModels extract sqlRaw into []*model.Escrow
func (et *EscrowTransactionQuery) BuildModels(rows *sql.Rows) ([]*model.Escrow, error) {
	var (
		escrows []*model.Escrow
		err     error
	)

	for rows.Next() {
		var escrow model.Escrow
		err = rows.Scan(
			&escrow.ID,
			&escrow.SenderAddress,
			&escrow.RecipientAddress,
			&escrow.ApproverAddress,
			&escrow.Amount,
			&escrow.Commission,
			&escrow.Timeout,
			&escrow.Status,
			&escrow.BlockHeight,
			&escrow.Latest,
			&escrow.Instruction,
		)
		if err != nil {
			return nil, err
		}
		escrows = append(escrows, &escrow)
	}
	return escrows, nil
}

// Scan extract sqlRaw *sql.Row into model.Escrow
func (et *EscrowTransactionQuery) Scan(escrow *model.Escrow, row *sql.Row) error {
	return row.Scan(
		&escrow.ID,
		&escrow.SenderAddress,
		&escrow.RecipientAddress,
		&escrow.ApproverAddress,
		&escrow.Amount,
		&escrow.Commission,
		&escrow.Timeout,
		&escrow.Status,
		&escrow.BlockHeight,
		&escrow.Latest,
		&escrow.Instruction,
	)
}

// Rollback delete records `WHERE height > "height"
func (et *EscrowTransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", et.getTableName()),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (id, block_height) IN (
				SELECT t2.id, MAX(t2.block_height)
				FROM %s as t2
				GROUP BY t2.id
			)`,
				et.TableName,
				et.TableName,
			),
			1,
			0,
		},
	}
}

func (et *EscrowTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (id, block_height) IN (SELECT t2.id, MAX(t2.block_height) FROM %s as t2 "+
			"WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 GROUP BY t2.id) ORDER BY block_height",
		strings.Join(et.Fields, ","),
		et.getTableName(),
		et.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (et *EscrowTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		et.TableName, fromHeight, toHeight)
}
