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
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetEscrowTransactionsFailed struct {
		service.EscrowTransactionServiceInterface
	}
	mockGetEscrowTransactionsSuccess struct {
		service.EscrowTransactionServiceInterface
	}
)

func (*mockGetEscrowTransactionsFailed) GetEscrowTransactions(request *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	return nil, errors.New("Error GetEscrowTransactions")
}

func (*mockGetEscrowTransactionsSuccess) GetEscrowTransactions(request *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	return &model.GetEscrowTransactionsResponse{}, nil
}

func TestEscrowTransactionHandler_GetEscrowTransactions(t *testing.T) {
	type fields struct {
		Service service.EscrowTransactionServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetEscrowTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetEscrowTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetEscrowTransactions:Error",
			fields: fields{
				Service: &mockGetEscrowTransactionsFailed{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetEscrowTransactions:Success",
			fields: fields{
				Service: &mockGetEscrowTransactionsSuccess{},
			},
			want:    &model.GetEscrowTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := &EscrowTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := eh.GetEscrowTransactions(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetEscrowTransactionError struct {
		service.EscrowTransactionServiceInterface
	}
	mockGetEscrowTransactionSuccess struct {
		service.EscrowTransactionServiceInterface
	}
)

func (*mockGetEscrowTransactionError) GetEscrowTransaction(request *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	return nil, errors.New("Error GetEscrowTransaction")
}
func (*mockGetEscrowTransactionSuccess) GetEscrowTransaction(request *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	return &model.Escrow{}, nil
}

func TestEscrowTransactionHandler_GetEscrowTransaction(t *testing.T) {
	type fields struct {
		Service service.EscrowTransactionServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetEscrowTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Escrow
		wantErr bool
	}{
		{
			name: "GetEscrowTransaction:Error",
			fields: fields{
				Service: &mockGetEscrowTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetEscrowTransaction:Success",
			fields: fields{
				Service: &mockGetEscrowTransactionSuccess{},
			},
			want:    &model.Escrow{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := &EscrowTransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := eh.GetEscrowTransaction(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscrowTransactionHandler.GetEscrowTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
