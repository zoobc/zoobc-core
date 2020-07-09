package strategy

import (
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestPeerStrategyHelper_GetRandomPeerWithoutRepetition(t *testing.T) {
	type args struct {
		peers map[string]*model.Peer
		mutex *sync.Mutex
	}
	peersMap := map[string]*model.Peer{
		"127.0.0.1:8000": {
			Info: &model.Node{
				Address: "127.0.0.1",
				Port:    8000,
			},
		},
		"127.0.0.2:8000": {
			Info: &model.Node{
				Address: "127.0.0.2",
				Port:    8000,
			},
		},
		"127.0.0.3:8000": {
			Info: &model.Node{
				Address: "127.0.0.3",
				Port:    8000,
			},
		},
	}
	tests := []struct {
		name string
		ps   *PeerStrategyHelper
		args args
		want *model.Peer
	}{
		{
			name: "GetRandomPeerWithoutRepetition:success",
			args: args{
				peers: peersMap,
				mutex: &sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PeerStrategyHelper{}
			i := 0
			peersCount := len(peersMap)
			for len(peersMap) > 0 {
				i++
				ps.GetRandomPeerWithoutRepetition(tt.args.peers, tt.args.mutex)
			}
			if i != peersCount {
				t.Error("PeerStrategyHelper.GetRandomPeerWithoutRepetition()")
			}
		})
	}
}
