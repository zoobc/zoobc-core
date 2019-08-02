package service

import (
	"context"
	"net"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type (
	mockServerService struct {
		ServerService
	}
)

var lis *bufconn.Listener

func init() {
	const bufSize = 1024 * 1024
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	service.RegisterP2PCommunicationServer(s, &mockServerService{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
func mockGrpcDialer(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	return conn, err
}
func mockFailedGrpcDialer(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(nativeUtil.GetFullAddressPeer(destinationPeer), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// -----
func (mss *mockServerService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return &model.Node{
		Address: "127.0.0.1",
		Port:    8001,
	}, nil
}

func (mss *mockServerService) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	return &model.GetMorePeersResponse{
		Peers: []*model.Node{
			{
				Address: "127.0.0.1",
				Port:    8002,
			},
			{
				Address: "118.99.96.66",
				Port:    3001,
			},
		},
	}, nil
}

func (mss *mockServerService) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	return &model.Empty{}, nil
}

func (mss *mockServerService) SendBlock(ctx context.Context, req *model.Block) (*model.Empty, error) {
	return &model.Empty{}, nil
}

func (mss *mockServerService) SendTransaction(ctx context.Context, req *model.SendTransactionRequest) (*model.Empty, error) {
	return &model.Empty{}, nil
}

func TestNewPeerServiceClient(t *testing.T) {
	tests := []struct {
		name string
		want *PeerServiceClient
	}{
		// Add test cases.
		{
			name: "TestNewPeerServiceClient",
			want: &PeerServiceClient{
				Dialer: func(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
					conn, err := grpc.Dial(nativeUtil.GetFullAddressPeer(destinationPeer), grpc.WithInsecure())
					if err != nil {
						return nil, err
					}
					return conn, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPeerServiceClient()
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewPeerServiceClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerServiceClient_GetPeerInfo(t *testing.T) {
	type fields struct {
		Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)
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
		// Add test cases.
		{
			name: "TestPeerServiceClient_GetPeerInfo",
			fields: fields{
				Dialer: mockGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
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
			name: "TestPeerServiceClient_GetPeerInfo:error",
			fields: fields{
				Dialer: mockFailedGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psc := &PeerServiceClient{
				Dialer: tt.fields.Dialer,
			}
			got, err := psc.GetPeerInfo(tt.args.destPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerServiceClient.GetPeerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerServiceClient.GetPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerServiceClient_GetMorePeers(t *testing.T) {
	type fields struct {
		Dialer Dialer
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
		// Add test cases.
		{
			name: "TestPeerServiceClient_GetMorePeers:success",
			fields: fields{
				Dialer: mockGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
			},
			want: &model.GetMorePeersResponse{
				Peers: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8002,
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
			name: "TestPeerServiceClient_GetMorePeers:error",
			fields: fields{
				Dialer: mockFailedGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psc := &PeerServiceClient{
				Dialer: tt.fields.Dialer,
			}
			got, err := psc.GetMorePeers(tt.args.destPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerServiceClient.GetMorePeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerServiceClient.GetMorePeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerServiceClient_SendPeers(t *testing.T) {
	type fields struct {
		Dialer Dialer
	}
	type args struct {
		destPeer  *model.Peer
		peersInfo []*model.Node
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
			name: "TestPeerServiceClient_SendPeers:success",
			fields: fields{
				Dialer: mockGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				peersInfo: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8002,
					},
					{
						Address: "127.0.0.1",
						Port:    8003,
					},
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
		{
			name: "TestPeerServiceClient_SendPeers:error",
			fields: fields{
				Dialer: mockFailedGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				peersInfo: []*model.Node{
					{
						Address: "127.0.0.1",
						Port:    8002,
					},
					{
						Address: "127.0.0.1",
						Port:    8003,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psc := &PeerServiceClient{
				Dialer: tt.fields.Dialer,
			}
			got, err := psc.SendPeers(tt.args.destPeer, tt.args.peersInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerServiceClient.SendPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerServiceClient.SendPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerServiceClient_SendBlock(t *testing.T) {
	type fields struct {
		Dialer Dialer
	}
	type args struct {
		destPeer *model.Peer
		block    *model.Block
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
			name: "TestPeerServiceClient_SendBlock:success",
			fields: fields{
				Dialer: mockGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				block: &model.Block{
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
		{
			name: "TestPeerServiceClient_SendBlock:error",
			fields: fields{
				Dialer: mockFailedGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				block: &model.Block{
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
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psc := &PeerServiceClient{
				Dialer: tt.fields.Dialer,
			}
			got, err := psc.SendBlock(tt.args.destPeer, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerServiceClient.SendBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerServiceClient.SendBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeerServiceClient_SendTransaction(t *testing.T) {
	type fields struct {
		Dialer Dialer
	}
	type args struct {
		destPeer         *model.Peer
		transactionBytes []byte
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
			name: "TestPeerServiceClient_SendTransaction:success",
			fields: fields{
				Dialer: mockGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				transactionBytes: []byte{
					2, 0, 1, 218, 138, 66, 93, 0, 0, 0, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107,
					68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 0, 0, 66, 67, 90, 75,
					76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106,
					102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 96, 0, 0, 0, 0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226,
					235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112,
					80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112,
					84, 106, 103, 116, 78, 9, 49, 50, 55, 46, 48, 46, 48, 46, 49, 160, 134, 1, 0, 0, 0, 0, 0, 118, 96, 0, 82, 83, 206, 138, 84, 224, 106,
					207, 135, 30, 2, 186, 237, 239, 131, 229, 86, 45, 235, 250, 248, 8, 166, 83, 102, 108, 132, 208, 227, 121, 235, 59, 31, 146, 98, 125,
					173, 86, 83, 138, 34, 164, 165, 200, 3, 149, 209, 190, 117, 102, 152, 173, 38, 151, 0, 212, 64, 253, 97, 123, 12,
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
		{
			name: "TestPeerServiceClient_SendTransaction:error",
			fields: fields{
				Dialer: mockFailedGrpcDialer,
			},
			args: args{
				destPeer: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
				},
				transactionBytes: []byte{
					2, 0, 1, 218, 138, 66, 93, 0, 0, 0, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107,
					68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 0, 0, 66, 67, 90, 75,
					76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106,
					102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 96, 0, 0, 0, 0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226,
					235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112,
					80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112,
					84, 106, 103, 116, 78, 9, 49, 50, 55, 46, 48, 46, 48, 46, 49, 160, 134, 1, 0, 0, 0, 0, 0, 118, 96, 0, 82, 83, 206, 138, 84, 224, 106,
					207, 135, 30, 2, 186, 237, 239, 131, 229, 86, 45, 235, 250, 248, 8, 166, 83, 102, 108, 132, 208, 227, 121, 235, 59, 31, 146, 98, 125,
					173, 86, 83, 138, 34, 164, 165, 200, 3, 149, 209, 190, 117, 102, 152, 173, 38, 151, 0, 212, 64, 253, 97, 123, 12,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psc := &PeerServiceClient{
				Dialer: tt.fields.Dialer,
			}
			got, err := psc.SendTransaction(tt.args.destPeer, tt.args.transactionBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerServiceClient.SendTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerServiceClient.SendTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
