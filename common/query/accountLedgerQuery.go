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

	"github.com/zoobc/zoobc-core/common/model"
)

type (

	// AccountLedgerQuery schema of AccountLedger
	AccountLedgerQuery struct {
		Fields    []string
		TableName string
	}
	// AccountLedgerQueryInterface includes interface methods for AccountLedgerQuery
	AccountLedgerQueryInterface interface {
		ExtractModel(accountLedger *model.AccountLedger) []interface{}
		BuildModel(accountLedgers []*model.AccountLedger, rows *sql.Rows) ([]*model.AccountLedger, error)
		InsertAccountLedger(accountLedger *model.AccountLedger) (qStr string, args []interface{})
	}
)

// NewAccountLedgerQuery func that return AccountLedger schema with value
func NewAccountLedgerQuery() *AccountLedgerQuery {
	return &AccountLedgerQuery{
		Fields: []string{
			"account_address",
			"balance_change",
			"block_height",
			"transaction_id",
			"event_type",
			"timestamp",
		},
		TableName: "account_ledger",
	}
}

func (q *AccountLedgerQuery) getTableName() interface{} {
	return q.TableName
}

// InsertAccountLedger represents insert query for AccountLedger
func (q *AccountLedgerQuery) InsertAccountLedger(accountLedger *model.AccountLedger) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			q.getTableName(),
			strings.Join(q.Fields, ", "),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(q.Fields)-1)),
		),
		q.ExtractModel(accountLedger)
}

// ExtractModel will extract accountLedger model to []interface
func (*AccountLedgerQuery) ExtractModel(accountLedger *model.AccountLedger) []interface{} {
	return []interface{}{
		accountLedger.GetAccountAddress(),
		accountLedger.GetBalanceChange(),
		accountLedger.GetBlockHeight(),
		accountLedger.GetTransactionID(),
		accountLedger.GetEventType(),
		accountLedger.GetTimestamp(),
	}
}

// BuildModel will create or build models that extracted from rows
func (*AccountLedgerQuery) BuildModel(accountLedgers []*model.AccountLedger, rows *sql.Rows) ([]*model.AccountLedger, error) {
	for rows.Next() {
		var (
			accountLedger model.AccountLedger
			err           error
		)
		err = rows.Scan(
			&accountLedger.AccountAddress,
			&accountLedger.BalanceChange,
			&accountLedger.BlockHeight,
			&accountLedger.TransactionID,
			&accountLedger.EventType,
			&accountLedger.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		accountLedgers = append(accountLedgers, &accountLedger)
	}
	return accountLedgers, nil
}

// Rollback represents delete query in block_height n
func (q *AccountLedgerQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", q.getTableName()),
			height,
		},
	}
}
