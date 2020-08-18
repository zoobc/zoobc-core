package smith

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/storage"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		ct                      chaintype.ChainType
		blocksmith              *model.Blocksmith
		blockService            service.BlockServiceInterface
		logger                  *log.Logger
		blockchainStatusService service.BlockchainStatusServiceInterface
		blockStateStorage       storage.CacheStorageInterface
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
				blockchainStatusService: &service.BlockchainStatusService{},
				blockStateStorage:       storage.NewBlockStateStorage(),
			},
			want: &BlockchainProcessor{
				ChainType:               &chaintype.MainChain{},
				Generator:               &model.Blocksmith{},
				BlockService:            &service.BlockService{},
				BlockchainStatusService: &service.BlockchainStatusService{},
				BlockStateStorage:       storage.NewBlockStateStorage(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainProcessor(
				tt.args.ct,
				tt.args.blocksmith,
				tt.args.blockService,
				tt.args.logger,
				tt.args.blockchainStatusService,
				tt.args.blockStateStorage,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}
