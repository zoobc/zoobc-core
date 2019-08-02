package native

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/native/service"
)

func TestService_InitService(t *testing.T) {
	type mockService struct {
		Service
	}
	type fields struct {
		HostService *service.HostService
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
		want    contract.P2PType
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
			want: &Service{
				HostService: service.CreateHostService(&model.Host{
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
			s := &Service{
				HostService: tt.fields.HostService,
				Observer:    tt.fields.Observer,
			}
			got, err := s.InitService(tt.args.myAddress, tt.args.port, tt.args.wellknownPeers, tt.args.obsr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.InitService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.InitService() = %v, want %v", got, tt.want)
			}
		})
	}
}
