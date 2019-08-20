package service

import (
	"bytes"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountAddress string) (*model.ProofOfOwnership, error)
		ValidateProofOfOwnership(poown *model.ProofOfOwnership, nodePublicKey []byte) error
	}

	// NodeAdminServiceHelpersInterface mockable service methods
	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
	}
)

func NewNodeAdminService(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	blockService BlockServiceInterface) *NodeAdminService {
	return &NodeAdminService{
		queryExecutor,
		blockQuery,
		signature,
		blockService,
	}
}

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(
	accountAddress string) (*model.ProofOfOwnership, error) {

	lastBlock, err := nas.BlockService.GetLastBlock()
	if err != nil {
		return nil, err
	}
	lastBlockHash, err := util.GetBlockHash(lastBlock)
	if err != nil {
		return nil, err
	}

	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      lastBlockHash,
		BlockHeight:    lastBlock.Height,
	}
	messageBytes := commonUtils.GetProofOfOwnershipMessageBytes(poownMessage)
	nodeSecretPhrase := viper.GetString("nodeSecretPhrase")
	poownSignature := crypto.NewSignature().SignByNode(messageBytes, nodeSecretPhrase)
	return &model.ProofOfOwnership{
		MessageBytes: messageBytes,
		Signature:    poownSignature,
	}, nil
}

// ValidateProofOfOwnership validates a proof of ownership message
func (nas *NodeAdminService) ValidateProofOfOwnership(poown *model.ProofOfOwnership, nodePublicKey []byte) error {

	if !crypto.NewSignature().VerifyNodeSignature(poown.MessageBytes, poown.Signature, nodePublicKey) {
		return blocker.NewBlocker(blocker.AppErr, "InvalidSignature")
	}

	message, err := commonUtils.ParseProofOfOwnershipMessageBytes(poown.MessageBytes)
	if err != nil {
		return err
	}

	lastBlock, err := nas.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	// Expiration, in number of blocks, of a proof of ownership message
	if lastBlock.Height-message.BlockHeight > constant.ProofOfOwnershipExpiration {
		return blocker.NewBlocker(blocker.AppErr, "ProofOfOwnershipExpired")
	}

	poownBlockRef, err := nas.BlockService.GetBlockByHeight(message.BlockHeight)
	if err != nil {
		return err
	}
	poownBlockHashRef, err := util.GetBlockHash(poownBlockRef)
	if err != nil {
		return err
	}
	if !bytes.Equal(poownBlockHashRef, message.BlockHash) {
		return blocker.NewBlocker(blocker.AppErr, "InvalidProofOfOwnershipBlockHash")
	}
	return nil
}
