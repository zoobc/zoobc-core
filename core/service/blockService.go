package service

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/chaintype"
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
	}

	BlockService struct {
		Chaintype chaintype.Chaintype
		Blocks    []model.Block
	}
)

func NewBlockService(chaintype chaintype.Chaintype) *BlockService {
	return &BlockService{
		Chaintype: chaintype,
		Blocks:    []model.Block{},
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
		bs.Blocks = append(bs.Blocks, block)
		fmt.Print("got new block")
		return nil

	} else {
		bs.Blocks = append(bs.Blocks, block)
		return nil
	}

	// log.Printf("\npushing block to in memory block list\n%v\n", block.GetBaseTarget())

	// apply transactions

	// broadcast block
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (model.Block, error) {
	if len(bs.Blocks) > 0 {
		return bs.Blocks[len(bs.Blocks)-1], nil
	}
	return model.Block{
		ID: -1,
	}, errors.New("No Block Yet")
}

// GetBlocks return all pushed blocks
func (bs *BlockService) GetBlocks() ([]model.Block, error) {
	return bs.Blocks, nil
}
