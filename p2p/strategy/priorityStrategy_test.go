package strategy

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
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

	p2pP1 = &model.Peer{
		Info: &model.Node{
			ID:      1111,
			Port:    8080,
			Address: "127.0.0.1",
		},
	}
	p2pP2 = &model.Peer{
		Info: &model.Node{
			ID:      2222,
			Port:    9090,
			Address: "127.0.0.2",
		},
	}
	p2pChunk1Bytes = []byte{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	}
	p2pChunk2Bytes = []byte{
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	}
	p2pChunk2InvalidBytes = []byte{
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0,
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
	p2pMockNodeRegistraionService struct {
		coreService.NodeRegistrationService
		successGetNodeRegistrationByNodePublicKey bool
		successGetNodeAddressesInfoFromDb         bool
		successGenerateNodeAddressInfo            bool
		nodeRegistration                          *model.NodeRegistration
		nodeAddressesInfo                         []*model.NodeAddressInfo
		nodeAddresesInfo                          *model.NodeAddressInfo
		addressInfoUpdated                        bool
	}
	p2pMockPeerServiceClient struct {
		client.PeerServiceClient
		noFailedDownloads bool
		downloadErr       bool
		returnInvalidData bool
	}
	p2pMockNodeStatusService struct {
		coreService.NodeStatusService
		host *model.Host
	}
)

func (p2pNssMock *p2pMockNodeStatusService) GetHost() *model.Host {
	if p2pNssMock.host != nil {
		return p2pNssMock.host
	}
	return &model.Host{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	}
}

func (p2pMpsc *p2pMockPeerServiceClient) SendNodeAddressInfo(
	destPeer *model.Peer,
	nodeAddressInfo *model.NodeAddressInfo,
) (*model.Empty, error) {
	return nil, nil
}

func (p2pNr *p2pMockNodeRegistraionService) UpdateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (updated bool, err error) {
	return p2pNr.addressInfoUpdated, nil
}

func (p2pNr *p2pMockNodeRegistraionService) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error) {
	if p2pNr.successGetNodeRegistrationByNodePublicKey {
		if p2pNr.nodeRegistration != nil {
			return p2pNr.nodeRegistration, nil
		}
		return &model.NodeRegistration{
			NodeID:             111,
			AccountAddress:     "OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			LockedBalance:      10000000,
			RegistrationHeight: 10,
			RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
			Latest:             true,
			NodePublicKey:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Height:             100,
		}, nil
	}
	return nil, errors.New("MockedError")
}

func (p2pNr *p2pMockNodeRegistraionService) GetNodeAddressesInfoFromDb(nodeIDs []int64) ([]*model.NodeAddressInfo, error) {
	if p2pNr.successGetNodeAddressesInfoFromDb {
		if len(p2pNr.nodeAddressesInfo) > 0 {
			return p2pNr.nodeAddressesInfo, nil
		}
		return []*model.NodeAddressInfo{
			{
				NodeID:           111,
				Address:          "192.168.1.1",
				Port:             8080,
				Signature:        make([]byte, 64),
				BlockHash:        make([]byte, 32),
				BlockHeight:      100,
				UpdatedTimestamp: 1234567890,
			},
		}, nil
	}
	return nil, errors.New("MockedError")
}

func (p2pNr *p2pMockNodeRegistraionService) GenerateNodeAddressInfo(
	nodeID int64,
	nodeAddress string,
	port uint32,
	nodeSecretPhrase string) (*model.NodeAddressInfo, error) {
	if p2pNr.successGenerateNodeAddressInfo {
		if p2pNr.nodeAddresesInfo != nil {
			return p2pNr.nodeAddresesInfo, nil
		}
		return &model.NodeAddressInfo{
			NodeID:           111,
			Address:          "192.168.1.1",
			Port:             8080,
			Signature:        make([]byte, 64),
			BlockHash:        make([]byte, 32),
			BlockHeight:      100,
			UpdatedTimestamp: 1234567890,
		}, nil
	}
	return nil, errors.New("MockedError")
}

func (*mockPeerServiceClientSuccess) DeleteConnection(destPeer *model.Peer) error {
	return nil
}

