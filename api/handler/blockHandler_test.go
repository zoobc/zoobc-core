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
	mockGetBlockError struct {
		service.BlockServiceInterface
	}
	mockGetBlockSuccess struct {
		service.BlockServiceInterface
	}
)

func (*mockGetBlockError) GetBlockByID(chainType chaintype.ChainType, id int64) (*model.GetBlockResponse, error) {
	return nil, errors.New("Error GetBlockByID")
}

func (*mockGetBlockError) GetBlockByHeight(chainType chaintype.ChainType, height uint32) (*model.GetBlockResponse, error) {
	return nil, errors.New("Error GetBlockByHeight")
}

func (*mockGetBlockSuccess) GetBlockByID(chainType chaintype.ChainType, id int64) (*model.GetBlockResponse, error) {
	return &model.GetBlockResponse{}, nil
}

func (*mockGetBlockSuccess) GetBlockByHeight(chainType chaintype.ChainType, height uint32) (*model.GetBlockResponse, error) {
	return &model.GetBlockResponse{}, nil
}

func TestBlockHandler_GetBlock(t *testing.T) {
	type fields struct {
		Service service.BlockServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetBlockRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlockResponse
		wantErr bool
	}{
		{
			name: "GetBlock:Error",
			fields: fields{
				Service: &mockGetBlockError{},
			},
			args: args{
				req: &model.GetBlockRequest{
					ID:     1,
					Height: 1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlock:Success",
			fields: fields{
				Service: &mockGetBlockSuccess{},
			},
			args: args{
				req: &model.GetBlockRequest{
					ID:     1,
					Height: 1,
				},
			},
			want:    &model.GetBlockResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockHandler{
				Service: tt.fields.Service,
			}
			got, err := bs.GetBlock(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHandler.GetBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockHandler.GetBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetBlocksError struct {
		service.BlockServiceInterface
	}
	mockGetBlocksSucess struct {
		service.BlockServiceInterface
	}
)

func (*mockGetBlocksError) GetBlocks(chainType chaintype.ChainType, count, height uint32) (*model.GetBlocksResponse, error) {
	return nil, errors.New("Error GetBlocks")
}

func (*mockGetBlocksSucess) GetBlocks(chainType chaintype.ChainType, count, height uint32) (*model.GetBlocksResponse, error) {
	return &model.GetBlocksResponse{}, nil
}

func TestBlockHandler_GetBlocks(t *testing.T) {
	type fields struct {
		Service service.BlockServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetBlocksRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlocksResponse
		wantErr bool
	}{
		{
			name: "GetBlocks:LimitExceeded",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 1000,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocks:Error",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 500,
				},
			},
			fields: fields{
				Service: &mockGetBlocksError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocks:Success",
			args: args{
				req: &model.GetBlocksRequest{
					Limit: 500,
				},
			},
			fields: fields{
				Service: &mockGetBlocksSucess{},
			},
			want:    &model.GetBlocksResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockHandler{
				Service: tt.fields.Service,
			}
			got, err := bs.GetBlocks(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHandler.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockHandler.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
