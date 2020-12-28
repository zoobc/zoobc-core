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
	mockGetAccountDatasetsError struct {
		service.AccountDatasetServiceInterface
	}
	mockGetAccountDatasetsSuccess struct {
		service.AccountDatasetServiceInterface
	}
)

func (*mockGetAccountDatasetsError) GetAccountDatasets(request *model.GetAccountDatasetsRequest) (*model.GetAccountDatasetsResponse, error) {
	return nil, errors.New("Error GetAccountDatasets")
}
func (*mockGetAccountDatasetsSuccess) GetAccountDatasets(request *model.GetAccountDatasetsRequest) (*model.GetAccountDatasetsResponse, error) {
	return &model.GetAccountDatasetsResponse{}, nil
}

func TestAccountDatasetHandler_GetAccountDatasets(t *testing.T) {
	type fields struct {
		Service service.AccountDatasetServiceInterface
	}
	type args struct {
		in0     context.Context
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
			name: "GetAccountDatasets:LimitExceeded",
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(600),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDatasets:Error",
			fields: fields{
				Service: &mockGetAccountDatasetsError{},
			},
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(250),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDatasets:Success",
			fields: fields{
				Service: &mockGetAccountDatasetsSuccess{},
			},
			args: args{
				request: &model.GetAccountDatasetsRequest{
					Pagination: &model.Pagination{
						Limit: uint32(250),
					},
				},
			},
			want:    &model.GetAccountDatasetsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adh := &AccountDatasetHandler{
				Service: tt.fields.Service,
			}
			got, err := adh.GetAccountDatasets(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountDatasetHandler.GetAccountDatasets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetHandler.GetAccountDatasets() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetAccountDatasetError struct {
		service.AccountDatasetServiceInterface
	}
	mockGetAccountDatasetSuccess struct {
		service.AccountDatasetServiceInterface
	}
)

func (*mockGetAccountDatasetError) GetAccountDataset(request *model.GetAccountDatasetRequest) (*model.AccountDataset, error) {
	return nil, errors.New("Error GetAccountDataset")
}
func (*mockGetAccountDatasetSuccess) GetAccountDataset(request *model.GetAccountDatasetRequest) (*model.AccountDataset, error) {
	return &model.AccountDataset{}, nil
}

func TestAccountDatasetHandler_GetAccountDataset(t *testing.T) {
	type fields struct {
		Service service.AccountDatasetServiceInterface
	}
	type args struct {
		in0     context.Context
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
			name: "GetAccountDataset:InvalidRequest",
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: nil,
					Property:                "",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDataset:Error",
			fields: fields{
				Service: &mockGetAccountDatasetError{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112,
						113, 140, 183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountDataset:Success",
			fields: fields{
				Service: &mockGetAccountDatasetSuccess{},
			},
			args: args{
				request: &model.GetAccountDatasetRequest{
					RecipientAccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112,
						113, 140, 183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want:    &model.AccountDataset{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adh := &AccountDatasetHandler{
				Service: tt.fields.Service,
			}
			got, err := adh.GetAccountDataset(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountDatasetHandler.GetAccountDataset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountDatasetHandler.GetAccountDataset() = %v, want %v", got, tt.want)
			}
		})
	}
}
