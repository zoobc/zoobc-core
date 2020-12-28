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
package service

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockAccountLedgerQuery = query.NewAccountLedgerQuery()
	mockAccountLedger      = &model.AccountLedger{
		AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		BalanceChange: 10,
		BlockHeight:   2,
		TransactionID: -9127118158999748858,
		EventType:     model.EventType_EventClaimNodeRegistrationTransaction,
		Timestamp:     1562117271,
	}
)

type (
	mockQueryAccountLedgersSuccess struct {
		query.Executor
	}
)

func (*mockQueryAccountLedgersSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("1"))
	return db.QueryRow(qStr), nil
}
func (*mockQueryAccountLedgersSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
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
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(rowsMock)
	return db.Query(qStr, args...)
}

func TestAccountLedgerService_GetAccountLedgers(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountLedgersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountLedgersResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryAccountLedgersSuccess{},
			},
			args: args{
				request: &model.GetAccountLedgersRequest{
					AccountAddress: mockAccountLedger.GetAccountAddress(),
					Pagination: &model.Pagination{
						Limit:      30,
						OrderField: "account_address",
						OrderBy:    model.OrderBy_DESC,
					},
				},
			},
			want: &model.GetAccountLedgersResponse{
				Total:          1,
				AccountLedgers: []*model.AccountLedger{mockAccountLedger},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &AccountLedgerService{
				Query: tt.fields.Query,
			}
			got, err := al.GetAccountLedgers(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountLedgers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountLedgers() got = %v, want %v", got, tt.want)
			}
		})
	}
}
