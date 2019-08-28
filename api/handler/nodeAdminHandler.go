package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

// PoownHandler handles requests related to poown
type NodeAdminHandler struct {
	Service service.NodeAdminServiceInterface
}

// GetPoown handles request to get data of a proof of ownership
func (gp *NodeAdminHandler) GetProofOfOwnership(ctx context.Context,
	req *model.GetProofOfOwnershipRequest) (*model.ProofOfOwnership, error) {
	response, err := gp.Service.GetProofOfOwnership()
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetPoown handles request to get data of a proof of ownership
func (gp *NodeAdminHandler) GenerateNodeKey(ctx context.Context,
	req *model.GenerateNodeKeyRequest) (*model.GenerateNodeKeyResponse, error) {
	nodePublicKey, err := gp.Service.GenerateNodeKey(util.GetSecureRandomSeed())
	if err != nil {
		return nil, err
	}
	response := &model.GenerateNodeKeyResponse{
		NodePublicKey: nodePublicKey,
	}

	return response, nil
}
