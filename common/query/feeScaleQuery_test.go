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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockFeeScale = &model.FeeScale{
		FeeScale:    constant.OneZBC,
		BlockHeight: 10,
		Latest:      true,
	}
)

func TestFeeScaleQuery_Scan(t *testing.T) {
	t.Run("FeeScaleQuery - Scan", func(t *testing.T) {
		feeScaleQuery := NewFeeScaleQuery()
		var result model.FeeScale
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(feeScaleQuery.Fields).
			AddRow(mockFeeScale.FeeScale, mockFeeScale.BlockHeight, mockFeeScale.Latest))
		row := db.QueryRow("foo")
		err := feeScaleQuery.Scan(&result, row)
		if err != nil {
			t.Error(err.Error())
		}
		if !reflect.DeepEqual(&result, mockFeeScale) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", result, mockFeeScale)
		}

	})
}

func TestFeeScaleQuery_BuildModel(t *testing.T) {
	t.Run("FeeScaleQuery-BuildModel:success", func(t *testing.T) {
		feeScaleQuery := NewFeeScaleQuery()
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(feeScaleQuery.Fields).
			AddRow(mockFeeScale.FeeScale, mockFeeScale.BlockHeight, mockFeeScale.Latest))
		rows, _ := db.Query("foo")
		var tempFeeScales []*model.FeeScale
		res, err := feeScaleQuery.BuildModel(tempFeeScales, rows)
		if err != nil {
			t.Error(err.Error())
		}
		if !reflect.DeepEqual(res[0], mockFeeScale) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res[0], &mockFeeScale)
		}
	})
}

func TestFeeScaleQuery_ExtractModel(t *testing.T) {
	t.Run("ExtractModel:success", func(t *testing.T) {
		feeScaleQuery := NewFeeScaleQuery()
		res := feeScaleQuery.ExtractModel(mockFeeScale)
		want := []interface{}{
			mockFeeScale.FeeScale, mockFeeScale.BlockHeight, mockFeeScale.Latest,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}

var (
	mockFeeScaleQuery = &FeeScaleQuery{
		Fields: []string{
			"fee_scale",
			"block_height",
			"latest",
		},
		TableName: "fee_scale",
	}
)

func TestNewFeeScaleQuery(t *testing.T) {
	tests := []struct {
		name string
		want *FeeScaleQuery
	}{
		{
			name: "NewFeeScaleQuery",
			want: mockFeeScaleQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeeScaleQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeeScaleQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeScaleQuery_InsertFeeScale(t *testing.T) {
	t.Run("InsertFeeScale", func(t *testing.T) {
		insertQueries := mockFeeScaleQuery.InsertFeeScale(mockFeeScale)
		expect := [][]interface{}{
			{

				"UPDATE fee_scale SET latest = ? WHERE latest = ? AND block_height IN (SELECT MAX(t2.block_height) FROM fee_scale as t2)",
				0,
				1,
			},
			append(
				[]interface{}{
					fmt.Sprintf(
						"INSERT INTO fee_scale (%s) VALUES(%s)",
						strings.Join(mockFeeScaleQuery.Fields, ", "),
						fmt.Sprintf("? %s", strings.Repeat(", ?", len(mockFeeScaleQuery.Fields)-1)),
					),
				},
				mockFeeScaleQuery.ExtractModel(mockFeeScale)...,
			),
		}

		if !reflect.DeepEqual(insertQueries, expect) {
			t.Errorf("expect: %v\n got: %v\n", expect, insertQueries)
		}
	})

}

func TestFeeScaleQuery_GetLatestFeeScale(t *testing.T) {
	t.Run("getLatestFeeScale", func(t *testing.T) {
		qry := mockFeeScaleQuery.GetLatestFeeScale()
		expectQuery := "SELECT fee_scale, block_height, latest FROM fee_scale WHERE latest = true"
		if qry != expectQuery {
			t.Errorf("expect query: %s get: %s", expectQuery, qry)
		}
	})
}
