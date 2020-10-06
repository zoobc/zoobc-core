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

	mockRows := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mockRows.AddRow(
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
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
					Property: "AccountDatasetEscrowApproval",
					Value:    "Message",
					RecipientAccountAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75, 213,
						137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
					SetterAccountAddress: nil,
					Height:               0,
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
						SetterAccountAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
							213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
						RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
							72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
						Property: "AccountDatasetEscrowApproval",
						Value:    "Message",
						Height:   5,
						Latest:   true,
						IsActive: true,
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
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
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
				SetterAccountAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
					213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
				RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
					72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				Property: "AccountDatasetEscrowApproval",
				Value:    "Message",
				Height:   5,
				Latest:   true,
				IsActive: true,
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
