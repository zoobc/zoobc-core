package model

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/constant"
)

// Blocksmith is wrapper for the account in smithing process
type Blocksmith struct {
	NodeID        int64
	NodePublicKey []byte
	NodeOrder     *big.Int
	SmithOrder    *big.Int
	Score         *big.Int
	SmithTime     int64
	BlockSeed     *big.Int
	SecretPhrase  string
	Deadline      uint32
}

// InitGenerator initiate generator
func NewBlocksmith(nodeSecretPhrase string, nodePublicKey []byte) *Blocksmith {
	blocksmith := &Blocksmith{
		Score:         big.NewInt(constant.DefaultParticipationScore),
		SecretPhrase:  nodeSecretPhrase,
		NodePublicKey: nodePublicKey,
	}
	return blocksmith
}

// GetTimestamp max timestamp allowed block to be smithed
func (blocksmith *Blocksmith) GetTimestamp(smithMax int64) int64 {
	elapsed := smithMax - blocksmith.SmithTime
	if elapsed > 3600 {
		return smithMax
	}
	return blocksmith.SmithTime + 1
}
