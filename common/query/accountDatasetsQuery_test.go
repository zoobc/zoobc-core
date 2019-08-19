package query

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
)

var mockDatasetQuery = &AccountDatasetsQuery{
	PrimaryFields: []string{
		"setter_account_address",
		"recipient_account_address",
		"property",
		"height",
	},
	OrdinaryFields: []string{
		"value",
		"timestamp_starts",
		"timestamp_expires",
		"latest",
	},
	TableName: "account_datasets",
}

var mockDataset = &model.Dataset{
	SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
	RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
	Property:                "Admin",
	Value:                   "You Welcome",
	TimestampStarts:         1565942932686,
	TimestampExpires:        1565943056129,
	Latest:                  true,
	Height:                  5,
}

func TestNewAccountDatasetsQuery(t *testing.T) {
	tests := []struct {
		name string
		want *AccountDatasetsQuery
	}{
		{
			name: "success",
			want: mockDatasetQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountDatasetsQuery(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountDatasetsQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetsQuery_GetDatasetsByRecipientAccountAddress(t *testing.T) {
	type args struct {
		RecipientAccountAddress string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantArgs  interface{}
	}{
		{
			name: "success",
			args: args{
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			},
			wantQuery: "SELECT setter_account_address,recipient_account_address,property,height,value,timestamp_starts,timestamp_expires,latest " +
				"FROM account_datasets " +
				"WHERE recipient_account_address = ? AND latest = 1",
			wantArgs: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := mockDatasetQuery.GetDatasetsByRecipientAccountAddress(tt.args.RecipientAccountAddress)
			if gotQuery != tt.wantQuery {
				t.Errorf("AccountDatasetsQuery.GetDatasetsByRecipientAccountAddress() gotQuery = \n%v want \n%v", gotQuery, tt.wantQuery)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("AccountDatasetsQuery.GetDatasetsByRecipientAccountAddress() gotArgs = \n%v want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountDatasetsQuery_GetLastDataset(t *testing.T) {
	type args struct {
		SetterAccountAddress    string
		RecipientAccountAddress string
		property                string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "success",
			args: args{
				SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				property:                "Admin",
			},
			wantQuery: "SELECT setter_account_address,recipient_account_address,property,height,value,timestamp_starts,timestamp_expires,latest " +
				"FROM account_datasets " +
				"WHERE setter_account_address = ? AND recipient_account_address = ? AND property = ? AND latest = 1 ORDER BY height DESC LIMIT 1",
			wantArgs: []interface{}{"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN", "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J", "Admin"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := mockDatasetQuery.GetLastDataset(tt.args.SetterAccountAddress, tt.args.RecipientAccountAddress, tt.args.property)
			if gotQuery != tt.wantQuery {
				t.Errorf("AccountDatasetsQuery.GetLastDataset() gotQuery = \n%v want \n%v", gotQuery, tt.wantQuery)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("AccountDatasetsQuery.GetLastDataset() gotArgs = \n%v want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAccountDatasetsQuery_AddDataset(t *testing.T) {
	var want [][]interface{}
	type args struct {
		dataset *model.Dataset
	}

	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{
			name: "success",
			args: args{
				dataset: mockDataset,
			},
			want: append(want, append([]interface{}{fmt.Sprintf(`
		UPDATE %s SET (%s) = 
		(
			SELECT '%s', %d, 
				%d + CASE 
					WHEN timestamp_expires - %d < 0 THEN 0
					ELSE timestamp_expires - %d END 
			FROM %s 
			WHERE %s AND latest = 1
			ORDER BY height DESC LIMIT 1
		) 
		WHERE %s AND latest = 1
	`,
				mockDatasetQuery.getTableName(),
				strings.Join(mockDatasetQuery.OrdinaryFields[:3], ","),
				mockDataset.GetValue(),
				mockDataset.GetTimestampStarts(),
				mockDataset.GetTimestampExpires(),
				mockDataset.GetTimestampStarts(),
				mockDataset.GetTimestampStarts(),
				mockDatasetQuery.getTableName(),
				fmt.Sprintf("%s = ? ", strings.Join(mockDatasetQuery.PrimaryFields, " = ? AND ")),
				fmt.Sprintf("%s = ? ", strings.Join(mockDatasetQuery.PrimaryFields, " = ? AND ")),
			)}, append(mockDatasetQuery.ExtractModel(mockDataset)[:4], mockDatasetQuery.ExtractModel(mockDataset)[:4]...)...),
				append([]interface{}{fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT %s,
			%d + IFNULL((
				SELECT CASE
					WHEN timestamp_expires - %d < 0
						THEN 0
					ELSE timestamp_expires - %d END
				FROM %s
				WHERE %s AND latest = 1
				ORDER BY height DESC LIMIT 1
			), 0),
			1
		WHERE NOT EXISTS (
			SELECT %s FROM %s
			WHERE %s
		)
	`,
					mockDatasetQuery.getTableName(),
					strings.Join(mockDatasetQuery.getFields(), ", "),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(mockDatasetQuery.getFields()[:6])-1)),
					mockDataset.GetTimestampExpires(),
					mockDataset.GetTimestampStarts(),
					mockDataset.GetTimestampStarts(),
					mockDatasetQuery.getTableName(),
					fmt.Sprintf("%s != ? ", strings.Join(mockDatasetQuery.PrimaryFields, " = ? AND ")),
					mockDatasetQuery.PrimaryFields[0],
					mockDatasetQuery.getTableName(),
					fmt.Sprintf("%s = ? ", strings.Join(mockDatasetQuery.PrimaryFields, " = ? AND ")),
				)}, append(mockDatasetQuery.ExtractModel(mockDataset)[:6],
					append(mockDatasetQuery.ExtractModel(mockDataset)[:4], mockDatasetQuery.ExtractModel(mockDataset)[:4]...)...)...),
				append([]interface{}{fmt.Sprintf(
					"UPDATE %s SET latest = false WHERE %s AND latest = 1",
					mockDatasetQuery.getTableName(),
					fmt.Sprintf("%s != ? ", strings.Join(mockDatasetQuery.PrimaryFields, " = ? AND ")), // where clause
				)},
					mockDatasetQuery.ExtractModel(mockDataset)[:4]...,
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mockDatasetQuery.AddDataset(tt.args.dataset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetsQuery.AddDataset() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetsQuery_ExtractModel(t *testing.T) {
	type args struct {
		dataset *model.Dataset
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "success",
			args: args{
				dataset: mockDataset,
			},
			want: []interface{}{
				mockDataset.SetterAccountAddress,
				mockDataset.RecipientAccountAddress,
				mockDataset.Property,
				mockDataset.Height,
				mockDataset.Value,
				mockDataset.TimestampStarts,
				mockDataset.TimestampExpires,
				mockDataset.Latest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mockDatasetQuery.ExtractModel(tt.args.dataset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetsQuery.ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountDatasetsQuery_BuildModel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows([]string{
			"SetterAccountAddress", "RecipientAccountAddress", "Property", "Height", "Value", "TimestampStarts", "TimestampExpires", "Latest"}).
			AddRow(
				mockDataset.SetterAccountAddress,
				mockDataset.RecipientAccountAddress,
				mockDataset.Property,
				mockDataset.Height,
				mockDataset.Value,
				mockDataset.TimestampStarts,
				mockDataset.TimestampExpires,
				mockDataset.Latest,
			))
		rows, _ := db.Query("foo")
		var tempDataset []*model.Dataset
		if got := mockDatasetQuery.BuildModel(tempDataset, rows); !reflect.DeepEqual(got[0], mockDataset) {
			t.Errorf("AccountDatasetsQuery.BuildModel() = \n%v want \n%v", got, mockDataset)
		}
	})
}
