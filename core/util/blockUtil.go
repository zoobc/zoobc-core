// util package contain basic utilities commonly used across the core package
package util

import (
	"bytes"
	"math/big"

	"github.com/zoobc/zoobc-core/common/query"

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
	return new(big.Int).SetBytes(seed[:8]), nil
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
func CalculateSmithScale(
	previousBlock, block *model.Block,
	smithingPeriod int64,
	blockQuery query.BlockQueryInterface,
	executor query.ExecutorInterface,
) *model.Block {
	switch {
	case block.Height < constant.AverageSmithingBlockHeight:
		prevSmithScale := previousBlock.GetSmithScale()
		smithScaleMul := new(big.Int).Mul(big.NewInt(prevSmithScale), big.NewInt(block.GetTimestamp()-previousBlock.GetTimestamp()))
		block.SmithScale = new(big.Int).Div(smithScaleMul, big.NewInt(smithingPeriod)).Int64()
		if block.GetSmithScale() < 0 || block.GetSmithScale() > constant.MaxSmithScale {
			block.SmithScale = constant.MaxSmithScale
		}
		if block.GetSmithScale() < prevSmithScale/2 {
			block.SmithScale = prevSmithScale / 2
		}
		if block.GetSmithScale() == 0 {
			block.SmithScale = 1
		}
		twoFoldCurSmithScale := new(big.Int).Mul(big.NewInt(prevSmithScale), big.NewInt(2))
		if twoFoldCurSmithScale.Cmp(big.NewInt(0)) < 0 {
			twoFoldCurSmithScale = big.NewInt(constant.MaxSmithScale)
		}
		if big.NewInt(block.GetSmithScale()).Cmp(twoFoldCurSmithScale) > 0 {
			block.SmithScale = twoFoldCurSmithScale.Int64()
		}
	case block.Height%2 == 0:
		var prev2Block model.Block
		prev2BlockQ := blockQuery.GetBlockByHeight(previousBlock.Height - 2)
		row := executor.ExecuteSelectRow(prev2BlockQ)
		err := blockQuery.Scan(&prev2Block, row)
		if err != nil {
			return nil
		}
		blockTimeAverage := (block.Timestamp - prev2Block.Timestamp) / 3
		if blockTimeAverage > smithingPeriod {
			if blockTimeAverage < constant.MaximumBlocktimeLimit {
				block.SmithScale = (previousBlock.SmithScale * blockTimeAverage) / smithingPeriod
			} else {
				block.SmithScale = (previousBlock.SmithScale * constant.MaximumBlocktimeLimit) / smithingPeriod
			}
		} else {
			if blockTimeAverage > constant.MinimumBlocktimeLimit {
				block.SmithScale = previousBlock.SmithScale - previousBlock.SmithScale*constant.SmithscaleGamma*
					(smithingPeriod-blockTimeAverage)/(100*smithingPeriod)
			} else {
				block.SmithScale = previousBlock.SmithScale - previousBlock.SmithScale*constant.SmithscaleGamma*
					(smithingPeriod-constant.MinimumBlocktimeLimit)/(100*smithingPeriod)
			}
		}
		if block.SmithScale < 0 || block.SmithScale > constant.MaxSmithScale2 {
			block.SmithScale = constant.MaxSmithScale2
		}
		if block.SmithScale < constant.MinSmithScale {
			block.SmithScale = constant.MinSmithScale
		}
	default:
		block.SmithScale = previousBlock.GetSmithScale()
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

func IsBlockIDExist(blockIds []int64, expectedBlockID int64) bool {
	for _, blockID := range blockIds {
		if blockID == expectedBlockID {
			return true
		}
	}
	return false
}

// CalculateSmithOrder calculate the blocksmith order parameter, used to sort/select the next blocksmith
func CalculateSmithOrder(score, blockSeed *big.Int, nodeID int64) *big.Int {
	nodePK := new(big.Int).SetUint64(uint64(nodeID))
	rndNodePK := uint64(nodePK.Int64() ^ blockSeed.Int64())
	return new(big.Int).Mul(score, new(big.Int).SetUint64(rndNodePK))
}

// CalculateNodeOrder calculate the Node order parameter, used to sort/select the group of blocksmith rewarded for a given block
func CalculateNodeOrder(score, blockSeed *big.Int, nodeID int64) *big.Int {
	nodePK := new(big.Int).SetUint64(uint64(nodeID))
	rndNodePK := uint64(nodePK.Int64() ^ blockSeed.Int64())
	return new(big.Int).Div(new(big.Int).SetUint64(rndNodePK), score)
}
