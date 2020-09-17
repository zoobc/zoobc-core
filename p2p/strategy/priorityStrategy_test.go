package strategy

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"
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
		ID:      int64(111),
		Address: "127.0.0.1",
		Port:    8000,
	}

	mockPeer = &model.Peer{
		Info: &model.Node{
			ID:      int64(222),
			Address: "127.0.0.1",
			Port:    3001,
		},
	}

	mockGoodScrambledNodes = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					ID:      int64(111),
					Address: "127.0.0.1",
					Port:    8000,
				},
			},
			1: {
				Info: &model.Node{
					ID:      int64(222),
					Address: "127.0.0.1",
					Port:    3001,
				},
			},
		},
		IndexNodes: map[string]*int{
			"111": &indexScramble[0],
			"222": &indexScramble[1],
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
)

type (
	mockBlockMainServiceSuccess struct {
		coreService.BlockServiceInterface
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
	}
	p2pMockNodeConfigurationService struct {
		coreService.NodeConfigurationService
		host *model.Host
	}
	p2pMockSignature struct {
		crypto.Signature
	}
)

func (p2pSigMock *p2pMockSignature) SignByNode(payload []byte, nodeSeed string) []byte {
	return make([]byte, 64)
}

func (p2pNssMock *p2pMockNodeConfigurationService) SetMyAddress(nodeAddress string, port uint32) {
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetNodeSecretPhrase() string {
	return "itsasecret"
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetNodePublicKey() []byte {
	return []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetHost() *model.Host {
	if p2pNssMock.host != nil {
		return p2pNssMock.host
	}
	return &model.Host{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
		ResolvedPeers:    make(map[string]*model.Peer),
		BlacklistedPeers: make(map[string]*model.Peer),
		KnownPeers:       make(map[string]*model.Peer),
		UnresolvedPeers:  make(map[string]*model.Peer),
	}
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetMyAddress() (string, error) {
	return "127.0.0.1", nil
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetHostID() (int64, error) {
	return 111, nil
}

func (p2pNssMock *p2pMockNodeConfigurationService) GetMyPeerPort() (uint32, error) {

	return 8001, nil
}

func (p2pMpsc *p2pMockPeerServiceClient) SendNodeAddressInfo(
	destPeer *model.Peer,
	nodeAddressInfo *model.NodeAddressInfo,
) (*model.Empty, error) {
	return nil, nil
}

func (p2pNr *p2pMockNodeRegistraionService) GetNodeAddressInfoByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	return []*model.NodeAddressInfo{
		{
			NodeID: int64(111),
			Status: model.NodeAddressStatus_NodeAddressPending,
		},
	}, nil
}

func (p2pNr *p2pMockNodeRegistraionService) UpdateNodeAddressInfo(
	nodeAddressMessage *model.NodeAddressInfo,
	status model.NodeAddressStatus) (updated bool, err error) {
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

func (p2pNr *p2pMockNodeRegistraionService) GetNodeAddressesInfoFromDb(nodeIDs []int64,
	addressStatus []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	if p2pNr.successGetNodeAddressesInfoFromDb {
		if len(p2pNr.nodeAddressesInfo) > 0 {
			return p2pNr.nodeAddressesInfo, nil
		}
		return []*model.NodeAddressInfo{
			{
				NodeID:      111,
				Address:     "192.168.1.1",
				Port:        8080,
				Signature:   make([]byte, 64),
				BlockHash:   make([]byte, 32),
				BlockHeight: 100,
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
			NodeID:      111,
			Address:     "192.168.1.1",
			Port:        8080,
			Signature:   make([]byte, 64),
			BlockHash:   make([]byte, 32),
			BlockHeight: 100,
		}, nil
	}
	return nil, errors.New("MockedError")
}

var (
	nrsAddress1    = "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
	nrsNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
)

func (p2pNr *p2pMockNodeRegistraionService) GetRegisteredNodes() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			NodeID:             int64(111),
			NodePublicKey:      nrsNodePubKey1,
			AccountAddress:     nrsAddress1,
			RegistrationHeight: 10,
			LockedBalance:      100000000,
			RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
			Latest:             true,
			Height:             200,
		},
	}, nil
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
func (*mockBlockMainServiceSuccess) GetLastBlock() (*model.Block, error) {
	return mockGoodBlock, nil
}

func TestPriorityStrategy_GetResolvedPeers(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetResolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetResolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetAnyResolvedPeer(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetResolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetAnyResolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetAnyResolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToResolvedPeer(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if err := ps.AddToResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.AddToResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_RemoveResolvedPeer(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if err := ps.RemoveResolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.RemoveResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_GetUnresolvedPeers(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: goodResolvedPeers,
		},
		{
			name: "wantUnresolvedPeersPopulatedWithKnownPeers",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetUnresolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetUnresolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetAnyUnresolvedPeer(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
			},
			want: priorityStrategyGoodHostInstance.GetUnresolvedPeers()["127.0.0.1:3000"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if got := ps.GetAnyUnresolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetAnyUnresolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToUnresolvedPeer(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
				NodeRegistrationService:  &mockNodeRegistrationService{},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
			}
			if err := ps.AddToUnresolvedPeer(tt.args.peer); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.AddToUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_AddToUnresolvedPeers(t *testing.T) {
	type args struct {
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
		newNode                  *model.Node
		MaxUnresolvedPeers       int32
		toForceAdd               bool
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				log.New(), nil, tt.args.nodeConfigurationService, nil, nil)
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
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
		peerToRemove             *model.Peer
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil, nil,
				nil, tt.args.nodeConfigurationService, nil, nil)
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
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetBlacklistedPeersTest",
			args: args{
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nil, nil, tt.args.nodeConfigurationService, nil, nil)
			if got := ps.GetBlacklistedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlacklistedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_AddToBlacklistedPeer(t *testing.T) {
	type args struct {
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
		newPeer                  *model.Peer
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nodeConfigurationService: nil,
				newPeer:                  nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeConfigurationService, nil, nil)
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
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
		peerToRemove             *model.Peer
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
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nodeConfigurationService: nil,
				peerToRemove:             nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPriorityStrategy(nil, nil, nil, nil,
				nil, nil, tt.args.nodeConfigurationService, nil, nil)
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
		nodeConfigurationService coreService.NodeConfigurationServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyKnownPeerTest",
			args: args{
				nodeConfigurationService: &p2pMockNodeConfigurationService{
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
				nil, nil, tt.args.nodeConfigurationService, nil, nil)
			if got := ps.GetAnyKnownPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetExceedMaxUnresolvedPeers(t *testing.T) {
	mockNodeConfigurationService := &p2pMockNodeConfigurationService{
		host: &model.Host{
			UnresolvedPeers: make(map[string]*model.Peer),
		},
	}
	mockNodeRegistrationService := &p2pMockNodeRegistraionService{}
	ps := NewPriorityStrategy(nil, mockNodeRegistrationService, &mockNodeAddressInfoServiceSuccess{}, nil, nil,
		nil, mockNodeConfigurationService, nil, nil)
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
	mockNodeConfigurationService := &p2pMockNodeConfigurationService{
		host: &model.Host{
			ResolvedPeers: make(map[string]*model.Peer),
		},
	}
	ps := NewPriorityStrategy(nil, nil, nil, nil, nil, nil, mockNodeConfigurationService, nil, nil)
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
	mockNodeAddressInfoServiceSuccess struct {
		coreService.NodeAddressInfoServiceInterface
	}
)

func (*mockNodeAddressInfoServiceSuccess) GetAddressInfoByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	return nil, nil
}

func (*mockNodeAddressInfoServiceSuccess) GetAddressInfoByNodeID(
	nodeID int64,
	addressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	return nil, nil
}

func (*mockNodeRegistrationService) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockGoodScrambledNodes, nil
}

func (*mockNodeRegistrationService) GetNodeAddressInfoByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	return nil, nil
}

func TestPriorityStrategy_GetPriorityPeers(t *testing.T) {
	var mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]*model.Peer
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						Info: mockHostInfo,
					},
				},
			},
			want: map[string]*model.Peer{
				"222": mockPeer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				NodeRegistrationService:  mockNodeRegistrationServiceInstance,
				BlockMainService:         &mockBlockMainServiceSuccess{},
			}
			if got := ps.GetPriorityPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GetPriorityPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityStrategy_GetHostInfo(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		BlockMainService         coreService.BlockServiceInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.Node
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
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
				BlockMainService:        &mockBlockMainServiceSuccess{},
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
			"version":                         mockGoodScrambledNodes.AddressNodes[1].GetInfo().GetVersion(),
			"codename":                        mockGoodScrambledNodes.AddressNodes[1].GetInfo().GetCodeName(),
		}
		mockHeader                          = metadata.New(mockMetadata)
		mockGoodMetadata                    = metadata.NewIncomingContext(context.Background(), mockHeader)
		mockNodeRegistrationServiceInstance = &mockNodeRegistrationService{}
	)

	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		Logger                   *log.Logger
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						UnresolvedPeers: make(map[string]*model.Peer),
						Info:            mockGoodScrambledNodes.AddressNodes[0].GetInfo(),
					},
				},
				NodeRegistrationService:  mockNodeRegistrationServiceInstance,
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
			},
			args: args{
				ctx: mockGoodMetadata,
			},
			want: true,
		},
		{
			name: "wantSuccess:notScramble",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
				NodeRegistrationService:  mockNodeRegistrationServiceInstance,
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				BlockMainService:         &mockBlockMainServiceSuccess{},
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
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: priorityStrategyGoodHostInstance,
				},
				NodeRegistrationService:  mockNodeRegistrationServiceInstance,
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				MaxResolvedPeers:         2,
				MaxUnresolvedPeers:       2,
				Logger:                   log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
				Logger:                   tt.fields.Logger,
				BlockMainService:         &mockBlockMainServiceSuccess{},
			}
			ps.ConnectPriorityPeersGradually()
		})
	}
}

