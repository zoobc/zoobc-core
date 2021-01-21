// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package blockchainsync

import (
	"errors"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreService "github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type mockPeerExplorer struct {
	strategy.PeerExplorerStrategyInterface
}

func (*mockPeerExplorer) DisconnectPeer(peer *model.Peer) {}

type mockQueryServiceSuccess struct {
	query.ExecutorInterface
}

type mockP2pServiceSuccess struct {
	client.PeerServiceClientInterface
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
	client.PeerServiceClientInterface
}

func (*mockP2pServiceSuccessOneResult) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{1},
	}, nil
}

type mockP2pServiceSuccessNewResult struct {
	client.PeerServiceClientInterface
}

func (*mockP2pServiceSuccessNewResult) GetNextBlockIDs(destPeer *model.Peer, _ chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	return &model.BlockIdsResponse{
		BlockIds: []int64{3, 4},
	}, nil
}

type mockP2pServiceFail struct {
	client.PeerServiceClientInterface
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

type mockBlockServiceSuccess struct {
	coreService.BlockServiceInterface
}

func (*mockBlockServiceSuccess) GetChainType() chaintype.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockServiceSuccess) GetBlockByID(blockID int64, withAttachedData bool) (*model.Block, error) {
	if blockID == 1 || blockID == 2 {
		return &model.Block{
			ID: 1,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "block is not found")
}

func (*mockBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &model.Block{ID: 1}, nil
}

type mockBlockServiceFail struct {
	coreService.BlockServiceInterface
}

func (*mockBlockServiceFail) GetChainType() chaintype.ChainType {
	return &chaintype.MainChain{}
}

func (*mockBlockServiceFail) GetLastBlock() (*model.Block, error) {
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "block is not found")
}

type (
	mockBlockchainStatusService struct {
		coreService.BlockchainStatusService
	}
)

func (*mockBlockchainStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return true
}

func (*mockBlockchainStatusService) IsDownloading(ct chaintype.ChainType) bool {
	return true
}

func TestGetPeerCommonBlockID(t *testing.T) {
	type args struct {
		PeerServiceClient       client.PeerServiceClientInterface
		PeerExplorer            strategy.PeerExplorerStrategyInterface
		blockService            coreService.BlockServiceInterface
		queryService            query.ExecutorInterface
		logger                  *log.Logger
		blockchainStatusService coreService.BlockchainStatusServiceInterface
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
				PeerServiceClient:       &mockP2pServiceSuccess{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				logger:                  log.New(),
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want:    int64(1),
			wantErr: false,
		},
		{
			name: "wantErr:getPeerCommonBlockID get last block failed",
			args: args{
				PeerServiceClient:       &mockP2pServiceSuccess{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceFail{},
				queryService:            &mockQueryServiceSuccess{},
				logger:                  log.New(),
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want:    int64(0),
			wantErr: true,
		},
		{
			name: "wantErr:getPeerCommonBlockID grpc error",
			args: args{
				PeerServiceClient:       &mockP2pServiceFail{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				logger:                  log.New(),
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want:    int64(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainDownloader := &BlockchainDownloader{
				BlockService:            tt.args.blockService,
				PeerServiceClient:       tt.args.PeerServiceClient,
				PeerExplorer:            tt.args.PeerExplorer,
				Logger:                  tt.args.logger,
				BlockchainStatusService: tt.args.blockchainStatusService,
			}
			got, err := blockchainDownloader.getPeerCommonBlockID(
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
		PeerServiceClient       client.PeerServiceClientInterface
		PeerExplorer            strategy.PeerExplorerStrategyInterface
		blockService            coreService.BlockServiceInterface
		queryService            query.ExecutorInterface
		blockchainStatusService coreService.BlockchainStatusServiceInterface
	}

	tests := []struct {
		name string
		args args
		want []int64
	}{
		{
			name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon new)",
			args: args{
				PeerServiceClient:       &mockP2pServiceSuccessNewResult{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want: []int64{3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (some getBlockIdsAfterCommon already exists)",
			args: args{
				PeerServiceClient:       &mockP2pServiceSuccess{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want: []int64{2, 3, 4},
		},
		{
			name: "want:getBlockIdsAfterCommon (all getBlockIdsAfterCommon already exists)",
			args: args{
				PeerServiceClient:       &mockP2pServiceSuccessOneResult{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want: []int64{1},
		},
		{
			name: "want:getBlockIdsAfterCommon (GetNextBlockIDs produce error)",
			args: args{
				PeerServiceClient:       &mockP2pServiceFail{},
				PeerExplorer:            &mockPeerExplorer{},
				blockService:            &mockBlockServiceSuccess{},
				queryService:            &mockQueryServiceSuccess{},
				blockchainStatusService: &mockBlockchainStatusService{},
			},
			want: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchainDownloader := &BlockchainDownloader{
				BlockService:            tt.args.blockService,
				PeerServiceClient:       tt.args.PeerServiceClient,
				PeerExplorer:            tt.args.PeerExplorer,
				BlockchainStatusService: tt.args.blockchainStatusService,
			}
			got := blockchainDownloader.getBlockIdsAfterCommon(
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
	blockService := coreService.NewBlockMainService(
		&chaintype.MainChain{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&transaction.Util{},
		&coreUtil.ReceiptUtil{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	blockchainDownloader := &BlockchainDownloader{
		BlockService:            blockService,
		PeerServiceClient:       &mockP2pServiceSuccess{},
		PeerExplorer:            &mockPeerExplorer{},
		BlockchainStatusService: &mockBlockchainStatusService{},
	}

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
			got, err := blockchainDownloader.getNextBlocks(
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
