package smith

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/util"

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
	mockNodeRegistrationService struct {
		service.NodeRegistrationService
	}
	mockNodeRegistrationServiceFail struct {
		service.NodeRegistrationService
	}
)

var mockBlocksmiths = []*model.Blocksmith{
	{
		NodePublicKey: []byte{1},
		Score:         new(big.Int).SetInt64(8000),
		SmithTime:     0,
		BlockSeed:     nil,
		SecretPhrase:  "",
		Deadline:      0,
	},
	{
		NodePublicKey: []byte{2},
		Score:         new(big.Int).SetInt64(1000),
		SmithTime:     0,
		BlockSeed:     nil,
		SecretPhrase:  "",
		Deadline:      0,
	},
	{
		NodePublicKey: []byte{3},
		Score:         new(big.Int).SetInt64(5000),
		SmithTime:     0,
		BlockSeed:     nil,
		SecretPhrase:  "",
		Deadline:      0,
	},
}

func (*mockNodeRegistrationService) GetActiveNodes() ([]*model.Blocksmith, error) {
	return mockBlocksmiths, nil
}

func (*mockNodeRegistrationServiceFail) GetActiveNodes() ([]*model.Blocksmith, error) {
	return nil, errors.New("mockedError")
}

func TestBlockchainProcessor_SortBlocksmith(t *testing.T) {
	t.Run("SortBlocksmith:success", func(t *testing.T) {
		var sortedBlocksmiths []model.Blocksmith
		bProcessor := NewBlockchainProcessor(
			&chaintype.MainChain{},
			&model.Blocksmith{
				NodePublicKey: nil,
				Score:         nil,
				SmithTime:     0,
				BlockSeed:     nil,
				SecretPhrase:  "",
				Deadline:      0,
			},
			nil,
			&mockNodeRegistrationService{},
		)
		listener := bProcessor.SortBlocksmith(&sortedBlocksmiths)
		listener.OnNotify(&model.Block{
			BlockSeed: util.ConvertUint64ToBytes(10000000),
		}, &chaintype.MainChain{})

		for i, s := range sortedBlocksmiths {
			switch i {
			case 0:
				if !reflect.DeepEqual(s, *mockBlocksmiths[0]) {
					t.Error("invalid sort")
				}
			case 1:
				if !reflect.DeepEqual(s, *mockBlocksmiths[2]) {
					t.Error("invalid sort")
				}
			case 2:
				if !reflect.DeepEqual(s, *mockBlocksmiths[1]) {
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
			}
		}
	})
	t.Run("SortBlocksmith:getNodeFail", func(t *testing.T) {
		var sortedBlocksmiths []model.Blocksmith
		bProcessor := NewBlockchainProcessor(
			&chaintype.MainChain{},
			&model.Blocksmith{
				NodePublicKey: nil,
				Score:         nil,
				SmithTime:     0,
				BlockSeed:     nil,
				SecretPhrase:  "",
				Deadline:      0,
			},
			nil,
			&mockNodeRegistrationServiceFail{},
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