type (
	psMockNodeRegistrationService struct {
		coreService.NodeRegistrationServiceInterface
		validateAddressInfoSuccess bool
		currentNode                *model.NodeRegistration
		currentNodeAddressInfo     *model.NodeAddressInfo
		prevNodeAddressInfo        *model.NodeAddressInfo
	}
	psMockPeerStrategyHelper struct {
		PeerStrategyHelperInterface
	}
)

func (psMock *psMockNodeRegistrationService) GenerateNodeAddressInfo(
	nodeID int64,
	nodeAddress string,
	port uint32,
	nodeSecretPhrase string) (*model.NodeAddressInfo, error) {
	return &model.NodeAddressInfo{
		NodeID:      111,
		Address:     "192.168.1.1",
		Port:        8080,
		Signature:   make([]byte, 64),
		BlockHash:   make([]byte, 32),
		BlockHeight: 100,
	}, nil
}

func (psMock *psMockNodeRegistrationService) ValidateNodeAddressInfo(
	nodeAddressMessage *model.NodeAddressInfo,
) (bool, error) {
	if psMock.validateAddressInfoSuccess {
		return false, nil
	}
	return true, errors.New("MockedError")
}

func (psMock *psMockNodeRegistrationService) UpdateNodeAddressInfo(
	nodeAddressMessage *model.NodeAddressInfo,
	status model.NodeAddressStatus,
) (updated bool, err error) {
	return true, nil
}

