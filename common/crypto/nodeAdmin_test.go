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
	mockValidAuth = base64.StdEncoding.EncodeToString(bufferValid.Bytes())
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
