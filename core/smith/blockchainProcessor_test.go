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
	blockSeed = new(big.Int).SetUint64(10000000)
	score1    = new(big.Int).SetInt64(8000)
	nodeID1   = int64(12536845)
	score2    = new(big.Int).SetInt64(1000)
	nodeID2   = int64(12536845)
	score3    = new(big.Int).SetInt64(5000)
	nodeID3   = int64(12536845)
	score4    = new(big.Int).SetInt64(10000)
	nodeID4   = int64(12536845)
	score5    = new(big.Int).SetInt64(9000)
	nodeID5   = int64(12536845)
	score6    = new(big.Int).SetInt64(100000)
	nodeID6   = int64(12536845)
	score7    = new(big.Int).SetInt64(90000)
	nodeID7   = int64(12536845)
	score8    = new(big.Int).SetInt64(65000)
	nodeID8   = int64(12536845)
	score9    = new(big.Int).SetInt64(999)
	nodeID9   = int64(12536845)
)

func getMockBlocksmiths() []*model.Blocksmith {
	return []*model.Blocksmith{
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
		{
			NodeID:        nodeID5,
			NodePublicKey: []byte{5},
			Score:         score5,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score5, blockSeed, nodeID5),
			NodeOrder:     coreUtil.CalculateNodeOrder(score5, blockSeed, nodeID5),
		},
		{
			NodeID:        nodeID6,
			NodePublicKey: []byte{6},
			Score:         score6,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score6, blockSeed, nodeID6),
			NodeOrder:     coreUtil.CalculateNodeOrder(score6, blockSeed, nodeID6),
		},
		{
			NodeID:        nodeID7,
			NodePublicKey: []byte{7},
			Score:         score7,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score7, blockSeed, nodeID7),
			NodeOrder:     coreUtil.CalculateNodeOrder(score7, blockSeed, nodeID7),
		},
		{
			NodeID:        nodeID8,
			NodePublicKey: []byte{8},
			Score:         score8,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score8, blockSeed, nodeID8),
			NodeOrder:     coreUtil.CalculateNodeOrder(score8, blockSeed, nodeID8),
		},
		{
			NodeID:        nodeID9,
			NodePublicKey: []byte{9},
			Score:         score9,
			SmithTime:     0,
			BlockSeed:     blockSeed,
			SecretPhrase:  "",
			Deadline:      0,
			SmithOrder:    coreUtil.CalculateSmithOrder(score9, blockSeed, nodeID9),
			NodeOrder:     coreUtil.CalculateNodeOrder(score9, blockSeed, nodeID9),
		},
	}
}

func (*mockBlockService) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	return getMockBlocksmiths(), nil
}

func (*mockBlockServiceFail) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	return nil, errors.New("mockedError")
}

func TestBlockchainProcessor_SortBlocksmith_01(t *testing.T) {
	t.Run("SortBlocksmith_01:success", func(t *testing.T) {
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
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[8]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[1]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[2]) {
					t.Error("invalid sort")
				}
			case 3:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[0]) {
					t.Error("invalid sort")
				}
			case 4:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[4]) {
					t.Error("invalid sort")
				}
			case 5:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[3]) {
					t.Error("invalid sort")
				}
			case 6:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[7]) {
					t.Error("invalid sort")
				}
			case 7:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[6]) {
					t.Error("invalid sort")
				}
			case 8:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[5]) {
					t.Error("invalid sort")
				}
			}
		}
	})
}

func TestBlockchainProcessor_SortBlocksmith_02(t *testing.T) {
	t.Run("SortBlocksmith_02:success", func(t *testing.T) {
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
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		// sort with different seed
		blockSeed = new(big.Int).SetUint64(999335345294492)
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[8]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[1]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[2]) {
					t.Error("invalid sort")
				}
			case 3:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[0]) {
					t.Error("invalid sort")
				}
			case 4:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[4]) {
					t.Error("invalid sort")
				}
			case 5:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[3]) {
					t.Error("invalid sort")
				}
			case 6:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[7]) {
					t.Error("invalid sort")
				}
			case 7:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[6]) {
					t.Error("invalid sort")
				}
			case 8:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[5]) {
					t.Error("invalid sort")
				}
			}
		}
	})
}
func TestBlockchainProcessor_SortBlocksmith_03(t *testing.T) {
	t.Run("SortBlocksmith_03:success", func(t *testing.T) {
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
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		// sort randomizing node id between blocksmiths
		nodeID1 = int64(273458748935)
		nodeID2 = int64(4458748935)
		nodeID3 = int64(2233432423)
		nodeID4 = int64(89543289543289)
		nodeID5 = int64(4378432789435897)
		nodeID6 = int64(2985699456643)
		nodeID7 = int64(1032547846084)
		nodeID8 = int64(69023893290543834)
		nodeID9 = int64(409580358990)
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[8]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[1]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[4]) {
					t.Error("invalid sort")
				}
			case 3:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[0]) {
					t.Error("invalid sort")
				}
			case 4:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[2]) {
					t.Error("invalid sort")
				}
			case 5:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[3]) {
					t.Error("invalid sort")
				}
			case 6:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[5]) {
					t.Error("invalid sort")
				}
			case 7:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[7]) {
					t.Error("invalid sort")
				}
			case 8:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[6]) {
					t.Error("invalid sort")
				}
			}
		}
	})
}

func TestBlockchainProcessor_SortBlocksmith_04(t *testing.T) {
	t.Run("SortBlocksmith_04:success", func(t *testing.T) {
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
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		// sort randomizing node id between blocksmiths and changing blockseed
		blockSeed = new(big.Int).SetUint64(39053285908984532)
		listener.OnNotify(&model.Block{}, &chaintype.MainChain{})
		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[1]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[8]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[0]) {
					t.Error("invalid sort")
				}
			case 3:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[2]) {
					t.Error("invalid sort")
				}
			case 4:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[3]) {
					t.Error("invalid sort")
				}
			case 5:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[4]) {
					t.Error("invalid sort")
				}
			case 6:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[7]) {
					t.Error("invalid sort")
				}
			case 7:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[6]) {
					t.Error("invalid sort")
				}
			case 8:
				if !reflect.DeepEqual(s, *getMockBlocksmiths()[5]) {
					t.Error("invalid sort")
				}
			}
		}
	})
}

func TestBlockchainProcessor_SortBlocksmith_fail(t *testing.T) {
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
