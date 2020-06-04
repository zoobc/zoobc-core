package strategy

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/client"
)

var (
	mockNode = model.Node{
		SharedAddress: "127.0.0.1",
		Address:       "127.0.0.1",
		Port:          8001,
	}
	mockPeers = map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &mockNode,
		},
	}
)

type (
	mockGetMorePeersError struct {
		client.PeerServiceClientInterface
	}
	mockGetMorePeersSuccess struct {
		client.PeerServiceClientInterface
	}
)

func (*mockGetMorePeersError) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	return nil, errors.New("Error GetMorePeers")
}

func (*mockGetMorePeersSuccess) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	return &model.GetMorePeersResponse{}, nil
}

func TestNativeStrategy_GetMorePeersHandler(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.Peer
		wantErr bool
	}{
		{
			name: "GetMorePeersHandler:PeerIsNil",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: map[string]*model.Peer{},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "GetMorePeersHandler:Error",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
				PeerServiceClient: &mockGetMorePeersError{},
				Logger:            &log.Logger{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMorePeersHandler:Success",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
				PeerServiceClient: &mockGetMorePeersSuccess{},
			},
			want: &model.Peer{
				Info: &mockNode,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			got, err := ns.GetMorePeersHandler()
			if (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.GetMorePeersHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NativeStrategy.GetMorePeersHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNativeStrategy_GetResolvedPeers(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "GetResolvedPeers:Success",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
			},
			want: mockPeers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if got := ns.GetResolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NativeStrategy.GetResolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNativeStrategy_GetAnyResolvedPeer(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Peer
	}{
		{
			name: "GetAnyResolvedPeer:Nil",
			fields: fields{
				Host: &model.Host{},
			},
			want: nil,
		},
		{
			name: "GetAnyResolvedPeer:Success",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
			},
			want: &model.Peer{
				Info: &mockNode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if got := ns.GetAnyResolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NativeStrategy.GetAnyResolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNativeStrategy_AddToResolvedPeer(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "AddToResolvedPeer:Error",
			args:    args{},
			wantErr: true,
		},
		{
			name: "AddToResolvedPeer:Success",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
			},
			args: args{
				peer: &model.Peer{
					Info: &mockNode,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if err := ns.AddToResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.AddToResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNativeStrategy_RemoveResolvedPeer(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "RemoveResolvedPeer:Nil",
			wantErr: true,
		},
		{
			name: "RemoveResolvedPeer:Success",
			fields: fields{
				Host: &model.Host{
					ResolvedPeers: mockPeers,
				},
			},
			args: args{
				peer: &model.Peer{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if err := ns.RemoveResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.RemoveResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNativeStrategy_GetUnresolvedPeers(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "GetUnresolvedPeers:Success",
			fields: fields{
				Host: &model.Host{
					UnresolvedPeers: mockPeers,
				},
			},
			want: mockPeers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if got := ns.GetUnresolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NativeStrategy.GetUnresolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNativeStrategy_AddToUnresolvedPeer(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "AddToUnresolvedPeer:Nil",
			wantErr: true,
		},
		{
			name: "AddToUnresolvedPeer:Success",
			fields: fields{
				Host: &model.Host{
					UnresolvedPeers: mockPeers,
				},
			},
			args: args{
				peer: &model.Peer{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if err := ns.AddToUnresolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.AddToUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNativeStrategy_AddToUnresolvedPeers(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	type args struct {
		newNodes []*model.Node
		toForce  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "AddToUnresolvedPeers:Full",
			fields: fields{
				Host: &model.Host{
					UnresolvedPeers: mockPeers,
				},
				MaxUnresolvedPeers: int32(0),
			},
			args: args{
				toForce: false,
			},
			wantErr: true,
		},
		{
			name: "AddToUnresolvedPeers:Success",
			fields: fields{
				Host: &model.Host{
					UnresolvedPeers: mockPeers,
				},
				MaxUnresolvedPeers: int32(0),
			},
			args: args{
				toForce: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if err := ns.AddToUnresolvedPeers(tt.args.newNodes, tt.args.toForce); (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.AddToUnresolvedPeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNativeStrategy_RemoveUnresolvedPeer(t *testing.T) {
	type fields struct {
		Host                 *model.Host
		PeerServiceClient    client.PeerServiceClientInterface
		ResolvedPeersLock    sync.RWMutex
		UnresolvedPeersLock  sync.RWMutex
		BlacklistedPeersLock sync.RWMutex
		MaxUnresolvedPeers   int32
		MaxResolvedPeers     int32
		Logger               *log.Logger
	}
	type args struct {
		peer *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name:    "RemoveUnresolvedPeer:Nil",
			wantErr: true,
		},
		{
			name: "RemoveUnresolvedPeer:Success",
			fields: fields{
				Host: &model.Host{
					UnresolvedPeers: mockPeers,
				},
			},
			args: args{
				peer: &model.Peer{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &NativeStrategy{
				Host:                 tt.fields.Host,
				PeerServiceClient:    tt.fields.PeerServiceClient,
				ResolvedPeersLock:    tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:  tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock: tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:   tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:     tt.fields.MaxResolvedPeers,
				Logger:               tt.fields.Logger,
			}
			if err := ns.RemoveUnresolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("NativeStrategy.RemoveUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
