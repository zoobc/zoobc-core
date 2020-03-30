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

func (*mockGetAccountDatasetsExecutor) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1))
	return db.QueryRow(""), nil
}
func (*mockGetAccountDatasetsExecutor) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(query.NewAccountDatasetsQuery().GetFields())
	mockRows.AddRow(
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"AccountDatasetEscrowApproval",
		5,
		"Message",
		1565942932686,
		1565943056129,
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)

	return db.Query("")
}

func TestAccountDatasetService_GetAccountDatasets(t *testing.T) {
	type fields struct {
		AccountDatasetQuery *query.AccountDatasetsQuery
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
					RecipientAccountAddress: "BCZAbcasdljasd_123876123",
					SetterAccountAddress:    "",
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
						SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
						RecipientAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						Property:                "AccountDatasetEscrowApproval",
						Value:                   "Message",
						TimestampStarts:         1565942932686,
						TimestampExpires:        1565943056129,
						Height:                  5,
						Latest:                  true,
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

	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().GetFields())
	mockRow.AddRow(
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"AccountDatasetEscrowApproval",
		5,
		"Message",
		1565942932686,
		1565943056129,
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow(""), nil
}
func (*mockExecutorGetAccountDatasetErr) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().GetFields())
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow(""), nil
}

func TestAccountDatasetService_GetAccountDataset(t *testing.T) {
	type fields struct {
		AccountDatasetQuery *query.AccountDatasetsQuery
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
				SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				RecipientAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Property:                "AccountDatasetEscrowApproval",
				Value:                   "Message",
				TimestampStarts:         1565942932686,
				TimestampExpires:        1565943056129,
				Height:                  5,
				Latest:                  true,
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
