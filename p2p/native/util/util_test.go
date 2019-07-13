package util

import (
	"reflect"
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
			name: "NewHostTest:NewHost",
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
					"127.0.0.1:3001": {
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          3001,
						},
					},
					"192.168.5.1:3002": {
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
				UnresolvedPeers: map[string]*model.Peer{
					"127.0.0.1:3001": {
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          3001,
						},
					},
					"192.168.5.1:3002": {
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

func TestGetAnyPeer(t *testing.T) {
	type args struct {
		hs *model.Host
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		// TODO: Add test cases.
		{
			name: "GetANyPeersTest:GetMorePeer",
			args: args{
				hs: &model.Host{
					Peers: map[string]*model.Peer{
						"127.0.0.1:3000": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          3000,
							},
						},
					},
				},
			},
			want: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAnyPeer(tt.args.hs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKnownPeer(t *testing.T) {
	type args struct {
		address string
		port    int
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		// TODO: Add test cases.
		{
			name: "NewKnownPeer:success",
			args: args{
				address: "127.0.0.1",
				port:    8001,
			},
			want: &model.Peer{
				Info: &model.Node{
					Address: "127.0.0.1",
					Port:    8001,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKnownPeer(tt.args.address, tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFullAddressPeer(t *testing.T) {
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "GetFullAddressPeer:success",
			args: args{
				peer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
			},
			want: "127.0.0.1:8001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFullAddressPeer(tt.args.peer); got != tt.want {
				t.Errorf("GetFullAddressPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToResolvedPeer(t *testing.T) {
	type args struct {
		host *model.Host
		peer *model.Peer
	}
	tests := []struct {
		name string
		args args
		want *model.Host
	}{
		// TODO: Add test cases.
		{
			name: "AddToResolvedPeer:success",
			args: args{
				host: &model.Host{
					Info:            &model.Node{},
					KnownPeers:      make(map[string]*model.Peer),
					Peers:           make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
				peer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
			},
			want: &model.Host{
				Info:       &model.Node{},
				KnownPeers: make(map[string]*model.Peer),
				Peers: map[string]*model.Peer{
					"127.0.0.1:8001": {
						Info: &model.Node{
							Address: "127.0.0.1",
							Port:    8001,
						},
					},
				},
				UnresolvedPeers: make(map[string]*model.Peer),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddToResolvedPeer(tt.args.host, tt.args.peer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddToResolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToUnresolvedPeers(t *testing.T) {
	type args struct {
		host     *model.Host
		newNodes []*model.Node
	}
	tests := []struct {
		name string
		args args
		want *model.Host
	}{
		// TODO: Add test cases.
		{
			name: "AddToUnresolvedPeers:success",
			args: args{
				host: &model.Host{
					Info:            &model.Node{},
					KnownPeers:      make(map[string]*model.Peer),
					Peers:           make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
				newNodes: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8001,
					},
					{
						Address: "192.168.1.5",
						Port:    8001,
					},
				},
			},
			want: &model.Host{
				Info:       &model.Node{},
				KnownPeers: make(map[string]*model.Peer),
				Peers:      make(map[string]*model.Peer),
				UnresolvedPeers: map[string]*model.Peer{
					"127.0.0.1:8001": {
						Info: &model.Node{
							Address: "127.0.0.1",
							Port:    8001,
						},
					},
					"192.168.1.5:8001": {
						Info: &model.Node{
							Address: "192.168.1.5",
							Port:    8001,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddToUnresolvedPeers(tt.args.host, tt.args.newNodes)
			if strings.Compare(got.String(), tt.want.String()) != 0 {
				t.Errorf("AddToUnresolvedPeers() = \n%v, want \n%v", got.String(), tt.want.String())
			}
		})
	}
}

func TestPeerUnblacklist(t *testing.T) {
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		// TODO: Add test cases.
		{
			name: "TestPeerUnblacklist:success",
			args: args{
				peer: &model.Peer{
					BlacklistingCause: "not connected",
					BlacklistingTime:  1234,
					State:             model.PeerState_BLACKLISTED,
				},
			},
			want: &model.Peer{
				BlacklistingCause: "",
				BlacklistingTime:  0,
				State:             model.PeerState_NON_CONNECTED,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PeerUnblacklist(tt.args.peer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerUnblacklist() = %v, want %v", got, tt.want)
			}
		})
	}
}
