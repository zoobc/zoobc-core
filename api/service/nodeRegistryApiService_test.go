package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewNodeRegistryService(t *testing.T) {
	type args struct {
		queryExecutor query.ExecutorInterface
	}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	tests := []struct {
		name string
		args args
		want *NodeRegistryService
	}{
		{
			name: "wantSuccess",
			args: args{
				queryExecutor: query.NewQueryExecutor(db),
			},
			want: &NodeRegistryService{
				Query: query.NewQueryExecutor(db),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeRegistryService(tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeRegistryService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetNodeRegistrationsFail struct {
		query.Executor
	}
	mockQueryGetNodeRegistrationsSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetNodeRegistrationsFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}

func (*mockQueryGetNodeRegistrationsSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "total_record") {
	case true:
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	default:
		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(query.NewNodeRegistrationQuery().Fields[1:]).
				AddRow(
					[]byte{1, 2},
					"AccountA",
					1,
					"127.0.0.1",
					1,
					true,
					1,
					1,
				),
			)
	}
	return db.Query("")
}

func TestNodeRegistryService_GetNodeRegistrations(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		params *model.GetNodeRegistrationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationsResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsSuccess{},
			},
			args: args{
				params: &model.GetNodeRegistrationsRequest{
					MaxRegistrationHeight: 1,
				},
			},
			want: &model.GetNodeRegistrationsResponse{
				Total: 1,
				NodeRegistrations: []*model.NodeRegistration{
					{
						NodePublicKey:      []byte{1, 2},
						AccountAddress:     "AccountA",
						RegistrationHeight: 1,
						NodeAddress:        "127.0.0.1",
						LockedBalance:      1,
						Queued:             true,
						Latest:             true,
						Height:             1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wantFail",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsFail{},
			},
			args: args{
				params: &model.GetNodeRegistrationsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NodeRegistryService{
				Query: tt.fields.Query,
			}
			got, err := ns.GetNodeRegistrations(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryService.GetNodeRegistrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryService.GetNodeRegistrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetNodeRegistrationFail struct {
		query.Executor
	}
	mockQueryGetNodeRegistrationSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetNodeRegistrationFail) ExecuteSelectRow(query string, args ...interface{}) *sql.Row {
	db, _, _ := sqlmock.New()
	return db.QueryRow(query)
}

func (*mockQueryGetNodeRegistrationSuccess) ExecuteSelectRow(query string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{
			"NodePublicKey",
			"AccountAddress",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"Queued",
			"Latest",
			"Height",
		}).AddRow([]byte{1, 1}, "AccountA", 1, "127.0.0.1", 1, true, true, 1))
	return db.QueryRow(query)
}

func TestNodeRegistryService_GetNodeRegistration(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		params *model.GetNodeRegistrationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationSuccess{},
			},
			args: args{
				params: &model.GetNodeRegistrationRequest{
					NodePublicKey:      []byte{1, 1},
					AccountAddress:     "AccountA",
					NodeAddress:        "127.0.0.1",
					RegistrationHeight: 1,
				},
			},
			want: &model.GetNodeRegistrationResponse{
				NodeRegistration: &model.NodeRegistration{
					NodePublicKey:      []byte{1, 1},
					AccountAddress:     "AccountA",
					RegistrationHeight: 1,
					NodeAddress:        "127.0.0.1",
					LockedBalance:      1,
					Queued:             true,
					Latest:             true,
					Height:             1,
				},
			},
			wantErr: false,
		},
		{
			name: "wantFail",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationFail{},
			},
			args: args{
				params: &model.GetNodeRegistrationRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NodeRegistryService{
				Query: tt.fields.Query,
			}
			got, err := ns.GetNodeRegistration(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryService.GetNodeRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryService.GetNodeRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}
