package strategy

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
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
	goodResolvedPeers = map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          3000,
			},
		},
	}
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
		UnresolvedPeers: goodResolvedPeers,
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

	indexScramble = []int{
		0: 0,
		1: 1,
	}

	mockHostInfo = &model.Node{
		Address: "127.0.0.1",
		Port:    8000,
	}

	mockPeer = &model.Peer{
		Info: &model.Node{
			Address: "127.0.0.1",
			Port:    3001,
		},
	}

	mockGoodScrambledNodes = &model.ScrambledNodes{
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
			"127.0.0.1:8000": &indexScramble[0],
			"127.0.0.1:3001": &indexScramble[1],
		},
	}

	mockGoodBlock = &model.Block{
		ID:                   0,
		BlockHash:            nil,
		PreviousBlockHash:    nil,
		Height:               0,
		Timestamp:            0,
		BlockSeed:            nil,
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		SmithScale:           0,
		BlocksmithPublicKey:  nil,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Version:              0,
		PayloadLength:        0,
		PayloadHash:          nil,
		Transactions:         nil,
		PublishedReceipts:    nil,
	}
)

type (
	mockQueryExecutorSuccess struct {
		query.Executor
	}
	mockPeerServiceClientSuccess struct {
		client.PeerServiceClient
	}

	mockPeerServiceClientFail struct {
		client.PeerServiceClient
	}
)

func (*mockPeerServiceClientSuccess) DeleteConnection(destPeer *model.Peer) error {
	return nil
}

func (*mockPeerServiceClientFail) DeleteConnection(destPeer *model.Peer) error {
	return errors.New("mockedError")
}
func (*mockQueryExecutorSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(qe).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetTimestamp(),
			mockGoodBlock.GetBlockSeed(),
			mockGoodBlock.GetBlockSignature(),
			mockGoodBlock.GetCumulativeDifficulty(),
			mockGoodBlock.GetSmithScale(),
			mockGoodBlock.GetPayloadLength(),
			mockGoodBlock.GetPayloadHash(),
			mockGoodBlock.GetBlocksmithPublicKey(),
			mockGoodBlock.GetTotalAmount(),
			mockGoodBlock.GetTotalFee(),
			mockGoodBlock.GetTotalCoinBase(),
			mockGoodBlock.GetVersion(),
		),
	)
	return db.QueryRow(qe), nil
}

func TestNewPriorityStrategy(t *testing.T) {
	type args struct {
		host              *model.Host
		peerServiceClient client.PeerServiceClientInterface
		queryExecutor     query.ExecutorInterface
		Logger            *log.Logger
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
				MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
				MaxResolvedPeers:   constant.MaxResolvedPeers,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPriorityStrategy(tt.args.host, tt.args.peerServiceClient, nil, tt.args.queryExecutor, nil, tt.args.Logger)
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
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
				Host:              priorityStrategyGoodHostInstance,
				PeerServiceClient: &mockPeerServiceClientSuccess{},
			},
			args: args{
				peer: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
			},
			wantErr: false,
		},
		{
			name: "wantFail",
			fields: fields{
				Host:              priorityStrategyGoodHostInstance,
				PeerServiceClient: &mockPeerServiceClientFail{},
			},
			args: args{
				peer: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
			want: goodResolvedPeers,
		},
		{
			name: "wantUnresolvedPeersPopulatedWithKnownPeers",
			fields: fields{
				Host: &model.Host{
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
					UnresolvedPeers: make(map[string]*model.Peer),
				},
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
		{
			name: "wantUnresolvedPeersPopulatedWithKnownPeersWithoutDuplication",
			fields: fields{
				Host: &model.Host{
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
						"127.0.0.1:8000": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8000,
							},
						},
						"127.0.0.1:8001": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8001,
							},
						},
						"127.0.0.1:8002": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8002,
							},
						},
					},
					UnresolvedPeers: make(map[string]*model.Peer),
					ResolvedPeers: map[string]*model.Peer{
						"127.0.0.1:8001": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8001,
							},
						},
					},
					BlacklistedPeers: map[string]*model.Peer{
						"127.0.0.1:8002": {
							Info: &model.Node{
								SharedAddress: "127.0.0.1",
								Address:       "127.0.0.1",
								Port:          8002,
							},
						},
					},
				},
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
			ps := &PriorityStrategy{
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, logrus.New())
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, nil)
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, nil)
			if got := ps.GetBlacklistedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlacklistedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToBlacklistPeer(t *testing.T) {
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
			name: "Host:AddToBlacklistPeer success",
			args: args{
				hostInstance: priorityStrategyGoodHostInstance,
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
					BlacklistingCause: "error",
				},
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
				BlacklistingCause: "error",
				BlacklistingTime:  uint64(time.Now().Unix()),
			},
			wantErr: false,
		},
		{
			name: "Host:AddToBlacklistPeer fails",
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, nil)
			err := ps.AddToBlacklistPeer(tt.args.newPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToBlacklistPeer error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetBlacklistedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToBlacklistPeer() = %v, want %v", peers, tt.wantContain)
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, nil)
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
			ps := NewPriorityStrategy(tt.args.hostInstance, nil, nil, nil, nil, nil)
			if got := ps.GetAnyKnownPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetExceedMaxUnresolvedPeers(t *testing.T) {
	ps := NewPriorityStrategy(&model.Host{
		UnresolvedPeers: make(map[string]*model.Peer),
	}, nil, nil, nil, nil, nil)
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
	}, nil, nil, nil, nil, nil)
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

