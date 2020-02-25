package smith

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewBlockchainProcessor(t *testing.T) {
	type args struct {
		blocksmith          *model.Blocksmith
		blockService        service.BlockServiceInterface
		logger              *log.Logger
		blockStatusServices map[int32]service.BlockStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		{
			name: "wantSuccess",
			args: args{
				blocksmith:          &model.Blocksmith{},
				blockService:        &service.BlockService{},
				blockStatusServices: make(map[int32]service.BlockStatusServiceInterface),
			},
			want: &BlockchainProcessor{
				BlockService:        &service.BlockService{},
				Generator:           &model.Blocksmith{},
				BlockStatusServices: make(map[int32]service.BlockStatusServiceInterface),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainProcessor(
				tt.args.blocksmith,
				tt.args.blockService,
				tt.args.logger,
				tt.args.blockStatusServices,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}
