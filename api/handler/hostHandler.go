package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type HostHandler struct {
	Service        service.HostServiceInterface
	P2pHostService contract.P2PType
}

func (hh *HostHandler) GetHostInfo(ctx context.Context, req *model.Empty) (*model.Host, error) {
	return hh.P2pHostService.GetHostInstance(), nil
}
