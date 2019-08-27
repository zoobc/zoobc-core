package service

import (
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	NodeAdminServiceInterface interface {
		GetProofOfOwnership(accountAddress string, timestamp int64, signature []byte, timeout int64) (*model.ProofOfOwnership, error)
		GenerateNodeKey(accountAddress string, timestamp int64, signature []byte, timeout int64, seed string) ([]byte, error)
	}

	// TransactionService represents struct of TransactionService
	NodeAdminService struct {
		Query                query.ExecutorInterface
		NodeAdminCoreService coreService.NodeAdminServiceInterface
	}
)

var nodeAdminServiceInstance *NodeAdminService

// NewBlockService create a singleton instance of BlockService
func NewNodeAdminService(queryExecutor query.ExecutorInterface) *NodeAdminService {
	if nodeAdminServiceInstance == nil {
		mainchain := &chaintype.MainChain{}
		blockService := coreService.NewBlockService(
			mainchain,
			queryExecutor,
			query.NewBlockQuery(mainchain),
			nil,
			query.NewTransactionQuery(mainchain),
			crypto.NewSignature(),
			nil,
			nil,
			nil,
			nil,
		)

		configPath := viper.GetString("configPath")
		nodeKeyFile := viper.GetString("nodeKeyFile")
		nodeKeyFilePath := filepath.Join(configPath, nodeKeyFile)

		nodeAdminCoreService := coreService.NewNodeAdminService(queryExecutor,
			query.NewBlockQuery(mainchain), crypto.NewSignature(), blockService, nodeKeyFilePath)
		nodeAdminServiceInstance = &NodeAdminService{
			Query:                queryExecutor,
			NodeAdminCoreService: nodeAdminCoreService,
		}
	}
	return nodeAdminServiceInstance
}

// GetProofOfOwnership GetProof of ownership
func (nas *NodeAdminService) GetProofOfOwnership(accountAddress string, timestamp int64,
	signature []byte, timeout int64) (*model.ProofOfOwnership, error) {
	// validate timestamp
	if timestamp > time.Now().Unix()+timeout || timestamp < time.Now().Unix()-timeout {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "TimeStampExpired")
	}
	// validate signature: message (the account address bytes+timestamp bytes..) must be signed by accountAddress
	message := append([]byte(accountAddress), util.ConvertUint64ToBytes(uint64(timestamp))...)
	if !crypto.NewSignature().VerifySignature(message, signature, accountAddress) {
		return nil, blocker.NewBlocker(blocker.AuthErr, "PoownAccountNotNodeOwner")
	}
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	if ownerAccountAddress != accountAddress {
		return nil, blocker.NewBlocker(blocker.AuthErr, "PoownAccountNotNodeOwner")
	}

	poown, err := nas.NodeAdminCoreService.GenerateProofOfOwnership(accountAddress)
	if err != nil {
		return nil, err
	}
	return poown, nil
}

// GenerateNodeKey api to request the node to generate a new key pairs
func (nas *NodeAdminService) GenerateNodeKey(accountAddress string, timestamp int64,
	signature []byte, timeout int64, seed string) ([]byte, error) {

	// validate timestamp
	if timestamp > time.Now().Unix()+timeout || timestamp < time.Now().Unix()-timeout {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "TimeStampExpired")
	}
	// validate signature: message (the account address bytes+timestamp bytes..) must be signed by accountAddress
	message := append([]byte(accountAddress), util.ConvertUint64ToBytes(uint64(timestamp))...)
	if !crypto.NewSignature().VerifySignature(message, signature, accountAddress) {
		return nil, blocker.NewBlocker(blocker.AuthErr, "PoownAccountNotNodeOwner")
	}
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	if ownerAccountAddress != accountAddress {
		return nil, blocker.NewBlocker(blocker.AuthErr, "PoownAccountNotNodeOwner")
	}

	// generate a node key pairs, store the private and return the public key
	nodePublicKey, err := nas.NodeAdminCoreService.GenerateNodeKey(seed)
	if err != nil {
		return nil, err
	}
	return nodePublicKey, nil
}
