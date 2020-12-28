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
	mockGetPendingTransactionsError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionsSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionsError) GetPendingTransactions(param *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	return nil, errors.New("Error GetPendingTransactions")
}

func (*mockGetPendingTransactionsSuccess) GetPendingTransactions(param *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	return &model.GetPendingTransactionsResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactions(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactions:ErrorPageLessThanOne",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Error",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Success",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsSuccess{},
			},
			want:    &model.GetPendingTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetPendingTransactionDetailByTransactionHashError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionDetailByTransactionHashSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionDetailByTransactionHashError) GetPendingTransactionDetailByTransactionHash(
	param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return nil, errors.New("Error GetPendingTransactionDetailByTransactionHash")
}

func (*mockGetPendingTransactionDetailByTransactionHashSuccess) GetPendingTransactionDetailByTransactionHash(
	param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return &model.GetPendingTransactionDetailByTransactionHashResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactionDetailByTransactionHash(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionDetailByTransactionHashRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionDetailByTransactionHashResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionDetailByTransactionHash:Error",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash:Success",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashSuccess{},
			},
			want:    &model.GetPendingTransactionDetailByTransactionHashResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactionDetailByTransactionHash(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetMultisignatureInfoError struct {
		service.MultisigServiceInterface
	}
	mockGetMultisignatureInfoSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetMultisignatureInfoError) GetMultisignatureInfo(param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	return nil, errors.New("Error GetMultisignatureInfo")
}

func (*mockGetMultisignatureInfoSuccess) GetMultisignatureInfo(param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	return &model.GetMultisignatureInfoResponse{}, nil
}

func TestMultisigHandler_GetMultisignatureInfo(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetMultisignatureInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisignatureInfoResponse
		wantErr bool
	}{
		{
			name: "GetMultisignatureInfo:ErrorPageLessThanOne",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:ErrorLimitMoreThan30",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 31,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:Error",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetMultisignatureInfoError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:Success",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetMultisignatureInfoSuccess{},
			},
			want:    &model.GetMultisignatureInfoResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetMultisignatureInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetMultisignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetMultisignatureInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetMultisigAddressByParticipantAddressError struct {
		service.MultisigServiceInterface
	}
	mockGetMultisigAddressByParticipantAddressSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetMultisigAddressByParticipantAddressError,
) GetMultisigAddressByParticipantAddress(param *model.GetMultisigAddressByParticipantAddressRequest,
) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	return nil, errors.New("Error GetMultisigAddressByParticipantAddress")
}
func (*mockGetMultisigAddressByParticipantAddressSuccess,
) GetMultisigAddressByParticipantAddress(param *model.GetMultisigAddressByParticipantAddressRequest,
) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	return &model.GetMultisigAddressByParticipantAddressResponse{}, nil
}

func TestMultisigHandler_GetMultisigAddressByParticipantAddress(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetMultisigAddressByParticipantAddressRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisigAddressByParticipantAddressResponse
		wantErr bool
	}{
		{
			name: "GetMultisigAddressByParticipantAddress:Error",
			fields: fields{
				MultisigService: &mockGetMultisigAddressByParticipantAddressError{},
			},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisigAddressByParticipantAddress:Success",
			fields: fields{
				MultisigService: &mockGetMultisigAddressByParticipantAddressSuccess{},
			},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressRequest{},
			},
			want:    &model.GetMultisigAddressByParticipantAddressResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetMultisigAddressByParticipantAddress(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetMultisigAddressByParticipantAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetMultisigAddressByParticipantAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
