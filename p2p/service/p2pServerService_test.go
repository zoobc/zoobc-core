package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc/metadata"
)

var (
	mockNode = model.Node{
		SharedAddress: "127.0.0.1",
		Address:       "127.0.0.1",
		Port:          8001,
		Version:       "1.0.0",
		CodeName:      "ZBC_main",
	}
	mockPeers = map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &mockNode,
		},
	}
	mockBlock = model.Block{
		ID:                   1,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		PayloadLength:        0,
		PayloadHash:          []byte{0, 0, 0, 1},
		BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		TotalAmount:   100000000,
		TotalFee:      10000000,
		TotalCoinBase: 1,
		Version:       0,
	}
	mockChainType = chaintype.MainChain{}
)

type (
	mockNodeRegistrationServiceGetNodeAddressesInfoSuccess struct {
		nodeaddressesInfo []*model.NodeAddressInfo
		coreService.NodeRegistrationServiceInterface
	}
	p2pSrvMockNodeAddressInfoService struct {
		nodeaddressesInfo []*model.NodeAddressInfo
		coreService.NodeAddressInfoService
	}
	mockPeerExplorerStrategySuccess struct {
		strategy.PriorityStrategy
	}
	mockPeerExplorerStrategyValidateRequestFail struct {
		strategy.PriorityStrategy
	}
	mockPeerExplorerStrategyAddToUnresolvedPeersFail struct {
		strategy.PriorityStrategy
	}

	mockBlockServiceSuccess struct {
		coreService.BlockService
	}
	mockBlockServiceGetLastBlockFailed struct {
		coreService.BlockService
	}

	mockNodeConfigurationService struct {
		coreService.NodeConfigurationServiceInterface
	}
)

func (mockNais *p2pSrvMockNodeAddressInfoService) GetAddressInfoTableWithConsolidatedAddresses(
	preferredStatus model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	return mockNais.nodeaddressesInfo, nil
}

func (mock *mockNodeRegistrationServiceGetNodeAddressesInfoSuccess) GetNodeAddressesInfoFromDb(
	nodeIDs []int64, addressStatus []model.NodeAddressStatus) ([]*model.NodeAddressInfo, error) {
	return mock.nodeaddressesInfo, nil
}
func (*mockPeerExplorerStrategySuccess) GetHostInfo() *model.Node {
	return &mockNode
}
func (*mockPeerExplorerStrategySuccess) ValidateRequest(ctx context.Context) bool {
	return true
}
func (*mockPeerExplorerStrategySuccess) GetResolvedPeers() map[string]*model.Peer {
	return mockPeers
}
func (*mockPeerExplorerStrategySuccess) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	return nil
}

func (*mockPeerExplorerStrategyValidateRequestFail) ValidateRequest(ctx context.Context) bool {
	return false
}

func (*mockPeerExplorerStrategyAddToUnresolvedPeersFail) ValidateRequest(ctx context.Context) bool {
	return true
}
func (*mockPeerExplorerStrategyAddToUnresolvedPeersFail) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	return errors.New("mock Error")
}
func (*mockPeerExplorerStrategyAddToUnresolvedPeersFail) GetHostInfo() *model.Node {
	return &model.Node{
		Version:  "1.0.0",
		CodeName: "ZBC_main",
	}
}

func (*mockBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &mockBlock, nil
}
func (*mockBlockServiceGetLastBlockFailed) GetLastBlock() (*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockNodeConfigurationService) GetHost() *model.Host {
	return &model.Host{Info: &mockNode}
}

