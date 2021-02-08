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
	"database/sql"
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAccountBalanceHelperExecutorAddSpendableFail struct {
		query.ExecutorInterface
	}
	mockAccountBalanceHelperExecutorAddSpendableSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockAccountBalanceHelperExecutorAddSpendableFail) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperExecutorAddSpendableFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperExecutorAddSpendableSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockAccountBalanceHelperExecutorAddSpendableSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestAccountBalanceHelper_AddAccountSpendableBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		address []byte
		amount  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "executorError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableFail{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "executeSuccess",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableSuccess{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHelper{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := abh.AddAccountSpendableBalance(tt.args.address, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("AddAccountSpendableBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountBalanceHelper_AddAccountBalance(t *testing.T) {
	type fields struct {
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		address     []byte
		amount      int64
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "executorError",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableFail{},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "executeSuccess",
			fields: fields{
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				QueryExecutor:       &mockAccountBalanceHelperExecutorAddSpendableSuccess{},
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHelper{
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				AccountLedgerQuery:  tt.fields.AccountLedgerQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			if err := abh.AddAccountBalance(tt.args.address, tt.args.amount, 0, tt.args.blockHeight, 0, 0); (err != nil) != tt.wantErr {
				t.Errorf("AddAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperSuccess struct {
		AccountBalanceHelper
	}
	mockAccountBalanceHelperFail struct {
		AccountBalanceHelper
	}
)

func (*mockAccountBalanceHelperSuccess) AddAccountSpendableBalance(address []byte, amount int64) error {
	return nil
}
func (*mockAccountBalanceHelperSuccess) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return true, nil
}
func (*mockAccountBalanceHelperFail) AddAccountSpendableBalance(address []byte, amount int64) error {
	return sql.ErrTxDone
}
func (*mockAccountBalanceHelperFail) HasEnoughSpendableBalance(
	dbTX bool, address []byte, compareBalance int64,
) (enough bool, err error) {
	return false, nil
}
func (*mockAccountBalanceHelperSuccess) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return nil
}
func (*mockAccountBalanceHelperFail) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64, blockTimestamp uint64,
) error {
	return sql.ErrTxDone
}
