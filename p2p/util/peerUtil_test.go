package util

import (
	"reflect"
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

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
			name: "NewknownPeerTest:success",
			args: args{
				address: "192.168.5.99",
				port:    5000,
			},
			want: &model.Peer{
				Info: &model.Node{
					Address: "192.168.5.99",
					Port:    5000,
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

func TestParseKnownPeers(t *testing.T) {
	type args struct {
		peers []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*model.Peer
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ParseKnownPeersTest:success",
			args: args{
				peers: []string{"192.168.1.2:2001", "192.168.5.123:3000"},
			},
			want:    append([]*model.Peer{}, NewKnownPeer("192.168.1.2", 2001), NewKnownPeer("192.168.5.123", 3000)),
			wantErr: false,
		},
		{
			name: "ParseKnownPeersTest:true",
			args: args{
				peers: []string{"192.168.1.2:2001a", "192.168.5.123:3000a"},
			},
			want:    []*model.Peer{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKnownPeers(tt.args.peers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKnownPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseKnownPeers() = %v, want %v", got, tt.want)
				}
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
						Address: "192.168.5.99",
						Port:    5000,
					},
				},
			},
			want: "192.168.5.99:5000",
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
			name: "AddToResolvedPeerTest:success",
			args: args{
				host: &model.Host{
					Info: &model.Node{
						Address: "192.168.5.1",
						Port:    3002,
					},
					KnownPeers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
					UnresolvedPeers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
					Peers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
				},
				peer: &model.Peer{
					Info: &model.Node{
						Address: "192.168.5.1",
						Port:    3002,
					},
				},
			},

			want: &model.Host{
				Info: &model.Node{
					Address: "192.168.5.1",
					Port:    3002,
				},
				KnownPeers: map[string]*model.Peer{
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
				Peers: map[string]*model.Peer{
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							Address: "192.168.5.1",
							Port:    3002,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddToResolvedPeer(tt.args.host, tt.args.peer)
			if strings.Compare(got.String(), tt.want.String()) != 0 {
				t.Errorf("\n%v \n%v", got, tt.want)
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
			name: "AddToUnresolvedPeersTest:success",
			args: args{
				host: &model.Host{
					Info: &model.Node{
						Address: "192.168.5.1",
						Port:    3002,
					},
					KnownPeers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
					UnresolvedPeers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
					Peers: map[string]*model.Peer{
						"192.168.5.1:3002": &model.Peer{
							Info: &model.Node{
								SharedAddress: "192.168.5.1",
								Address:       "192.168.5.1",
								Port:          3002,
							},
						},
					},
				},
				newNodes: []*model.Node{
					{
						Address: "192.168.5.1",
						Port:    3002,
					},
				},
			},
			want: &model.Host{
				Info: &model.Node{
					Address: "192.168.5.1",
					Port:    3002,
				},
				KnownPeers: map[string]*model.Peer{
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
				UnresolvedPeers: map[string]*model.Peer{
					"192.168.5.1:3002": &model.Peer{
						Info: &model.Node{
							SharedAddress: "192.168.5.1",
							Address:       "192.168.5.1",
							Port:          3002,
						},
					},
				},
				Peers: map[string]*model.Peer{
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
			if got := AddToUnresolvedPeers(tt.args.host, tt.args.newNodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\n%v\n %v", got, tt.want)
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
			name: "PeerUnblacklistTest:success",
			args: args{
				peer: &model.Peer{
					Info: &model.Node{
						Address: "192.168.5.11",
						Port:    3000,
					},
					State: 3,
				},
			},

			want: &model.Peer{
				Info: &model.Node{
					Address: "192.168.5.11",
					Port:    3000,
				},
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
