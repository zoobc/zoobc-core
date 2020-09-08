package handler

import (
	"context"
	"io"
	"time"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NodeRegistryHandler handles requests related to node registry
type NodeRegistryHandler struct {
	Service       service.NodeRegistryServiceInterface
	NodePublicKey []byte
}

func (nrh NodeRegistryHandler) GetNodeRegistrations(
	ctx context.Context,
	req *model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	var (
		response   *model.GetNodeRegistrationsResponse
		pagination = req.GetPagination()
		err        error
	)

	if pagination == nil {
		pagination = &model.Pagination{
			OrderField: "registration_height",
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if pagination.GetLimit() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	req.Pagination = pagination

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

func (nrh NodeRegistryHandler) GetMyNodePublicKey(
	ctx context.Context,
	req *model.Empty,
) (*model.GetMyNodePublicKeyResponse, error) {
	return &model.GetMyNodePublicKeyResponse{
		NodePublicKey: nrh.NodePublicKey,
	}, nil
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

func (nrh *NodeRegistryHandler) GetPendingNodeRegistrations(
	stream rpcService.NodeRegistrationService_GetPendingNodeRegistrationsServer,
) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if in == nil {
		return nil
	}
	for {
		pendingNodes, err := nrh.Service.GetPendingNodeRegistrations(in)
		if err != nil {
			return err
		}
		err = stream.Send(pendingNodes)
		if err != nil {
			return err // close connection if sending response to client result in error
		}
		time.Sleep(5 * time.Second)
	}
}
