package service

import (
	"github.com/zoobc/zoobc-core/common/crypto"
	"reflect"
	"testing"
)

var (
	mockSignature    = crypto.NewSignature()
	mockOwnerAddress = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
		28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
)

func TestNewNodeHardwareService(t *testing.T) {
	type args struct {
		ownerAccountAddress []byte
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
