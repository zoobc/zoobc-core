package model

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/constant"
)

// Blocksmith is wrapper for the account in smithing process
type Blocksmith struct {
	NodeID        int64
	NodePublicKey []byte
	Score         *big.Int
	SecretPhrase  string
}

// NewBlocksmith initiate generator
func NewBlocksmith(nodeSecretPhrase string, nodePublicKey []byte, nodeID int64) *Blocksmith {
	blocksmith := &Blocksmith{
		Score:         big.NewInt(constant.DefaultParticipationScore),
		SecretPhrase:  nodeSecretPhrase,
		NodePublicKey: nodePublicKey,
		NodeID:        nodeID,
	}
	return blocksmith
}
