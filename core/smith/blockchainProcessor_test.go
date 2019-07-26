package smith

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		ct           contract.ChainType
		blocksmith   *Blocksmith
		blockService service.BlockServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:           &chaintype.MainChain{},
				blocksmith:   &Blocksmith{},
				blockService: &service.BlockService{},
			},
			want: &BlockchainProcessor{
				Chaintype:    &chaintype.MainChain{},
				BlockService: &service.BlockService{},
				Generator:    &Blocksmith{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainProcessor(tt.args.ct, tt.args.blocksmith, tt.args.blockService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}
