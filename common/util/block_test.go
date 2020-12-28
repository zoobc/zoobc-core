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
package util

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockBlockData = model.Block{
		Version:              uint32(1),
		PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
		BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
		Timestamp:            int64(15875592),
		TotalAmount:          int64(0),
		TotalFee:             int64(0),
		TotalCoinBase:        int64(0),
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		CumulativeDifficulty: "355353517378119",
		BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	}
)

func TestGetBlockByte(t *testing.T) {
	type args struct {
		block *model.Block
		sign  bool
		ct    chaintype.ChainType
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetBlockByte:one",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
				},
				sign: false,
				ct:   &chaintype.MainChain{},
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125,
				75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 2, 65, 76, 32, 76, 12,
				12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
			wantErr: false,
		},
		{
			name: "GetBlockByte:withSignature",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					BlockSignature:       []byte{1, 3, 4, 54, 65, 76, 3, 3, 54, 12, 5, 64, 23, 12, 21},
				},
				sign: true,
				ct:   &chaintype.MainChain{},
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108,
				229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134,
				144, 246, 37, 144, 213, 135, 2, 65, 76, 32, 76, 12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89,
				86, 3, 6, 22, 1, 3, 4, 54, 65, 76, 3, 3, 54, 12, 5, 64, 23, 12, 21},
			wantErr: false,
		},
		{
			name: "GetBlockByte:error-{sign true without signature}",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
				},
				sign: true,
				ct:   &chaintype.MainChain{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockByte(tt.args.block, tt.args.sign, tt.args.ct)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockExecutorGetLastBlockSuccess struct {
		query.Executor
	}
	mockExecutorGetLastBlockNoRow struct {
		query.Executor
	}
	mockExecutorGetLastBlockFail struct {
		query.Executor
	}
)

func (*mockExecutorGetLastBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows([]string{"fake"})
	mockRows.AddRow("1")
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorGetLastBlockNoRow) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorGetLastBlockSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mockRows.AddRow(
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.QueryRow(""), nil
}

func TestGetLastBlock(t *testing.T) {
	type args struct {
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "wantFail:Fail",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockFail{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:NoRow",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockNoRow{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockSuccess{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    &mockBlockData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetLastBlock(tt.args.queryExecutor, tt.args.blockQuery)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestUtil_getMinRollbackHeight(t *testing.T) {
	type fields struct {
		Height uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
		{
			name: "GetMinRollbackHeight Successful",
			fields: fields{
				Height: constant.MinRollbackBlocks - 1,
			},
			want: 0,
		},
		{
			name: "GetMinROllbackHeight Failed",
			fields: fields{
				Height: constant.MinRollbackBlocks + 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMinRollbackHeight(tt.fields.Height)
			if got != tt.want {
				t.Errorf("Service.getMinRollbackHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockedQueryExecutorGetBlockByHeightNoRows struct {
		query.Executor
	}
	mockedQueryExecutorGetBlockByHeightSuccess struct {
		query.Executor
	}
)

func (*mockedQueryExecutorGetBlockByHeightNoRows) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields))
	return db.QueryRow(qe), nil
}

func (*mockedQueryExecutorGetBlockByHeightSuccess) ExecuteSelectRow(qe string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields).
		AddRow(
			mockBlockData.GetHeight(),
			int64(100),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qe), nil
}

func TestGetBlockByHeight(t *testing.T) {
	type args struct {
		height        uint32
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "WantErr:NoRows",
			args: args{
				height:        100,
				queryExecutor: &mockedQueryExecutorGetBlockByHeightNoRows{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			wantErr: true,
		},
		{
			name: "WantSuccess",
			args: args{
				height:        100,
				queryExecutor: &mockedQueryExecutorGetBlockByHeightSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
				ID:                   100,
				Version:              uint32(1),
				PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
				BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
				Timestamp:            int64(15875592),
				TotalAmount:          int64(0),
				TotalFee:             int64(0),
				TotalCoinBase:        int64(0),
				PayloadHash:          []byte{},
				CumulativeDifficulty: "355353517378119",
				BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockByHeight(tt.args.height, tt.args.queryExecutor, tt.args.blockQuery)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockByHeight() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
