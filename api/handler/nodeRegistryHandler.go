package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
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
