package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeAddressInfoHandler struct {
	Service service.NodeAddressInfoAPIServiceInterface
}

// GetNodeAddressInfo handles request to get one or a list of address info from db
func (naih *NodeAddressInfoHandler) GetNodeAddressInfo(ctx context.Context,
	req *model.GetNodeAddressesInfoRequest) (*model.GetNodeAddressesInfoResponse, error) {
	response, err := naih.Service.GetNodeAddressesInfo(req)
	if err != nil {
		return nil, err
	}
	if len(req.NodeIDs) == 0 {
		return nil, status.Error(codes.Internal, "nodeIDs required")
	}

	return response, nil
}

// GetNodeAddressesInfo dummy method to implement interface (shared with p2p rpc api service)
func (naih *NodeAddressInfoHandler) GetNodeAddressesInfo(ctx context.Context,
	req *model.GetNodeAddressesInfoRequest) ([]*model.NodeAddressInfo, error) {
	return nil, nil
}
