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

func (*mockQueryGetNodeRegistrationsFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("want error")
}

func (*mockQueryGetNodeRegistrationsSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows(query.NewNodeRegistrationQuery().Fields).
			AddRow(
				1,
				[]byte{1, 2},
				"AccountA",
				1,
				1,
				uint32(model.NodeRegistrationState_NodeQueued),
				true,
				1,
			),
		)
	return db.Query("")
}

func (*mockQueryGetNodeRegistrationsSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch strings.Contains(qStr, "total_record") {
	case true:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	default:
		return nil, nil
	}
	return db.QueryRow(qStr), nil
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
						NodeID:             1,
						NodePublicKey:      []byte{1, 2},
						AccountAddress:     "AccountA",
						RegistrationHeight: 1,
						LockedBalance:      1,
						RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
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
				t.Errorf("NodeRegistryService.GetNodeRegistrations() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryService.GetNodeRegistrations() = \n%v, want \n%v", got, tt.want)
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

func (*mockQueryGetNodeRegistrationFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, _, _ := sqlmock.New()
	return db.QueryRow(query), nil
}

func (*mockQueryGetNodeRegistrationSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewNodeRegistrationQuery().Fields,
		).AddRow(
			1,
			[]byte{1, 1},
			"AccountA",
			1,
			1,
			uint32(model.NodeRegistrationState_NodeQueued),
			true,
			1,
		))
	return db.QueryRow(qStr), nil
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
					RegistrationHeight: 1,
				},
			},
			want: &model.GetNodeRegistrationResponse{
				NodeRegistration: &model.NodeRegistration{
					NodeID:             1,
					NodePublicKey:      []byte{1, 1},
					AccountAddress:     "AccountA",
					RegistrationHeight: 1,
					LockedBalance:      1,
					RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
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
				t.Errorf("NodeRegistryService.GetNodeRegistration() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistryService_GetPendingNodeRegistrations(t *testing.T) {
	type fields struct {
		Query                 query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		req *model.GetPendingNodeRegistrationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingNodeRegistrationsResponse
		wantErr bool
	}{
		{
			name: "wantError",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsFail{},
			},
			args: args{
				req: &model.GetPendingNodeRegistrationsRequest{
					Limit: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsSuccess{},
			},
			args: args{
				req: &model.GetPendingNodeRegistrationsRequest{
					Limit: 1,
				},
			},
			want: &model.GetPendingNodeRegistrationsResponse{
				NodeRegistrations: []*model.NodeRegistration{
					{
						NodeID:             1,
						NodePublicKey:      []byte{1, 2},
						AccountAddress:     "AccountA",
						RegistrationHeight: 1,
						LockedBalance:      1,
						RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
						Latest:             true,
						Height:             1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NodeRegistryService{
				Query:                 tt.fields.Query,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			got, err := ns.GetPendingNodeRegistrations(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryService.GetPendingNodeRegistrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryService.GetPendingNodeRegistrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetNodeRegistrationsByNodePublicKeysFail struct {
		query.Executor
	}
	mockQueryGetNodeRegistrationsByNodePublicKeysSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetNodeRegistrationsByNodePublicKeysFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}

func (*mockQueryGetNodeRegistrationsByNodePublicKeysSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows(query.NewNodeRegistrationQuery().Fields).
			AddRow(
				1,
				[]byte{1, 2},
				"AccountA",
				1,
				1,
				uint32(model.NodeRegistrationState_NodeQueued),
				true,
				1,
			),
		)
	return db.Query("")
}

func TestNodeRegistryService_GetNodeRegistrationsByNodePublicKeys(t *testing.T) {
	type fields struct {
		Query                 query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		params *model.GetNodeRegistrationsByNodePublicKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationsByNodePublicKeysResponse
		wantErr bool
	}{
		{
			name: "wantError",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsByNodePublicKeysFail{},
			},
			args: args{
				params: &model.GetNodeRegistrationsByNodePublicKeysRequest{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsByNodePublicKeysSuccess{},
			},
			args: args{
				params: &model.GetNodeRegistrationsByNodePublicKeysRequest{},
			},
			want: &model.GetNodeRegistrationsByNodePublicKeysResponse{
				NodeRegistrations: []*model.NodeRegistration{
					{
						NodeID:             1,
						NodePublicKey:      []byte{1, 2},
						AccountAddress:     "AccountA",
						RegistrationHeight: 1,
						LockedBalance:      1,
						RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
						Latest:             true,
						Height:             1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NodeRegistryService{
				Query:                 tt.fields.Query,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			got, err := ns.GetNodeRegistrationsByNodePublicKeys(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryService.GetNodeRegistrationsByNodePublicKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryService.GetNodeRegistrationsByNodePublicKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
