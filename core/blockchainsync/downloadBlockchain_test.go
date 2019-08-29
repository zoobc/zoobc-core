package blockchainsync

import (
	"errors"
	"fmt"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

var (
	mockCommonMilestoneBlockIdResponseSuccess = &model.GetCommonMilestoneBlockIdsResponse{
		BlockIds: []int64{1},
		Last:     true,
	}
	mockBlockIdsResponseSuccess = &model.BlockIdsResponse{
		BlockIds: []int64{1, 2, 3},
	}
)

type (
	// peer service client mocks
	mockPeerServiceClientSuccess struct {
		client.PeerServiceClient
	}
	mockPeerServiceClientFail struct {
		client.PeerServiceClient
	}
	// peer explorer mock
	mockPeerExplorerSuccess struct {
		strategy.NativeStrategy
	}
	// p2p service mocks
	mockP2pServiceSuccess struct {
		p2p.Peer2PeerServiceInterface
	}
	mockP2pServiceSuccessNewResult struct {
		p2p.Peer2PeerServiceInterface
	}

	// block service mocks
	mockBlockServiceSuccess struct {
		coreService.BlockServiceInterface
	}

	mockBlockServiceFail struct {
		coreService.BlockServiceInterface
	}

	mockBlockServiceGetBlockByIDFail struct {
		mockBlockServiceSuccess
	}
)

func (*mockPeerServiceClientSuccess) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
	astMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return mockCommonMilestoneBlockIdResponseSuccess, nil
}

func (*mockPeerServiceClientSuccess) GetNextBlockIDs(
	destPeer *model.Peer, chaintype chaintype.ChainType, blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return mockBlockIdsResponseSuccess, nil
}

func (*mockPeerServiceClientFail) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
	astMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return nil, errors.New("mockErr")
}

func (*mockPeerServiceClientFail) GetNextBlockIDs(
	destPeer *model.Peer, chaintype chaintype.ChainType, blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return nil, errors.New("mockErr")
}

func (*mockPeerExplorerSuccess) DisconnectPeer(peer *model.Peer) {

}

func (*mockBlockServiceGetBlockByIDFail) GetBlockByID(int64) (*model.Block, error) {
	return nil, blocker.NewBlocker(blocker.DBErr, "mockErr")
}

func (*mockP2pServiceSuccess) GetNextBlocks(destPeer *model.Peer, _ chaintype.ChainType,
	blockIDs []int64, blockID int64) (*model.BlocksData, error) {
	return &model.BlocksData{
		NextBlocks: []*model.Block{
			{
				ID: int64(123),
			},
			{
				ID: int64(234),
			},
		},
	}, nil
}

func (*mockP2pServiceSuccess) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

func (*mockP2pServiceSuccess) GetCommonMilestoneBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return &model.GetCommonMilestoneBlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

type mockP2pServiceSuccessOneResult struct {
	p2p.Peer2PeerServiceInterface
}

func (*mockP2pServiceSuccessOneResult) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1},
	}, nil
}

func (*mockP2pServiceSuccessNewResult) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{3, 4},
	}, nil
}

type mockP2pServiceFail struct {
	p2p.Peer2PeerServiceInterface
}

func (*mockP2pServiceFail) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType, blockID int64,
	limit uint32) (*model.BlockIdsResponse, error) {
	return nil, errors.New("simulating error")
}

func (*mockP2pServiceFail) GetCommonMilestoneBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return nil, errors.New("mock error")
}

func (*mockP2pServiceFail) DisconnectPeer(peer *model.Peer) {}

func (*mockBlockServiceSuccess) GetChainType() chaintype.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockServiceSuccess) GetBlockByID(blockID int64) (*model.Block, error) {
	if blockID == 1 || blockID == 2 {
		return &model.Block{
			ID: 1,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block is not found"))
}

func (*mockBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 1}, nil
}

func (*mockBlockServiceFail) GetChainType() chaintype.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockServiceFail) GetLastBlock() (*model.Block, error) {
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block is not found"))
}

