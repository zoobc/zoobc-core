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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockGetAccountDatasetsExecutor struct {
		query.ExecutorInterface
	}
)

var (
	accDatasetSetterAccount1 = []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
		213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54}
	accDatasetRecipientAccount1 = []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75, 213,
		137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54}
)

func (*mockGetAccountDatasetsExecutor) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}
func (*mockGetAccountDatasetsExecutor) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mockRows.AddRow(
		accDatasetSetterAccount1,
		accDatasetRecipientAccount1,
		"AccountDatasetEscrowApproval",
		"Message",
		true,
		true,
		5,
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)

	return db.Query("")
}

func TestAccountDatasetService_GetAccountDatasets(t *testing.T) {
	type fields struct {
		AccountDatasetQuery *query.AccountDatasetQuery
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountDatasetsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountDatasetsResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &mockGetAccountDatasetsExecutor{},
			},
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Property:                "AccountDatasetEscrowApproval",
					Value:                   "Message",
					RecipientAccountAddress: accDatasetRecipientAccount1,
					SetterAccountAddress:    nil,
					Height:                  0,
					Pagination: &model.Pagination{
						OrderField: "height",
						OrderBy:    model.OrderBy_ASC,
						Page:       0,
						Limit:      500,
					},
				},
			},
			want: &model.GetAccountDatasetsResponse{
				Total: 1,
				AccountDatasets: []*model.AccountDataset{
					{
						SetterAccountAddress:    accDatasetSetterAccount1,
						RecipientAccountAddress: accDatasetRecipientAccount1,
						Property:                "AccountDatasetEscrowApproval",
						Value:                   "Message",
						Height:                  5,
						Latest:                  true,
						IsActive:                true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ads := &AccountDatasetService{
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			got, err := ads.GetAccountDatasets(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountDatasets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountDatasets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockExecutorGetAccountDataset struct {
		query.ExecutorInterface
	}
	mockExecutorGetAccountDatasetErr struct {
		query.ExecutorInterface
	}
)

func (*mockExecutorGetAccountDataset) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mockRow.AddRow(
		accDatasetSetterAccount1,
		accDatasetRecipientAccount1,
		"AccountDatasetEscrowApproval",
		"Message",
		true,
		true,
		5,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow(""), nil
}
func (*mockExecutorGetAccountDatasetErr) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow(""), nil
}

func TestAccountDatasetService_GetAccountDataset(t *testing.T) {
	type fields struct {
		AccountDatasetQuery *query.AccountDatasetQuery
		QueryExecutor       query.ExecutorInterface
	}
	type args struct {
		request *model.GetAccountDatasetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.AccountDataset
		wantErr bool
	}{
		{
			name: "wantError:NoRows",
			fields: fields{
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &mockExecutorGetAccountDatasetErr{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					Property: "AccountDatasetEscrowApproval",
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       &mockExecutorGetAccountDataset{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					Property: "AccountDatasetEscrowApproval",
				},
			},
			want: &model.AccountDataset{
				SetterAccountAddress:    accDatasetSetterAccount1,
				RecipientAccountAddress: accDatasetRecipientAccount1,
				Property:                "AccountDatasetEscrowApproval",
				Value:                   "Message",
				Height:                  5,
				Latest:                  true,
				IsActive:                true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ads := &AccountDatasetService{
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
				QueryExecutor:       tt.fields.QueryExecutor,
			}
			got, err := ads.GetAccountDataset(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountDataset() got = %v, want %v", got, tt.want)
			}
		})
	}
}
