package service

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"reflect"
	"testing"
)

var (
	mockSignature    = crypto.NewSignature()
	mockOwnerAddress = "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
	mockOwnerSeed    = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	mockAuth         = &model.Auth{
		RequestType: 0,
		Timestamp:   10000,
		Signature:   nil,
	}
	mockInvalidAuth = &model.Auth{
		RequestType: 0,
		Timestamp:   0,
		Signature:   nil,
	}
	mockValidSignature   []byte
	mockInvalidSignature []byte
)

func setupNodeHardwareService() {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(mockAuth.RequestType)))
	buffer.Write(util.ConvertUint64ToBytes(mockAuth.Timestamp))
	mockValidSignature = (&crypto.Signature{}).Sign(
		buffer.Bytes(),
		constant.NodeSignatureTypeDefault,
		mockOwnerSeed,
	)
	mockInvalidSignature = []byte{
		0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
	}
}

func TestNewNodeHardwareService(t *testing.T) {
	type args struct {
		ownerAccountAddress string
		signature           crypto.SignatureInterface
	}
	tests := []struct {
		name string
		args args
		want *NodeHardwareService
	}{
		{
			name: "NewNodeHardwareService:success",
			args: args{
				ownerAccountAddress: mockOwnerAddress,
				signature:           mockSignature,
			},
			want: &NodeHardwareService{
				OwnerAccountAddress: mockOwnerAddress,
				Signature:           mockSignature,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeHardwareService(tt.args.ownerAccountAddress, tt.args.signature); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeHardwareService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeHardwareService_GetNodeHardware(t *testing.T) {
	setupNodeHardwareService()
	type fields struct {
		OwnerAccountAddress string
		Signature           crypto.SignatureInterface
	}
	type args struct {
		request *model.GetNodeHardwareRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeHardwareResponse
		wantErr bool
	}{
		{
			name: "GetNodeHardware:fail-invalidAuth",
			fields: fields{
				OwnerAccountAddress: mockOwnerAddress,
				Signature:           mockSignature,
			},
			args: args{
				&model.GetNodeHardwareRequest{
					Auth: mockInvalidAuth,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nhs := &NodeHardwareService{
				OwnerAccountAddress: tt.fields.OwnerAccountAddress,
				Signature:           tt.fields.Signature,
			}
			got, err := nhs.GetNodeHardware(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNodeHardware() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeHardware() got = %v, want %v", got, tt.want)
			}
		})
	}
}
