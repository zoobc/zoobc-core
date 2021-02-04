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
package accounttype

import (
	"encoding/hex"
	"testing"
)

func TestETHAccountType_VerifySignature(t *testing.T) {
	payload, _ := hex.DecodeString("0100000001df6f1b60000000000400000011f2b30c9479ccaa639962e943ca7cfd3498705258ddb49d" +
		"fe25bba00a555e48cb35a79f3d084ce26dbac0e6bb887463774817cb80e89b20c0990bc47f9075d500000000e12c84a0fd461cbbec5" +
		"956a66b2ebad0499491cff77f75b583d041d757d87fff00e1f505000000000800000000e1f505000000000200000000000000")
	signature, _ := hex.DecodeString("c79984b222e95f095df054be5533fbc92f95f078b375d2985472bc96012176da2442dcbfe274ffe6a" +
		"0f4bf31bfc6093554aae00f105a37add43257c569eb8fe91c")
	wrongPublicKey, _ := hex.DecodeString("10f2b30c9479ccaa639962e943ca7cfd3498705258ddb49dfe25bba00a555e48cb35a79f3d084c" +
		"e26dbac0e6bb887463774817cb80e89b20c0990bc47f9075d5")
	publicKey, _ := hex.DecodeString("11f2b30c9479ccaa639962e943ca7cfd3498705258ddb49dfe25bba00a555e48cb35a79f3d084ce26db" +
		"ac0e6bb887463774817cb80e89b20c0990bc47f9075d5")

	type fields struct {
		privateKey  []byte
		publicKey   []byte
		fullAddress []byte
	}
	type args struct {
		payload        []byte
		signature      []byte
		accountAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantErr:invalid signature",
			fields: fields{
				publicKey: wrongPublicKey,
			},
			args: args{
				payload:   payload,
				signature: signature,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				publicKey: publicKey,
			},
			args: args{
				payload:   payload,
				signature: signature,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ETHAccountType{
				privateKey:  tt.fields.privateKey,
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if err := acc.VerifySignature(tt.args.payload, tt.args.signature, tt.args.accountAddress); (err != nil) != tt.wantErr {
				t.Errorf("ETHAccountType.VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestETHAccountType_GetEncodedAddress(t *testing.T) {
	publicKey, _ := hex.DecodeString("11f2b30c9479ccaa639962e943ca7cfd3498705258ddb49dfe25bba00a555e48cb35a79f3d084ce26dbac0e6bb887463774817cb80e89b20c0990bc47f9075d5")

	type fields struct {
		privateKey  []byte
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "wantError",
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				publicKey: publicKey,
			},
			want: "0xc2524c08e0166f6a3b8d9925f8864c8ee18cb729",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ETHAccountType{
				privateKey:  tt.fields.privateKey,
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.GetEncodedAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("ETHAccountType.GetEncodedAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ETHAccountType.GetEncodedAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