func (psMock *psMockNodeRegistrationService) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error) {
	if psMock.currentNode != nil {
		return psMock.currentNode, nil
	}
	return nil, nil
}

func (psMock *psMockNodeRegistrationService) GetNodeRegistrationByNodeID(nodeID int64) (*model.NodeRegistration, error) {
	if psMock.currentNode != nil {
		return psMock.currentNode, nil
	}
	return nil, nil
}

func (psMock *psMockNodeRegistrationService) GetNodeAddressesInfoFromDb(nodeIDs []int64,
	addressStatus []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	if psMock.currentNodeAddressInfo != nil {
		return []*model.NodeAddressInfo{psMock.currentNodeAddressInfo}, nil
	}
	return nil, nil
}

func (psMock *psMockNodeRegistrationService) GetNodeAddressInfoByAddressPort(
	address string,
	port uint32,
	nodeAddressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	if psMock.prevNodeAddressInfo != nil {
		return []*model.NodeAddressInfo{psMock.currentNodeAddressInfo}, nil
	}
	return nil, nil
}

// GetRandomPeerWithoutRepetition spy on method instead of mock it, so we can mock other service methods while returning real values for this
func (psMock *psMockPeerStrategyHelper) GetRandomPeerWithoutRepetition(peers map[string]*model.Peer, mutex *sync.Mutex) *model.Peer {
	realService := NewPeerStrategyHelper()
	return realService.GetRandomPeerWithoutRepetition(peers, mutex)
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
				NodeID:      int64(111),
				Address:     "127.0.0.1",
				Port:        3000,
				Signature:   make([]byte, 64),
				BlockHash:   make([]byte, 32),
				BlockHeight: 10,
			},
		},
	}, nil
}

