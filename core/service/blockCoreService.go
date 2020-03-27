package service

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
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
			empty bool,
		) (*model.Block, error)
		ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error
		ValidatePayloadHash(block *model.Block) error
		GetPayloadHashAndLength(block *model.Block) (payloadHash []byte, payloadLength uint32, err error)
		PushBlock(previousBlock, block *model.Block, broadcast, persist bool) error
		GetBlockByID(id int64, withAttachedData bool) (*model.Block, error)
		GetBlockByHeight(uint32) (*model.Block, error)
		GetBlocksFromHeight(startHeight, limit uint32, withAttachedData bool) ([]*model.Block, error)
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
			peer *model.Peer,
		) (*model.BatchReceipt, error)
		GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error)
		PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error)
		GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface
		ReceivedValidatedBlockTransactionsListener() observer.Listener
		BlockTransactionsRequestedListener() observer.Listener
		WillSmith(
			blocksmith *model.Blocksmith,
			blockchainProcessorLastBlockID int64,
		) (lastBlockID, blocksmithIndex int64, err error)
	}
)
