package service

import (
	"bytes"
	"errors"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GetMessageSize() uint32
		GetBytesFromMessage(poown *model.ProofOfOwnershipMessage) ([]byte, error)
		ParseMessageBytes(messageBytes []byte) (*model.ProofOfOwnershipMessage, error)
		GenerateProofOfOwnership(accountType uint32, accountAddress string, signature []byte) (*model.ProofOfOwnership, error)
		ValidateProofOfOwnershipRequest(accountType uint32, accountAddress string, signature []byte) bool
		ValidateProofOfOwnership(nodeMessages, signature, publicKey []byte)
	}

	// NodeAdminServiceHelpersInterface mockable service methods
	NodeAdminServiceHelpersInterface interface {
		LoadOwnerAccountFromConfig() (ownerAccountType uint32, ownerAccountAddress string, err error)
		LoadNodeSeedFromConfig() (nodeSeed string, err error)
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
	accountType := 2
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
	if uint32(len(messageBytes)) != nas.GetMessageSize() {
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
	messageBytes, err := nas.GetBytesFromMessage(poownMessage)
	if err != nil {
		return nil, err
	}
	poownSignature, err := nas.SignPoownMessageBytes(messageBytes)
	if err != nil {
		return nil, err
	}
	log.Println(messageBytes)
	log.Println(poownSignature)

	return &model.ProofOfOwnership{
		MessageBytes: messageBytes,
		Signature:    poownSignature,
	}, nil
}

// SignPoownMessageBytes sign poown message bytes with the node private key
func (nas *NodeAdminService) SignPoownMessageBytes(messageBytes []byte) ([]byte, error) {
	nodeSecretPhrase, err := nas.Helpers.LoadNodeSeedFromConfig()
	if err != nil {
		return nil, err
	}
	poownSignature := crypto.NewSignature().SignByNode(messageBytes, nodeSecretPhrase)
	return poownSignature, nil
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

func readNodeMessages(buf *bytes.Buffer, nBytes int) ([]byte, error) {
	nextBytes := buf.Next(nBytes)
	if len(nextBytes) < nBytes {
		return nil, errors.New("EndOfBufferReached")
	}
	return nextBytes, nil
}

// validate proof of ownership
func (nas *NodeAdminService) ValidateProofOfOwnership(nodeMessages, signature, nodePublicKey []byte) error {

	v1 := crypto.NewSignature().VerifyNodeSignature(nodeMessages, signature, nodePublicKey)
	if !v1 {
		return errors.New("InvalidSignature")
	}

	buffer := bytes.NewBuffer(nodeMessages)

	accountID, err := readNodeMessages(buffer, 46)
	if err != nil {
		return err
	}
	if accountID == nil {
		fmt.Println(err)
	}

	lastBlockHash, err := readNodeMessages(buffer, 64)
	if err != nil {
		return err
	}
	blockHeightBytes, err := readNodeMessages(buffer, 4)
	if err != nil {
		return err
	}

	blockHeight := commonUtil.ConvertBytesToUint32([]byte{blockHeightBytes[0], 0, 0, 0})
	fmt.Printf("block height %v\n", blockHeight)
	err2 := nas.ValidateHeight(blockHeight)
	if err2 != nil {
		return err2
	}

	err3 := nas.ValidateBlockHash(blockHeight, lastBlockHash)
	fmt.Printf("err3 %v\n", err3)
	if err3 != nil {
		return err3
	}

	return nil

}

func (nas *NodeAdminService) ValidateHeight(blockHeight uint32) error {
	rows, _ := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetLastBlock())
	fmt.Printf("lastblock %v\n", rows)
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)

	if blockHeight > blocks[0].Height {
		return errors.New("block is older")
	}

	return nil
}
func (nas *NodeAdminService) ValidateBlockHash(blockHeight uint32, lastBlockHash []byte) error {

	rows, _ := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetLastBlock())
	fmt.Printf("rows : %v\n", rows)
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)
	fmt.Printf("blocks : %v\n", blocks)
	digest := sha3.New512()
	blockByte, _ := util.GetBlockByte(blocks[0], true)
	_, _ = digest.Write(blockByte)
	hash := digest.Sum([]byte{})

	if !bytes.Equal(hash, lastBlockHash) {
		return errors.New("hash didn't same")
	}

	return nil
}
