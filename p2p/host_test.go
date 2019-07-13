package p2p

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestInitHostService(t *testing.T) {
	type args struct {
		myAddress      string
		port           uint32
		wellknownPeers []string
	}
	tests := []struct {
		name    string
		args    args
		want    *HostService
		wantErr bool
	}{
		//TODO: Add test cases.
		{
			name: "InitHostService:error",
			args: args{
				myAddress:      "192.168.9.7",
				port:           3002,
				wellknownPeers: []string{"192.168.5.167:40OO", "192.168.5.54:4000s"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "InitHostService:success",
			args: args{
				myAddress:      "192.168.9.7",
				port:           3002,
				wellknownPeers: []string{"192.168.5.167:4000", "192.168.5.54:4001"},
			},
			want: &HostService{
				Host: &model.Host{
					Info: &model.Node{
						Address: "192.168.9.7",
						Port:    3002,
					},
					KnownPeers: map[string]*model.Peer{
						"192.168.5.167:4000": &model.Peer{
							Info: &model.Node{
								Address: "192.168.5.167",
								Port:    4000,
							},
						},
						"192.168.5.54:4001": &model.Peer{
							Info: &model.Node{
								Address: "192.168.5.54",
								Port:    4001,
							},
						},
					},
					UnresolvedPeers: map[string]*model.Peer{
						"192.168.5.167:4000": &model.Peer{
							Info: &model.Node{
								Address: "192.168.5.167",
								Port:    4000,
							},
						},
						"192.168.5.54:4001": &model.Peer{
							Info: &model.Node{
								Address: "192.168.5.54",
								Port:    4001,
							},
						},
					},
				},
				ChainType:  nil,
				GrpcServer: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitHostService(tt.args.myAddress, tt.args.port, tt.args.wellknownPeers)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitHostService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cmp.Equal(got, tt.want) {
					t.Errorf("\n%v\n%v", got, tt.want)
				}
			}
		})
	}
}
