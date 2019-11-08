package auth

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	ProofOfOwnershipValidationInterface interface {
		ValidateProofOfOwnership(
			poown *model.ProofOfOwnership,
			nodePublicKey []byte,
			queryExecutor query.ExecutorInterface,
			blockQuery query.BlockQueryInterface,
		) error
	}

	// Signature object handle signing and verifying different signature
	ProofOfOwnershipValidation struct {
	}
)

// ValidateProofOfOwnership validates a proof of ownership message
func (*ProofOfOwnershipValidation) ValidateProofOfOwnership(
	poown *model.ProofOfOwnership,
	nodePublicKey []byte,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) error {

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
