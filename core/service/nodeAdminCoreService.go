package service

import (
	"github.com/zoobc/zoobc-core/common/blocker"
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

	// get the node seed (private key)
	nodeKeyConfig := util.NewNodeKeyConfig()
	nodeKeys, _ := nodeKeyConfig.ParseKeysFile()
	nodeKey := nodeKeyConfig.GetLastNodeKey(nodeKeys)
	if nodeKey == nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "MissingNodePrivateKey")
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
		AccountAddress: accountAddress,
		BlockHash:      lastBlockHash,
		BlockHeight:    lastBlock.Height,
	}

	messageBytes := commonUtils.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(messageBytes, nodeKey.Seed)
	return &model.ProofOfOwnership{
		MessageBytes: messageBytes,
		Signature:    poownSignature,
	}, nil
}
