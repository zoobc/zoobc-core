package strategy

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/p2p/client"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc/metadata"
)

func changeMaxUnresolvedPeers(hostServiceInstance *PriorityStrategy, newValue int32) {
	hostServiceInstance.MaxUnresolvedPeers = newValue
}

func changeMaxResolvedPeers(hostServiceInstance *PriorityStrategy, newValue int32) {
	hostServiceInstance.MaxResolvedPeers = newValue
}

var (
	priorityStrategyGoodHostInstance = &model.Host{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          8000,
		},
		KnownPeers: map[string]*model.Peer{
			"127.0.0.1:3000": {
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
		ResolvedPeers: map[string]*model.Peer{
			"127.0.0.1:3000": {
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
		UnresolvedPeers: map[string]*model.Peer{
			"127.0.0.1:3000": {
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
		BlacklistedPeers: map[string]*model.Peer{
			"127.0.0.1:3000": {
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}

	indexScrumble = []int{
		0: 0,
		1: 1,
	}

	mockGoodScrumbleNode = ScrambleNode{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					Address: "127.0.0.1",
					Port:    8000,
				},
			},
			1: {
				Info: &model.Node{
					Address: "127.0.0.1",
					Port:    3001,
				},
			},
		},
		IndexNodes: map[string]*int{
			"127.0.0.1:8000": &indexScrumble[0],
			"127.0.0.1:3001": &indexScrumble[1],
		},
	}

	db, mock, _ = sqlmock.New()
)

func TestNewPriorityStrategy(t *testing.T) {
	type args struct {
		host                  *model.Host
		peerServiceClient     client.PeerServiceClientInterface
		queryExecutor         query.ExecutorInterface
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *PriorityStrategy
	}{
		{
			name: "wantSuccess",
			args: args{
				host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			want: &PriorityStrategy{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
				ScrambleNode: ScrambleNode{
					IndexNodes: make(map[string]*int),
				},
				MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
				MaxResolvedPeers:   constant.MaxResolvedPeers,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPriorityStrategy(
				tt.args.host, tt.args.peerServiceClient,
				tt.args.queryExecutor,
				tt.args.nodeRegistrationQuery,
				tt.args.Logger)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPriorityStrategy() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetResolvedPeers(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetResolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetResolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetAnyResolvedPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetAnyResolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetAnyResolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToResolvedPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
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
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			args: args{
				peer: &model.Peer{
					Info: &model.Node{
						Address: "18.0.0.1",
						Port:    8001,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if err := ps.AddToResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.AddToResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_RemoveResolvedPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
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
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			args: args{
				peer: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if err := ps.RemoveResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.RemoveResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_GetUnresolvedPeers(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			want: priorityStrategyGoodHostInstance.GetUnresolvedPeers(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetUnresolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetUnresolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetAnyUnresolvedPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			want: priorityStrategyGoodHostInstance.GetUnresolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetAnyUnresolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetAnyUnresolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToUnresolvedPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
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
			name: "Host:AddToUnresolvedPeer success",
			fields: fields{
				Host: priorityStrategyGoodHostInstance,
			},
			args: args{
				peer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if err := ps.AddToUnresolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.AddToUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_AddToUnresolvedPeers(t *testing.T) {
	type args struct {
		hostInstance       *model.Host
		newNode            *model.Node
		MaxUnresolvedPeers int32
		toForceAdd         bool
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToUnresolvedPeers success",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				newNode: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
				MaxUnresolvedPeers: 100,
				toForceAdd:         true,
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToUnresolvedPeers fail",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				newNode: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
				MaxUnresolvedPeers: 1,
				toForceAdd:         false,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			changeMaxUnresolvedPeers(ps, tt.args.MaxUnresolvedPeers)
			err := ps.AddToUnresolvedPeers([]*model.Node{tt.args.newNode}, tt.args.toForceAdd)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToUnresolvedPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetUnresolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToUnresolvedPeers() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestPriorityStrategy_RemoveUnresolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		peerToRemove *model.Peer
	}
	tests := []struct {
		name           string
		args           args
		wantNotContain *model.Peer
		wantErr        bool
	}{
		{
			name: "Host:RemoveUnresolvedPeer success",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				peerToRemove: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			wantNotContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:RemoveUnresolvedPeer fails",
			args: args{
				hostInstance: nil,
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			err := ps.RemoveUnresolvedPeer(tt.args.peerToRemove)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetUnresolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantNotContain) {
						t.Errorf("RemoveUnresolvedPeer() = %v, want %v", peers, tt.wantNotContain)
					}
				}
			}
		})
	}
}

func TestPriorityStrategy_GetBlacklistedPeers(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetBlacklistedPeersTest",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3000": {
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			if got := ps.GetBlacklistedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlacklistedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToBlacklistedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		newPeer      *model.Peer
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToBlacklistedPeer success",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToBlacklistedPeer fails",
			args: args{
				hostInstance: nil,
				newPeer:      nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			err := ps.AddToBlacklistedPeer(tt.args.newPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToBlacklistedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetBlacklistedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToBlacklistedPeer() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestPriorityStrategy_RemoveBlacklistedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		peerToRemove *model.Peer
	}
	tests := []struct {
		name           string
		args           args
		wantNotContain *model.Peer
		wantErr        bool
	}{
		{
			name: "Host:RemoveBlacklistedPeer success",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				peerToRemove: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			wantNotContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:RemoveBlacklistedPeer fails",
			args: args{
				hostInstance: nil,
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			err := ps.RemoveBlacklistedPeer(tt.args.peerToRemove)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveBlacklistedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetBlacklistedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantNotContain) {
						t.Errorf("RemoveBlacklistedPeer() = %v, want %v", peers, tt.wantNotContain)
					}
				}
			}
		})
	}
}

func TestPriorityStrategy_GetAnyKnownPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyKnownPeerTest",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
			},
			want: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil)
			if got := ps.GetAnyKnownPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetExceedMaxUnresolvedPeers(t *testing.T) {
	ps := NewPriorityStrategy(&model.Host{
		UnresolvedPeers: make(map[string]*model.Peer),
	}, nil, nil, nil, nil)
	changeMaxUnresolvedPeers(ps, 1)

	var expectedResult, exceedMaxUnresolvedPeers int32

	expectedResult = int32(0)
	exceedMaxUnresolvedPeers = ps.GetExceedMaxUnresolvedPeers()
	if exceedMaxUnresolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxUnresolvedPeers() = %v, want %v", exceedMaxUnresolvedPeers, expectedResult)
	}

	_ = ps.AddToUnresolvedPeer(&model.Peer{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	})

	expectedResult = int32(1)
	exceedMaxUnresolvedPeers = ps.GetExceedMaxUnresolvedPeers()
	if exceedMaxUnresolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxUnresolvedPeers() = %v, want %v", exceedMaxUnresolvedPeers, expectedResult)
	}
}

func TestPriorityStrategy_GetExceedMaxResolvedPeers(t *testing.T) {
	ps := NewPriorityStrategy(&model.Host{
		ResolvedPeers: make(map[string]*model.Peer),
	}, nil, nil, nil, nil)
	changeMaxResolvedPeers(ps, 1)

	var expectedResult, exceedMaxResolvedPeers int32

	expectedResult = int32(0)
	exceedMaxResolvedPeers = ps.GetExceedMaxResolvedPeers()
	if exceedMaxResolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxResolvedPeers() = %v, want %v", exceedMaxResolvedPeers, expectedResult)
	}

	_ = ps.AddToResolvedPeer(&model.Peer{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	})

	expectedResult = int32(1)
	exceedMaxResolvedPeers = ps.GetExceedMaxResolvedPeers()
	if exceedMaxResolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxResolvedPeers() = %v, want %v", exceedMaxResolvedPeers, expectedResult)
	}
}

func TestPriorityStrategy_GetPriorityPeers(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
				},
				ScrambleNode: mockGoodScrumbleNode,
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3001": mockGoodScrumbleNode.AddressNodes[1],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetPriorityPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetPriorityPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetHostInfo(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Node
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
				},
				ScrambleNode: mockGoodScrumbleNode,
			},
			want: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetHostInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetHostInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ValidatePriorityPeer(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
	}
	type args struct {
		host *model.Node
		peer *model.Node
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ScrambleNode: mockGoodScrumbleNode,
			},
			args: args{
				host: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
				peer: mockGoodScrumbleNode.AddressNodes[1].GetInfo(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
			}
			if got := ps.ValidatePriorityPeer(tt.args.host, tt.args.peer); got != tt.want {
				t.Errorf("PriorityStrategy.ValidatePriorityPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ValidateRangePriorityPeers(t *testing.T) {
	type args struct {
		peerIndex          int
		hostStartPeerIndex int
		hostEndPeerIndex   int
	}
	type test struct {
		name string
		args args
		want bool
	}

	var (
		Tests        = []test{}
		successCases = []args{
			0: {
				peerIndex:          1,
				hostStartPeerIndex: 1,
				hostEndPeerIndex:   2,
			},
			1: {
				peerIndex:          1,
				hostStartPeerIndex: 1,
				hostEndPeerIndex:   1,
			},
			2: {
				peerIndex:          0,
				hostStartPeerIndex: 3,
				hostEndPeerIndex:   1,
			},
			3: {
				peerIndex:          4,
				hostStartPeerIndex: 4,
				hostEndPeerIndex:   1,
			},
		}
		failedCases = []args{
			0: {
				peerIndex:          0,
				hostStartPeerIndex: 1,
				hostEndPeerIndex:   4,
			},
			1: {
				peerIndex:          1,
				hostStartPeerIndex: 4,
				hostEndPeerIndex:   0,
			},
		}
	)

	for _, args := range successCases {
		newTest := test{
			name: "wantSuccess",
			args: args,
			want: true,
		}
		Tests = append(Tests, newTest)
	}

	for _, args := range failedCases {
		newTest := test{
			name: "wantFail",
			args: args,
			want: false,
		}
		Tests = append(Tests, newTest)
	}

	for _, tt := range Tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{}
			if got := ps.ValidateRangePriorityPeers(tt.args.peerIndex, tt.args.hostStartPeerIndex, tt.args.hostEndPeerIndex); got != tt.want {
				t.Errorf("PriorityStrategy.ValidateRangePriorityPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ValidateRequest(t *testing.T) {
	var (
		mockMetadata = map[string]string{
			p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(mockGoodScrumbleNode.AddressNodes[1].GetInfo()),
		}
		mockHeader       = metadata.New(mockMetadata)
		mockGoodMetadata = metadata.NewIncomingContext(context.Background(), mockHeader)
	)

	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
		Logger                *log.Logger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
				},
				ScrambleNode: mockGoodScrumbleNode,
			},
			args: args{
				ctx: mockGoodMetadata,
			},
			want: true,
		},
		{
			name: "wantSuccess:notScrumble",
			fields: fields{
				Host:         priorityStrategyGoodHostInstance,
				ScrambleNode: mockGoodScrumbleNode,
			},
			args: args{
				ctx: mockGoodMetadata,
			},
			want: true,
		},
		{
			name:   "wantFail:nilDefaultConnectionMetadata",
			fields: fields{},
			args: args{
				ctx: context.Background(),
			},
			want: false,
		},
		{
			name:   "wantFail:nilContext",
			fields: fields{},
			args: args{
				ctx: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
				Logger:                tt.fields.Logger,
			}
			if got := ps.ValidateRequest(tt.args.ctx); got != tt.want {
				t.Errorf("PriorityStrategy.ValidateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	executorBuildScrambleNodesSuccess struct {
		query.Executor
	}
	executorBuildScrambleNodesFail struct {
		query.Executor
	}
)

func (*executorBuildScrambleNodesSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewNodeRegistrationQuery().Fields,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76}, "accountA", 0, "127.0.0.1:3000", 0, false, true, 0,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 78}, "accountB", 0, "127.0.0.1:3001", 0, false, true, 0,
	))

	return db.Query(qStr, 1)
}

func (*executorBuildScrambleNodesFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func TestPriorityStrategy_BuildScrambleNodes(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
		Logger                *log.Logger
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host: &model.Host{
					Info: mockGoodScrumbleNode.AddressNodes[0].GetInfo(),
				},
				QueryExecutor: &executorBuildScrambleNodesSuccess{
					query.Executor{
						Db: db,
					},
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				ScrambleNode: ScrambleNode{
					IndexNodes: make(map[string]*int),
				},
				Logger: log.New(),
			},
			args: args{
				block: &model.Block{
					Height:    1,
					BlockSeed: []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
				},
			},
		},
		{
			name: "wantFail",
			fields: fields{
				QueryExecutor: &executorBuildScrambleNodesFail{
					query.Executor{
						Db: db,
					},
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args: args{
				block: &model.Block{
					Height:    1,
					BlockSeed: []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
				Logger:                tt.fields.Logger,
			}
			ps.BuildScrambleNodes(tt.args.block)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPriorityStrategy_ConnectPriorityPeersGradually(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
		Logger                *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host:               priorityStrategyGoodHostInstance,
				ScrambleNode:       mockGoodScrumbleNode,
				MaxResolvedPeers:   2,
				MaxUnresolvedPeers: 2,
				Logger:             log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
				Logger:                tt.fields.Logger,
			}
			ps.ConnectPriorityPeersGradually()
		})
	}
}

type (
	executorGetBlockBuildScrumbleNodeSuccess struct {
		query.Executor
	}
)

var mockBlockData = model.Block{
	ID:                1,
	PreviousBlockHash: []byte{},
	Height:            11,
	Timestamp:         1,
	BlockSeed: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "1",
	SmithScale:           1,
	PayloadLength:        1,
	PayloadHash:          []byte{},
	BlocksmithPublicKey:  []byte{},
	TotalAmount:          1000,
	TotalFee:             0,
	TotalCoinBase:        1,
	Version:              0,
}

func (*executorGetBlockBuildScrumbleNodeSuccess) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetSmithScale(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	return db.QueryRow(qStr)
}

func TestPriorityStrategy_GetBlockBuildScrumbleNode(t *testing.T) {
	type fields struct {
		Host                  *model.Host
		PeerServiceClient     client.PeerServiceClientInterface
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		ScrambleNode          ScrambleNode
		MaxUnresolvedPeers    int32
		MaxResolvedPeers      int32
		Logger                *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.Block
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor: &executorGetBlockBuildScrumbleNodeSuccess{
					query.Executor{
						Db: db,
					},
				},
			},
			want:    &mockBlockData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                  tt.fields.Host,
				PeerServiceClient:     tt.fields.PeerServiceClient,
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambleNode:          tt.fields.ScrambleNode,
				MaxUnresolvedPeers:    tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:      tt.fields.MaxResolvedPeers,
				Logger:                tt.fields.Logger,
			}
			got, err := ps.GetBlockBuildScrumbleNode()
			if (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.GetBlockBuildScrumbleNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetBlockBuildScrumbleNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
