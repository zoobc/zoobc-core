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
		want *PeerService
	}{
		// Test cases.
		{name: "ClientPeerService:success",
			args: args{
				chaintype: new(chaintype.MainChain),
			},
			want: &PeerService{
				Peer:      nil,
				ChainType: new(chaintype.MainChain),
			},
		},
		{name: "ClientPeerService:fail",
			args: args{
				chaintype: nil,
			},
			want: &PeerService{
				Peer:      nil,
				ChainType: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClientPeerService(tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientPeerService() = %v, want %v", got, tt.want)
			}
		})
	}
}
