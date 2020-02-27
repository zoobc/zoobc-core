package smith

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		ct                     chaintype.ChainType
		blocksmith             *model.Blocksmith
		blockService           service.BlockServiceInterface
		logger                 *log.Logger
		blockTypeStatusService service.BlockTypeStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:                     &chaintype.MainChain{},
				blocksmith:             &model.Blocksmith{},
				blockService:           &service.BlockService{},
				blockTypeStatusService: service.NewBlockTypeStatusService(false),
			},
			want: &BlockchainProcessor{
				ChainType:              &chaintype.MainChain{},
				BlockService:           &service.BlockService{},
				Generator:              &model.Blocksmith{},
				BlockTypeStatusService: service.NewBlockTypeStatusService(false),
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
				tt.args.blockTypeStatusService,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}
