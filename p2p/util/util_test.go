// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
		version    string
		codename   string
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
				version:  "1.0.0",
				codename: "ZBC_main",
			},
			want: &model.Host{
				Info: &model.Node{
					Address:  "127.0.0.1",
					Port:     3000,
					Version:  "1.0.0",
					CodeName: "ZBC_main",
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
			got := NewHost(tt.args.address, tt.args.port, tt.args.knownPeers, tt.args.version, tt.args.codename)
			if strings.Compare(got.String(), tt.want.String()) != 0 {
				t.Errorf("\n%v \n%v", got.String(), tt.want.String())
			}
		})
	}
}

func TestNewPeer(t *testing.T) {
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
			name: "NewPeer:success",
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
			if got := NewPeer(tt.args.address, tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPeer() = %v, want %v", got, tt.want)
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
			want:    append([]*model.Peer{}, NewPeer("192.168.1.2", 2001), NewPeer("192.168.5.123", 3000)),
			wantErr: false,
		},
		{
			name: "ParseKnownPeersTest:true",
			args: args{
				peers: []string{"192.168.1.2:2001xa", "192.168.5.123:3000a"},
			},
			want:    nil,
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

func TestCheckPeerCompatibility(t *testing.T) {
	type args struct {
		host *model.Node
		peer *model.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "wantFailed: unmatched version (breaking change)",
			args: args{
				host: &model.Node{
					Version: "1.0.0",
				},
				peer: &model.Node{
					Version: "2.0.0",
				},
			},
			wantErr: true,
		},
		{
			name: "wantFailed: unmatched CodeName",
			args: args{
				host: &model.Node{
					CodeName: "ZBC_main",
				},
				peer: &model.Node{
					CodeName: "ZBC_test",
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess: even though different minor version",
			args: args{
				host: &model.Node{
					Version:  "1.0.0",
					CodeName: "ZBC_main",
				},
				peer: &model.Node{
					Version:  "1.2.0",
					CodeName: "ZBC_main",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckPeerCompatibility(tt.args.host, tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("CheckPeerCompatibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
