package service

import (
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	BlockServiceInterface interface {
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, blockReceipts []*model.PublishedReceipt, spinePublicKeys []*model.SpinePublicKey,
			payloadHash []byte, payloadLength uint32, cumulativeDifficulty *big.Int, genesisSignature []byte) (*model.Block, error)
		GenerateBlock(
			previousBlock *model.Block,
			secretPhrase string,
			timestamp int64,
		) (*model.Block, error)
		ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error
		PushBlock(previousBlock, block *model.Block, broadcast bool) error
		GetBlockByID(id int64, withAttachedData bool) (*model.Block, error)
		GetBlockByHeight(uint32) (*model.Block, error)
		GetBlocksFromHeight(uint32, uint32) ([]*model.Block, error)
		GetLastBlock() (*model.Block, error)
		GetBlockHash(block *model.Block) ([]byte, error)
		GetBlocks() ([]*model.Block, error)
		PopulateBlockData(block *model.Block) error
		GetGenesisBlock() (*model.Block, error)
		GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error)
		AddGenesis() error
		CheckGenesis() bool
		GetChainType() chaintype.ChainType
		ChainWriteLock(int)
		ChainWriteUnlock(actionType int)
		ReceiveBlock(
			senderPublicKey []byte,
			lastBlock,
			block *model.Block,
			nodeSecretPhrase string,
		) (*model.BatchReceipt, error)
		GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error)
		PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error)
		GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface
	}
)

func NewBlockService(
	ct chaintype.ChainType,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	mainBlockQuery, spineBlockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	spinePublicKeyQuery query.SpinePublicKeyQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	receiptService ReceiptServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	obsr *observer.Observer,
	blocksmithStrategy strategy.BlocksmithStrategyInterface,
	logger *log.Logger,
	accountLedgerQuery query.AccountLedgerQueryInterface,
	megablockQuery query.MegablockQueryInterface,
	fileChunkQuery query.FileChunkQueryInterface,
	blockIncompleteQueueService BlockIncompleteQueueServiceInterface,
) BlockServiceInterface {
	switch ct.(type) {
	case *chaintype.MainChain:
		return &BlockService{
			Chaintype:                   ct,
			KVExecutor:                  kvExecutor,
			QueryExecutor:               queryExecutor,
			BlockQuery:                  mainBlockQuery,
			MempoolQuery:                mempoolQuery,
			TransactionQuery:            transactionQuery,
			MerkleTreeQuery:             merkleTreeQuery,
			PublishedReceiptQuery:       publishedReceiptQuery,
			SkippedBlocksmithQuery:      skippedBlocksmithQuery,
			SpinePublicKeyQuery:         spinePublicKeyQuery,
			Signature:                   signature,
			MempoolService:              mempoolService,
			ReceiptService:              receiptService,
			NodeRegistrationService:     nodeRegistrationService,
			ActionTypeSwitcher:          txTypeSwitcher,
			AccountBalanceQuery:         accountBalanceQuery,
			ParticipationScoreQuery:     participationScoreQuery,
			NodeRegistrationQuery:       nodeRegistrationQuery,
			BlocksmithStrategy:          blocksmithStrategy,
			Observer:                    obsr,
			Logger:                      logger,
			AccountLedgerQuery:          accountLedgerQuery,
			BlockIncompleteQueueService: blockIncompleteQueueService,
		}
	case *chaintype.SpineChain:
		return &BlockSpineService{
			Chaintype: ct,
			// KVExecutor:            kvExecutor,
			QueryExecutor: queryExecutor,
			BlockQuery:    spineBlockQuery,
			// SpinePublicKeyQuery:   spinePublicKeyQuery,
			Signature: signature,
			// NodeRegistrationQuery: nodeRegistrationQuery,
			BlocksmithStrategy: blocksmithStrategy,
			Observer:           obsr,
			Logger:             logger,
			SpinePublicKeyService: &BlockSpinePublicKeyService{
				Logger:                logger,
				NodeRegistrationQuery: nodeRegistrationQuery,
				QueryExecutor:         queryExecutor,
				Signature:             signature,
				SpinePublicKeyQuery:   spinePublicKeyQuery,
			},
			MegablockService: NewMegablockService(
				queryExecutor,
				megablockQuery,
				fileChunkQuery,
				logger,
			),
		}
	}
	return nil
}
