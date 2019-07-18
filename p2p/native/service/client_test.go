package service

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/p2p/native/util"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type (
	mockHostService struct {
		HostService
	}
)

var lis *bufconn.Listener

func init() {
	const bufSize = 1024 * 1024
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	service.RegisterP2PCommunicationServer(s, &mockHostService{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func mockGrpcDialer() *grpc.ClientConn {
	ctx := context.Background()
	conn, _ := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	return conn
}

func (mhs *mockHostService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return &model.Node{
		Address: "127.0.0.1",
		Port:    8001,
	}, nil
}

func (mhs *mockHostService) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	return &model.GetMorePeersResponse{
		Peers: []*model.Node{
			{
				Address: "127.0.0.1",
				Port:    8001,
			},
			{
				Address: "118.99.96.66",
				Port:    3001,
			},
		},
	}, nil
}

func TestClientPeerService(t *testing.T) {
	type args struct {
		chaintype contract.ChainType
	}
	tests := []struct {
		name string
		args args
		want *PeerService
	}{
		// TODO: Add test cases.
		{
			name: "TestClientPeerService",
			args: args{
				chaintype: &chaintype.MainChain{},
			},
			want: &PeerService{
				ChainType: &chaintype.MainChain{},
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

var faildDial, _ = util.GrpcDialer(&model.Peer{
	Info: &model.Node{
		Address: "127.0.0.1",
		Port:    3000,
	},
})

func TestPeerService_GetPeerInfo(t *testing.T) {
	type fields struct {
		Peer      *model.Peer
		ChainType contract.ChainType
	}
	type args struct {
		connection *grpc.ClientConn
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
			name:   "TestPeerService_GetPeerInfo:success",
			fields: fields{},
			args: args{
				connection: mockGrpcDialer(),
			},
			want: &model.Node{
				Address: "127.0.0.1",
				Port:    8001,
			},
			wantErr: false,
		},
		{
			name:   "TestPeerService_GetPeerInfo:failed",
			fields: fields{},
			args: args{
				connection: faildDial,
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
			got, err := cs.GetPeerInfo(tt.args.connection)
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
		connection *grpc.ClientConn
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
			name:   "TestPeerService_GetMorePeers:success",
			fields: fields{},
			args: args{
				connection: mockGrpcDialer(),
			},
			want: &model.GetMorePeersResponse{
				Peers: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8001,
					},
					{
						Address: "118.99.96.66",
						Port:    3001,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "TestPeerService_GetMorePeers:success",
			fields: fields{},
			args: args{
				connection: faildDial,
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
			got, err := cs.GetMorePeers(tt.args.connection)
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
