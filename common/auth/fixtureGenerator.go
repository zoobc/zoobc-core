package auth

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	//STEF TODO: insert real accountAddress here
	senderAddress1 = make([]byte, 36)
	// var senderSeed1 = "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"
	nodeSeed1   = "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
	nodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	block1 = &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
)

func GetFixturesProofOfOwnershipValidation(height uint32, hash []byte, block *model.Block) (poown *model.ProofOfOwnership) {
	if block == nil {
		block = block1
	}
	if block.Transactions == nil {
		block.Transactions = make([]*model.Transaction, 0)
	}
	if len(hash) == 0 {
		blockHash, _ := util.GetBlockHash(block, &chaintype.MainChain{})
		hash = blockHash
	}
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      hash,
		BlockHeight:    height,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	return
}
