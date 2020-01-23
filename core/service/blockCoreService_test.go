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
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
)

func TestNewMainBlockService(t *testing.T) {
	type args struct {
		ct                          chaintype.ChainType
		kvExecutor                  kvdb.KVExecutorInterface
		queryExecutor               query.ExecutorInterface
		blockQuery                  query.BlockQueryInterface
		mempoolQuery                query.MempoolQueryInterface
		transactionQuery            query.TransactionQueryInterface
		merkleTreeQuery             query.MerkleTreeQueryInterface
		publishedReceiptQuery       query.PublishedReceiptQueryInterface
		skippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
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
		blockIncompleteQueueService BlockIncompleteQueueServiceInterface
		transactionUtil             transaction.UtilInterface
		receiptUtil                 coreUtil.ReceiptUtilInterface
		transactionCoreService      TransactionCoreServiceInterface
	}
	transactionUtil := &transaction.Util{}
	receiptUtil := &coreUtil.ReceiptUtil{}
	transactionCoreService := &TransactionCoreService{}

	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:                     &chaintype.MainChain{},
				obsr:                   observer.NewObserver(),
				transactionUtil:        transactionUtil,
				receiptUtil:            receiptUtil,
				transactionCoreService: transactionCoreService,
			},
			want: &BlockService{
				Chaintype:              &chaintype.MainChain{},
				Observer:               observer.NewObserver(),
				TransactionUtil:        transactionUtil,
				ReceiptUtil:            receiptUtil,
				TransactionCoreService: transactionCoreService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMainBlockService(tt.args.ct, tt.args.kvExecutor, tt.args.queryExecutor, tt.args.blockQuery,
				tt.args.mempoolQuery, tt.args.transactionQuery, tt.args.merkleTreeQuery, tt.args.publishedReceiptQuery,
				tt.args.skippedBlocksmithQuery, tt.args.signature, tt.args.mempoolService,
				tt.args.receiptService, tt.args.nodeRegistrationService, tt.args.txTypeSwitcher, tt.args.accountBalanceQuery,
				tt.args.participationScoreQuery, tt.args.nodeRegistrationQuery, tt.args.obsr, tt.args.blocksmithStrategyMain,
				tt.args.logger, tt.args.accountLedgerQuery, tt.args.blockIncompleteQueueService, tt.args.transactionUtil,
				tt.args.receiptUtil, tt.args.transactionCoreService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestNewSpineBlockService(t *testing.T) {
	type args struct {
		ct                     chaintype.ChainType
		kvExecutor             kvdb.KVExecutorInterface
		queryExecutor          query.ExecutorInterface
		blockQuery             query.BlockQueryInterface
		spinePublicKeyQuery    query.SpinePublicKeyQueryInterface
		signature              crypto.SignatureInterface
		nodeRegistrationQuery  query.NodeRegistrationQueryInterface
		obsr                   *observer.Observer
		blocksmithStrategyMain strategy.BlocksmithStrategyInterface
		logger                 *log.Logger
	}

	tests := []struct {
		name string
		args args
		want *BlockSpineService
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:   &chaintype.SpineChain{},
				obsr: observer.NewObserver(),
			},
			want: &BlockSpineService{
				Chaintype: &chaintype.SpineChain{},
				Observer:  observer.NewObserver(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSpineBlockService(tt.args.ct, tt.args.kvExecutor, tt.args.queryExecutor, tt.args.blockQuery,
				tt.args.spinePublicKeyQuery, tt.args.signature, tt.args.nodeRegistrationQuery, tt.args.obsr,
				tt.args.blocksmithStrategyMain, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
