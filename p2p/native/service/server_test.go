package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"google.golang.org/grpc"
)

func TestServerService_GetPeerInfo(t *testing.T) {
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
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
					ResolvedPeers:   make(map[string]*model.Peer),
					KnownPeers:      make(map[string]*model.Peer),
					UnresolvedPeers: make(map[string]*model.Peer),
				},
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
			hostServiceInstance = CreateHostService(tt.fields.Host)
			obsr := &observer.Observer{}
			ss := NewServerService(obsr)
			got, err := ss.GetPeerInfo(tt.args.ctx, tt.args.req)
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

func TestServerService_GetMorePeers(t *testing.T) {
	type fields struct {
		Host       *model.Host
		GrpcServer *grpc.Server
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
					ResolvedPeers: map[string]*model.Peer{
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
			hostServiceInstance = CreateHostService(tt.fields.Host)
			obsr := &observer.Observer{}
			ss := NewServerService(obsr)
			got, err := ss.GetMorePeers(tt.args.ctx, tt.args.req)
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

func TestServerService_SendPeers(t *testing.T) {
	type fields struct {
		Observer   *observer.Observer
		Host       *model.Host
		GrpcServer *grpc.Server
	}
	type args struct {
		ctx context.Context
		req *model.SendPeersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		// Add test cases.
		{
			name: "TestServerService_SendPeers:Success",
			fields: fields{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          8001,
					},
					ResolvedPeers: map[string]*model.Peer{
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
				Observer: observer.NewObserver(),
			},
			args: args{
				ctx: context.Background(),
				req: &model.SendPeersRequest{
					Peers: []*model.Node{
						{
							Address:       "192.168.55.3",
							SharedAddress: "192.168.55.3",
							Port:          2001,
						},
					},
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &ServerService{
				Observer: tt.fields.Observer,
			}
			hostServiceInstance = CreateHostService(tt.fields.Host)
			got, err := ss.SendPeers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerService.SendPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServerService.SendPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerService_SendBlock(t *testing.T) {
	type fields struct {
		Observer *observer.Observer
	}
	type args struct {
		ctx context.Context
		req *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		// Add test cases.
		{
			name: "TestServerService_SendBlock:Success",
			fields: fields{
				Observer: observer.NewObserver(),
			},
			args: args{
				ctx: context.Background(),
				req: &model.Block{
					ID:                   0,
					Height:               0,
					Version:              1,
					CumulativeDifficulty: "",
					SmithScale:           0,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         make([]byte, 32),
					Timestamp:            12345678,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					PayloadLength:        0,
					BlockSignature:       []byte{},
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &ServerService{
				Observer: tt.fields.Observer,
			}
			got, err := ss.SendBlock(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerService.SendBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServerService.SendBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerService_SendTransaction(t *testing.T) {
	type fields struct {
		Observer *observer.Observer
	}
	type args struct {
		ctx context.Context
		req *model.SendTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		// Add test cases.
		{
			name: "TestServerService_SendTransaction:Success",
			fields: fields{
				Observer: observer.NewObserver(),
			},
			args: args{
				ctx: context.Background(),
				req: &model.SendTransactionRequest{
					TransactionBytes: []byte{
						1, 0, 1, 53, 119, 58, 93, 0, 0, 0, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77,
						84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116,
						78, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107,
						106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 8, 0,
						0, 0, 16, 39, 0, 0, 0, 0, 0, 0, 32, 85, 34, 198, 89, 78, 166, 142, 59, 148, 243, 133, 69, 66, 123, 219, 2, 3, 229, 172,
						221, 35, 185, 208, 43, 44, 172, 96, 166, 116, 205, 93, 78, 194,
					},
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &ServerService{
				Observer: tt.fields.Observer,
			}
			got, err := ss.SendTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerService.SendTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServerService.SendTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