func TestGetPeerCommonBlockID(t *testing.T) {
	type args struct {
		p2pService        p2p.Peer2PeerServiceInterface
		blockService      coreService.BlockServiceInterface
		peerServiceClient client.PeerServiceClientInterface
		peerExplorer      strategy.PeerExplorerStrategyInterface
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "wantErr:getPeerCommonBlockID get last block failed",
			args: args{
				p2pService:   &mockP2pServiceSuccess{},
				blockService: &mockBlockServiceFail{},
			},
			want:    int64(0),
			wantErr: true,
		},
		{
			name: "wantErr:getPeerCommonBlockID client error",
			args: args{
				p2pService:        &mockP2pServiceFail{},
				blockService:      &mockBlockServiceSuccess{},
				peerServiceClient: &mockPeerServiceClientFail{},
				peerExplorer:      &mockPeerExplorerSuccess{},
			},
			want:    int64(0),
			wantErr: true,
		},
		{
			name: "wantErr:getPeerCommonBlockID - { ! not found }",
			args: args{
				p2pService:        &mockP2pServiceSuccess{},
				blockService:      &mockBlockServiceGetBlockByIDFail{},
				peerServiceClient: &mockPeerServiceClientSuccess{},
				peerExplorer:      &mockPeerExplorerSuccess{},
			},
			wantErr: true,
			want:    int64(0),
		},
		{
			name: "want:getPeerCommonBlockID successfully return common block ID",
			args: args{
				p2pService:        &mockP2pServiceSuccess{},
				blockService:      &mockBlockServiceSuccess{},
				peerServiceClient: &mockPeerServiceClientSuccess{},
				peerExplorer:      nil,
			},
			want:    int64(1),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainSyncService := NewBlockchainSyncService(
				tt.args.blockService,
				tt.args.p2pService,
				tt.args.peerServiceClient,
				tt.args.peerExplorer,
			)
			got, err := blockchainSyncService.getPeerCommonBlockID(
				&model.Peer{},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPeerCommonBlockID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPeerCommonBlockID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlockIdsAfterCommon(t *testing.T) {
	type args struct {
		p2pService        p2p.Peer2PeerServiceInterface
		blockService      coreService.BlockServiceInterface
		peerServiceClient client.PeerServiceClientInterface
		peerExplorer      strategy.PeerExplorerStrategyInterface
	}

	tests := []struct {
		name string
		args args
		want []int64
	}{
		//{
		//	name: "want:getBlockIdsAfterCommon (GetNextBlockIDs produce error)",
		//	args: args{
		//		p2pService:        &mockP2pServiceFail{},
		//		blockService:      &mockBlockServiceSuccess{},
		//		peerServiceClient: &mockPeerServiceClientFail{},
		//	},
		//	want: []int64{},
		//},
		{
			name: "wantErr:getBlockIdsAfterCommon - {GetBlockByID fail}",
			args: args{
				p2pService:        &mockP2pServiceSuccessNewResult{},
				blockService:      &mockBlockServiceFail{},
				peerServiceClient: &mockPeerServiceClientSuccess{},
			},
			want: []int64{},
		},
		//{
		//	name: "want:getBlockIdsAfterCommon (some getBlockIdsAfterCommon already exists)",
		//	args: args{
		//		p2pService:   &mockP2pServiceSuccess{},
		//		blockService: &mockBlockServiceSuccess{},
		//	},
		//	want: []int64{2, 3, 4},
		//},
		//{
		//	name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon already exists)",
		//	args: args{
		//		p2pService:   &mockP2pServiceSuccessOneResult{},
		//		blockService: &mockBlockServiceSuccess{},
		//	},
		//	want: []int64{1},
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainSyncService := NewBlockchainSyncService(
				tt.args.blockService,
				tt.args.p2pService,
				tt.args.peerServiceClient,
				tt.args.peerExplorer,
			)
			got := blockchainSyncService.getBlockIdsAfterCommon(
				&model.Peer{},
				0,
			)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNextBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextBlocks(t *testing.T) {
	blockService := coreService.NewBlockService(&chaintype.MainChain{}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	blockchainSyncService := NewBlockchainSyncService(
		blockService,
		&mockP2pServiceSuccess{},
		nil,
		nil,
	)
	type args struct {
		maxNextBlocks uint32
		peerUsed      *model.Peer
		blockIDs      []int64
		blockID       int64
		start         uint32
		stop          uint32
	}

	tests := []struct {
		name    string
		args    args
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "wantSuccess:GetNextBlocks",
			args: args{
				maxNextBlocks: uint32(2),
				peerUsed: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
					},
				},
				blockIDs: []int64{1, 2, 3},
				blockID:  int64(1),
				start:    uint32(1),
				stop:     uint32(2),
			},
			want: []*model.Block{
				{
					ID: int64(123),
				},
				{
					ID: int64(234),
				},
			},
			wantErr: false,
		},
		{
			name: "wantError:GetNextBlocks Too many blocks returned",
			args: args{
				maxNextBlocks: uint32(0),
				peerUsed: &model.Peer{
					Info: &model.Node{
						Address: "127.0.0.1",
					},
				},
				blockIDs: []int64{1, 2, 3},
				blockID:  int64(1),
				start:    uint32(1),
				stop:     uint32(2),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blockchainSyncService.getNextBlocks(
				tt.args.maxNextBlocks,
				tt.args.peerUsed,
				tt.args.blockIDs,
				tt.args.start,
				tt.args.stop,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNextBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNextBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
