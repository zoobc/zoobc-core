package service

import (
	"bytes"
	"errors"

	proto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountType uint32, accountAddress string, signature []byte) (*model.ProofOfOwnership, error)
		ValidateProofOfOwnership(poown *model.ProofOfOwnership, nodePublicKey []byte)
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
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
		Helpers       NodeAdminServiceHelpersInterface
		BlockService  BlockServiceInterface
	}
)

func NewNodeAdminService(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	accountQuery query.AccountQueryInterface,
	signature crypto.SignatureInterface,
	helpers NodeAdminServiceHelpersInterface,
	blockService BlockServiceInterface) *NodeAdminService {
	return &NodeAdminService{
		queryExecutor,
		blockQuery,
		accountQuery,
		signature,
		helpers,
		blockService,
	}
}

// GetMessageSize return the message size in bytes
// note: it can be used to validate a message
func (*NodeAdminService) GetMessageSize() uint32 {
	accountType := 4
	//TODO: this is valid for account type = 0
	accountAddress := 44
	blockHash := 64
	blockHeight := 4
	return uint32(accountType + accountAddress + blockHash + blockHeight)
}

// GetBytesFromMessage wrapper around proto.marshal function. returns the message's bytes
func (*NodeAdminService) GetBytesFromMessage(poown *model.ProofOfOwnershipMessage) ([]byte, error) {
	b, err := proto.Marshal(poown)
	if err != nil {
		return nil, errors.New("InvalidPoownMessage")
	}
	return b, nil
}

// GetBytesFromMessage wrapper around proto.marshal function. returns the message's bytes
func (nas *NodeAdminService) ParseMessageBytes(messageBytes []byte) (*model.ProofOfOwnershipMessage, error) {
	message := new(model.ProofOfOwnershipMessage)
	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return nil, errors.New("InvalidPoownMessageBytes")
	}
	return message, nil
}

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(accountType uint32,
	accountAddress string) (*model.ProofOfOwnership, error) {

	ownerAccountType := viper.GetUint32("ownerAccountType")
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	if ownerAccountAddress != accountAddress && ownerAccountType != accountType {
		return nil, errors.New("PoownAccountNotNodeOwner")
	}

	lastBlock, err := nas.BlockService.GetLastBlock()
	if err != nil {
		return nil, err
	}
	lastBlockHash, err := util.GetBlockHash(lastBlock)
	if err != nil {
		return nil, err
	}

	poownMessage := &model.ProofOfOwnershipMessage{
		AccountType:    accountType,
		AccountAddress: accountAddress,
		BlockHash:      lastBlockHash,
		BlockHeight:    lastBlock.Height,
	}
	messageBytes, err := nas.Helpers.GetBytesFromMessage(poownMessage)
	log.Println(messageBytes)
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
		return errors.New("InvalidSignature")
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
		return errors.New("ProofOfOwnershipExpired")
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
		return errors.New("InvalidProofOfOwnershipBlockHash")
	}
	return nil
}
