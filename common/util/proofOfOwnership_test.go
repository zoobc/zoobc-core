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
	"github.com/zoobc/zoobc-core/common/accounttype"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestParseProofOfOwnershipBytes(t *testing.T) {
	poown := &model.ProofOfOwnership{
		MessageBytes: make([]byte, GetProofOfOwnershipSize(&accounttype.ZbcAccountType{}, false)),
		Signature:    make([]byte, constant.NodeSignature),
	}
	poownBytes := GetProofOfOwnershipBytes(poown)
	type args struct {
		poownBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "ParseProofOfOwnershipBytes - fail (empty bytes)",
			args: args{
				poownBytes: []byte{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipBytes - fail (wrong poown size)",
			args: args{
				poownBytes: poownBytes[:10],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipBytes - fail (no signature / wrong signature size)",
			args: args{
				poownBytes: poownBytes[:GetProofOfOwnershipSize(&accounttype.ZbcAccountType{}, false)],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipBytes - success",
			args:    args{poownBytes: poownBytes},
			want:    poown,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProofOfOwnershipBytes(tt.args.poownBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProofOfOwnershipBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProofOfOwnershipBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProofOfOwnershipSize(t *testing.T) {
	t.Run("WithAndWithoutSignature-Gap", func(t *testing.T) {
		withSig := GetProofOfOwnershipSize(&accounttype.ZbcAccountType{}, true)
		withoutSig := GetProofOfOwnershipSize(&accounttype.ZbcAccountType{}, false)
		if withSig-withoutSig != constant.NodeSignature {
			t.Errorf("GetPoownSize with and without signature should have %d difference",
				constant.NodeSignature)
		}
	})
}

func TestParseProofOfOwnershipMessageBytes(t *testing.T) {
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		BlockHash:   make([]byte, constant.BlockHash),
		BlockHeight: 0,
	}
	poownMessageBytes := GetProofOfOwnershipMessageBytes(poownMessage)
	type args struct {
		poownMessageBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ProofOfOwnershipMessage
		wantErr bool
	}{
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - no bytes",
			args:    args{poownMessageBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - wrong account address",
			args:    args{poownMessageBytes: poownMessageBytes[:10]},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - no block hash",
			args:    args{poownMessageBytes: poownMessageBytes[:len(poownMessage.AccountAddress)]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipMessageBytes:fail - no block height",
			args: args{
				poownMessageBytes: poownMessageBytes[:(len(poownMessage.AccountAddress) +
					int(constant.BlockHash))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipMessageBytes:fail - success",
			args: args{
				poownMessageBytes: poownMessageBytes,
			},
			want:    poownMessage,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProofOfOwnershipMessageBytes(tt.args.poownMessageBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProofOfOwnershipMessageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProofOfOwnershipMessageBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
