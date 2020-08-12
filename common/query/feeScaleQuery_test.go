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
