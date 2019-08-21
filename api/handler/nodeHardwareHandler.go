package handler

import (
	"github.com/zoobc/zoobc-core/api/service"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"io"
)

type NodeHardwareHandler struct {
	Service service.NodeHardwareServiceInterface
}

func (nhh *NodeHardwareHandler) GetNodeHardware(
	stream rpcService.NodeHardwareService_GetNodeHardwareServer,
) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if in == nil {
			return nil
		}
		nodeHardware, err := nhh.Service.GetNodeHardware(in)
		if err != nil {
			return err
		}
		err = stream.Send(nodeHardware)
		if err != nil {
			return err
		}
	}
}