func TestNewP2PServerService(t *testing.T) {
	type args struct {
		nodeRegistrationService coreService.NodeRegistrationServiceInterface
		fileService             coreService.FileServiceInterface
		peerExplorer            strategy.PeerExplorerStrategyInterface
		blockServices           map[int32]coreService.BlockServiceInterface
		mempoolServices         map[int32]coreService.MempoolServiceInterface
		nodeSecretPhrase        string
		observer                *observer.Observer
	}
	tests := []struct {
		name string
		args args
		want *P2PServerService
	}{
		{
			name: "wantSuccess",
			args: args{
				nodeRegistrationService: nil,
				fileService:             nil,
				peerExplorer:            nil,
				blockServices:           make(map[int32]coreService.BlockServiceInterface),
				mempoolServices:         make(map[int32]coreService.MempoolServiceInterface),
				nodeSecretPhrase:        "",
				observer:                nil,
			},
			want: &P2PServerService{
				BlockServices:    make(map[int32]coreService.BlockServiceInterface),
				MempoolServices:  make(map[int32]coreService.MempoolServiceInterface),
				NodeSecretPhrase: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewP2PServerService(tt.args.nodeRegistrationService, tt.args.fileService, nil,
				nil, tt.args.peerExplorer,
				tt.args.blockServices,
				tt.args.mempoolServices,
				tt.args.nodeSecretPhrase, tt.args.observer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewP2PServerService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_GetPeerInfo(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx context.Context
		req *model.GetPeerInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPeerInfoResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.GetPeerInfoRequest{},
			},
			want: &model.GetPeerInfoResponse{
				HostInfo: &mockNode,
			},
			wantErr: false,
		},
		{
			name: "wnatFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.GetPeerInfoRequest{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetPeerInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetPeerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_GetMorePeers(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx context.Context
		req *model.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Node
		wantErr bool
	}{
		{
			name: "wantFailed:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.Empty{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx: context.Background(),
				req: &model.Empty{},
			},
			want: []*model.Node{
				0: &mockNode,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetMorePeers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetMorePeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetMorePeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_SendPeers(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx   context.Context
		peers []*model.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args: args{ctx: context.Background(),
				peers: []*model.Node{
					0: &mockNode,
				}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFailed:AddToUnresolvedPeers",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyAddToUnresolvedPeersFail{},
			},
			args: args{
				ctx: context.Background(),
				peers: []*model.Node{
					0: &mockNode,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx: context.Background(),
				peers: []*model.Node{
					0: &mockNode,
				},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.SendPeers(tt.args.ctx, tt.args.peers)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.SendPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.SendPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_GetCumulativeDifficulty(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx       context.Context
		chainType chaintype.ChainType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetCumulativeDifficultyResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetLastBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockBlockServiceGetLastBlockFailed{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want: &model.GetCumulativeDifficultyResponse{
				CumulativeDifficulty: mockBlock.GetCumulativeDifficulty(),
				Height:               mockBlock.GetHeight(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetCumulativeDifficulty(tt.args.ctx, tt.args.chainType)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetCumulativeDifficulty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetCumulativeDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetCommonMilestoneBlockIDsBlockServiceSuccess struct {
		coreService.BlockService
	}
)

var (
	mockGetCommonMilestoneBlockIDsSameLastBlockID                     = mockBlock.GetID()
	mockGetCommonMilestoneBlockIDsSameLastBlockIDFail                 = mockBlock.GetID() + 1
	mockGetCommonMilestoneBlockIDsLastMilestoneBlockID                = mockBlock.GetID() + 2
	mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDFail            = mockBlock.GetID() + 3
	mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDSuccessByHeight = mockBlock.GetID() + 4
	mockGetCommonMilestoneBlockIDsLastMilestoneBlock                  = model.Block{
		ID:                mockGetCommonMilestoneBlockIDsLastMilestoneBlockID,
		PreviousBlockHash: mockBlock.GetPayloadHash(),
		Height:            mockBlock.GetHeight() + 1,
	}
	mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight = model.Block{
		ID:                mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDSuccessByHeight,
		PreviousBlockHash: mockBlock.GetPayloadHash(),
		Height:            mockBlock.GetHeight(),
	}
	mockGetCommonMilestoneBlockIDsGenesisBlock = model.Block{
		ID:                mockBlock.GetID() + 5,
		PreviousBlockHash: mockBlock.GetPayloadHash(),
		Height:            0,
	}
)

func (*mockGetCommonMilestoneBlockIDsBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &mockBlock, nil
}
func (*mockGetCommonMilestoneBlockIDsBlockServiceSuccess) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	switch id {
	case mockGetCommonMilestoneBlockIDsSameLastBlockID:
		return &mockBlock, nil
	case mockGetCommonMilestoneBlockIDsLastMilestoneBlockID:
		return &mockGetCommonMilestoneBlockIDsLastMilestoneBlock, nil
	case mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDSuccessByHeight:
		return &mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight, nil
	default:
		return nil, errors.New("mock Error")
	}
}

func (*mockGetCommonMilestoneBlockIDsBlockServiceSuccess) GetBlockByHeight(blockHeight uint32) (*model.Block, error) {
	switch blockHeight {
	case mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight.GetHeight():
		return &mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight, nil
	case mockGetCommonMilestoneBlockIDsGenesisBlock.GetHeight():
		return &mockGetCommonMilestoneBlockIDsGenesisBlock, nil
	default:
		return nil, errors.New("mock Error")
	}

}

func TestP2PServerService_GetCommonMilestoneBlockIDs(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx                  context.Context
		chainType            chaintype.ChainType
		lastBlockID          int64
		lastMilestoneBlockID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetCommonMilestoneBlockIdsResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:SupliedLastBlockID&LastMilestoneBlockID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          0,
				lastMilestoneBlockID: 0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetLastBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockBlockServiceGetLastBlockFailed{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          1,
				lastMilestoneBlockID: 2,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:SameLastBlockID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetCommonMilestoneBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          mockGetCommonMilestoneBlockIDsSameLastBlockID,
				lastMilestoneBlockID: 2,
			},
			want: &model.GetCommonMilestoneBlockIdsResponse{
				BlockIds: []int64{mockGetCommonMilestoneBlockIDsSameLastBlockID},
				Last:     true,
			},
			wantErr: false,
		},
		{
			name: "wantFail:GetLastMilestoneBlockID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetCommonMilestoneBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          mockGetCommonMilestoneBlockIDsSameLastBlockIDFail,
				lastMilestoneBlockID: mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDFail,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetBlockByHeight",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetCommonMilestoneBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          mockGetCommonMilestoneBlockIDsSameLastBlockIDFail,
				lastMilestoneBlockID: mockGetCommonMilestoneBlockIDsLastMilestoneBlockID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:lastMilestoneBlockID!=0",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetCommonMilestoneBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          mockGetCommonMilestoneBlockIDsSameLastBlockIDFail,
				lastMilestoneBlockID: mockGetCommonMilestoneBlockIDsLastMilestoneBlockIDSuccessByHeight,
			},
			want: &model.GetCommonMilestoneBlockIdsResponse{
				BlockIds: []int64{
					mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight.GetID(),
					mockGetCommonMilestoneBlockIDsGenesisBlock.GetID(),
				},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:lastMilestoneBlockID==0",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetCommonMilestoneBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:                  context.Background(),
				chainType:            &mockChainType,
				lastBlockID:          mockGetCommonMilestoneBlockIDsSameLastBlockIDFail,
				lastMilestoneBlockID: 0,
			},
			want: &model.GetCommonMilestoneBlockIdsResponse{
				BlockIds: []int64{
					mockGetCommonMilestoneBlockIDsLastMilestoneBlockSuccessByHeight.GetID(),
					mockGetCommonMilestoneBlockIDsGenesisBlock.GetID(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetCommonMilestoneBlockIDs(tt.args.ctx, tt.args.chainType, tt.args.lastBlockID, tt.args.lastMilestoneBlockID)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetCommonMilestoneBlockIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetCommonMilestoneBlockIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNextBlockIDsBlockServiceGetBlockByIDFail struct {
		coreService.BlockService
	}
	mockGetNextBlockIDsBlockServiceGetBlocksFromHeightFail struct {
		coreService.BlockService
	}
	mockGetNextBlockIDsBlockServiceSuccess struct {
		coreService.BlockService
	}
)

var (
	mockGetNextBlockIDsLimit   = constant.PeerGetBlocksLimit - (constant.PeerGetBlocksLimit - 1)
	mockGetNextBlockIDsSuccess = model.Block{
		ID:                mockBlock.GetID(),
		PreviousBlockHash: mockBlock.GetPayloadHash(),
		Height:            mockBlock.GetHeight() + 1,
	}
)

func (*mockGetNextBlockIDsBlockServiceGetBlockByIDFail) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockGetNextBlockIDsBlockServiceGetBlocksFromHeightFail) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return &mockGetNextBlockIDsSuccess, nil

}
func (*mockGetNextBlockIDsBlockServiceGetBlocksFromHeightFail) GetBlocksFromHeight(
	startHeight, limit uint32,
	withAttachedData bool,
) ([]*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockGetNextBlockIDsBlockServiceSuccess) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return &mockGetNextBlockIDsSuccess, nil
}

func (*mockGetNextBlockIDsBlockServiceSuccess) GetBlocksFromHeight(
	startHeight, limit uint32,
	withAttachedData bool,
) ([]*model.Block, error) {
	return []*model.Block{&mockGetNextBlockIDsSuccess}, nil

}

func TestP2PServerService_GetNextBlockIDs(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx        context.Context
		chainType  chaintype.ChainType
		reqLimit   uint32
		reqBlockID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetBlockByID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlockIDsBlockServiceGetBlockByIDFail{},
				},
			},
			args: args{
				ctx:        context.Background(),
				chainType:  &mockChainType,
				reqLimit:   mockGetNextBlockIDsLimit,
				reqBlockID: mockBlock.GetID(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetBlocksFromHeight",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlockIDsBlockServiceGetBlocksFromHeightFail{},
				},
			},
			args: args{
				ctx:        context.Background(),
				chainType:  &mockChainType,
				reqLimit:   mockGetNextBlockIDsLimit,
				reqBlockID: mockBlock.GetID(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlockIDsBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:        context.Background(),
				chainType:  &mockChainType,
				reqLimit:   mockGetNextBlockIDsLimit,
				reqBlockID: mockBlock.GetID(),
			},
			want:    []int64{mockGetNextBlockIDsSuccess.GetID()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetNextBlockIDs(tt.args.ctx, tt.args.chainType, tt.args.reqLimit, tt.args.reqBlockID)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetNextBlockIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetNextBlockIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNextBlocksBlockServiceGetBlockByIDFail struct {
		coreService.BlockService
	}
	mockGetNextBlocksBlockServiceGetBlocksFromHeightFail struct {
		coreService.BlockService
	}

	mockGetNextBlocksBlockServiceSuccess struct {
		coreService.BlockService
	}
)

var (
	mockGetNextBlocksSuccess = model.Block{
		ID:                mockBlock.GetID(),
		PreviousBlockHash: mockBlock.GetPayloadHash(),
		Height:            mockBlock.GetHeight() + 1,
	}
)

func (*mockGetNextBlocksBlockServiceGetBlockByIDFail) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockGetNextBlocksBlockServiceGetBlocksFromHeightFail) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return &mockGetNextBlocksSuccess, nil
}
func (*mockGetNextBlocksBlockServiceGetBlocksFromHeightFail) GetBlocksFromHeight(
	startHeight, limit uint32,
	withAttachedData bool,
) ([]*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockGetNextBlocksBlockServiceSuccess) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	return &mockGetNextBlocksSuccess, nil
}
func (*mockGetNextBlocksBlockServiceSuccess) GetBlocksFromHeight(
	startHeight, limit uint32,
	withAttachedData bool,
) ([]*model.Block, error) {
	return []*model.Block{&mockGetNextBlocksSuccess}, nil
}
func (*mockGetNextBlocksBlockServiceSuccess) PopulateBlockData(block *model.Block) error {
	return nil
}

func TestP2PServerService_GetNextBlocks(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx         context.Context
		chainType   chaintype.ChainType
		blockID     int64
		blockIDList []int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlocksData
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetBlockByID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlocksBlockServiceGetBlockByIDFail{},
				},
			},
			args: args{
				ctx:         context.Background(),
				chainType:   &mockChainType,
				blockID:     mockBlock.GetID(),
				blockIDList: []int64{mockBlock.GetID()},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetBlocksFromHeight",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlocksBlockServiceGetBlocksFromHeightFail{},
				},
			},
			args: args{
				ctx:         context.Background(),
				chainType:   &mockChainType,
				blockID:     mockBlock.GetID(),
				blockIDList: []int64{mockBlock.GetID()},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:MissMatchBlockID",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlocksBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:         context.Background(),
				chainType:   &mockChainType,
				blockID:     mockBlock.GetID(),
				blockIDList: []int64{mockGetNextBlocksSuccess.GetID() + 1},
			},
			want:    &model.BlocksData{},
			wantErr: false,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockGetNextBlocksBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:         context.Background(),
				chainType:   &mockChainType,
				blockID:     mockBlock.GetID(),
				blockIDList: []int64{mockGetNextBlocksSuccess.GetID()},
			},
			want:    &model.BlocksData{NextBlocks: []*model.Block{&mockGetNextBlocksSuccess}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.GetNextBlocks(tt.args.ctx, tt.args.chainType, tt.args.blockID, tt.args.blockIDList)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetNextBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetNextBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendBlockBlockServiceGetLastBlockFail struct {
		coreService.BlockService
	}

	mockSendBlockBlockServiceReceiveBlockFail struct {
		coreService.BlockService
	}
	mockSendBlockBlockServiceSuccess struct {
		coreService.BlockService
	}
)

func (*mockSendBlockBlockServiceGetLastBlockFail) GetLastBlock() (*model.Block, error) {
	return nil, errors.New("mock Error")
}

func (*mockSendBlockBlockServiceReceiveBlockFail) GetLastBlock() (*model.Block, error) {
	return &mockBlock, nil
}
func (*mockSendBlockBlockServiceReceiveBlockFail) ReceiveBlock(
	[]byte, *model.Block, *model.Block, string, *model.Peer,
) (*model.Receipt, error) {
	return nil, errors.New("mock Error")
}

func (*mockSendBlockBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &mockBlock, nil
}
func (*mockSendBlockBlockServiceSuccess) ReceiveBlock(
	[]byte, *model.Block, *model.Block, string, *model.Peer,
) (*model.Receipt, error) {
	return &model.Receipt{
		SenderPublicKey: []byte{1},
	}, nil
}

func TestP2PServerService_SendBlock(t *testing.T) {
	var (
		mockMetaData      = map[string]string{p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(&mockNode)}
		mockHeaderContext = metadata.New(mockMetaData)
		mockContext       = metadata.NewIncomingContext(context.Background(), mockHeaderContext)

		mockMetaDataPeerFail      = map[string]string{p2pUtil.DefaultConnectionMetadata: "127.0.0.1:fail"}
		mockHeaderContextPeerFail = metadata.New(mockMetaDataPeerFail)
		mockContextPeerFail       = metadata.NewIncomingContext(context.Background(), mockHeaderContextPeerFail)
	)
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx             context.Context
		chainType       chaintype.ChainType
		block           *model.Block
		senderPublicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendBlockResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidContext",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ParsePeer",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx:       mockContextPeerFail,
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       mockContext,
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetLastBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendBlockBlockServiceGetLastBlockFail{},
				},
			},
			args: args{
				ctx:       mockContext,
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ReceiveBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendBlockBlockServiceReceiveBlockFail{},
				},
			},
			args: args{
				ctx:       mockContext,
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "waantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendBlockBlockServiceSuccess{},
				},
			},
			args: args{
				ctx:       mockContext,
				chainType: &mockChainType,
			},
			want: &model.SendBlockResponse{
				Receipt: &model.Receipt{SenderPublicKey: []byte{1}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.SendBlock(tt.args.ctx, tt.args.chainType, tt.args.block, tt.args.senderPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.SendBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.SendBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendTransactionBlockServiceGetLastBlockFail struct {
		coreService.BlockService
	}
	mockSendTransactionBlockServiceSuccess struct {
		coreService.BlockService
	}
	mockSendTransactionMempoolServiceReceivedTransactionFail struct {
		coreService.MempoolService
	}
	mockSendTransactionMempoolServiceSuccess struct {
		coreService.MempoolService
	}
)

func (*mockSendTransactionBlockServiceGetLastBlockFail) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return nil, errors.New("mock Error")
}
func (*mockSendTransactionBlockServiceSuccess) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{
		ID:        mockBlock.ID,
		Height:    mockBlock.Height,
		BlockHash: mockBlock.BlockHash,
	}, nil
}
func (*mockSendTransactionMempoolServiceReceivedTransactionFail) ReceivedTransaction(
	[]byte, []byte, *storage.BlockCacheObject, string,
) (*model.Receipt, error) {
	return nil, errors.New("mock Error")
}

func (*mockSendTransactionMempoolServiceSuccess) ReceivedTransaction(
	[]byte, []byte, *storage.BlockCacheObject, string,
) (*model.Receipt, error) {
	return &model.Receipt{
		SenderPublicKey: []byte{1},
	}, nil
}

func TestP2PServerService_SendTransaction(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx              context.Context
		chainType        chaintype.ChainType
		transactionBytes []byte
		senderPublicKey  []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendTransactionResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType_BlockService",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetLastBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionBlockServiceGetLastBlockFail{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType_MempoolService",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ReceiveTransaction",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionMempoolServiceReceivedTransactionFail{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionMempoolServiceSuccess{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want: &model.SendTransactionResponse{
				Receipt: &model.Receipt{SenderPublicKey: []byte{1}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.SendTransaction(tt.args.ctx, tt.args.chainType, tt.args.transactionBytes, tt.args.senderPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.SendTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.SendTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendTransactionsBlockServiceGetLastBlockFail struct {
		coreService.BlockService
	}
	mockSendTransactionsBlockServiceSuccess struct {
		coreService.BlockService
	}
	mockSendTransactionsMempoolServiceReceivedTransactionsFail struct {
		coreService.MempoolService
	}
	mockSendTransactionsMempoolServiceSuccess struct {
		coreService.MempoolService
	}
)

func (*mockSendTransactionsBlockServiceGetLastBlockFail) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return nil, errors.New("mock Error")
}
func (*mockSendTransactionsBlockServiceSuccess) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{
		ID:        mockBlock.ID,
		Height:    mockBlock.Height,
		BlockHash: mockBlock.BlockHash,
	}, nil
}
func (*mockSendTransactionsMempoolServiceReceivedTransactionsFail) ReceivedBlockTransactions(
	[]byte, [][]byte, *storage.BlockCacheObject, string,
) ([]*model.Receipt, error) {
	return nil, errors.New("mock Error")
}
func (*mockSendTransactionsMempoolServiceSuccess) ReceivedBlockTransactions(
	[]byte, [][]byte, *storage.BlockCacheObject, string,
) ([]*model.Receipt, error) {
	return []*model.Receipt{{
		SenderPublicKey: []byte{1},
	}}, nil
}
func TestP2PServerService_SendBlockTransactions(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx               context.Context
		chainType         chaintype.ChainType
		transactionsBytes [][]byte
		senderPublicKey   []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendBlockTransactionsResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType_BlockService",
			fields: fields{
				PeerExplorer:  &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:GetLastBlock",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsBlockServiceGetLastBlockFail{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidChainType_MempoolService",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ReceiveTransactions",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsMempoolServiceReceivedTransactionsFail{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				BlockServices: map[int32]coreService.BlockServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsBlockServiceSuccess{},
				},
				MempoolServices: map[int32]coreService.MempoolServiceInterface{
					mockChainType.GetTypeInt(): &mockSendTransactionsMempoolServiceSuccess{},
				},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want: &model.SendBlockTransactionsResponse{
				Receipts: []*model.Receipt{{
					SenderPublicKey: []byte{1},
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.SendBlockTransactions(tt.args.ctx, tt.args.chainType, tt.args.transactionsBytes, tt.args.senderPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.SendBlockTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.SendBlockTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_RequestBlockTransactions(t *testing.T) {
	var (
		mockMetaData      = map[string]string{p2pUtil.DefaultConnectionMetadata: p2pUtil.GetFullAddress(&mockNode)}
		mockHeaderContext = metadata.New(mockMetaData)
		mockContext       = metadata.NewIncomingContext(context.Background(), mockHeaderContext)

		mockMetaDataPeerFail      = map[string]string{p2pUtil.DefaultConnectionMetadata: "127.0.0.1:fail"}
		mockHeaderContextPeerFail = metadata.New(mockMetaDataPeerFail)
		mockContextPeerFail       = metadata.NewIncomingContext(context.Background(), mockHeaderContextPeerFail)
	)
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx             context.Context
		chainType       chaintype.ChainType
		blockID         int64
		transactionsIDs []int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidContext",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx:       context.Background(),
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ParsePeer",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
			},
			args: args{
				ctx:       mockContextPeerFail,
				chainType: &mockChainType,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				Observer:     observer.NewObserver(),
			},
			args: args{
				ctx:       mockContext,
				chainType: &mockChainType,
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.RequestBlockTransactions(tt.args.ctx, tt.args.chainType, tt.args.blockID, tt.args.transactionsIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.RequestBlockTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.RequestBlockTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockRequestDownloadFileFileServiceReadFileByNameFail struct {
		coreService.FileService
	}
	mockRequestDownloadFileFileServiceReadFileByNameSuccess struct {
		coreService.FileService
	}
)

var (
	mockRequestDownloadFilePath = "./mockPath"
	mockRequestDownloadFileName = "mockName"
)

func (*mockRequestDownloadFileFileServiceReadFileByNameFail) ReadFileFromDir(dir, fileName string) ([]byte, error) {
	return nil, errors.New("mock Error")
}

func (*mockRequestDownloadFileFileServiceReadFileByNameFail) GetDownloadPath() string {
	return mockRequestDownloadFilePath
}

func (*mockRequestDownloadFileFileServiceReadFileByNameSuccess) ReadFileFromDir(dir, fileName string) ([]byte, error) {
	return []byte{1}, nil
}
func (*mockRequestDownloadFileFileServiceReadFileByNameSuccess) GetDownloadPath() string {
	return mockRequestDownloadFilePath
}
func TestP2PServerService_RequestDownloadFile(t *testing.T) {
	type fields struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
	type args struct {
		ctx            context.Context
		fileChunkNames []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.FileDownloadResponse
		wantErr bool
	}{
		{
			name: "wantFail:ValidateRequest",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:ReadFileByNameFail",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				FileService:  &mockRequestDownloadFileFileServiceReadFileByNameFail{},
			},
			args: args{
				ctx:            context.Background(),
				fileChunkNames: []string{mockRequestDownloadFileName},
			},
			want: &model.FileDownloadResponse{
				FileChunks: make([][]byte, 0),
				Failed:     []string{mockRequestDownloadFileName},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:ReadFileByNameSuccess",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				FileService:  &mockRequestDownloadFileFileServiceReadFileByNameSuccess{},
			},
			args: args{
				ctx:            context.Background(),
				fileChunkNames: []string{mockRequestDownloadFileName},
			},
			want: &model.FileDownloadResponse{
				FileChunks: [][]byte{{1}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				FileService:      tt.fields.FileService,
				PeerExplorer:     tt.fields.PeerExplorer,
				BlockServices:    tt.fields.BlockServices,
				MempoolServices:  tt.fields.MempoolServices,
				NodeSecretPhrase: tt.fields.NodeSecretPhrase,
				Observer:         tt.fields.Observer,
			}
			got, err := ps.RequestDownloadFile(tt.args.ctx, nil, tt.args.fileChunkNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.RequestDownloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.RequestDownloadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestP2PServerService_GetNodeAddressesInfo(t *testing.T) {
	type fields struct {
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		NodeAddressInfoService  coreService.NodeAddressInfoServiceInterface
		FileService             coreService.FileServiceInterface
		PeerExplorer            strategy.PeerExplorerStrategyInterface
		BlockServices           map[int32]coreService.BlockServiceInterface
		MempoolServices         map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase        string
		Observer                *observer.Observer
	}
	type args struct {
		ctx context.Context
		req *model.GetNodeAddressesInfoRequest
	}
	nodeAddressesInfo := []*model.NodeAddressInfo{
		{
			NodeID:      int64(111),
			Signature:   make([]byte, 64),
			BlockHash:   make([]byte, 32),
			BlockHeight: 1,
			Port:        8080,
			Address:     "192.168.1.1",
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetNodeAddressesInfoResponse
		wantErr bool
	}{
		{
			name: "GetNodeAddressesInfo:fail-{ValidateRequest}",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategyValidateRequestFail{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				PeerExplorer: &mockPeerExplorerStrategySuccess{},
				NodeAddressInfoService: &p2pSrvMockNodeAddressInfoService{
					nodeaddressesInfo: nodeAddressesInfo,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &model.GetNodeAddressesInfoRequest{
					NodeIDs: []int64{111},
				},
			},
			want: &model.GetNodeAddressesInfoResponse{
				NodeAddressesInfo: nodeAddressesInfo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &P2PServerService{
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				NodeAddressInfoService:  tt.fields.NodeAddressInfoService,
				FileService:             tt.fields.FileService,
				PeerExplorer:            tt.fields.PeerExplorer,
				BlockServices:           tt.fields.BlockServices,
				MempoolServices:         tt.fields.MempoolServices,
				NodeSecretPhrase:        tt.fields.NodeSecretPhrase,
				Observer:                tt.fields.Observer,
			}
			got, err := ps.GetNodeAddressesInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerService.GetNodeAddressesInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerService.GetNodeAddressesInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
