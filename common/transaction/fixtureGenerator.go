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
		AccountAddress: senderAddress1,
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
		AccountAddress: senderAddress1,
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
		AccountAddress: senderAddress1,
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

func GetFixturesForRemoveNoderegistration() (txBody *model.RemoveNodeRegistrationTransactionBody, txBodyBytes []byte) {

	txBody = &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86,
			211, 123, 72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
	}
	nr := RemoveNodeRegistration{
		Body: txBody,
	}
	txBodyBytes = nr.GetBodyBytes()
	return txBody, txBodyBytes
}
