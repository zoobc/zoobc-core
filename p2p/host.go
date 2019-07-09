package p2p

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
)

/*
HostService is
*/
type HostService struct {
	Host            *model.Host
	GrpcServer      *grpc.Server
	ChainTypeNumber int32
}

var hostServiceInstance *HostService

// InitHostService to start peer process
func InitHostService(myAddress string, port uint32, wellknownPeers []string) (*HostService, error) {
	if hostServiceInstance == nil {
		knownPeersResult, err := util.ParseKnownPeers(wellknownPeers)
		if err != nil {
			return nil, err
		}

		host := util.NewHost(myAddress, port, knownPeersResult)
		hostServiceInstance = &HostService{
			Host: host,
		}
	}
	return hostServiceInstance, nil
}
