package service

import (
	"encoding/json"
	"io/ioutil"
	"os"

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
		ParseKeysFile() ([]*model.NodeKey, error)
		GetLastNodeKey(nodeKeys []*model.NodeKey) *model.NodeKey
		GenerateNodeKey(seed string) ([]byte, error)
	}

	// NodeAdminServiceHelpersInterface mockable service methods
	NodeAdminService struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
	}
)

func NewNodeAdminService(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	signature crypto.SignatureInterface,
	blockService BlockServiceInterface,
	nodeKeyFilePath string) *NodeAdminService {
	return &NodeAdminService{
		QueryExecutor: queryExecutor,
		BlockQuery:    blockQuery,
		Signature:     signature,
		BlockService:  blockService,
		FilePath:      nodeKeyFilePath,
	}
}

// generate proof of ownership
func (nas *NodeAdminService) GenerateProofOfOwnership(
	accountAddress string) (*model.ProofOfOwnership, error) {

	// get the node seed (private key)
	nodeKeys, err := nas.ParseKeysFile()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorParseKeysFile: ,"+err.Error())
	}
	nodeKey := nas.GetLastNodeKey(nodeKeys)
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

// ParseNodeKeysFile read the node key file and parses it into an array of NodeKey stuct
func (nas *NodeAdminService) ParseKeysFile() ([]*model.NodeKey, error) {
	file, err := ioutil.ReadFile(nas.FilePath)
	if err != nil && os.IsNotExist(err) {
		return nil, blocker.NewBlocker(blocker.AppErr, "NodeKeysFileNotExist")
	}
	data := make([]*model.NodeKey, 0)
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "InvalidNodeKeysFile")
	}
	return data, nil
}

// GetLastNodeKey retrieves the last node key object from the node_key configuration file
func (*NodeAdminService) GetLastNodeKey(nodeKeys []*model.NodeKey) *model.NodeKey {
	if len(nodeKeys) == 0 {
		return nil
	}
	max := nodeKeys[0]
	for _, nodeKey := range nodeKeys {
		if nodeKey.ID > max.ID {
			max = nodeKey
		}
	}
	return max
}

// GenerateNodeKey generates a new node ket from its seed and store it, together with relative public key into node_keys file
func (nas *NodeAdminService) GenerateNodeKey(seed string) ([]byte, error) {
	publicKey := commonUtils.GetPublicKeyFromSeed(seed)
	nodeKey := &model.NodeKey{
		Seed:      seed,
		PublicKey: publicKey,
	}

	nodeKeys := make([]*model.NodeKey, 0)

	_, err := os.Stat(nas.FilePath)
	if !(err != nil && os.IsNotExist(err)) {
		// if there are previous keys, get the new id
		nodeKeys, err = nas.ParseKeysFile()
		if err != nil {
			return nil, err
		}
		lastNodeKey := nas.GetLastNodeKey(nodeKeys)
		nodeKey.ID = lastNodeKey.ID + 1
	}

	// append generated key to previous keys array
	nodeKeys = append(nodeKeys, nodeKey)
	file, err := json.MarshalIndent(nodeKeys, "", " ")
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorMarshalingNodeKeys: "+err.Error())
	}
	err = ioutil.WriteFile(nas.FilePath, file, 0644)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "ErrorWritingNodeKeysFile: "+err.Error())
	}

	return publicKey, nil
}
