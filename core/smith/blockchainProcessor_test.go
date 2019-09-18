package smith

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		ct           chaintype.ChainType
		blocksmith   *model.Blocksmith
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
				blocksmith:   &model.Blocksmith{},
				blockService: &service.BlockService{},
			},
			want: &BlockchainProcessor{
				Chaintype:    &chaintype.MainChain{},
				BlockService: &service.BlockService{},
				Generator:    &model.Blocksmith{},
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
