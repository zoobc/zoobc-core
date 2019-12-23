package model

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
)

// Blocksmith is wrapper for the account in smithing process
type Blocksmith struct {
	Chaintype     chaintype.ChainType
	NodeID        int64
	NodePublicKey []byte
	NodeOrder     *big.Int
	SmithOrder    uint32
	Score         *big.Int
	SmithTime     int64
	BlockSeed     int64
	SecretPhrase  string
	Deadline      uint32
}

// NewBlocksmith initiate generator
func NewBlocksmith(chaintype chaintype.ChainType, nodeSecretPhrase string, nodePublicKey []byte, nodeID int64) *Blocksmith {
	blocksmith := &Blocksmith{
		Chaintype:     chaintype,
		Score:         big.NewInt(constant.DefaultParticipationScore),
		SecretPhrase:  nodeSecretPhrase,
		NodePublicKey: nodePublicKey,
		NodeID:        nodeID,
	}
	return blocksmith
}
