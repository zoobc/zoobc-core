package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NodeRegistryHandler handles requests related to node registry
type NodeRegistryHandler struct {
	Service service.NodeRegistryServiceInterface
}

func (nrh NodeRegistryHandler) GetNodeRegistrations(
	ctx context.Context,
	req *model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	var (
		response *model.GetNodeRegistrationsResponse
		err      error
	)
	response, err = nrh.Service.GetNodeRegistrations(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (nrh NodeRegistryHandler) GetNodeRegistration(
	ctx context.Context,
	req *model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	var (
		response *model.GetNodeRegistrationResponse
		err      error
	)
	response, err = nrh.Service.GetNodeRegistration(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (nrh NodeRegistryHandler) GetNodeRegistrationsByNodePublicKeys(
	ctx context.Context,
	req *model.GetNodeRegistrationsByNodePublicKeysRequest,
) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error) {
	var (
		response *model.GetNodeRegistrationsByNodePublicKeysResponse
		err      error
	)
	if len(req.NodePublicKeys) == 0 {
		return nil, status.Error(codes.InvalidArgument, "At least 1 node public key is required")
	}
	response, err = nrh.Service.GetNodeRegistrationsByNodePublicKeys(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
