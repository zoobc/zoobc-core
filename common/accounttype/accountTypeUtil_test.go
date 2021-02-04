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
	"bytes"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetAccountTypes(t *testing.T) {
	var (
		zbcAccount   = &ZbcAccountType{}
		dummyAccount = &BTCAccountType{}
		emptyAccount = &EmptyAccountType{}
	)
	tests := []struct {
		name string
		want map[uint32]AccountTypeInterface
	}{
		{
			name: "TestGetAccountTypes:success",
			want: map[uint32]AccountTypeInterface{
				uint32(zbcAccount.GetTypeInt()):   zbcAccount,
				uint32(dummyAccount.GetTypeInt()): dummyAccount,
				uint32(emptyAccount.GetTypeInt()): dummyAccount,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAccountTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAccountType(t *testing.T) {
	zbcAccType := &ZbcAccountType{}
	zbcAccType.SetAccountPublicKey([]byte{1, 2, 3})
	btcAccType := &BTCAccountType{}
	btcAccType.SetAccountPublicKey([]byte{1, 2, 3})
	emptyAccType := &EmptyAccountType{}
	emptyAccType.SetAccountPublicKey([]byte{1, 2, 3})
	estoniaEidAccType := &EstoniaEidAccountType{}
	estoniaEidAccType.SetAccountPublicKey([]byte{1, 2, 3})
	ethAccType := &ETHAccountType{}
	ethAccType.SetAccountPublicKey([]byte{17, 242, 179, 12, 148, 121, 204, 170, 99, 153, 98, 233, 67, 202, 124, 253, 52, 152, 112, 82, 88, 221, 180, 157, 254, 37, 187, 160, 10, 85, 94, 72, 203, 53, 167, 159, 61, 8, 76, 226, 109, 186, 192, 230, 187, 136, 116, 99, 119, 72, 23, 203, 128, 232, 155, 32, 192, 153, 11, 196, 127, 144, 117, 213})
	type args struct {
		accTypeInt int32
		accPubKey  []byte
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestNewAccountType:success/ZbcAccountType",
			args: args{
				accPubKey:  []byte{1, 2, 3},
				accTypeInt: 0,
			},
			want: zbcAccType,
		},
		{
			name: "TestNewAccountType:success/BTCAccountType",
			args: args{
				accPubKey:  []byte{1, 2, 3},
				accTypeInt: 1,
			},
			want: btcAccType,
		},
		{
			name: "TestNewAccountType:success/EmptyAccountType",
			args: args{
				accPubKey:  []byte{1, 2, 3},
				accTypeInt: 2,
			},
			want: emptyAccType,
		},
		{
			name: "TestNewAccountType:success/EstoniaEidAccountType",
			args: args{
				accPubKey:  []byte{1, 2, 3},
				accTypeInt: 3,
			},
			want: estoniaEidAccType,
		},
		{
			name: "TestNewAccountType:success/ETHAccountType",
			args: args{
				accPubKey: []byte{17, 242, 179, 12, 148, 121, 204, 170, 99, 153, 98, 233, 67, 202, 124, 253, 52, 152, 112,
					82, 88, 221, 180, 157, 254, 37, 187, 160, 10, 85, 94, 72, 203, 53, 167, 159, 61, 8, 76, 226, 109, 186,
					192, 230, 187, 136, 116, 99, 119, 72, 23, 203, 128, 232, 155, 32, 192, 153, 11, 196, 127, 144, 117, 213},
				accTypeInt: 4,
			},
			want: ethAccType,
		},
		{
			name: "TestNewAccountType:fail-{invalidAccountType}",
			args: args{
				accPubKey:  []byte{},
				accTypeInt: 99,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountType(tt.args.accTypeInt, tt.args.accPubKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccountType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAccountTypeFromAccount(t *testing.T) {
	var (
		accountAddressPubKey = []byte{149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23,
			39, 95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
		accountAddress = append([]byte{0, 0, 0, 0}, accountAddressPubKey...)
		accType        = &ZbcAccountType{}
	)
	accType.SetAccountPublicKey(accountAddressPubKey)
	type args struct {
		accountAddress []byte
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestNewAccountTypeFromAccount:success",
			args: args{
				accountAddress: accountAddress,
			},
			want: accType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountTypeFromAccount(tt.args.accountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccountTypeFromAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountTypeFromAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBytesToAccountType(t *testing.T) {
	var (
		accPubKey = []byte{149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
		buffer = []byte{0, 0, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108, 1, 2, 3, 4, 5, 6, 0, 0, 0, 0}
		bufferInvalidAccType = []byte{255, 255, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108, 1, 2, 3, 4, 5, 6, 0, 0, 0, 0}
		bufferInvalidPubKeyLength = []byte{255, 255, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221,
			109, 23, 39,
			95, 248, 147, 114, 91}
		accTypeRes = &ZbcAccountType{}
	)
	accTypeRes.SetAccountPublicKey(accPubKey)
	type args struct {
		buffer *bytes.Buffer
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestParseBytesToAccountType:success",
			args: args{
				buffer: bytes.NewBuffer(buffer),
			},
			want: accTypeRes,
		},
		{
			name: "TestParseBytesToAccountType:fail-{InvalidAccountType}",
			args: args{
				buffer: bytes.NewBuffer(bufferInvalidAccType),
			},
			wantErr: true,
		},
		{
			name: "TestParseBytesToAccountType:fail-{InvalidAccountPubKey}",
			args: args{
				buffer: bytes.NewBuffer(bufferInvalidPubKeyLength),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytesToAccountType(tt.args.buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytesToAccountType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBytesToAccountType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEncodedAccountToAccountAddress(t *testing.T) {
	var (
		encAddress1  = "ZBC_SUAW4BPA_S2CFKO6N_FWUGXD6R_262523IX_E5P7RE3S_LNZUWMY7_SRWCMI2J"
		fullAddress1 = []byte{0, 0, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
	)
	type args struct {
		accTypeInt            int32
		encodedAccountAddress string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestParseEncodedAccountToAccountAddress:success",
			args: args{
				encodedAccountAddress: encAddress1,
				accTypeInt:            int32(model.AccountType_ZbcAccountType),
			},
			want: fullAddress1,
		},
		{
			name: "TestParseEncodedAccountToAccountAddress:success-btcImplementation",
			args: args{
				encodedAccountAddress: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
				accTypeInt:            int32(model.AccountType_BTCAccountType),
			},
			want: []byte{0, 0, 0, 0, 13, 137, 40, 212, 218, 119, 144, 80, 70, 113, 150, 129, 2, 84, 45, 144, 145, 17, 64, 134},
		},
		{
			name: "TestParseEncodedAccountToAccountAddress:fail-{InvalidAccountType}",
			args: args{
				encodedAccountAddress: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
				accTypeInt:            99,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEncodedAccountToAccountAddress(tt.args.accTypeInt, tt.args.encodedAccountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEncodedAccountToAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEncodedAccountToAccountAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
