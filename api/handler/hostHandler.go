package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type HostHandler struct {
	Service service.HostServiceInterface
}

func (hh *HostHandler) GetHostInfo(ctx context.Context, req *model.Empty) (*model.HostInfo, error) {
	return hh.Service.GetHostInfo()
}

func (hh *HostHandler) GetHostPeers(context.Context, *model.Empty) (*model.GetHostPeersResponse, error) {
	return hh.Service.GetHostPeers()
}
