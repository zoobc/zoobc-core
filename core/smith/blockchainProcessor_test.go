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
		blocksmith   *model.Blocksmith
		blockService service.BlockServiceInterface
		logger       *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *BlockchainProcessor
	}{
		{
			name: "wantSuccess",
			args: args{
				blocksmith:   &model.Blocksmith{},
				blockService: &service.BlockService{},
			},
			want: &BlockchainProcessor{
				BlockService: &service.BlockService{},
				Generator:    &model.Blocksmith{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockchainProcessor(
				tt.args.blocksmith,
				tt.args.blockService,
				tt.args.logger,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}
