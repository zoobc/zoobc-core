package util

import (
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewHost(t *testing.T) {

	type args struct {
		address    string
		port       uint32
		knownPeers []*model.Peer
	}
	tests := []struct {
		name string
		args args
		want *model.Host
	}{
		// TODO: Add test cases.
		{
			name: "NewHostTest:one",
			args: args{
				address: "127.0.0.1",
				port:    3000,
				knownPeers: []*model.Peer{
					{
						Info: &model.Node{
							Address:       "127.0.0.1",
							Port:          3001,
							SharedAddress: "127.0.0.1",
						},
					},
					{
						Info: &model.Node{
							Address:       "192.168.5.1",
							Port:          3002,
							SharedAddress: "192.168.5.1",
						},
					},
				},
			},
			want: &model.Host{
				Info: &model.Node{
					Address: "127.0.0.1",
					Port:    3000,
				},
				KnownPeers: map[string]*model.Peer{
					"127.0.0.1:3001": &model.Peer{
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          3001,
						},
					},
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
				UnresolvedPeers: map[string]*model.Peer{
					"127.0.0.1:3001": &model.Peer{
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          3001,
						},
					},
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHost(tt.args.address, tt.args.port, tt.args.knownPeers)
			if strings.Compare(got.String(), tt.want.String()) != 0 {
				t.Errorf("\n%v \n%v", got.String(), tt.want.String())
			}
		})
	}
}
