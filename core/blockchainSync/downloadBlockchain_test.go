package blockchainSync

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type mockP2pService_getNextBlocks struct {
	p2p.P2pServiceInterface
}

func (*mockP2pService_getNextBlocks) GetNextBlocks(destPeer *model.Peer, chaintype contract.ChainType, blockIds []int64, blockID int64) (*model.BlocksData, error) {
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

func TestDownloadBlockchain_getNextBlocks(t *testing.T) {
	blockService := coreService.NewBlockService(&chaintype.MainChain{}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	blockchainSyncService := NewBlockchainSyncService(blockService, &mockP2pService_getNextBlocks{})
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
			name: "wantError:Too many blocks returned",
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
				tt.args.blockID,
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
