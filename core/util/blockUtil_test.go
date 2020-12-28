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
// util package contain basic utilities commonly used across the core package
package util

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetBlockSeed(t *testing.T) {
	resultOne := int64(7942030951238827391)
	resulTwo := int64(5467201322837561108)
	type args struct {
		publicKey    []byte
		nodeID       int64
		block        *model.Block
		secretPhrase string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GetBlockSeed:one",
			args: args{
				publicKey: []byte{1, 2, 3, 4},
				nodeID:    10,
				block: &model.Block{
					BlockSeed: []byte{1, 3, 4, 5},
				},
				secretPhrase: "randomsecretphrase",
			},
			want:    resultOne,
			wantErr: false,
		},
		{
			name: "GetBlockSeed:two",
			args: args{
				nodeID: 20,
				publicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
					255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				block: &model.Block{
					BlockSeed: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229},
				},
				secretPhrase: "randomsecretphrase",
			},
			want:    resulTwo,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockSeed(tt.args.nodeID, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlockID(t *testing.T) {
	type args struct {
		block *model.Block
		ct    chaintype.ChainType
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "GetBlockID:one",
			args: args{
				block: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithPublicKey:  []byte{},
					Timestamp:            15875392,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "341353517378119",
					BlockSignature:       []byte{},
				},
				ct: &chaintype.MainChain{},
			},
			want: -4663951010383348858,
		},
		{
			name: "GetBlockID:two",
			args: args{
				block: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey:  []byte{},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					BlockSignature:       []byte{},
				},
				ct: &chaintype.MainChain{},
			},
			want: -7100824243827680979,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBlockID(tt.args.block, tt.args.ct); got != tt.want {
				t.Errorf("GetBlockID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBlockIDExist(t *testing.T) { //todo:update test after applying signature related functionalities
	type args struct {
		blockIds        []int64
		expectedBlockID int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "wantSuccess:BlockIDFound",
			args: args{
				blockIds:        []int64{1, 2, 3, 4},
				expectedBlockID: int64(1),
			},
			want: true,
		},
		{
			name: "wantSuccess:InvalidTimestamp",
			args: args{
				blockIds:        []int64{1, 2, 3, 4},
				expectedBlockID: int64(5),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlockIDExist(tt.args.blockIds, tt.args.expectedBlockID); got != tt.want {
				t.Errorf("TestIsBlockIDExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGenesis(t *testing.T) {
	type args struct {
		previousBlockID int64
		block           *model.Block
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not genesis",
			args: args{
				previousBlockID: 1,
				block: &model.Block{
					ID:                   2,
					PreviousBlockHash:    nil,
					Height:               2,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "",
				},
			},
			want: false,
		},
		{
			name: "genesis",
			args: args{
				previousBlockID: -1,
				block: &model.Block{
					ID:                   1,
					PreviousBlockHash:    nil,
					Height:               2,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "11111",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGenesis(tt.args.previousBlockID, tt.args.block); got != tt.want {
				t.Errorf("IsGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateNodeOrder(t *testing.T) {
	type args struct {
		score     *big.Int
		blockSeed int64
		nodeID    int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "CalculateNodeOrder:success",
			args: args{
				score:     new(big.Int).SetInt64(1000),
				blockSeed: 1000,
				nodeID:    int64(1000),
			},
			want: "7357219233906154824",
		},
		{
			name: "CalculateNodeOrder:success2",
			args: args{
				score:     new(big.Int).SetInt64(2000),
				blockSeed: 3000,
				nodeID:    int64(10),
			},
			want: "4922650802765643438",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateNodeOrder(tt.args.score, tt.args.blockSeed,
				tt.args.nodeID); !reflect.DeepEqual(fmt.Sprintf("%d", got), tt.want) {
				t.Errorf("CalculateNodeOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
