package transaction

import (
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var senderAddress1 = "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
var senderSeed1 = "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"

func GetFixturesForNoderegistration() (poownMessage *model.ProofOfOwnershipMessage, poown *model.ProofOfOwnership,
	txBody *model.NodeRegistrationTransactionBody, txBodyBytes []byte) {

	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		BlockHash: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224,
			101, 127, 241, 62, 152, 187, 255, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89,
			107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75},
		BlockHeight: 0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().Sign(
		poownMessageBytes,
		senderAddress1,
		senderSeed1,
	)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.NodeRegistrationTransactionBody{
		NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43,
			63, 224, 101, 127, 241, 62, 152, 187, 255},
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		NodeAddress:    "10.10.0.1",
		LockedBalance:  10000000000,
		Poown:          poown,
	}
	nr := NodeRegistration{
		Body: txBody,
	}
	txBodyBytes = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}

func GetFixturesForUpdateNoderegistration() (poownMessage *model.ProofOfOwnershipMessage, poown *model.ProofOfOwnership,
	txBody *model.UpdateNodeRegistrationTransactionBody, txBodyBytes []byte) {

	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		BlockHash: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224,
			101, 127, 241, 62, 152, 187, 255, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89,
			107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75},
		BlockHeight: 0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().Sign(
		poownMessageBytes,
		senderAddress1,
		senderSeed1,
	)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43,
			63, 224, 101, 127, 241, 62, 152, 187, 255},
		NodeAddress:   "10.10.0.1",
		LockedBalance: 10000000000,
		Poown:         poown,
	}
	nr := UpdateNodeRegistration{
		Body: txBody,
	}
	txBodyBytes = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}