func (*mockPeerServiceClientSuccess) SendNodeAddressInfo(destPeer *model.Peer, nodeAddressInfo *model.NodeAddressInfo) (*model.Empty, error) {
	return nil, nil
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
		host               *model.Host
		peerServiceClient  client.PeerServiceClientInterface
		queryExecutor      query.ExecutorInterface
		logger             *log.Logger
		peerStrategyHelper PeerStrategyHelperInterface
		nodeStatusService  coreService.NodeStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *PriorityStrategy
	}{
		{
			name: "wantSuccess",
			args: args{
				peerStrategyHelper: NewPeerStrategyHelper(),
				nodeStatusService:  &coreService.NodeStatusService{},
			},
			want: &PriorityStrategy{
				MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
				MaxResolvedPeers:   constant.MaxResolvedPeers,
				PeerStrategyHelper: NewPeerStrategyHelper(),
				NodeStatusService:  &coreService.NodeStatusService{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPriorityStrategy(tt.args.peerServiceClient, nil,
				tt.args.queryExecutor, nil, tt.args.logger, tt.args.peerStrategyHelper, tt.args.nodeStatusService)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPriorityStrategy() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetResolvedPeers(t *testing.T) {
	type fields struct {
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: goodResolvedPeers,
		},
		{
			name: "wantUnresolvedPeersPopulatedWithKnownPeers",
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
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
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetUnresolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService:  tt.fields.NodeStatusService,
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
		nodeStatusService  coreService.NodeStatusServiceInterface
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
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				log.New(), nil, tt.args.nodeStatusService)
			changeMaxUnresolvedPeers(ps, tt.args.MaxUnresolvedPeers)
			err := ps.AddToUnresolvedPeers([]*model.Node{tt.args.newNode}, tt.args.toForceAdd)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToUnresolvedPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := ps.GetUnresolvedPeers()
				for fullAddressPeer, peer := range peers {
					if p2pUtil.GetFullAddressPeer(tt.wantContain) == fullAddressPeer {
						tt.wantContain.UnresolvingTime = peer.GetUnresolvingTime()
						if reflect.DeepEqual(peer, tt.wantContain) {
							return
						}
					}
				}
				t.Errorf("AddToUnresolvedPeers() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestPriorityStrategy_RemoveUnresolvedPeer(t *testing.T) {
	type args struct {
		nodeStatusService coreService.NodeStatusServiceInterface
		peerToRemove      *model.Peer
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
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				nodeStatusService: nil,
				peerToRemove:      nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeStatusService)
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
		nodeStatusService coreService.NodeStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetBlacklistedPeersTest",
			args: args{
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
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
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeStatusService)
			if got := ps.GetBlacklistedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlacklistedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToBlacklistedPeer(t *testing.T) {
	type args struct {
		nodeStatusService coreService.NodeStatusServiceInterface
		newPeer           *model.Peer
	}
	tests := []struct {
		name        string
		args        args
		reason      string
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToBlacklistedPeer success",
			args: args{
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			reason: "error",
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
			name: "Host:AddToBlacklistedPeer fails",
			args: args{
				nodeStatusService: nil,
				newPeer:           nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeStatusService)
			err := ps.AddToBlacklistedPeer(tt.args.newPeer, tt.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToBlacklistedPeer error = %v, wantErr %v", err, tt.wantErr)
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
		nodeStatusService coreService.NodeStatusServiceInterface
		peerToRemove      *model.Peer
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
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				nodeStatusService: nil,
				peerToRemove:      nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeStatusService)
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
		nodeStatusService coreService.NodeStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyKnownPeerTest",
			args: args{
				nodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeStatusService)
			if got := ps.GetAnyKnownPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetExceedMaxUnresolvedPeers(t *testing.T) {
	mockNodeStatusService := &p2pMockNodeStatusService{
		host: &model.Host{
			UnresolvedPeers: make(map[string]*model.Peer),
		},
	}
	ps := NewPriorityStrategy(nil, nil, nil, nil, nil, nil, mockNodeStatusService)
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
	mockNodeStatusService := &p2pMockNodeStatusService{
		host: &model.Host{
			ResolvedPeers: make(map[string]*model.Peer),
		},
	}
	ps := NewPriorityStrategy(nil, nil, nil, nil, nil, nil, mockNodeStatusService)
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
		NodeStatusService coreService.NodeStatusServiceInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						Info: mockHostInfo,
					},
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
				NodeStatusService:       tt.fields.NodeStatusService,
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
		NodeStatusService     coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						Info: mockHostInfo,
					},
				},
			},
			want: mockHostInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:  tt.fields.NodeStatusService,
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
		NodeStatusService       coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						UnresolvedPeers: make(map[string]*model.Peer),
						Info:            mockGoodScrambledNodes.AddressNodes[0].GetInfo(),
					},
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService:       tt.fields.NodeStatusService,
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
		NodeStatusService       coreService.NodeStatusServiceInterface
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
				NodeStatusService: &p2pMockNodeStatusService{
					host: priorityStrategyGoodHostInstance,
				},
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
				NodeStatusService:       tt.fields.NodeStatusService,
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

type (
	psMockNodeRegistrationService struct {
		coreService.NodeRegistrationServiceInterface
		validateAddressInfoSuccess bool
	}
	psMockPeerStrategyHelper struct {
		PeerStrategyHelperInterface
		peer *model.Peer
	}
)

func (psMock *psMockNodeRegistrationService) ValidateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) error {
	if psMock.validateAddressInfoSuccess {
		return nil
	}
	return errors.New("MockedError")
}

func (psMock *psMockNodeRegistrationService) UpdateNodeAddressInfo(nodeAddressMessage *model.NodeAddressInfo) (updated bool, err error) {
	return true, nil
}

func (psMock *psMockPeerStrategyHelper) GetRandomPeerWithoutRepetition(peers map[string]*model.Peer, mutex *sync.Mutex) *model.Peer {
	if psMock.peer != nil {
		return psMock.peer
	}
	return &model.Peer{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	}
}

func (*mockPeerServiceClientFail) GetNodeAddressesInfo(
	destPeer *model.Peer,
	nodeRegistrations []*model.NodeRegistration,
) (*model.GetNodeAddressesInfoResponse, error) {
	return nil, errors.New("MockedError")
}

func (*mockPeerServiceClientSuccess) GetNodeAddressesInfo(
	destPeer *model.Peer,
	nodeRegistrations []*model.NodeRegistration,
) (*model.GetNodeAddressesInfoResponse, error) {
	return &model.GetNodeAddressesInfoResponse{
		NodeAddressesInfo: []*model.NodeAddressInfo{
			{
				Address:     "127.0.0.1",
				Port:        3000,
				Signature:   make([]byte, 64),
				BlockHash:   make([]byte, 32),
				BlockHeight: 10,
			},
		},
	}, nil
}

func TestPriorityStrategy_GetNodeAddressesInfo(t *testing.T) {
	type fields struct {
		NodeStatusService       coreService.NodeStatusServiceInterface
		PeerServiceClient       client.PeerServiceClientInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		ResolvedPeersLock       sync.RWMutex
		UnresolvedPeersLock     sync.RWMutex
		BlacklistedPeersLock    sync.RWMutex
		MaxUnresolvedPeers      int32
		MaxResolvedPeers        int32
		Logger                  *log.Logger
		PeerStrategyHelper      PeerStrategyHelperInterface
	}
	type args struct {
		nodeRegistrations []*model.NodeRegistration
	}

	nodeRegistrations := []*model.NodeRegistration{
		{
			NodeID:             111,
			Height:             10,
			NodePublicKey:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			Latest:             true,
			RegistrationStatus: 0,
			RegistrationHeight: 0,
			LockedBalance:      0,
			AccountAddress:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetNodeAddressesInfo:fail-{noResolvedPeers}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: make(map[string]*model.Peer, 0),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GetNodeAddressesInfo:fail-{ErrorGettingNodeAddressInfo}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper: NewPeerStrategyHelper(),
				PeerServiceClient:  &mockPeerServiceClientFail{},
				Logger:             log.New(),
			},
			want: make([]*model.NodeAddressInfo, 0),
		},
		{
			name: "GetNodeAddressesInfo:fail-{ErrorValidatingNodeAddressInfo}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper:      NewPeerStrategyHelper(),
				PeerServiceClient:       &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{},
				Logger:                  log.New(),
			},
			want: make([]*model.NodeAddressInfo, 0),
		},
		{
			name: "GetNodeAddressesInfo:success",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper: &psMockPeerStrategyHelper{},
				PeerServiceClient:  &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{
					validateAddressInfoSuccess: true,
				},
				Logger: log.New(),
			},
			want: []*model.NodeAddressInfo{
				{
					Address:     "127.0.0.1",
					Port:        3000,
					Signature:   make([]byte, 64),
					BlockHash:   make([]byte, 32),
					BlockHeight: 10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:       tt.fields.NodeStatusService,
				PeerServiceClient:       tt.fields.PeerServiceClient,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				ResolvedPeersLock:       tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:     tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock:    tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:      tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:        tt.fields.MaxResolvedPeers,
				Logger:                  tt.fields.Logger,
				PeerStrategyHelper:      tt.fields.PeerStrategyHelper,
			}
			got, err := ps.GetNodeAddressesInfo(tt.args.nodeRegistrations)
			if (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.GetNodeAddressesInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetNodeAddressesInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_ReceiveNodeAddressInfo(t *testing.T) {
	type fields struct {
		NodeStatusService       coreService.NodeStatusServiceInterface
		PeerServiceClient       client.PeerServiceClientInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		ResolvedPeersLock       sync.RWMutex
		UnresolvedPeersLock     sync.RWMutex
		BlacklistedPeersLock    sync.RWMutex
		MaxUnresolvedPeers      int32
		MaxResolvedPeers        int32
		Logger                  *log.Logger
		PeerStrategyHelper      PeerStrategyHelperInterface
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ReceiveNodeAddressInfo:success",
			args: args{
				nodeAddressInfo: &model.NodeAddressInfo{},
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerServiceClient:       &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{},
				Logger:                  log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:       tt.fields.NodeStatusService,
				PeerServiceClient:       tt.fields.PeerServiceClient,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				ResolvedPeersLock:       tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:     tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock:    tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:      tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:        tt.fields.MaxResolvedPeers,
				Logger:                  tt.fields.Logger,
				PeerStrategyHelper:      tt.fields.PeerStrategyHelper,
			}
			if err := ps.ReceiveNodeAddressInfo(tt.args.nodeAddressInfo); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.ReceiveNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_UpdateOwnNodeAddressInfo(t *testing.T) {
	type fields struct {
		NodeStatusService       coreService.NodeStatusServiceInterface
		PeerServiceClient       client.PeerServiceClientInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		ResolvedPeersLock       sync.RWMutex
		UnresolvedPeersLock     sync.RWMutex
		BlacklistedPeersLock    sync.RWMutex
		MaxUnresolvedPeers      int32
		MaxResolvedPeers        int32
		Logger                  *log.Logger
		PeerStrategyHelper      PeerStrategyHelperInterface
	}
	type args struct {
		nodeAddress      string
		port             uint32
		nodeSecretPhrase string
	}
	peers := make(map[string]*model.Peer)
	peers[p2pP1.Info.Address] = p2pP1
	peers[p2pP2.Info.Address] = p2pP2

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "UpdateOwnNodeAddressInfo:success-{recordNotUpdated}",
			args: args{
				nodeAddress:      "192.0.0.1",
				port:             8080,
				nodeSecretPhrase: "itsasecret",
			},
			fields: fields{
				NodeRegistrationService: &p2pMockNodeRegistraionService{
					successGetNodeRegistrationByNodePublicKey: true,
					successGenerateNodeAddressInfo:            true,
					successGetNodeAddressesInfoFromDb:         true,
				},
			},
		},
		{
			name: "UpdateOwnNodeAddressInfo:success-{recordUpdated}",
			args: args{
				nodeAddress:      "192.0.0.2",
				port:             8080,
				nodeSecretPhrase: "itsasecret",
			},
			fields: fields{
				NodeStatusService: &p2pMockNodeStatusService{
					host: &model.Host{
						ResolvedPeers: peers,
					},
				},
				NodeRegistrationService: &p2pMockNodeRegistraionService{
					successGetNodeRegistrationByNodePublicKey: true,
					successGenerateNodeAddressInfo:            true,
					successGetNodeAddressesInfoFromDb:         true,
					addressInfoUpdated:                        true,
				},
				PeerServiceClient: &p2pMockPeerServiceClient{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeStatusService:       tt.fields.NodeStatusService,
				PeerServiceClient:       tt.fields.PeerServiceClient,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				ResolvedPeersLock:       tt.fields.ResolvedPeersLock,
				UnresolvedPeersLock:     tt.fields.UnresolvedPeersLock,
				BlacklistedPeersLock:    tt.fields.BlacklistedPeersLock,
				MaxUnresolvedPeers:      tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:        tt.fields.MaxResolvedPeers,
				Logger:                  tt.fields.Logger,
				PeerStrategyHelper:      tt.fields.PeerStrategyHelper,
			}
			if err := ps.UpdateOwnNodeAddressInfo(tt.args.nodeAddress, tt.args.port, tt.args.nodeSecretPhrase); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.UpdateOwnNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
