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
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetTransactionError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionError) GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error) {
	return nil, errors.New("Error GetTransaction")
}
func (*mockGetTransactionSuccess) GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}

func TestTransactionHandler_GetTransaction(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransaction:Failed",
			fields: fields{
				Service: &mockGetTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransaction:Success",
			fields: fields{
				Service: &mockGetTransactionSuccess{},
			},
			want:    &model.Transaction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetTransactionsError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionsSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionsError) GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	return nil, errors.New("Error GetTransactions")
}

func (*mockGetTransactionsSuccess) GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	return &model.GetTransactionsResponse{}, nil
}

func TestTransactionHandler_GetTransactions(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetTransactions:Failed",
			fields: fields{
				Service: &mockGetTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactions:Success",
			fields: fields{
				Service: &mockGetTransactionsSuccess{},
			},
			want:    &model.GetTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockPostTransactionError struct {
		service.TransactionServiceInterface
	}
	mockPostTransactionSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockPostTransactionError) PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error) {
	return nil, errors.New("Error PostTransaction")
}
func (*mockPostTransactionSuccess) PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}

func TestTransactionHandler_PostTransaction(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.PostTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.PostTransactionResponse
		wantErr bool
	}{
		{
			name: "PostTransaction:Failed",
			fields: fields{
				Service: &mockPostTransactionError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "PostTransaction:Success",
			fields: fields{
				Service: &mockPostTransactionSuccess{},
			},
			want: &model.PostTransactionResponse{
				Transaction: &model.Transaction{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.PostTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.PostTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.PostTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetTransactionMinimumFeeError struct {
		service.TransactionServiceInterface
	}
	mockGetTransactionMinimumFeeSuccess struct {
		service.TransactionServiceInterface
	}
)

func (*mockGetTransactionMinimumFeeError) GetTransactionMinimumFee(request *model.GetTransactionMinimumFeeRequest,
) (*model.GetTransactionMinimumFeeResponse, error) {
	return nil, errors.New("Error GetTransactionMinimumFee")
}
func (*mockGetTransactionMinimumFeeSuccess) GetTransactionMinimumFee(request *model.GetTransactionMinimumFeeRequest,
) (*model.GetTransactionMinimumFeeResponse, error) {
	return &model.GetTransactionMinimumFeeResponse{}, nil
}

func TestTransactionHandler_GetTransactionMinimumFee(t *testing.T) {
	type fields struct {
		Service service.TransactionServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetTransactionMinimumFeeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetTransactionMinimumFeeResponse
		wantErr bool
	}{

		{
			name: "GetTransactionMinimumFee:Failed",
			fields: fields{
				Service: &mockGetTransactionMinimumFeeError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionMinimumFee:Success",
			fields: fields{
				Service: &mockGetTransactionMinimumFeeSuccess{},
			},
			want:    &model.GetTransactionMinimumFeeResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHandler{
				Service: tt.fields.Service,
			}
			got, err := th.GetTransactionMinimumFee(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionHandler.GetTransactionMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionHandler.GetTransactionMinimumFee() = %v, want %v", got, tt.want)
			}
		})
	}
}
