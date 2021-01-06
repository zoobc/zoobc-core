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
package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// AccountBalanceHelperInterface methods collection for transaction helper, it for account balance stuff and account ledger also
	// It better to use with QueryExecutor.BeginTX()
	AccountBalanceHelperInterface interface {
		AddAccountSpendableBalance(address []byte, amount int64) error
		AddAccountBalance(address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64,
			blockTimestamp uint64) error
		GetBalanceByAccountAddress(accountBalance *model.AccountBalance, address []byte, dbTx bool) error
		HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error)
	}
	// AccountBalanceHelper fields for AccountBalanceHelperInterface for transaction helper
	AccountBalanceHelper struct {
		// accountBalance cache when get from db, use this for validation only.
		accountBalance      model.AccountBalance
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
)

func NewAccountBalanceHelper(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	accountLedgerQuery query.AccountLedgerQueryInterface,
) *AccountBalanceHelper {
	return &AccountBalanceHelper{
		AccountBalanceQuery: accountBalanceQuery,
		AccountLedgerQuery:  accountLedgerQuery,
		QueryExecutor:       queryExecutor,
	}
}

// AddAccountSpendableBalance add spendable_balance field to the address provided, must be executed inside db transaction
// scope
func (abh *AccountBalanceHelper) AddAccountSpendableBalance(address []byte, amount int64) error {
	accountBalanceSenderQ, accountBalanceSenderQArgs := abh.AccountBalanceQuery.AddAccountSpendableBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
		},
	)
	err := abh.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err == nil {
		abh.accountBalance = model.AccountBalance{}
	}
	return err
}

// AddAccountBalance add balance and spendable_balance field to the address provided at blockHeight, must be executed
// inside db transaction scope, there process is:
//      - Add new record into account_balance
//      - Add new record into account_ledger
func (abh *AccountBalanceHelper) AddAccountBalance(
	address []byte,
	amount int64,
	event model.EventType,
	blockHeight uint32,
	transactionID int64,
	blockTimestamp uint64,
) error {

	var queries [][]interface{}

	addAccountBalanceQ := abh.AccountBalanceQuery.AddAccountBalance(
		amount,
		map[string]interface{}{
			"account_address": address,
			"block_height":    blockHeight,
		},
	)
	queries = append(queries, addAccountBalanceQ...)

	accountLedgerQ, accountLedgerArgs := abh.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: address,
		BalanceChange:  amount,
		TransactionID:  transactionID,
		BlockHeight:    blockHeight,
		EventType:      event,
		Timestamp:      blockTimestamp,
	})
	queries = append(queries, append([]interface{}{accountLedgerQ}, accountLedgerArgs...))
	err := abh.QueryExecutor.ExecuteTransactions(queries)
	if err == nil {
		abh.accountBalance = model.AccountBalance{}
	}
	return err
}

// GetBalanceByAccountAddress fetching the balance of an account from database
func (abh *AccountBalanceHelper) GetBalanceByAccountAddress(accountBalance *model.AccountBalance, address []byte, dbTx bool) error {
	var (
		row *sql.Row
		err error
	)

	qry, args := abh.AccountBalanceQuery.GetAccountBalanceByAccountAddress(address)
	row, err = abh.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	err = abh.AccountBalanceQuery.Scan(accountBalance, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "TXSenderNotFound")
	}
	return nil
}

// HasEnoughSpendableBalance check if account has enough has spendable balance and will save
func (abh *AccountBalanceHelper) HasEnoughSpendableBalance(dbTX bool, address []byte, compareBalance int64) (enough bool, err error) {
	if bytes.Equal(abh.accountBalance.GetAccountAddress(), address) {
		return abh.accountBalance.GetSpendableBalance() >= compareBalance, nil
	}
	var (
		row            *sql.Row
		accountBalance model.AccountBalance
	)
	qry, args := abh.AccountBalanceQuery.GetAccountBalanceByAccountAddress(address)
	row, err = abh.QueryExecutor.ExecuteSelectRow(qry, dbTX, args...)
	if err != nil {
		return enough, err
	}
	err = abh.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return enough, err
	}
	abh.accountBalance = accountBalance
	return accountBalance.GetSpendableBalance() >= compareBalance, nil
}
