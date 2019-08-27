package handler

import (
	"context"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/blocker"
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
	// validate mandatory fields
	if req.AccountAddress == "" {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "AccountAddressRequired")
	}
	if req.Timestamp == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "TimestampRequired")
	}
	if len(req.Signature) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "SignatureRequired")
	}

	timeout := viper.GetInt64("apiReqTimeoutSec")
	response, err := gp.Service.GetProofOfOwnership(req.AccountAddress, req.Timestamp, req.Signature, timeout)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetPoown handles request to get data of a proof of ownership
func (gp *NodeAdminHandler) GenerateNodeKey(ctx context.Context,
	req *model.GenerateNodeKeyRequest) (*model.GenerateNodeKeyResponse, error) {
	// validate mandatory fields
	if req.AccountAddress == "" {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "AccountAddressRequired")
	}
	if req.Timestamp == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "TimestampRequired")
	}
	if len(req.Signature) == 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "SignatureRequired")
	}

	timeout := viper.GetInt64("apiReqTimeoutSec")
	nodePublicKey, err := gp.Service.GenerateNodeKey(req.AccountAddress, req.Timestamp, req.Signature, timeout, util.GetSecureRandomSeed())
	if err != nil {
		return nil, err
	}
	response := &model.GenerateNodeKeyResponse{
		NodePublicKey: nodePublicKey,
	}

	return response, nil
}
