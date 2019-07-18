package handler

import (
	"context"
	"fmt"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type HostHandler struct {
	Service        service.HostServiceInterface
	P2pHostService contract.P2PType
}

func (hh *HostHandler) GetHostInfo(ctx context.Context, req *model.Empty) (*model.Host, error) {
	fmt.Printf("\n\n\n%v\n", hh.P2pHostService)
	return hh.P2pHostService.GetHostInstance(), nil
}
