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
package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetNodeRegistrationsFailed struct {
		service.NodeRegistryServiceInterface
	}
	mockGetNodeRegistrationsSuccess struct {
		service.NodeRegistryServiceInterface
	}
)

func (*mockGetNodeRegistrationsFailed) GetNodeRegistrations(*model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	return nil, errors.New("Error GetNodeRegistrations")
}
func (*mockGetNodeRegistrationsSuccess) GetNodeRegistrations(*model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	return &model.GetNodeRegistrationsResponse{}, nil
}

func TestNodeRegistryHandler_GetNodeRegistrations(t *testing.T) {
	type fields struct {
		Service service.NodeRegistryServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeRegistrationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationsResponse
		wantErr bool
	}{
		{
			name: "GetNodeRegistrations:Error",
			fields: fields{
				Service: &mockGetNodeRegistrationsFailed{},
			},
			args: args{
				req: &model.GetNodeRegistrationsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistrations:Success",
			fields: fields{
				Service: &mockGetNodeRegistrationsSuccess{},
			},
			args: args{
				req: &model.GetNodeRegistrationsRequest{},
			},
			want:    &model.GetNodeRegistrationsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrh := NodeRegistryHandler{
				Service: tt.fields.Service,
			}
			got, err := nrh.GetNodeRegistrations(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNodeRegistrationFailed struct {
		service.NodeRegistryServiceInterface
	}
	mockGetNodeRegistrationSuccess struct {
		service.NodeRegistryServiceInterface
	}
)

func (*mockGetNodeRegistrationFailed) GetNodeRegistration(*model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	return nil, errors.New("Error GetNodeRegistration")
}

func (*mockGetNodeRegistrationSuccess) GetNodeRegistration(*model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	return &model.GetNodeRegistrationResponse{}, nil
}

func TestNodeRegistryHandler_GetNodeRegistration(t *testing.T) {
	type fields struct {
		Service service.NodeRegistryServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeRegistrationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationResponse
		wantErr bool
	}{
		{
			name: "GetNodeRegistration:Failed",
			fields: fields{
				Service: &mockGetNodeRegistrationFailed{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistration:Success",
			fields: fields{
				Service: &mockGetNodeRegistrationSuccess{},
			},
			args: args{
				req: &model.GetNodeRegistrationRequest{},
			},
			want:    &model.GetNodeRegistrationResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrh := NodeRegistryHandler{
				Service: tt.fields.Service,
			}
			got, err := nrh.GetNodeRegistration(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryHandler.GetNodeRegistration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryHandler.GetNodeRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNodeRegistrationsByNodePublicKeysError struct {
		service.NodeRegistryServiceInterface
	}
	mockGetNodeRegistrationsByNodePublicKeysSuccess struct {
		service.NodeRegistryServiceInterface
	}
)

func (*mockGetNodeRegistrationsByNodePublicKeysError) GetNodeRegistrationsByNodePublicKeys(*model.GetNodeRegistrationsByNodePublicKeysRequest,
) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error) {
	return nil, errors.New("Error GetNodeRegistrationsByNodePublicKeys")
}
func (*mockGetNodeRegistrationsByNodePublicKeysSuccess) GetNodeRegistrationsByNodePublicKeys(*model.GetNodeRegistrationsByNodePublicKeysRequest,
) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error) {
	return &model.GetNodeRegistrationsByNodePublicKeysResponse{}, nil
}

func TestNodeRegistryHandler_GetNodeRegistrationsByNodePublicKeys(t *testing.T) {
	type fields struct {
		Service service.NodeRegistryServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeRegistrationsByNodePublicKeysRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeRegistrationsByNodePublicKeysResponse
		wantErr bool
	}{
		{
			name: "GetNodeRegistrationsByNodePublicKeys:InvalidArgument",
			args: args{
				req: &model.GetNodeRegistrationsByNodePublicKeysRequest{
					NodePublicKeys: [][]byte{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistrationsByNodePublicKeys:Error",
			args: args{
				req: &model.GetNodeRegistrationsByNodePublicKeysRequest{
					NodePublicKeys: [][]byte{
						{1, 0},
						{1, 0, 1},
					},
				},
			},
			fields: fields{
				Service: &mockGetNodeRegistrationsByNodePublicKeysError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNodeRegistrationsByNodePublicKeys:Success",
			args: args{
				req: &model.GetNodeRegistrationsByNodePublicKeysRequest{
					NodePublicKeys: [][]byte{
						{1, 0},
						{1, 0, 1},
					},
				},
			},
			fields: fields{
				Service: &mockGetNodeRegistrationsByNodePublicKeysSuccess{},
			},
			want:    &model.GetNodeRegistrationsByNodePublicKeysResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrh := NodeRegistryHandler{
				Service: tt.fields.Service,
			}
			got, err := nrh.GetNodeRegistrationsByNodePublicKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrationsByNodePublicKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistryHandler.GetNodeRegistrationsByNodePublicKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
