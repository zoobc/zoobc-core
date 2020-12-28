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
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewSkippedBlockSmithService(t *testing.T) {
	type args struct {
		skippedBlocksmithQuery *query.SkippedBlocksmithQuery
		queryExecutor          query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want SkippedBlockSmithServiceInterface
	}{
		{
			name: "wantSuccess",
			args: args{
				skippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
				queryExecutor:          &query.Executor{},
			},
			want: &SkippedBlockSmithService{
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
				QueryExecutor:          &query.Executor{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSkippedBlockSmithService(tt.args.skippedBlocksmithQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSkippedBlockSmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetSkippedBlockSmithsSelectFail struct {
		query.Executor
	}
	mockGetSkippedBlockSmithsSelectSuccess struct {
		query.Executor
	}
)

func (*mockGetSkippedBlockSmithsSelectFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}
func (*mockGetSkippedBlockSmithsSelectFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}
func (*mockGetSkippedBlockSmithsSelectSuccess) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockRows := mock.NewRows(query.NewSkippedBlocksmithQuery().Fields)
	mockRows.AddRow(
		[]byte{1},
		1,
		1,
		1,
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}
func (*mockGetSkippedBlockSmithsSelectSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}

func TestSkippedBlockSmithService_GetSkippedBlockSmiths(t *testing.T) {
	type fields struct {
		QueryExecutor          query.ExecutorInterface
		SkippedBlocksmithQuery *query.SkippedBlocksmithQuery
	}
	type args struct {
		req *model.GetSkippedBlocksmithsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetSkippedBlocksmithsResponse
		wantErr bool
	}{
		{
			name: "wantFail:",
			fields: fields{
				QueryExecutor:          &mockGetSkippedBlockSmithsSelectFail{},
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
			},
			args: args{
				req: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 1,
					BlockHeightEnd:   2,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "WantSuccess",
			fields: fields{
				QueryExecutor:          &mockGetSkippedBlockSmithsSelectSuccess{},
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
			},
			args: args{
				req: &model.GetSkippedBlocksmithsRequest{
					BlockHeightStart: 1,
					BlockHeightEnd:   2,
				},
			},
			want: &model.GetSkippedBlocksmithsResponse{
				Total: 1,
				SkippedBlocksmiths: []*model.SkippedBlocksmith{
					{
						BlocksmithPublicKey: []byte{1},
						POPChange:           1,
						BlockHeight:         1,
						BlocksmithIndex:     1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbs := &SkippedBlockSmithService{
				QueryExecutor:          tt.fields.QueryExecutor,
				SkippedBlocksmithQuery: tt.fields.SkippedBlocksmithQuery,
			}
			got, err := sbs.GetSkippedBlockSmiths(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkippedBlockSmithService.GetSkippedBlockSmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SkippedBlockSmithService.GetSkippedBlockSmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}