type (
	mockNodeRegistrationService struct {
		coreService.NodeRegistrationServiceInterface
	}
)

func (*mockNodeRegistrationService) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockGoodScrambledNodes, nil
}

func TestPriorityStrategy_GetPriorityPeers(t *testing.T) {
	var mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	type fields struct {
		Host *model.Host
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
					Info: mockHostInfo,
				},
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3001": mockPeer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                    tt.fields.Host,
				NodeRegistrationService: mockNodeRegistrationServiceInstance,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:           &mockQueryExecutorSuccess{},
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
					Info: mockHostInfo,
				},
			},
			want: mockHostInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:               tt.fields.Host,
				PeerServiceClient:  tt.fields.PeerServiceClient,
				QueryExecutor:      tt.fields.QueryExecutor,
				MaxUnresolvedPeers: tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:   tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetHostInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetHostInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ValidatePriorityPeer(t *testing.T) {
	var mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	type args struct {
		host           *model.Node
		peer           *model.Node
		scrambledNodes *model.ScrambledNodes
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "wantSuccess",
			args: args{
				host:           mockGoodScrambledNodes.AddressNodes[0].GetInfo(),
				peer:           mockGoodScrambledNodes.AddressNodes[1].GetInfo(),
				scrambledNodes: mockGoodScrambledNodes,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeRegistrationService: mockNodeRegistrationServiceInstance,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:           &mockQueryExecutorSuccess{},
			}
			if got := ps.ValidatePriorityPeer(tt.args.scrambledNodes, tt.args.host, tt.args.peer); got != tt.want {
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
			p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(mockGoodScrambledNodes.AddressNodes[1].GetInfo()),
		}
		mockHeader                          = metadata.New(mockMetadata)
		mockGoodMetadata                    = metadata.NewIncomingContext(context.Background(), mockHeader)
		mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	)

	type fields struct {
		Host                    *model.Host
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		Logger                  *log.Logger
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
					Info: mockGoodScrambledNodes.AddressNodes[0].GetInfo(),
				},
				NodeRegistrationService: mockNodeRegistrationServiceInstance,
			},
			args: args{
				ctx: mockGoodMetadata,
			},
			want: true,
		},
		{
			name: "wantSuccess:notScramble",
			fields: fields{
				Host:                    priorityStrategyGoodHostInstance,
				NodeRegistrationService: mockNodeRegistrationServiceInstance,
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
				Host:                    tt.fields.Host,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:           &mockQueryExecutorSuccess{},
			}
			if got := ps.ValidateRequest(tt.args.ctx); got != tt.want {
				t.Errorf("PriorityStrategy.ValidateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ConnectPriorityPeersGradually(t *testing.T) {
	var mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	type fields struct {
		Host                    *model.Host
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		MaxUnresolvedPeers      int32
		MaxResolvedPeers        int32
		Logger                  *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Host:                    priorityStrategyGoodHostInstance,
				NodeRegistrationService: mockNodeRegistrationServiceInstance,
				MaxResolvedPeers:        2,
				MaxUnresolvedPeers:      2,
				Logger:                  log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				Host:                    tt.fields.Host,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				MaxUnresolvedPeers:      tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:        tt.fields.MaxResolvedPeers,
				Logger:                  tt.fields.Logger,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:           &mockQueryExecutorSuccess{},
			}
			ps.ConnectPriorityPeersGradually()
		})
	}
}
