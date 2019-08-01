package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
)

func TestClientPeerService(t *testing.T) {
	type args struct {
		chaintype contract.ChainType
	}

	tests := []struct {
		name string
		args args
		want *PeerServiceClient
	}{
		// Test cases.
		{name: "ClientPeerService:success",
			args: args{
				chaintype: new(chaintype.MainChain),
			},
			want: &PeerServiceClient{},
		},
		{name: "ClientPeerService:fail",
			args: args{
				chaintype: nil,
			},
			want: &PeerServiceClient{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPeerServiceClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientPeerService() = %v, want %v", got, tt.want)
			}
		})
	}
}
