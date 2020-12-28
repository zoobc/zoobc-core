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
	mockGetAccountBalanceError struct {
		service.AccountBalanceServiceInterface
	}
	mockGetAccountBalanceSuccess struct {
		service.AccountBalanceServiceInterface
	}
)

func (*mockGetAccountBalanceError) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	return nil, errors.New("error GetAccountBalance")
}
func (*mockGetAccountBalanceSuccess) GetAccountBalance(request *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error) {
	return &model.GetAccountBalanceResponse{
		AccountBalance: &model.AccountBalance{
			AccountAddress: request.AccountAddress,
		},
	}, nil
}

func TestAccountBalanceHandler_GetAccountBalance(t *testing.T) {
	type fields struct {
		Service service.AccountBalanceServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetAccountBalanceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalanceResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalance:fail",
			fields: fields{
				Service: &mockGetAccountBalanceError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalance:success",
			fields: fields{
				Service: &mockGetAccountBalanceSuccess{},
			},
			args: args{
				request: &model.GetAccountBalanceRequest{
					AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140,
						183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			want: &model.GetAccountBalanceResponse{
				AccountBalance: &model.AccountBalance{
					AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140,
						183, 229, 116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHandler{
				Service: tt.fields.Service,
			}
			got, err := abh.GetAccountBalance(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceHandler.GetAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceHandler.GetAccountBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetAccountBalancesSuccess struct {
		service.AccountBalanceServiceInterface
	}
)

func (*mockGetAccountBalancesSuccess) GetAccountBalances(request *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error) {
	return &model.GetAccountBalancesResponse{
		AccountBalances: []*model.AccountBalance{},
	}, nil
}

func TestAccountBalanceHandler_GetAccountBalances(t *testing.T) {
	type fields struct {
		Service service.AccountBalanceServiceInterface
	}
	type args struct {
		ctx     context.Context
		request *model.GetAccountBalancesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetAccountBalancesResponse
		wantErr bool
	}{
		{
			name: "GetAccountBalancesHandler:fail",
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: [][]byte{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetAccountBalancesHandler:success",
			args: args{
				request: &model.GetAccountBalancesRequest{
					AccountAddresses: [][]byte{
						{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
							116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
					},
				},
			},
			fields: fields{
				Service: &mockGetAccountBalancesSuccess{},
			},
			want: &model.GetAccountBalancesResponse{
				AccountBalances: []*model.AccountBalance{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abh := &AccountBalanceHandler{
				Service: tt.fields.Service,
			}
			got, err := abh.GetAccountBalances(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBalanceHandler.GetAccountBalances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalanceHandler.GetAccountBalances() = %v, want %v", got, tt.want)
			}
		})
	}
}
