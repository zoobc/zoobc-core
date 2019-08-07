package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	NodeAdminAPIServiceInterface interface {
		GetProofOfOwnership(accountType uint32, accountAddress string) (*model.ProofOfOwnership, error)
	}

	// TransactionService represents struct of TransactionService
	NodeAdmin struct {
		NodeAdminService service.NodeAdminServiceInterface
	}
)

// GetProof of ownership
func (nas *NodeAdmin) GetProofOfOwnership(accountType uint32, accountAddress string) (*model.ProofOfOwnership, error) {

	poown, err := nas.NodeAdminService.GenerateProofOfOwnership(accountType, accountAddress)
	if err != nil {
		return nil, err
	}
	return poown, nil
}
