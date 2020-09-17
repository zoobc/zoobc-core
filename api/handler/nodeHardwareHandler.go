package handler

import (
	"context"
	"io"
	"time"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/api/service"
	rpcService "github.com/zoobc/zoobc-core/common/service"
)

type NodeHardwareHandler struct {
	Service service.NodeHardwareServiceInterface
}

func (nhh *NodeHardwareHandler) GetNodeHardware(
	stream rpcService.NodeHardwareService_GetNodeHardwareServer,
) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if in == nil {
		return nil
	}
	for {
		nodeHardware, err := nhh.Service.GetNodeHardware(in)
		if err != nil {
			return err
		}
		err = stream.Send(nodeHardware)
		if err != nil {
			return err // close connection if sending response to client result in error
		}
		time.Sleep(5 * time.Second)
	}
}

func (nhh *NodeHardwareHandler) GetNodeTime(context.Context, *model.Empty) (*model.GetNodeTimeResponse, error) {
	t := time.Now().UTC().Unix()
	return &model.GetNodeTimeResponse{
		NodeTime: t,
	}, nil
}
