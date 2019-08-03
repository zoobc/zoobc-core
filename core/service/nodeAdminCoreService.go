package service

import (
	"bytes"
	"errors"

	proto "github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
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
		LoadOwnerAccountFromConfig() (ownerAccountType uint32, ownerAccountAddress string, err error)
		LoadNodeSeedFromConfig() (nodeSeed string, err error)
		GetBytesFromMessage(poown *model.ProofOfOwnershipMessage) ([]byte, error)
		ParseMessageBytes(messageBytes []byte) (*model.ProofOfOwnershipMessage, error)
		// TODO: to be implemented: method to validate a request for a new proof of ownership, coming from the client
		// ValidateProofOfOwnershipRequest(accountType uint32, accountAddress string, signature []byte) bool
		GetMessageSize() uint32
	}

	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
		Helpers       NodeAdminServiceHelpersInterface
	}
)

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
	messageLength := len(messageBytes)
	messageLengthRef := nas.GetMessageSize()
	if uint32(messageLength) != messageLengthRef {
		return nil, errors.New("InvalidPownMessageSize")
	}
	message := new(model.ProofOfOwnershipMessage)
	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return nil, errors.New("InvalidPoownMessageBytes")
	}
	return message, nil
}

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(accountType uint32,
	accountAddress string) (*model.ProofOfOwnership, error) {

	ownerAccountType, ownerAccountAddress, err := nas.Helpers.LoadOwnerAccountFromConfig()
	if ownerAccountAddress != accountAddress && ownerAccountType != accountType {
		return nil, errors.New("PoownAccountNotNodeOwner")
	}

	if err != nil {
		return nil, err
	}
	mainChain := &chaintype.MainChain{}
	blockService := NewBlockService(
		mainChain,
		nas.QueryExecutor,
		query.NewBlockQuery(mainChain),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	lastBlock, err := blockService.GetLastBlock()
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
	if err != nil {
		return nil, err
	}
	nodeSecretPhrase, err := nas.Helpers.LoadNodeSeedFromConfig()
	if err != nil {
		return nil, err
	}
	poownSignature := crypto.NewSignature().SignByNode(messageBytes, nodeSecretPhrase)
	if err != nil {
		return nil, err
	}

	return &model.ProofOfOwnership{
		MessageBytes: messageBytes,
		Signature:    poownSignature,
	}, nil
}

func (*NodeAdminService) LoadOwnerAccountFromConfig() (ownerAccountType uint32, ownerAccountAddress string, err error) {
	err = nil
	if err = commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		err = errors.New("NodeConfigFileNotFound")
		return
	}
	ownerAccountType = viper.GetUint32("ownerAccountType")
	ownerAccountAddress = viper.GetString("ownerAccountAddress")
	return
}

func (*NodeAdminService) LoadNodeSeedFromConfig() (nodeSeed string, err error) {
	err = nil
	if err = commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		err = errors.New("NodeConfigFileNotFound")
		return
	}
	nodeSeed = viper.GetString("nodeSecretPhrase")
	return
}

// ValidateProofOfOwnership validates a proof of ownership message
func (nas *NodeAdminService) ValidateProofOfOwnership(poown *model.ProofOfOwnership, nodePublicKey []byte) error {

	v1 := crypto.NewSignature().VerifyNodeSignature(poown.MessageBytes, poown.Signature, nodePublicKey)
	if !v1 {
		return errors.New("InvalidSignature")
	}

	message, err := nas.Helpers.ParseMessageBytes(poown.MessageBytes)
	if err != nil {
		return err
	}

	// validate height
	mainChain := &chaintype.MainChain{}
	blockService := NewBlockService(
		mainChain,
		nas.QueryExecutor,
		query.NewBlockQuery(mainChain),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	lastBlock, err := blockService.GetLastBlock()
	if err != nil {
		return err
	}

	// FIXME: create an application-wide constant for this
	// Expiration, in number of blocks, of a proof of ownership message
	poownExpiration := uint32(100)
	if lastBlock.Height-message.BlockHeight > poownExpiration {
		return errors.New("ProofOfOwnershipExpired")
	}

	poownBlockRef, err := blockService.GetBlockByHeight(message.BlockHeight)
	if err != nil {
		return err
	}
	poownBlockHashRef, err := util.GetBlockHash(poownBlockRef)
	if err != nil {
		return err
	}
	if bytes.Compare(poownBlockHashRef, message.BlockHash) != 0 {
		return errors.New("InvalidProofOfOwnershipBlockHash")
	}
	return nil
}
