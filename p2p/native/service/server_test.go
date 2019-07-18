package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc"
)

func TestHostService_GetPeerInfo(t *testing.T) {
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
		ChainType  contract.ChainType
	}
	type args struct {
		ctx context.Context
		req *model.GetPeerInfoRequest
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
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
					Peers:           make(map[string]*model.Peer),
					KnownPeers:      make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.GetPeerInfoRequest{
					Version: "v1.0.1",
				},
			},
			want: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          8001,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HostService{
				Host:       tt.fields.Host,
				GrpcServer: tt.fields.GrpcServer,
				ChainType:  tt.fields.ChainType,
			}
			got, err := hs.GetPeerInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostService.GetPeerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HostService.GetPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostService_GetMorePeers(t *testing.T) {
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
		ChainType  contract.ChainType
	}
	type args struct {
		ctx context.Context
		req *model.Empty
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
			name: "TestHostService_GetMorePeers:success",
			fields: fields{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
					Peers: map[string]*model.Peer{
						"192.168.55.3:2001": {
							Info: &model.Node{
								SharedAddress: "192.168.55.3",
								Address:       "192.168.55.3",
								Port:          2001,
							},
						},
					},
					KnownPeers:      make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.Empty{},
			},
			want: &model.GetMorePeersResponse{
				Peers: []*model.Node{
					{
						Address:       "192.168.55.3",
						SharedAddress: "192.168.55.3",
						Port:          2001,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HostService{
				Host:       tt.fields.Host,
				GrpcServer: tt.fields.GrpcServer,
				ChainType:  tt.fields.ChainType,
			}
			got, err := hs.GetMorePeers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostService.GetMorePeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HostService.GetMorePeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostService_ResolvePeers(t *testing.T) {
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
		ChainType  contract.ChainType
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
					Peers: map[string]*model.Peer{
						"192.168.55.3:2001": {
							Info: &model.Node{
								SharedAddress: "192.168.55.3",
								Address:       "192.168.55.3",
								Port:          2001,
							},
						},
					},
					KnownPeers:      make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HostService{
				Host:       tt.fields.Host,
				GrpcServer: tt.fields.GrpcServer,
				ChainType:  tt.fields.ChainType,
			}
			hs.ResolvePeers()
		})
	}
}

func TestHostService_GetMorePeersHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
		ChainType  contract.ChainType
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
					Peers: map[string]*model.Peer{
						"192.168.55.3:2001": {
							Info: &model.Node{
								SharedAddress: "192.168.55.3",
								Address:       "192.168.55.3",
								Port:          2001,
							},
						},
					},
					KnownPeers:      make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HostService{
				Host:       tt.fields.Host,
				GrpcServer: tt.fields.GrpcServer,
				ChainType:  tt.fields.ChainType,
			}
			hs.GetMorePeersHandler()
		})
	}
}
