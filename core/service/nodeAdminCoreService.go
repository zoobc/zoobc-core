package service

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountType uint32, accountAddress string, signature []byte)
		ValidateProofOfOwnership(nodeMessages, signature, publicKey []byte)
	}

	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
	}
)

var (
	ownerAccountAddress string
	nodeSecretPhrase    string
	sign                []byte
)

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(accountType uint32,
	accountAddress string, signature []byte) (nodeMessages, proofOfOwnershipSign []byte) {

	lastBlock, lastBlockHash, _ := nas.LookupLastBlock()
	ownerAccountAddress := nas.LookupOwnerAccount()

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(commonUtil.ConvertUint32ToBytes(accountType)[:2])
	buffer.Write([]byte(accountAddress))
	buffer.Write(lastBlockHash)
	buffer.Write(commonUtil.ConvertUint32ToBytes(lastBlock.Height))

	nodeMessages = buffer.Bytes()
	proofOfOwnershipSign = nas.SignData(nodeMessages)
	if ownerAccountAddress == accountAddress {
		return nodeMessages, proofOfOwnershipSign
	}
	return nil, nil
}

// GetLastBlock return the last pushed block

func (nas *NodeAdminService) LookupLastBlock() (*model.Block, []byte, error) {
	rows, err := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetLastBlock())
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, nil, err
	}
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {

		digest := sha3.New512()
		blockByte, _ := util.GetBlockByte(blocks[0], true)
		_, _ = digest.Write(blockByte)
		hash := digest.Sum([]byte{})

		return blocks[0], hash, nil
	}
	return nil, nil, errors.New("BlockNotFound")

}
func (nas *NodeAdminService) LookupOwnerAccount() string {
	if err := commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	} else {
		ownerAccountAddress = viper.GetString("ownerAccountAddress")
	}
	return ownerAccountAddress
}

func ed25519GetPrivateKeyFromSeed(seed string) []byte {
	// Convert seed (secret phrase) to byte array
	seedBuffer := []byte(seed)
	// Compute SHA3-256 hash of seed (secret phrase)
	seedHash := sha3.Sum256(seedBuffer)
	// Generate a private key from the hash of the seed
	return ed25519.NewKeyFromSeed(seedHash[:])
}

func (nas *NodeAdminService) SignData(payload []byte) []byte {
	if err := commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	} else {
		nodeSecretPhrase = viper.GetString("nodeSecretPhrase")
	}
	nodePrivateKey := ed25519GetPrivateKeyFromSeed(nodeSecretPhrase)
	sign = ed25519.Sign(nodePrivateKey, payload)

	return sign
}

func readNodeMessages(buf *bytes.Buffer, nBytes int) ([]byte, error) {
	nextBytes := buf.Next(nBytes)
	if len(nextBytes) < nBytes {
		return nil, errors.New("EndOfBufferReached")
	}
	return nextBytes, nil
}

// validate proof of ownership
func (nas *NodeAdminService) ValidateProofOfOwnership(nodeMessages, signature []byte, accountAddress string) error {

	buffer := bytes.NewBuffer(nodeMessages)

	accountId, err := readNodeMessages(buffer, 46)

	if err != nil {
		return err
	}
	fmt.Printf("accountId %v\n", accountId)
	lastBlockHash, err := readNodeMessages(buffer, 64)
	if err != nil {
		return err
	}
	blockHeightBytes, err := readNodeMessages(buffer, 4)
	if err != nil {
		return err
	}

	blockHeight := commonUtil.ConvertBytesToUint32([]byte{blockHeightBytes[0], 0, 0, 0})

	err1 := nas.ValidateSignature(signature, nodeMessages, accountAddress)
	fmt.Printf("err1 %v\n", err1)
	if err1 != nil {
		return err1
	}

	err2 := nas.ValidateHeight(blockHeight)
	fmt.Printf("err2 %v\n", err2)
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
func (nas *NodeAdminService) ValidateSignature(signature, payload []byte, accountAddress string) error {

	nodePublicKey, _ := commonUtil.GetPublicKeyFromAddress(accountAddress)
	fmt.Printf("public key %v\n", nodePublicKey)
	fmt.Printf("signature %v\n", signature)
	// result := ed25519.Verify(nodePublicKey, payload, signature)
	// if !result {
	// 	return errors.New("signature not valid")
	// }

	return nil
}
func (nas *NodeAdminService) ValidateHeight(blockHeight uint32) error {
	rows, _ := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetLastBlock())
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)

	if blockHeight > blocks[0].Height {
		return errors.New("block is older")
	}

	return nil
}
func (nas *NodeAdminService) ValidateBlockHash(blockHeight uint32, lastBlockHash []byte) error {

	rows, _ := nas.QueryExecutor.ExecuteSelect(nas.BlockQuery.GetBlockByHeight(blockHeight))
	var blocks []*model.Block
	blocks = nas.BlockQuery.BuildModel(blocks, rows)

	digest := sha3.New512()
	blockByte, _ := util.GetBlockByte(blocks[0], true)
	_, _ = digest.Write(blockByte)
	hash := digest.Sum([]byte{})

	if !bytes.Equal(hash, lastBlockHash) {
		return errors.New("hash didn't same")
	}

	return nil
}
