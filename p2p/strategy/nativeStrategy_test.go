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
