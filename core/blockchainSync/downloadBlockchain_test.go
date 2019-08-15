package blockchainsync

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type mockP2pServiceSuccess struct {
	p2p.ServiceInterface
}

func (*mockP2pServiceSuccess) GetNextBlocks(destPeer *model.Peer, _ contract.ChainType,
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

func (*mockP2pServiceSuccess) GetNextBlockIDs(destPeer *model.Peer, _ contract.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

func (*mockP2pServiceSuccess) GetCommonMilestoneBlockIDs(destPeer *model.Peer, _ contract.ChainType,
	lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return &model.GetCommonMilestoneBlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

type mockP2pServiceSuccessOneResult struct {
	p2p.ServiceInterface
}

func (*mockP2pServiceSuccessOneResult) GetNextBlockIDs(destPeer *model.Peer, _ contract.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1},
	}, nil
}

type mockP2pServiceSuccessNewResult struct {
	p2p.ServiceInterface
}

func (*mockP2pServiceSuccessNewResult) GetNextBlockIDs(destPeer *model.Peer, _ contract.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{3, 4},
	}, nil
}

type mockP2pServiceFail struct {
	p2p.ServiceInterface
}

func (*mockP2pServiceFail) GetNextBlockIDs(destPeer *model.Peer, _ contract.ChainType, blockID int64,
	limit uint32) (*model.BlockIdsResponse, error) {
	return nil, errors.New("simulating error")
}

func (*mockP2pServiceFail) GetCommonMilestoneBlockIDs(destPeer *model.Peer, _ contract.ChainType,
	lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return nil, errors.New("mock error")
}

func (*mockP2pServiceFail) DisconnectPeer(peer *model.Peer) {}

type mockBlockServiceSuccess struct {
	coreService.BlockServiceInterface
}

func (*mockBlockServiceSuccess) GetChainType() contract.ChainType {
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

type mockBlockServiceFail struct {
	coreService.BlockServiceInterface
}

func (*mockBlockServiceFail) GetChainType() contract.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockServiceFail) GetLastBlock() (*model.Block, error) {
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block is not found"))
}

func TestGetPeerCommonBlockID(t *testing.T) {
	type args struct {
		p2pService   p2p.ServiceInterface
		blockService coreService.BlockServiceInterface
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "want:getPeerCommonBlockID successfully return common block ID",
			args: args{
				p2pService:   &mockP2pServiceSuccess{},
				blockService: &mockBlockServiceSuccess{},
			},
			want:    int64(1),
			wantErr: false,
		},
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
			name: "wantErr:getPeerCommonBlockID grpc error",
			args: args{
				p2pService:   &mockP2pServiceFail{},
				blockService: &mockBlockServiceSuccess{},
			},
			want:    int64(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainSyncService := NewBlockchainSyncService(tt.args.blockService, tt.args.p2pService)
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
		p2pService   p2p.ServiceInterface
		blockService coreService.BlockServiceInterface
	}

	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon new)",
			args: args{
				p2pService:   &mockP2pServiceSuccessNewResult{},
				blockService: &mockBlockServiceSuccess{},
			},
			want: []int64{3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (some getBlockIdsAfterCommon already exists)",
			args: args{
				p2pService:   &mockP2pServiceSuccess{},
				blockService: &mockBlockServiceSuccess{},
			},
			want: []int64{2, 3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon already exists)",
			args: args{
				p2pService:   &mockP2pServiceSuccessOneResult{},
				blockService: &mockBlockServiceSuccess{},
			},
			want: []int64{1},
		},
		{
			name: "want:getBlockIdsAfterCommon (GetNextBlockIDs produce error)",
			args: args{
				p2pService:   &mockP2pServiceFail{},
				blockService: &mockBlockServiceSuccess{},
			},
			want: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainSyncService := NewBlockchainSyncService(tt.args.blockService, tt.args.p2pService)
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
	blockchainSyncService := NewBlockchainSyncService(blockService, &mockP2pServiceSuccess{})
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
