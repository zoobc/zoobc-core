package service

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	NodeAdminServiceInterface interface {
		GetProofOfOwnership(accountAddress string, signature []byte) (*model.ProofOfOwnership, error)
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

		nodeAdminCoreService := coreService.NewNodeAdminService(queryExecutor,
			query.NewBlockQuery(mainchain), crypto.NewSignature(), &coreService.NodeAdminService{}, blockService)
		nodeAdminServiceInstance = &NodeAdminService{
			Query:                queryExecutor,
			NodeAdminCoreService: nodeAdminCoreService,
		}
	}
	return nodeAdminServiceInstance
}

// GetProof of ownership
func (nas *NodeAdminService) GetProofOfOwnership(accountAddress string, signature []byte) (*model.ProofOfOwnership, error) {
	// validate signature: message (the account address..) must be signed by accountAddress
	if !crypto.NewSignature().VerifySignature([]byte(accountAddress), signature, accountAddress) {
		return nil, errors.New("PoownAccountNotNodeOwner")
	}
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	if ownerAccountAddress != accountAddress {
		return nil, errors.New("PoownAccountNotNodeOwner")
	}

	poown, err := nas.NodeAdminCoreService.GenerateProofOfOwnership(accountAddress)
	if err != nil {
		return nil, err
	}
	return poown, nil
}
