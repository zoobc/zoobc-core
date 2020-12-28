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
				[]byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
					28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
						NodeID:        1,
						NodePublicKey: []byte{1, 2},
						AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
							28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
			[]byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
				28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
					NodePublicKey: []byte{1, 1},
					AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
						28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
					RegistrationHeight: 1,
				},
			},
			want: &model.GetNodeRegistrationResponse{
				NodeRegistration: &model.NodeRegistration{
					NodeID:        1,
					NodePublicKey: []byte{1, 1},
					AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
						28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
						NodeID:        1,
						NodePublicKey: []byte{1, 2},
						AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
							28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
				[]byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
					28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
				params: &model.GetNodeRegistrationsByNodePublicKeysRequest{
					NodePublicKeys: [][]byte{
						{1, 2, 3},
						{3, 2, 1},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryGetNodeRegistrationsByNodePublicKeysSuccess{},
			},
			args: args{
				params: &model.GetNodeRegistrationsByNodePublicKeysRequest{
					NodePublicKeys: [][]byte{
						{1, 2},
					},
				},
			},
			want: &model.GetNodeRegistrationsByNodePublicKeysResponse{
				NodeRegistrations: []*model.NodeRegistration{
					{
						NodeID: 1,
						NodePublicKey: []byte{
							1, 2,
						},
						AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
							28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
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
