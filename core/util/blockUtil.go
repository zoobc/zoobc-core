// util package contain basic utilities commonly used across the core package
package util

import (
	"bytes"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	commonUtils "github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

// GetBlockSeed calculate seed value, the first 8 byte of the digest(previousBlockSeed, publicKey)
func GetBlockSeed(publicKey []byte, block *model.Block, secretPhrase string) (*big.Int, error) {
	digest := sha3.New256()
	_, err := digest.Write(block.GetBlockSeed())
	if err != nil {
		return nil, err
	}

	previousSeedHash := digest.Sum([]byte{})
	payload := bytes.NewBuffer([]byte{})
	payload.Write(publicKey)
	payload.Write(previousSeedHash)
	signature := (&crypto.Signature{}).SignByNode(payload.Bytes(), secretPhrase)
	seed := sha3.Sum256(signature)
	return new(big.Int).SetBytes([]byte{
		seed[7],
		seed[6],
		seed[5],
		seed[4],
		seed[3],
		seed[2],
		seed[1],
		seed[0],
	}), nil
}

// GetSmithTime calculate smith time of a blocksmith
func GetSmithTime(score, seed *big.Int, block *model.Block) int64 {
	if score.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	staticTarget := new(big.Int).Mul(big.NewInt(block.SmithScale), score)
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
		digest := sha3.New256()
		blockByte, _ := commonUtils.GetBlockByte(block, true)
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
	digest := sha3.New256()
	blockByte, _ := commonUtils.GetBlockByte(block, true)
	_, err := digest.Write(blockByte)
	if err != nil {
		return nil, err
	}
	return digest.Sum([]byte{}), nil
}

// ValidateBlock validate block to be pushed into the blockchain
func ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	if block.GetTimestamp() > curTime+15 {
		return blocker.NewBlocker(blocker.BlockErr, "invalid timestamp")
	}
	if GetBlockID(block) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "invalid ID")
	}
	// Verify Signature
	sig := new(crypto.Signature)
	blockByte, err := commonUtils.GetBlockByte(block, false)
	if err != nil {
		return err
	}

	if !sig.VerifyNodeSignature(
		blockByte,
		block.BlockSignature,
		block.BlocksmithPublicKey,
	) {
		return blocker.NewBlocker(blocker.BlockErr, "invalid signature")
	}
	// Verify previous block hash
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
		return blocker.NewBlocker(blocker.BlockErr, "invalid previous block hash")
	}
	return nil
}

func IsBlockIDExist(blockIds []int64, expectedBlockID int64) bool {
	for _, blockID := range blockIds {
		if blockID == expectedBlockID {
			return true
		}
	}
	return false
}
