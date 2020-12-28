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
package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	mockOwnerAddress = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
		81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	mockOwnerSeed = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	mockValidAuth,
	mockInvalidTimestampAuth,
	mockInvalidRequestTypeAuth,
	mockInvalidSignatureAuth string
)

func setupVerifyAuthAPI() {
	var (
		bufferValid              *bytes.Buffer
		bufferInvalidTimestamp   *bytes.Buffer
		bufferInvalidRequestType *bytes.Buffer
	)
	bufferValid = bytes.NewBuffer([]byte{})
	bufferInvalidRequestType = bytes.NewBuffer([]byte{})
	bufferInvalidTimestamp = bytes.NewBuffer([]byte{})
	bufferValid.Write(util.ConvertUint64ToBytes(10000))
	bufferValid.Write(util.ConvertUint32ToBytes(0))
	bufferInvalidRequestType.Write(util.ConvertUint64ToBytes(10000))
	bufferInvalidRequestType.Write(util.ConvertUint32ToBytes(10000))
	bufferInvalidTimestamp.Write(util.ConvertUint64ToBytes(0))
	bufferInvalidTimestamp.Write(util.ConvertUint32ToBytes(0))
	validSignature, _ := (&Signature{}).Sign(
		bufferValid.Bytes(),
		model.AccountType_ZbcAccountType,
		mockOwnerSeed,
	)
	bufferValid.Write(validSignature)
	validAuthBytes := bufferValid.Bytes()
	mockValidAuth = base64.StdEncoding.EncodeToString(validAuthBytes)
	bufferValid.Write([]byte{1, 2})
	mockInvalidSignatureAuth = base64.StdEncoding.EncodeToString(bufferValid.Bytes())
	mockInvalidTimestampAuth = base64.StdEncoding.EncodeToString(bufferInvalidTimestamp.Bytes())
	mockInvalidRequestTypeAuth = base64.StdEncoding.EncodeToString(bufferInvalidRequestType.Bytes())
}

func TestVerifyAuthAPI(t *testing.T) {
	setupVerifyAuthAPI()
	type args struct {
		ownerAddress []byte
		auth         string
		requestType  model.RequestType
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "VerifyAuthAPI:fail-invalid-auth",
			args: args{
				ownerAddress: nil,
				auth:         "aaaaaaa",
				requestType:  0,
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:fail-invalid-requestType",
			args: args{
				ownerAddress: nil,
				auth:         mockInvalidRequestTypeAuth,
				requestType:  0,
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:fail-invalid-timestamp",
			args: args{
				ownerAddress: nil,
				auth:         mockInvalidTimestampAuth,
				requestType:  0,
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:fail-invalid-signature",
			args: args{
				ownerAddress: mockOwnerAddress,
				auth:         mockInvalidSignatureAuth,
				requestType:  0,
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:success",
			args: args{
				ownerAddress: mockOwnerAddress,
				auth:         mockValidAuth,
				requestType:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyAuthAPI(tt.args.ownerAddress, tt.args.auth,
				tt.args.requestType); (err != nil) != tt.wantErr {
				t.Errorf("VerifyAuthAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
			LastRequestTimestamp = 0
		})
	}
}
