package handler

import (
	"context"
	"errors"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

// PoownHandler handles requests related to poown
type NodeAdminHandler struct {
	Service service.NodeAdminServiceInterface
}

// GetPoown handles request to get data of a proof of ownership
func (gp *NodeAdminHandler) GetProofOfOwnership(ctx context.Context,
	req *model.GetProofOfOwnershipRequest) (*model.ProofOfOwnership, error) {
	// validate mandatory fields
	if req.AccountAddress == "" {
		return nil, errors.New("AccountAddressRequired")
	}
	if len(req.Signature) == 0 {
		return nil, errors.New("SignatureRequired")
	}

	response, err := gp.Service.GetProofOfOwnership(req.AccountAddress, req.Signature)
	if err != nil {
		return nil, err
	}

	return response, nil
}
