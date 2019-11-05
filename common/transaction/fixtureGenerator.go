package transaction

import (
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	senderAddress1 = "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
	// var senderSeed1 = "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"
	nodeSeed1   = "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
	nodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	nodePubKey2 = []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234,
		181, 100, 163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220}
	block1 = &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		SmithScale:           1,
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
)

func GetFixturesForNoderegistration(nodeRegistrationQuery query.NodeRegistrationQueryInterface) (
	poownMessage *model.ProofOfOwnershipMessage,
	poown *model.ProofOfOwnership,
	txBody *model.NodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {
	blockHash, _ := util.GetBlockHash(block1)
	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.NodeRegistrationTransactionBody{
		NodePublicKey:  nodePubKey1,
		AccountAddress: senderAddress1,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.0.1",
		},
		LockedBalance: 10000000000,
		Poown:         poown,
	}
	nr := NodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
	txBodyBytes = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}

func GetFixturesForUpdateNoderegistration(nodeRegistrationQuery query.NodeRegistrationQueryInterface) (
	poownMessage *model.ProofOfOwnershipMessage,
	poown *model.ProofOfOwnership,
	txBody *model.UpdateNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {
	blockHash, _ := util.GetBlockHash(block1)

	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.0.1",
		},
		LockedBalance: 10000000000,
		Poown:         poown,
	}
	nr := UpdateNodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
	txBodyBytes = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}

func GetFixturesForClaimNoderegistration() (
	poown *model.ProofOfOwnership,
	txBody *model.ClaimNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {

	blockHash, _ := util.GetBlockHash(block1)
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
		Poown:         poown,
	}
	nr := ClaimNodeRegistration{
		Body: txBody,
	}
	txBodyBytes = nr.GetBodyBytes()
	return
}

func GetFixturesForRemoveNoderegistration() (
	txBody *model.RemoveNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {

	txBody = &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
	}
	nr := RemoveNodeRegistration{
		Body: txBody,
	}
	txBodyBytes = nr.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixturesForSetupAccountDataset() (
	txBody *model.SetupAccountDatasetTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.SetupAccountDatasetTransactionBody{
		SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		Property:                "Admin",
		Value:                   "Welcome",
		MuchTime:                123,
	}

	sa := SetupAccountDataset{
		Body: txBody,
	}
	return txBody, sa.GetBodyBytes()
}

func GetFixturesForRemoveAccountDataset() (
	txBody *model.RemoveAccountDatasetTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.RemoveAccountDatasetTransactionBody{
		SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		Property:                "Admin",
		Value:                   "Good bye",
	}

	ra := RemoveAccountDataset{
		Body: txBody,
	}
	return txBody, ra.GetBodyBytes()
}
