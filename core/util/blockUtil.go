// util package contain basic utilities commonly used across the core package
package util

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// GetBlockSeed calculate seed value, the first 8 byte of the digest(previousBlockSeed, publicKey)
func GetBlockSeed(publicKey []byte, block *model.Block) (*big.Int, error) {
	digest := sha3.New512()
	_, err := digest.Write(block.GetBlockSeed())
	if err != nil {
		return nil, err
	}
	_, err = digest.Write(publicKey)

	if err != nil {
		return nil, err
	}

	blockSeedHash := digest.Sum([]byte{})
	res := new(big.Int)
	return res.SetBytes([]byte{
		blockSeedHash[7],
		blockSeedHash[6],
		blockSeedHash[5],
		blockSeedHash[4],
		blockSeedHash[3],
		blockSeedHash[2],
		blockSeedHash[1],
		blockSeedHash[0],
	}), nil
}

// GetSmithTime calculate smith time of a blocksmith
func GetSmithTime(balance, seed *big.Int, block *model.Block) int64 {
	if balance.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	staticTarget := new(big.Int).Mul(big.NewInt(block.SmithScale), balance)
	elapsedFromLastBlock := new(big.Int).Div(seed, staticTarget).Int64()
	return block.GetTimestamp() + elapsedFromLastBlock
}

// CalculateSmithScale base target of block and return modified block
func CalculateSmithScale(previousBlock, block *model.Block, smithingDelayTime int64) *model.Block {
	prevSmithScale := previousBlock.GetSmithScale()
	smithScaleMul := new(big.Int).Mul(big.NewInt(prevSmithScale), big.NewInt(block.GetTimestamp()-previousBlock.GetTimestamp()))
	block.SmithScale = new(big.Int).Div(smithScaleMul, big.NewInt(smithingDelayTime)).Int64()
	if big.NewInt(block.GetSmithScale()).Cmp(big.NewInt(0)) < 0 || big.NewInt(block.GetSmithScale()).Cmp(
		big.NewInt(constant.MaxSmithScale)) > 0 {
		block.SmithScale = constant.MaxSmithScale
	}
	if big.NewInt(block.GetSmithScale()).Cmp(new(big.Int).Div(big.NewInt(prevSmithScale), big.NewInt(2))) < 0 {
		block.SmithScale = prevSmithScale / 2
	}
	if big.NewInt(block.GetSmithScale()).Cmp(big.NewInt(0)) == 0 {
		block.SmithScale = 1
	}
	twoFoldCurSmithScale := new(big.Int).Mul(big.NewInt(prevSmithScale), big.NewInt(2))
	if twoFoldCurSmithScale.Cmp(big.NewInt(0)) < 0 {
		twoFoldCurSmithScale = big.NewInt(constant.MaxSmithScale)
	}
	if big.NewInt(block.GetSmithScale()).Cmp(twoFoldCurSmithScale) > 0 {
		block.SmithScale = twoFoldCurSmithScale.Int64()
	}

	two64, _ := new(big.Int).SetString(constant.Two64, 0)
	previousBlockCumulativeDifficulty, _ := new(big.Int).SetString(previousBlock.GetCumulativeDifficulty(), 10)
	block.CumulativeDifficulty = new(big.Int).Add(
		previousBlockCumulativeDifficulty,
		new(big.Int).Div(two64, big.NewInt(block.GetSmithScale()))).String()
	return block
}

// GetBlockID generate block ID value if haven't
// return the assigned ID if assigned
func GetBlockID(block *model.Block) int64 {
	if block.ID == 0 {
		digest := sha3.New512()
		blockByte, _ := GetBlockByte(block, true)
		_, _ = digest.Write(blockByte)
		hash, _ := GetBlockHash(block)
		block.ID = GetBlockIDFromHash(hash)
	}
	return block.ID
}

// GetBlockIdFromHash returns blockID from given hash
func GetBlockIDFromHash(blockHash []byte) int64 {
	res := new(big.Int)
	return res.SetBytes([]byte{
		blockHash[7],
		blockHash[6],
		blockHash[5],
		blockHash[4],
		blockHash[3],
		blockHash[2],
		blockHash[1],
		blockHash[0],
	}).Int64()
}

// GetBlockHash return the block's bytes hash.
// note: the block must be signed, otherwise this function returns an error
func GetBlockHash(block *model.Block) ([]byte, error) {
	digest := sha3.New512()
	blockByte, _ := GetBlockByte(block, true)
	_, err := digest.Write(blockByte)
	if err != nil {
		return nil, err
	}
	return digest.Sum([]byte{}), nil
}

// GetBlockByte generate value for `Bytes` field if not assigned yet
// return .`Bytes` if value assigned
func GetBlockByte(block *model.Block, signed bool) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(block.GetVersion()))
	buffer.Write(util.ConvertUint64ToBytes(uint64(block.GetTimestamp())))
	buffer.Write(util.ConvertIntToBytes(len(block.GetTransactions())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(block.GetTotalAmount())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(block.GetTotalFee())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(block.GetTotalCoinBase())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(block.GetPayloadLength())))
	buffer.Write(block.PayloadHash)
	buffer.Write(util.ConvertUint32ToBytes(block.BlocksmithAddressLength))
	buffer.Write([]byte(block.GetBlocksmithAddress()))
	buffer.Write(block.GetBlockSeed())
	buffer.Write(block.GetPreviousBlockHash())
	if signed {
		if block.BlockSignature == nil {
			return nil, errors.New("BlockNotSigned")
		}
		buffer.Write(block.BlockSignature)
	}
	return buffer.Bytes(), nil
}

// ValidateBlock validate block to be pushed into the blockchain
func ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	if block.GetTimestamp() > curTime+15 {
		return errors.New("InvalidTimestamp")
	}
	if GetBlockID(block) == 0 {
		return errors.New("duplicate block:TODO:conditionNotComplete")
	}

	previousBlockIDFromHash := new(big.Int)
	previousBlockIDFromHashInt := previousBlockIDFromHash.SetBytes([]byte{
		block.PreviousBlockHash[7],
		block.PreviousBlockHash[6],
		block.PreviousBlockHash[5],
		block.PreviousBlockHash[4],
		block.PreviousBlockHash[3],
		block.PreviousBlockHash[2],
		block.PreviousBlockHash[1],
		block.PreviousBlockHash[0],
	}).Int64()
	if previousLastBlock.ID != previousBlockIDFromHashInt {
		return errors.New("previous block ID does not match with current previous_block_hash value")
	}
	return nil
}

func IsBlockIDExist(blockIds []int64, expectedBlockId int64) bool {
	for _, blockId := range blockIds {
		if blockId == expectedBlockId {
			return true
		}
	}
	return false
}
