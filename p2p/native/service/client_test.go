package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
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

// Before run this test make sure application alredy running in port 8001
func TestPeerService_GetPeerInfo(t *testing.T) {
	type fields struct {
		Peer      *model.Peer
		ChainType contract.ChainType
	}
	type args struct {
		destPeer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Node
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "GetPeerInfo:success",
			fields: fields{
				Peer:      &model.Peer{},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
				},
			},
			want: &model.Node{
				Address: "127.0.0.1",
				Port:    8001,
			},
			wantErr: false,
		},
		{
			name: "GetPeerInfo:error",
			fields: fields{
				Peer:      &model.Peer{},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &PeerService{
				Peer:      tt.fields.Peer,
				ChainType: tt.fields.ChainType,
			}
			got, err := cs.GetPeerInfo(tt.args.destPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerService.GetPeerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerService.GetPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerService_GetMorePeers(t *testing.T) {
	type fields struct {
		Peer      *model.Peer
		ChainType contract.ChainType
	}
	type args struct {
		destPeer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMorePeersResponse
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "GetMorePeers:success",
			fields: fields{
				Peer:      &model.Peer{},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
				},
			},
			want: &model.GetMorePeersResponse{
				Peers: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8000,
					},
					{
						Address: "118.99.96.66",
						Port:    3001,
					},
					{
						Address: "192.168.5.2",
						Port:    3003,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GetMorePeers:error",
			fields: fields{
				Peer:      &model.Peer{},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &PeerService{
				Peer:      tt.fields.Peer,
				ChainType: tt.fields.ChainType,
			}
			got, err := cs.GetMorePeers(tt.args.destPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerService.GetMorePeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerService.GetMorePeers() = %v, want %v", got, tt.want)
			}
		})
	}
}
