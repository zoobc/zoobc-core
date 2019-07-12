package p2p

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestNewPeerService(t *testing.T) {
	type args struct {
		chaintypeNumber int32
	}

	tests := []struct {
		name string
		args args
		want *PeerService
	}{
		// TODO: Add test cases.
		{name: "NewPeerService:main",
			args: args{
				chaintypeNumber: 0,
			},
			want: &PeerService{
				Peer:      nil,
				ChainType: chaintype.GetChainType(0),
			},
		},
		{name: "NewPeerService:spine",
			args: args{
				chaintypeNumber: 1,
			},
			want: &PeerService{
				Peer:      nil,
				ChainType: chaintype.GetChainType(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPeerService(tt.args.chaintypeNumber); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPeerService() = %v, want %v", got, tt.want)
			}
		})
	}
}
