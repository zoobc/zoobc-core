package auth

import (
	"bytes"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	NodeAuthValidationInterface interface {
		ValidateProofOfOwnership(
			poown *model.ProofOfOwnership,
			nodePublicKey []byte,
			queryExecutor query.ExecutorInterface,
			blockQuery query.BlockQueryInterface,
		) error
		ValidateProofOfOrigin(
			poorig *model.ProofOfOrigin,
			nodePublicKey,
			challengeResponse []byte,
		) error
	}

	// Signature object handle signing and verifying different signature
	NodeAuthValidation struct {
		Signature crypto.SignatureInterface
	}
)

func NewNodeAuthValidation(
	signature crypto.SignatureInterface,
) *NodeAuthValidation {
	return &NodeAuthValidation{
		Signature: signature,
	}
}

// ValidateProofOfOwnership validates a proof of ownership message
func (nav *NodeAuthValidation) ValidateProofOfOwnership(
	poown *model.ProofOfOwnership,
	nodePublicKey []byte,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) error {

	// TODO: use composition instead, such as per ValidateProofOfOrigin
	if !crypto.NewSignature().VerifyNodeSignature(poown.MessageBytes, poown.Signature, nodePublicKey) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}

	message, err := util.ParseProofOfOwnershipMessageBytes(poown.MessageBytes)
	if err != nil {
		return err
	}

	lastBlock, err := util.GetLastBlock(queryExecutor, blockQuery)
	if err != nil {
		return err
	}
	// Expiration, in number of blocks, of a proof of ownership message
	if lastBlock.Height-message.BlockHeight > constant.ProofOfOwnershipExpiration {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOwnershipExpired")
	}

	poownBlockRef, err := util.GetBlockByHeight(message.BlockHeight, queryExecutor, blockQuery)
	if err != nil {
		return err
	}
	if !bytes.Equal(poownBlockRef.BlockHash, message.BlockHash) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHash")
	}
	return nil
}

// ValidateProofOfOrigin validates a proof of origin message
func (nav *NodeAuthValidation) ValidateProofOfOrigin(
	poorig *model.ProofOfOrigin,
	nodePublicKey,
	challengeResponse []byte,
) error {
	if poorig == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOriginNotProvided")
	}
	if poorig.Timestamp < time.Now().Unix() {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOriginExpired")
	}

	if !bytes.Equal(challengeResponse, poorig.MessageBytes) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidChallengeResponse")
	}

	if !nav.Signature.VerifyNodeSignature(
		util.GetProofOfOriginUnsignedBytes(poorig),
		poorig.Signature,
		nodePublicKey,
	) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}

	return nil
}
