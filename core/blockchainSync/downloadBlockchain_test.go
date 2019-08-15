package blockchainSync

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

type mockP2pService_success struct {
	p2p.P2pServiceInterface
}

func (*mockP2pService_success) GetNextBlocks(destPeer *model.Peer, chaintype contract.ChainType, blockIds []int64, blockID int64) (*model.BlocksData, error) {
	return &model.BlocksData{
		NextBlocks: []*model.Block{
			&model.Block{
				ID: int64(123),
			},
			&model.Block{
				ID: int64(234),
			},
		},
	}, nil
}

func (*mockP2pService_success) GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

func (*mockP2pService_success) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return &model.GetCommonMilestoneBlockIdsResponse{
		BlockIds: []int64{1, 2, 3, 4},
	}, nil
}

type mockP2pService_success_1result struct {
	p2p.P2pServiceInterface
}

func (*mockP2pService_success_1result) GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1},
	}, nil
}

type mockP2pService_success_newResult struct {
	p2p.P2pServiceInterface
}

func (*mockP2pService_success_newResult) GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{3, 4},
	}, nil
}

type mockP2pService_fail struct {
	p2p.P2pServiceInterface
}

func (*mockP2pService_fail) GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error) {
	return nil, errors.New("simulating error")
}

func (*mockP2pService_fail) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return nil, errors.New("mock error")
}

func (*mockP2pService_fail) DisconnectPeer(peer *model.Peer) {}

type mockBlockService_success struct {
	coreService.BlockServiceInterface
}

func (*mockBlockService_success) GetChainType() contract.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockService_success) GetBlockByID(blockID int64) (*model.Block, error) {
	if blockID == 1 || blockID == 2 {
		return &model.Block{
			ID: 1,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block is not found"))
}

func (*mockBlockService_success) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 1}, nil
}

type mockBlockService_fail struct {
	coreService.BlockServiceInterface
}

func (*mockBlockService_fail) GetChainType() contract.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockService_fail) GetLastBlock() (*model.Block, error) {
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block is not found"))
}

func TestGetPeerCommonBlockID(t *testing.T) {
	type args struct {
		p2pService   p2p.P2pServiceInterface
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
				p2pService:   &mockP2pService_success{},
				blockService: &mockBlockService_success{},
			},
			want:    int64(1),
			wantErr: false,
		},
		{
			name: "wantErr:getPeerCommonBlockID get last block failed",
			args: args{
				p2pService:   &mockP2pService_success{},
				blockService: &mockBlockService_fail{},
			},
			want:    int64(0),
			wantErr: true,
		},
		{
			name: "wantErr:getPeerCommonBlockID grpc error",
			args: args{
				p2pService:   &mockP2pService_fail{},
				blockService: &mockBlockService_success{},
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
		p2pService   p2p.P2pServiceInterface
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
				p2pService:   &mockP2pService_success_newResult{},
				blockService: &mockBlockService_success{},
			},
			want: []int64{3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (some getBlockIdsAfterCommon already exists)",
			args: args{
				p2pService:   &mockP2pService_success{},
				blockService: &mockBlockService_success{},
			},
			want: []int64{2, 3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon already exists)",
			args: args{
				p2pService:   &mockP2pService_success_1result{},
				blockService: &mockBlockService_success{},
			},
			want: []int64{1},
		},
		{
			name: "want:getBlockIdsAfterCommon (GetNextBlockIDs produce error)",
			args: args{
				p2pService:   &mockP2pService_fail{},
				blockService: &mockBlockService_success{},
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
	blockchainSyncService := NewBlockchainSyncService(blockService, &mockP2pService_success{})
	type args struct {
		maxNextBlocks uint32
		peerUsed      *model.Peer
		blockIds      []int64
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
				blockIds: []int64{1, 2, 3},
				blockID:  int64(1),
				start:    uint32(1),
				stop:     uint32(2),
			},
			want: []*model.Block{
				&model.Block{
					ID: int64(123),
				},
				&model.Block{
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
				blockIds: []int64{1, 2, 3},
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
				tt.args.blockIds,
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
