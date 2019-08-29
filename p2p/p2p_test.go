package p2p

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	"reflect"
	"testing"
)

func TestService_InitService(t *testing.T) {
	type mockService struct {
		Peer2PeerService
	}
	type fields struct {
		HostService *strategy.NativeStrategy
		Observer    *observer.Observer
	}
	type args struct {
		myAddress      string
		port           uint32
		wellknownPeers []string
		obsr           *observer.Observer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    p2p.Peer2PeerServiceInterface
		wantErr bool
	}{
		// Add test cases.
		{
			name: "TestService_InitService:success",
			fields: fields{
				HostService: nil,
				Observer:    observer.NewObserver(),
			},
			args: args{
				myAddress:      "127.0.0.1",
				port:           8001,
				wellknownPeers: []string{"127.0.0.1:8002", "127.0.0.1:8003"},
				obsr:           observer.NewObserver(),
			},
			want: &Peer2PeerService{
				PeerExplorer: strategy.NewNativeStrategy(&model.Host{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
					ResolvedPeers: make(map[string]*model.Peer),
					UnresolvedPeers: map[string]*model.Peer{
						"127.0.0.1:8002": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8002,
							},
						},
						"127.0.0.1:8003": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8003,
							},
						},
					},
				}),
				Observer: observer.NewObserver(),
			},
			wantErr: false,
		},
		{
			name: "TestService_InitService:failed",
			fields: fields{
				HostService: nil,
				Observer:    observer.NewObserver(),
			},
			args: args{
				myAddress:      "127.0.0.1",
				port:           8001,
				wellknownPeers: []string{"127.0.0.1:"},
				obsr:           observer.NewObserver(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Peer2PeerService{
				PeerExplorer: tt.fields.HostService,
				Observer:     tt.fields.Observer,
			}
			got, err := s.InitService(tt.args.myAddress, tt.args.port, tt.args.wellknownPeers, tt.args.obsr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Peer2PeerService.InitService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Peer2PeerService.InitService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetHostInstance(t *testing.T) {
	type fields struct {
		HostService *strategy.NativeStrategy
		Observer    *observer.Observer
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Host
	}{
		// Add test cases.
		{
			name: "TestService_GetHostInstance",
			fields: fields{
				HostService: strategy.NewNativeStrategy(&model.Host{
					Info: &model.Node{
						Address: "127.0.0.1",
						Port:    8001,
					},
					ResolvedPeers: make(map[string]*model.Peer),
					UnresolvedPeers: map[string]*model.Peer{
						"127.0.0.1:8002": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8002,
							},
						},
						"127.0.0.1:8003": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8003,
							},
						},
					},
				}),
				Observer: observer.NewObserver(),
			},
			want: &model.Host{
				Info: &model.Node{
					Address: "127.0.0.1",
					Port:    8001,
				},
				ResolvedPeers: make(map[string]*model.Peer),
				UnresolvedPeers: map[string]*model.Peer{
					"127.0.0.1:8002": {
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          8002,
						},
					},
					"127.0.0.1:8003": {
						Info: &model.Node{
							SharedAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Port:          8003,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Peer2PeerService{
				PeerExplorer: tt.fields.HostService,
				Observer:     tt.fields.Observer,
			}
			if got := s.GetHostInstance(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Peer2PeerService.GetHostInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_SendBlockListener(t *testing.T) {
	type fields struct {
		HostService *strategy.NativeStrategy
		Observer    *observer.Observer
	}
	tests := []struct {
		name   string
		fields fields
		want   observer.Listener
	}{
		{
			name:   "TestService_SendBlockListener",
			fields: fields{},
			want: observer.Listener{
				OnNotify: func(block interface{}, args interface{}) {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Peer2PeerService{
				PeerExplorer: tt.fields.HostService,
				Observer:     tt.fields.Observer,
			}
			got := s.SendBlockListener()
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Peer2PeerService.SendBlockListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_SendTransactionListener(t *testing.T) {
	type fields struct {
		HostService *strategy.NativeStrategy
		Observer    *observer.Observer
	}
	tests := []struct {
		name   string
		fields fields
		want   observer.Listener
	}{
		// Add test cases.
		{
			name:   "TestService_SendTransactionListener",
			fields: fields{},
			want: observer.Listener{
				OnNotify: func(block interface{}, args interface{}) {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Peer2PeerService{
				PeerExplorer: tt.fields.HostService,
				Observer:     tt.fields.Observer,
			}
			got := s.SendTransactionListener()
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Peer2PeerService.SendTransactionListener() = %v, want %v", got, tt.want)
			}
		})
	}
}
