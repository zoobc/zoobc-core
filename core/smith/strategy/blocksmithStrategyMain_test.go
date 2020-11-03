package strategy

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/storage"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestNewBlocksmithStrategy(t *testing.T) {
	type args struct {
		logger                  *log.Logger
		activeNodeRegistryCache storage.CacheStorageInterface
		chaintype               chaintype.ChainType
		rng                     *crypto.RandomNumberGenerator
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategyMain
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategyMain(
				nil, nil, nil, nil, nil,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategyMain(
				tt.args.logger, nil, tt.args.activeNodeRegistryCache, tt.args.rng,
				tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategyMain() = %v, want %v", got, tt.want)
			}
		})
	}
}
