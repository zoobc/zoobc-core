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
	"golang.org/x/crypto/sha3"
)

type (
	// NodeAdminServiceInterface represents interface for NodeAdminService
	NodeAdminServiceInterface interface {
		GenerateProofOfOwnership(accountType string, accountAddress uint32, signature []byte)
		ValidateProofOfOwnership()
	}

	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
	}
)

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(accountType uint32, accountAddress string, signature []byte) (nodeMessages []byte, proofOfOwnershipSign []byte) {

	lastBlock, lastBlockHash, _ := nas.LookupLastBlock()

	ownerAccount, _ := nas.LookupOwnerAccount()

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(commonUtil.ConvertUint32ToBytes(accountType)[:2])
	buffer.Write([]byte(accountAddress))
	buffer.Write(lastBlockHash)
	buffer.Write(commonUtil.ConvertUint32ToBytes(lastBlock.Height))

	if ownerAccount.AccountType == accountType && ownerAccount.Address == accountAddress {
		nodeMessages = buffer.Bytes()
		proofOfOwnershipSign = nas.SignData(nodeMessages)

		return nodeMessages, proofOfOwnershipSign
	}
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
		return nil, &model.Block{
			ID: -1,
		}, err
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
	return nil, &model.Block{
		ID: -1,
	}, errors.New("BlockNotFound")

}

func (nas *NodeAdminService) LookupOwnerAccount() (*model.Account, error) {
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
}

func (nas *NodeAdminService) SignData(payload []byte) (sign []byte) {
	nodeSecretPhrase := viper.GetString("nodeSecretPhrase")
	sign = nas.Signature.SignBlock(payload, nodeSecretPhrase)
	return sign
}

// validate proof of ownership
func (nas *NodeAdminService) ValidateProofOfOwnership(mpTx *model.MempoolTransaction) error {
	nas.ValidateSignature()
	nas.ValidateHeight()
	nas.LookupBlock()
}
func (nas *NodeAdminService) ValidateSignature() error {

}
func (nas *NodeAdminService) ValidateHeight() error {

}
func (nas *NodeAdminService) LookupBlock() error {

}
