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
// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

// ResetBlockService resets the singleton back to nil, used in test case teardown
func ResetBlockService() {
	blockServiceInstance = nil
}

var (
	mockGoodBlock = model.Block{
		ID:                   1,
		BlockHash:            []byte{},
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            10000,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{3},
		CumulativeDifficulty: "1",
		PayloadLength:        1,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  []byte{},
		TotalAmount:          1000,
		TotalFee:             0,
		TotalCoinBase:        1,
		Version:              0,
	}
)

type (
	mockBlockService struct {
		coreService.BlockService
	}
	mockQueryExecutorBlockByIDFail struct {
		query.Executor
	}

	mockQueryExecutorBlockByIDNotFound struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksSuccess struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksFail struct {
		query.Executor
	}
)

func (*mockBlockService) GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error) {
	return &model.BlockExtendedInfo{}, nil
}

func (*mockQueryExecutorBlockByIDFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2))
	return db.QueryRow(query), nil
}

func (*mockQueryExecutorBlockByIDNotFound) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2))
	return db.QueryRow(qe), nil
}

func (*mockQueryExecutorGetBlocksSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.MainChain{}).Fields,
	).AddRow(
		mockGoodBlock.GetHeight(),
		mockGoodBlock.GetID(),
		mockGoodBlock.GetBlockHash(),
		mockGoodBlock.GetPreviousBlockHash(),
		mockGoodBlock.GetTimestamp(),
		mockGoodBlock.GetBlockSeed(),
		mockGoodBlock.GetBlockSignature(),
		mockGoodBlock.GetCumulativeDifficulty(),
		mockGoodBlock.GetPayloadLength(),
		mockGoodBlock.GetPayloadHash(),
		mockGoodBlock.GetBlocksmithPublicKey(),
		mockGoodBlock.GetTotalAmount(),
		mockGoodBlock.GetTotalFee(),
		mockGoodBlock.GetTotalCoinBase(),
		mockGoodBlock.GetVersion(),
		mockGoodBlock.GetMerkleRoot(),
		mockGoodBlock.GetMerkleTree(),
		mockGoodBlock.GetReferenceBlockHeight(),
	).AddRow(
		mockGoodBlock.GetHeight(),
		mockGoodBlock.GetID(),
		mockGoodBlock.GetBlockHash(),
		mockGoodBlock.GetPreviousBlockHash(),
		mockGoodBlock.GetTimestamp(),
		mockGoodBlock.GetBlockSeed(),
		mockGoodBlock.GetBlockSignature(),
		mockGoodBlock.GetCumulativeDifficulty(),
		mockGoodBlock.GetPayloadLength(),
		mockGoodBlock.GetPayloadHash(),
		mockGoodBlock.GetBlocksmithPublicKey(),
		mockGoodBlock.GetTotalAmount(),
		mockGoodBlock.GetTotalFee(),
		mockGoodBlock.GetTotalCoinBase(),
		mockGoodBlock.GetVersion(),
		mockGoodBlock.GetMerkleRoot(),
		mockGoodBlock.GetMerkleTree(),
		mockGoodBlock.GetReferenceBlockHeight(),
	))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetBlocksFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

type (
	mockQueryGetBlockByIDSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlockByIDSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
			mockGoodBlock.GetTimestamp(),
			mockGoodBlock.GetBlockSeed(),
			mockGoodBlock.GetBlockSignature(),
			mockGoodBlock.GetCumulativeDifficulty(),
			mockGoodBlock.GetPayloadLength(),
			mockGoodBlock.GetPayloadHash(),
			mockGoodBlock.GetBlocksmithPublicKey(),
			mockGoodBlock.GetTotalAmount(),
			mockGoodBlock.GetTotalFee(),
			mockGoodBlock.GetTotalCoinBase(),
			mockGoodBlock.GetVersion(),
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}
func TestBlockService_GetBlockByID(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		id        int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlockResponse
		wantErr bool
	}{
		{
			name: "GetBlockByID:success",
			fields: fields{
				Query: &mockQueryGetBlockByIDSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        mockGoodBlock.ID,
			},
			wantErr: false,
			want: &model.GetBlockResponse{
				Block: &mockGoodBlock,
			},
		},
		{
			name: "GetBlockByID:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlockByID:fail-{Block.ID notfound}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDNotFound{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlockByID(tt.args.chainType, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByID() got = \n%v, want = \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetBlockByHeightSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlockByHeightSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
			mockGoodBlock.GetTimestamp(),
			mockGoodBlock.GetBlockSeed(),
			mockGoodBlock.GetBlockSignature(),
			mockGoodBlock.GetCumulativeDifficulty(),
			mockGoodBlock.GetPayloadLength(),
			mockGoodBlock.GetPayloadHash(),
			mockGoodBlock.GetBlocksmithPublicKey(),
			mockGoodBlock.GetTotalAmount(),
			mockGoodBlock.GetTotalFee(),
			mockGoodBlock.GetTotalCoinBase(),
			mockGoodBlock.GetVersion(),
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func TestBlockService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		height    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlockResponse
		wantErr bool
	}{
		{
			name: "GetBlockByHeight:success",
			fields: fields{
				Query: &mockQueryGetBlockByHeightSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    mockGoodBlock.Height,
			},
			wantErr: false,
			want: &model.GetBlockResponse{
				Block: &mockGoodBlock,
			},
		},
		{
			name: "GetBlockByHeight:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlockByHeight:fail-{Block.ID notfound}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDNotFound{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlockByHeight(tt.args.chainType, tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetBlocksSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
			mockGoodBlock.GetTimestamp(),
			mockGoodBlock.GetBlockSeed(),
			mockGoodBlock.GetBlockSignature(),
			mockGoodBlock.GetCumulativeDifficulty(),
			mockGoodBlock.GetPayloadLength(),
			mockGoodBlock.GetPayloadHash(),
			mockGoodBlock.GetBlocksmithPublicKey(),
			mockGoodBlock.GetTotalAmount(),
			mockGoodBlock.GetTotalFee(),
			mockGoodBlock.GetTotalCoinBase(),
			mockGoodBlock.GetVersion(),
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		),
	)
	return db.Query(qStr)
}

func TestBlockService_GetBlocks(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		blockSize uint32
		height    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlocksResponse
		wantErr bool
	}{
		{
			name: "GetBlocks:success",
			fields: fields{
				Query: &mockQueryGetBlocksSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				blockSize: 2,
				height:    1,
			},
			want: &model.GetBlocksResponse{
				Height: 1,
				Count:  1,
				Blocks: []*model.Block{
					&mockGoodBlock,
				},
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:success",
			fields: fields{
				Query: &mockQueryExecutorGetBlocksFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				blockSize: 2,
				height:    1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlocks(tt.args.chainType, tt.args.blockSize, tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlockService(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	type args struct {
		queryExecutor     query.ExecutorInterface
		blockCoreServices map[int32]coreService.BlockServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "NewBlockService:InitiateBlockServiceInstance",
			args: args{
				queryExecutor: query.NewQueryExecutor(db),
				blockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &coreService.BlockService{},
				},
			},
			want: &BlockService{
				Query: query.NewQueryExecutor(db),
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &coreService.BlockService{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(tt.args.queryExecutor, tt.args.blockCoreServices, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
		})
	}
}
