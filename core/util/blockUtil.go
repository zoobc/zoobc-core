// util package contain basic utilities commonly used across the core package
package util

import (
	"bytes"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// GetBlockSeed calculate seed value, the first 8 byte of the digest(previousBlockSeed, nodeID)
func GetBlockSeed(nodeID int64, block *model.Block) (int64, error) {
	digest := sha3.New256()
	_, err := digest.Write(block.GetBlockSeed())
	if err != nil {
		return 0, err
	}
	previousSeedHash := digest.Sum([]byte{})
	payload := bytes.NewBuffer([]byte{})
	payload.Write(commonUtils.ConvertUint64ToBytes(uint64(nodeID)))
	payload.Write(previousSeedHash)
	seed := sha3.Sum256(payload.Bytes())
	return new(big.Int).SetBytes(seed[:8]).Int64(), nil
}

// CalculateCumulativeDifficulty get the cumulative difficulty of the incoming block based on its blocksmith index
func CalculateCumulativeDifficulty(
	previousBlock *model.Block,
	blocksmithIndex int64,
) (string, error) {
	previousCumulativeDifficulty, ok := new(big.Int).SetString(previousBlock.CumulativeDifficulty, 10)
	if !ok {
		return "", blocker.NewBlocker(blocker.AppErr, "FailToCalculateCummulativeDifficulty")
	}
	currentCumulativeDifficulty := constant.CumulativeDifficultyDivisor / (blocksmithIndex + 1)

	newCumulativeDifficulty := new(big.Int).Add(
		previousCumulativeDifficulty, new(big.Int).SetInt64(currentCumulativeDifficulty),
	)
	return newCumulativeDifficulty.String(), nil
}

// GetBlockID generate block ID value if haven't
// return the assigned ID if assigned
func GetBlockID(block *model.Block, ct chaintype.ChainType) int64 {
	if block.ID == 0 {
		// Attention! make sure we have the full block data here (block + transactions/spine pub keys, etc...)
		hash, err := commonUtils.GetBlockHash(block, ct)
		if err != nil {
			return 0
		}
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

func IsBlockIDExist(blockIds []int64, expectedBlockID int64) bool {
	for _, blockID := range blockIds {
		if blockID == expectedBlockID {
			return true
		}
	}
	return false
}

// CalculateNodeOrder calculate the Node order parameter, used to sort/select the group of blocksmith rewarded for a given block
func CalculateNodeOrder(score *big.Int, blockSeed, nodeID int64) *big.Int {
	prn := crypto.PseudoRandomGenerator(uint64(nodeID), uint64(blockSeed), crypto.PseudoRandomSha3256)
	return new(big.Int).SetUint64(prn)
}

func IsGenesis(previousBlockID int64, block *model.Block) bool {
	return previousBlockID == -1 && block.CumulativeDifficulty != ""
}

// GetAddRemoveSpineKeyAction transcode nodeRegistrationStatus into relative spinekeypublickey acion
// eg. if node is deleted, the action for this spine public key is "RemoveKey", if registered "AddKey"
func GetAddRemoveSpineKeyAction(nodeRegistrationStatus uint32) (publicKeyAction model.SpinePublicKeyAction) {
	switch nodeRegistrationStatus {
	case uint32(model.NodeRegistrationState_NodeDeleted):
		publicKeyAction = model.SpinePublicKeyAction_RemoveKey
	case uint32(model.NodeRegistrationState_NodeRegistered):
		publicKeyAction = model.SpinePublicKeyAction_AddKey
	}
	return publicKeyAction
}
