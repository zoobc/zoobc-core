package service

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	"github.com/zoobc/zoobc-core/observer"
)

func TestNewBlockService(t *testing.T) {
	type args struct {
		ct                          chaintype.ChainType
		kvExecutor                  kvdb.KVExecutorInterface
		queryExecutor               query.ExecutorInterface
		spineBlockQuery             query.BlockQueryInterface
		mainBlockQuery              query.BlockQueryInterface
		mempoolQuery                query.MempoolQueryInterface
		transactionQuery            query.TransactionQueryInterface
		merkleTreeQuery             query.MerkleTreeQueryInterface
		publishedReceiptQuery       query.PublishedReceiptQueryInterface
		skippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		spinePublicKeyQuery         query.SpinePublicKeyQueryInterface
		signature                   crypto.SignatureInterface
		mempoolService              MempoolServiceInterface
		receiptService              ReceiptServiceInterface
		nodeRegistrationService     NodeRegistrationServiceInterface
		txTypeSwitcher              transaction.TypeActionSwitcher
		accountBalanceQuery         query.AccountBalanceQueryInterface
		participationScoreQuery     query.ParticipationScoreQueryInterface
		nodeRegistrationQuery       query.NodeRegistrationQueryInterface
		blocksmithStrategyMain      strategy.BlocksmithStrategyInterface
		obsr                        *observer.Observer
		logger                      *log.Logger
		accountLedgerQuery          query.AccountLedgerQueryInterface
		megablockQuery              query.MegablockQueryInterface
		fileChunkQuery              query.FileChunkQueryInterface
		blockIncompleteQueueService BlockIncompleteQueueServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:   &chaintype.MainChain{},
				obsr: observer.NewObserver(),
			},
			want: &BlockService{
				Chaintype: &chaintype.MainChain{},
				Observer:  observer.NewObserver(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(tt.args.ct, tt.args.kvExecutor, tt.args.queryExecutor, tt.args.mainBlockQuery,
				tt.args.spineBlockQuery, tt.args.mempoolQuery, tt.args.transactionQuery, tt.args.merkleTreeQuery,
				tt.args.publishedReceiptQuery, tt.args.skippedBlocksmithQuery, tt.args.spinePublicKeyQuery,
				tt.args.signature, tt.args.mempoolService, tt.args.receiptService, tt.args.nodeRegistrationService,
				tt.args.txTypeSwitcher, tt.args.accountBalanceQuery, tt.args.participationScoreQuery,
				tt.args.nodeRegistrationQuery, tt.args.obsr, tt.args.blocksmithStrategyMain, tt.args.logger,
				tt.args.accountLedgerQuery, tt.args.megablockQuery, tt.args.fileChunkQuery,
				tt.blockIncompleteQueueService); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
		})
	}
}
