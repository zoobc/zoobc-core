package smith

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		ct                      chaintype.ChainType
		blocksmith              *model.Blocksmith
		blockService            service.BlockServiceInterface
		nodeRegistrationService service.NodeRegistrationServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:                      &chaintype.MainChain{},
				blocksmith:              &model.Blocksmith{},
				blockService:            &service.BlockService{},
				nodeRegistrationService: &service.NodeRegistrationService{},
			},
			want: &BlockchainProcessor{
				Chaintype:               &chaintype.MainChain{},
				BlockService:            &service.BlockService{},
				Generator:               &model.Blocksmith{},
				NodeRegistrationService: &service.NodeRegistrationService{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainProcessor(
				tt.args.ct, tt.args.blocksmith, tt.args.blockService, tt.args.nodeRegistrationService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockBlockService struct {
		service.BlockService
	}
	mockBlockServiceFail struct {
		service.BlockService
	}
)

var (
	blockSeed       = new(big.Int).SetInt64(0)
	score1          = new(big.Int).SetInt64(8000)
	nodeID1         = int64(1)
	score2          = new(big.Int).SetInt64(1000)
	nodeID2         = int64(1)
	score3          = new(big.Int).SetInt64(5000)
	nodeID3         = int64(1)
	score4          = new(big.Int).SetInt64(5000)
	nodeID4         = int64(10)
	mockBlocksmiths = []*model.Blocksmith{
		{
			NodeID:        nodeID1,
			NodePublicKey: []byte{1},
			Score:         score1,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score1, blockSeed, nodeID1),
			NodeOrder:     coreUtil.CalculateNodeOrder(score1, blockSeed, nodeID1),
		},
		{
			NodeID:        nodeID2,
			NodePublicKey: []byte{2},
			Score:         score2,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score2, blockSeed, nodeID2),
			NodeOrder:     coreUtil.CalculateNodeOrder(score2, blockSeed, nodeID2),
		},
		{
			NodeID:        nodeID3,
			NodePublicKey: []byte{3},
			Score:         score3,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score3, blockSeed, nodeID3),
			NodeOrder:     coreUtil.CalculateNodeOrder(score3, blockSeed, nodeID3),
		},
		{
			NodeID:        nodeID4,
			NodePublicKey: []byte{4},
			Score:         score4,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score4, blockSeed, nodeID4),
			NodeOrder:     coreUtil.CalculateNodeOrder(score4, blockSeed, nodeID4),
		},
	}
)

func (*mockBlockService) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	return mockBlocksmiths, nil
}

func (*mockBlockServiceFail) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	return nil, errors.New("mockedError")
}

func TestBlockchainProcessor_SortBlocksmith(t *testing.T) {
	t.Run("SortBlocksmith:success", func(t *testing.T) {
		var sortedBlocksmiths []model.Blocksmith
		bProcessor := NewBlockchainProcessor(
			&chaintype.MainChain{},
			&model.Blocksmith{
				NodeID:        0,
				NodePublicKey: nil,
				Score:         nil,
				SmithTime:     0,
				BlockSeed:     nil,
				SecretPhrase:  "",
				Deadline:      0,
			},
			&mockBlockService{},
			nil,
		)
		listener := bProcessor.SortBlocksmith(&sortedBlocksmiths)
		listener.OnNotify(&model.Block{
			BlockSeed: util.ConvertUint64ToBytes(10000000),
		}, &chaintype.MainChain{})

		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *mockBlocksmiths[1]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *mockBlocksmiths[2]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *mockBlocksmiths[0]) {
					t.Error("invalid sort")
				}
			case 3:
				// note that this has a lower score than the previous, but a higher nodeID (randomization)
				if !reflect.DeepEqual(s, *mockBlocksmiths[3]) {
					t.Error("invalid sort")
				}
			}
		}
		// sort with different seed
		listener.OnNotify(&model.Block{
			BlockSeed: util.ConvertUint64ToBytes(119294492),
		}, &chaintype.MainChain{})
		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *mockBlocksmiths[1]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *mockBlocksmiths[2]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *mockBlocksmiths[0]) {
					t.Error("invalid sort")
				}
			case 3:
				if !reflect.DeepEqual(s, *mockBlocksmiths[3]) {
					t.Error("invalid sort")
				}
			}
		}
	})
	t.Run("SortBlocksmith:getNodeFail", func(t *testing.T) {
		var sortedBlocksmiths []model.Blocksmith
		bProcessor := NewBlockchainProcessor(
			&chaintype.MainChain{},
			&model.Blocksmith{
				NodeID:        0,
				NodePublicKey: nil,
				Score:         nil,
				SmithTime:     0,
				BlockSeed:     nil,
				SecretPhrase:  "",
				Deadline:      0,
			},
			&mockBlockServiceFail{},
			nil,
		)
		listener := bProcessor.SortBlocksmith(&sortedBlocksmiths)
		listener.OnNotify(&model.Block{
			BlockSeed: util.ConvertUint64ToBytes(10000000),
		}, &chaintype.MainChain{})
		if len(sortedBlocksmiths) > 0 {
			// note: if before there are success sort, the sorted blocksmiths will not be empty
			// but stay the same as previous sorted list
			t.Error("if get nodes fail, empty list won't be filled")
		}
	})

}
