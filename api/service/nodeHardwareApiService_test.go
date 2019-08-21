package service

import (
	"github.com/zoobc/zoobc-core/common/crypto"
	"reflect"
	"testing"
)

var (
	mockSignature    = crypto.NewSignature()
	mockOwnerAddress = "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
)

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
