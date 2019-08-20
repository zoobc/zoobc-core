package util

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"testing"
)

var (
	mockOwnerAddress = "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
	mockOwnerSeed    = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
)

func TestVerifyAuthAPI(t *testing.T) {
	type args struct {
		ownerAddress string
		auth         *model.Auth
		requestType  model.RequestType
		signature    crypto.SignatureInterface
	}
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(0))
	buffer.Write(ConvertUint64ToBytes(10000))
	validSignature := (&crypto.Signature{}).Sign(
		buffer.Bytes(),
		constant.NodeSignatureTypeDefault,
		mockOwnerSeed,
	)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "VerifyAuthAPI:fail-invalid-requestType",
			args: args{
				ownerAddress: "",
				auth: &model.Auth{
					RequestType: 1,
					Timestamp:   0,
					Signature:   nil,
				},
				requestType: 0,
				signature:   crypto.NewSignature(),
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:fail-invalid-timestamp",
			args: args{
				ownerAddress: "",
				auth: &model.Auth{
					RequestType: 0,
					Timestamp:   0,
					Signature:   nil,
				},
				requestType: 0,
				signature:   crypto.NewSignature(),
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:fail-invalid-signature",
			args: args{
				ownerAddress: mockOwnerAddress,
				auth: &model.Auth{
					RequestType: 0,
					Timestamp:   1000,
					Signature:   []byte{0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				requestType: 0,
				signature:   crypto.NewSignature(),
			},
			wantErr: true,
		},
		{
			name: "VerifyAuthAPI:success",
			args: args{
				ownerAddress: mockOwnerAddress,
				auth: &model.Auth{
					RequestType: 0,
					Timestamp:   10000,
					Signature:   validSignature,
				},
				requestType: 0,
				signature:   crypto.NewSignature(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyAuthAPI(tt.args.ownerAddress, tt.args.auth,
				tt.args.requestType, tt.args.signature); (err != nil) != tt.wantErr {
				t.Errorf("VerifyAuthAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
