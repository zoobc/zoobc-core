package query

import (
	"reflect"
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
		insertQry, args := mockFeeScaleQuery.InsertFeeScale(mockFeeScale)
		expectQuery := "INSERT INTO fee_scale (fee_scale, block_height, latest) VALUES(? , ?, ?)"
		if insertQry != expectQuery {
			t.Errorf("expect query: %s get: %s", expectQuery, insertQry)
		}
		if !reflect.DeepEqual(args, mockFeeScaleQuery.ExtractModel(mockFeeScale)) {
			t.Error("returns args doesn't match return of .ExtractModel call")
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