func TestPriorityStrategy_SyncNodeAddressInfoTable(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		BlockMainService         coreService.BlockServiceInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
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
		want    map[int64]*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetNodeAddressesInfo:fail-{noResolvedPeers}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: make(map[string]*model.Peer),
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper:       NewPeerStrategyHelper(),
				PeerServiceClient:        &mockPeerServiceClientFail{},
				NodeRegistrationService:  &psMockNodeRegistrationService{},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				Logger:                   log.New(),
			},
			want: make(map[int64]*model.NodeAddressInfo),
		},
		{
			name: "GetNodeAddressesInfo:fail-{ErrorValidatingNodeAddressInfo}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper:       NewPeerStrategyHelper(),
				PeerServiceClient:        &mockPeerServiceClientSuccess{},
				NodeRegistrationService:  &psMockNodeRegistrationService{},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				Logger:                   log.New(),
			},
			want: make(map[int64]*model.NodeAddressInfo),
		},
		{
			name: "GetNodeAddressesInfo:success",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper: &psMockPeerStrategyHelper{},
				PeerServiceClient:  &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{
					validateAddressInfoSuccess: true,
				},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				Logger:                   log.New(),
			},
			want: map[int64]*model.NodeAddressInfo{
				int64(111): {
					NodeID:      int64(111),
					Address:     "127.0.0.1",
					Port:        3000,
					Signature:   make([]byte, 64),
					BlockHash:   make([]byte, 32),
					BlockHeight: 10,
				},
			},
		},
		{
			name: "GetNodeAddressesInfo:success-{updateMissingOwnAddressInfo}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper: &psMockPeerStrategyHelper{},
				PeerServiceClient:  &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{
					validateAddressInfoSuccess: true,
					currentNode: &model.NodeRegistration{
						NodeID:             int64(111),
						RegistrationStatus: uint32(0),
					},
				},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				Logger:                   log.New(),
			},
			want: map[int64]*model.NodeAddressInfo{
				int64(111): {
					NodeID:      int64(111),
					Address:     "127.0.0.1",
					Port:        3000,
					Signature:   make([]byte, 64),
					BlockHash:   make([]byte, 32),
					BlockHeight: 10,
				},
			},
		},
		{
			name: "GetNodeAddressesInfo:success-{syncOwnAddressInfoWithPeers}",
			args: args{
				nodeRegistrations: nodeRegistrations,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: goodResolvedPeers,
					},
				},
				PeerStrategyHelper: &psMockPeerStrategyHelper{},
				PeerServiceClient:  &mockPeerServiceClientSuccess{},
				NodeRegistrationService: &psMockNodeRegistrationService{
					validateAddressInfoSuccess: true,
					currentNode: &model.NodeRegistration{
						NodeID:             int64(111),
						RegistrationStatus: uint32(0),
					},
					currentNodeAddressInfo: &model.NodeAddressInfo{
						NodeID:      111,
						Address:     "127.0.0.1",
						Port:        3000,
						Signature:   make([]byte, 64),
						BlockHash:   make([]byte, 32),
						BlockHeight: 100,
					},
				},
				NodeAddressesInfoService: &mockNodeAddressInfoServiceSuccess{},
				Logger:                   log.New(),
			},
			want: map[int64]*model.NodeAddressInfo{
				int64(111): {
					NodeID:      int64(111),
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
				Logger:                   tt.fields.Logger,
				PeerStrategyHelper:       tt.fields.PeerStrategyHelper,
			}
			got, err := ps.SyncNodeAddressInfoTable(tt.args.nodeRegistrations)
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
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		BlockMainService         coreService.BlockServiceInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
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
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
				Logger:                   tt.fields.Logger,
				PeerStrategyHelper:       tt.fields.PeerStrategyHelper,
			}
			if err := ps.ReceiveNodeAddressInfo(tt.args.nodeAddressInfo); (err != nil) != tt.wantErr {
				t.Errorf("PriorityStrategy.ReceiveNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_UpdateOwnNodeAddressInfo(t *testing.T) {
	type fields struct {
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		BlockMainService         coreService.BlockServiceInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
	}
	type args struct {
		nodeAddress    string
		port           uint32
		forceBroadcast bool
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
				nodeAddress: "192.0.0.1",
				port:        8080,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
					host: &model.Host{
						ResolvedPeers: peers,
					},
				},
				NodeRegistrationService: &p2pMockNodeRegistraionService{
					successGetNodeRegistrationByNodePublicKey: true,
					successGenerateNodeAddressInfo:            true,
					successGetNodeAddressesInfoFromDb:         true,
				},
				Logger: log.New(),
			},
		},
		{
			name: "UpdateOwnNodeAddressInfo:success-{recordUpdated}",
			args: args{
				nodeAddress: "192.0.0.2",
				port:        8080,
			},
			fields: fields{
				NodeConfigurationService: &p2pMockNodeConfigurationService{
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
				Logger:            log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
				Logger:                   tt.fields.Logger,
				PeerStrategyHelper:       tt.fields.PeerStrategyHelper,
			}
			if err := ps.UpdateOwnNodeAddressInfo(tt.args.nodeAddress, tt.args.port, tt.args.forceBroadcast); (err != nil) != tt.
				wantErr {
				t.Errorf("PriorityStrategy.UpdateOwnNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPriorityStrategy_GenerateProofOfOrigin(t *testing.T) {
	type fields struct {
		BlockchainStatusService  coreService.BlockchainStatusServiceInterface
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		NodeAddressesInfoService coreService.NodeAddressInfoServiceInterface
		BlockMainService         coreService.BlockServiceInterface
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
		Signature                crypto.SignatureInterface
	}
	type args struct {
		challenge        []byte
		timestamp        int64
		nodeSecretPhrase string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.ProofOfOrigin
	}{
		{
			name: "GenerateProofOfOrigin:success",
			args: args{
				challenge:        make([]byte, 32),
				timestamp:        int64(1562117271),
				nodeSecretPhrase: "shhhhh",
			},
			fields: fields{
				Signature: &p2pMockSignature{},
			},
			want: &model.ProofOfOrigin{
				MessageBytes: make([]byte, 32),
				Timestamp:    int64(1562117271),
				Signature:    make([]byte, 64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriorityStrategy{
				BlockchainStatusService:  tt.fields.BlockchainStatusService,
				NodeConfigurationService: tt.fields.NodeConfigurationService,
				PeerServiceClient:        tt.fields.PeerServiceClient,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				NodeAddressInfoService:   tt.fields.NodeAddressesInfoService,
				BlockMainService:         tt.fields.BlockMainService,
				MaxUnresolvedPeers:       tt.fields.MaxUnresolvedPeers,
				MaxResolvedPeers:         tt.fields.MaxResolvedPeers,
				Logger:                   tt.fields.Logger,
				PeerStrategyHelper:       tt.fields.PeerStrategyHelper,
				Signature:                tt.fields.Signature,
			}
			if got := ps.GenerateProofOfOrigin(tt.args.challenge, tt.args.timestamp, tt.args.nodeSecretPhrase); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PriorityStrategy.GenerateProofOfOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
