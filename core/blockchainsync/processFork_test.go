package blockchainsync

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestService_ProcessFork(t *testing.T) {
	type fields struct {
		NeedGetMoreBlocks          bool
		IsDownloading              bool
		LastBlockchainFeeder       *model.Peer
		LastBlockchainFeederHeight uint32
		PeerHasMore                bool
		ChainType                  chaintype.ChainType
		BlockService               service.BlockServiceInterface
		LastBlock                  model.Block
		TransactionQuery           query.TransactionQueryInterface
		ForkingProcess             ForkingProcessorInterface
		QueryExecutor              query.ExecutorInterface
		BlockQuery                 query.BlockQueryInterface
	}
	type args struct {
		forkBlocks  []*model.Block
		commonBlock *model.Block
		feederPeer  *model.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &ForkingProcessor{
				ChainType:    tt.fields.ChainType,
				BlockService: tt.fields.BlockService,
			}
			if err := fp.ProcessFork(tt.args.forkBlocks, tt.args.commonBlock, tt.args.feederPeer); (err != nil) != tt.wantErr {
				t.Errorf("Service.ProcessFork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
