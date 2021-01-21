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
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockAccountLedgerQuery = NewAccountLedgerQuery()
	mockAccountLedger      = &model.AccountLedger{
		AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		BalanceChange: 10000,
		BlockHeight:   1,
		TransactionID: -123123123123,
		EventType:     model.EventType_EventNodeRegistrationTransaction,
		Timestamp:     1562117271,
	}
)

func TestAccountLedgerQuery_InsertAccountLedger(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedger *model.AccountLedger
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedger: mockAccountLedger,
			},
			wantQStr: "INSERT INTO account_ledger (account_address, balance_change, block_height, transaction_id, event_type, timestamp) " +
				"VALUES(? , ?, ?, ?, ?, ?)",
			wantArgs: []interface{}{
				mockAccountLedger.GetAccountAddress(),
				mockAccountLedger.GetBalanceChange(),
				mockAccountLedger.GetBlockHeight(),
				mockAccountLedger.GetTransactionID(),
				mockAccountLedger.GetEventType(),
				mockAccountLedger.GetTimestamp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := q.InsertAccountLedger(tt.args.accountLedger)
			if gotQStr != tt.wantQStr {
				t.Errorf("InsertAccountLedger() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertAccountLedger() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountLedgerQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantMultiQueries [][]interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args:   args{height: 1},
			wantMultiQueries: [][]interface{}{
				{
					"DELETE FROM account_ledger WHERE block_height > ?",
					uint32(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := q.Rollback(tt.args.height); !reflect.DeepEqual(gotMultiQueries, tt.wantMultiQueries) {
				t.Errorf("Rollback() = %v, want %v", gotMultiQueries, tt.wantMultiQueries)
			}
		})
	}
}

func TestAccountLedgerQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedger *model.AccountLedger
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedger: mockAccountLedger,
			},
			want: []interface{}{
				mockAccountLedger.GetAccountAddress(),
				mockAccountLedger.GetBalanceChange(),
				mockAccountLedger.GetBlockHeight(),
				mockAccountLedger.GetTransactionID(),
				mockAccountLedger.GetEventType(),
				mockAccountLedger.GetTimestamp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := a.ExtractModel(tt.args.accountLedger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountLedgerQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountLedgerQuery_BuildModel(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rowsMock := sqlmock.NewRows(mockAccountLedgerQuery.Fields)
	rowsMock.AddRow(
		mockAccountLedger.GetAccountAddress(),
		mockAccountLedger.GetBalanceChange(),
		mockAccountLedger.GetBlockHeight(),
		mockAccountLedger.GetTransactionID(),
		mockAccountLedger.GetEventType(),
		mockAccountLedger.GetTimestamp(),
	)
	mock.ExpectQuery("").WillReturnRows(rowsMock)
	rows, _ := db.Query("")

	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		accountLedgers []*model.AccountLedger
		rows           *sql.Rows
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.AccountLedger
		wantErr bool
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockAccountLedgerQuery),
			args: args{
				accountLedgers: []*model.AccountLedger{},
				rows:           rows,
			},
			want: []*model.AccountLedger{mockAccountLedger},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AccountLedgerQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			got, err := a.BuildModel(tt.args.accountLedgers, tt.args.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountLedgerQuery.BuildModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountLedgerQuery.BuildModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
