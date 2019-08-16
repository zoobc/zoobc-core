package service

import (
	"bytes"

	proto "github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountAddress string) (*model.ProofOfOwnership, error)
		ValidateProofOfOwnership(poown *model.ProofOfOwnership, nodePublicKey []byte) error
	}

	// NodeAdminServiceHelpersInterface mockable service methods
	NodeAdminServiceHelpersInterface interface {
		GetBytesFromMessage(poown *model.ProofOfOwnershipMessage) ([]byte, error)
		ParseMessageBytes(messageBytes []byte) (*model.ProofOfOwnershipMessage, error)
		// TODO: to be implemented: method to validate a request for a new proof of ownership, coming from the client
		// ValidateProofOfOwnershipRequest(accountType uint32, accountAddress string, signature []byte) bool
	}

	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		Helpers       NodeAdminServiceHelpersInterface
		BlockService  BlockServiceInterface
	}
)

func NewNodeAdminService(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	helpers NodeAdminServiceHelpersInterface,
	blockService BlockServiceInterface) *NodeAdminService {
	return &NodeAdminService{
		queryExecutor,
		blockQuery,
		signature,
		helpers,
		blockService,
	}
}

// GetBytesFromMessage wrapper around proto.marshal function. returns the message's bytes
func (*NodeAdminService) GetBytesFromMessage(poown *model.ProofOfOwnershipMessage) ([]byte, error) {
	b, err := proto.Marshal(poown)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockErr, "InvalidPoownMessage")
	}
	return b, nil
}

// GetBytesFromMessage wrapper around proto.marshal function. returns the message's bytes
func (nas *NodeAdminService) ParseMessageBytes(messageBytes []byte) (*model.ProofOfOwnershipMessage, error) {
	message := new(model.ProofOfOwnershipMessage)
	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return nil, blocker.NewBlocker(blocker.BlockErr, "InvalidPoownMessageBytes")
	}
	return message, nil
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
	messageBytes, err := nas.Helpers.GetBytesFromMessage(poownMessage)
	if err != nil {
		return nil, err
	}
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
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSignature")
	}

	message, err := nas.Helpers.ParseMessageBytes(poown.MessageBytes)
	if err != nil {
		return err
	}

	lastBlock, err := nas.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	// Expiration, in number of blocks, of a proof of ownership message
	if lastBlock.Height-message.BlockHeight > constant.ProofOfOwnershipExpiration {
		return blocker.NewBlocker(blocker.BlockErr, "ProofOfOwnershipExpired")
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
		return blocker.NewBlocker(blocker.BlockErr, "InvalidProofOfOwnershipBlockHash")
	}
	return nil
}
