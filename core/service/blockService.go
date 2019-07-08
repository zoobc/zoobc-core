package service

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
	core_util "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockServiceInterface interface {
		VerifySeed(seed *big.Int, balance *big.Int, previousBlock model.Block, timestamp int64) bool
		NewBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
			hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, secretPhrase string) *model.Block
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
			hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, smithScale int64, cumulativeDifficulty *big.Int, genesisSignature []byte) *model.Block
		PushBlock(previousBlock, block model.Block) error
		GetLastBlock() (model.Block, error)
		GetBlocks() ([]model.Block, error)
		GetGenesisBlock() (model.Block, error)
	}

	BlockService struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Blocks        []model.Block
	}
)

func NewBlockService(chaintype contract.ChainType, queryExecutor query.ExecutorInterface, blockQuery query.BlockQueryInterface) *BlockService {
	return &BlockService{
		Chaintype:     chaintype,
		QueryExecutor: queryExecutor,
		BlockQuery:    blockQuery,
		Blocks:        []model.Block{},
	}
}

// NewBlock generate new block
func (*BlockService) NewBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
	hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
	transactions []*model.Transaction, payloadHash []byte, secretPhrase string) *model.Block {
	block := &model.Block{
		Version:           version,
		PreviousBlockHash: previousBlockHash,
		BlockSeed:         blockSeed,
		BlocksmithID:      blocksmithID,
		Height:            previousBlockHeight,
		Timestamp:         timestamp,
		TotalAmount:       totalAmount,
		TotalFee:          totalFee,
		TotalCoinBase:     totalCoinBase,
		Transactions:      transactions,
		PayloadHash:       payloadHash,
	}
	//block.BlockSignature = core.MakeSignature(block.Byte(), GetPrivateKeyFromSeed(secretPhrase))
	//block.Bytes = nil
	return block
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (*BlockService) NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
	hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
	transactions []*model.Transaction, payloadHash []byte, smithScale int64, cumulativeDifficulty *big.Int, genesisSignature []byte) *model.Block {
	block := &model.Block{
		Version:              version,
		PreviousBlockHash:    previousBlockHash,
		BlockSeed:            blockSeed,
		BlocksmithID:         blocksmithID,
		Height:               previousBlockHeight,
		Timestamp:            timestamp,
		TotalAmount:          totalAmount,
		TotalFee:             totalFee,
		TotalCoinBase:        totalCoinBase,
		Transactions:         transactions,
		PayloadHash:          payloadHash,
		SmithScale:           smithScale,
		CumulativeDifficulty: cumulativeDifficulty.String(),
		BlockSignature:       genesisSignature,
	}
	//block.Bytes = nil
	return block
}

// VerifySeed Verify a block can be forged (by a given account, using computed seed value and account balance).
// Can be used to check who's smithing the next block (lastBlock) or if last forged block
// (previousBlock) is acceptable by the network (meaning has been smithed by a valid blocksmith).
func (*BlockService) VerifySeed(seed *big.Int, balance *big.Int, previousBlock model.Block, timestamp int64) bool {
	elapsedTime := timestamp - previousBlock.GetTimestamp()
	effectiveBaseTarget := new(big.Int).Mul(balance, big.NewInt(previousBlock.GetSmithScale()))
	prevTarget := new(big.Int).Mul(big.NewInt(int64(elapsedTime-1)), effectiveBaseTarget)
	target := new(big.Int).Add(effectiveBaseTarget, prevTarget)
	return seed.Cmp(target) < 0 && (seed.Cmp(prevTarget) >= 0 || elapsedTime > 300)
}

// PushBlock push block into blockchain
func (bs *BlockService) PushBlock(previousBlock, block model.Block) error {
	if previousBlock.GetID() != -1 {
		block.Height = previousBlock.GetHeight() + 1
		block = core_util.CalculateSmithScale(previousBlock, block, bs.Chaintype.GetChainSmithingDelayTime())
	}
	result, err := bs.QueryExecutor.ExecuteStatement(bs.BlockQuery.InsertBlock(), bs.BlockQuery.ExtractModel(block)...)
	if err != nil {
		return err
	}
	fmt.Printf("got new block, %v", result)
	return nil
	// apply transactions

	// broadcast block
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetLastBlock())
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return model.Block{
			ID: -1,
		}, err
	}
	var lastBlock model.Block
	if rows.Next() {
		err = rows.Scan(&lastBlock.ID, &lastBlock.PreviousBlockHash, &lastBlock.Height, &lastBlock.Timestamp, &lastBlock.BlockSeed, &lastBlock.BlockSignature, &lastBlock.CumulativeDifficulty,
			&lastBlock.SmithScale, &lastBlock.PayloadLength, &lastBlock.PayloadHash, &lastBlock.BlocksmithID, &lastBlock.TotalAmount, &lastBlock.TotalFee, &lastBlock.TotalCoinBase, &lastBlock.Version)
		if err != nil {
			return model.Block{
				ID: -1,
			}, err
		}
		return lastBlock, nil
	} else {
		return model.Block{
			ID: -1,
		}, errors.New("BlockNotFound")
	}

}

// GetGenesis return the last pushed block
func (bs *BlockService) GetGenesisBlock() (model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetGenesisBlock())
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return model.Block{
			ID: -1,
		}, err
	}
	var lastBlock model.Block
	if rows.Next() {
		err = rows.Scan(&lastBlock.ID, &lastBlock.PreviousBlockHash, &lastBlock.Height, &lastBlock.Timestamp, &lastBlock.BlockSeed, &lastBlock.BlockSignature, &lastBlock.CumulativeDifficulty,
			&lastBlock.SmithScale, &lastBlock.PayloadLength, &lastBlock.PayloadHash, &lastBlock.BlocksmithID, &lastBlock.TotalAmount, &lastBlock.TotalFee, &lastBlock.TotalCoinBase, &lastBlock.Version)
		if err != nil {
			return model.Block{
				ID: -1,
			}, err
		}
		return lastBlock, nil
	} else {
		return model.Block{
			ID: -1,
		}, errors.New("BlockNotFound")
	}

}

// GetBlocks return all pushed blocks
func (bs *BlockService) GetBlocks() ([]model.Block, error) {
	var blocks []model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlocks(0, 100))
	defer func() {
		rows.Close()
	}()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var block model.Block
		err = rows.Scan(&block.ID, &block.PreviousBlockHash, &block.Height, &block.Timestamp, &block.BlockSeed, &block.BlockSignature, &block.CumulativeDifficulty,
			&block.SmithScale, &block.PayloadLength, &block.PayloadHash, &block.BlocksmithID, &block.TotalAmount, &block.TotalFee, &block.TotalCoinBase, &block.Version)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
