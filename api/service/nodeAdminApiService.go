package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	NodeAdminServiceInterface interface {
		GetProofOfOwnership() (*model.ProofOfOwnership, error)
		GenerateNodeKey(seed string) ([]byte, error)
	}

	// TransactionService represents struct of TransactionService
	NodeAdminService struct {
		Query                query.ExecutorInterface
		NodeAdminCoreService coreService.NodeAdminServiceInterface
		ownerAccountAddress  string
	}
)

var nodeAdminServiceInstance *NodeAdminService

// NewBlockService create a singleton instance of BlockService
func NewNodeAdminService(queryExecutor query.ExecutorInterface, ownerAccountAddress, nodeKeyFilePath string) *NodeAdminService {
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
func (nas *NodeAdminService) GetProofOfOwnership() (*model.ProofOfOwnership, error) {
	poown, err := nas.NodeAdminCoreService.GenerateProofOfOwnership(nas.ownerAccountAddress)
	if err != nil {
		return nil, err
	}
	return poown, nil
}

// GenerateNodeKey api to request the node to generate a new key pairs
func (nas *NodeAdminService) GenerateNodeKey(seed string) ([]byte, error) {
	// generate a node key pairs, store the private and return the public key
	nodePublicKey, err := nas.NodeAdminCoreService.GenerateNodeKey(seed)
	if err != nil {
		return nil, err
	}
	return nodePublicKey, nil
}
