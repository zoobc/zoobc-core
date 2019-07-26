package service

import (
	"bytes"
	"errors"

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

	if ownerAccountAddress == accountAddress {
		nodeMessages := buffer.Bytes()
		proofOfOwnershipSign := nas.SignData(nodeMessages)

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
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	return ownerAccountAddress
}

func (nas *NodeAdminService) SignData(payload []byte) (sign []byte) {
	nodeSecretPhrase := viper.GetString("nodeSecretPhrase")
	sign = nas.Signature.SignBlock(payload, nodeSecretPhrase)
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
func (nas *NodeAdminService) ValidateProofOfOwnership(nodeMessages, signature, publicKey []byte) error {

	buffer := bytes.NewBuffer(nodeMessages)

	blockHeightBytes, err := readNodeMessages(buffer, 5)
	blockHeight := commonUtil.ConvertBytesToUint32([]byte{blockHeightBytes[0], 0, 0, 0})
	if err != nil {
		return err
	}

	lastBlockHash, err := readNodeMessages(buffer, 4)
	if err != nil {
		return err
	}

	err1 := nas.ValidateSignature(signature, nodeMessages, publicKey)
	if err1 != nil {
		return err1
	}

	err2 := nas.ValidateHeight(blockHeight)
	if err2 != nil {
		return err2
	}

	err3 := nas.ValidateBlockHash(blockHeight, lastBlockHash)
	if err3 != nil {
		return err3
	}

	return nil

}
func (nas *NodeAdminService) ValidateSignature(signature, payload, publicKey []byte) error {

	result := ed25519.Verify(publicKey, payload, signature)

	if !result {
		return errors.New("signature not valid")
	}

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
