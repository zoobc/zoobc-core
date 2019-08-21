package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type (
	HostServiceInterface interface {
		GetHostInfo() (*model.HostInfo, error)
		GetHostPeers() (*model.GetHostPeersResponse, error)
	}

	HostService struct {
		Query         query.ExecutorInterface
		BlockServices map[int32]coreService.BlockServiceInterface
		P2pService    p2p.ServiceInterface
	}
)

var hostServiceInstance *HostService

// NewHostService create a singleton instance of HostService
func NewHostService(queryExecutor query.ExecutorInterface, p2pService p2p.ServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface) HostServiceInterface {
	if hostServiceInstance == nil {
		hostServiceInstance = &HostService{
			Query:         queryExecutor,
			P2pService:    p2pService,
			BlockServices: blockServices,
		}
	}
	return hostServiceInstance
}

func (hs *HostService) GetHostInfo() (*model.HostInfo, error) {
	chainStatuses := []*model.ChainStatus{}
	for chainType, blockService := range hs.BlockServices {
		lastBlock, err := blockService.GetLastBlock()
		if lastBlock == nil || err != nil {
			continue
		}
		chainStatuses = append(chainStatuses, &model.ChainStatus{
			ChainType: chainType,
			Height:    lastBlock.Height,
			LastBlock: lastBlock,
		})
	}
	return &model.HostInfo{
		Host:          hs.P2pService.GetHostInstance(),
		ChainStatuses: chainStatuses,
	}, nil
}

func (hs *HostService) GetHostPeers() (*model.GetHostPeersResponse, error) {
	host := hs.P2pService.GetHostInstance()
	return &model.GetHostPeersResponse{
		ConnectedPeers:    host.GetResolvedPeers(),
		DisconnectedPeers: host.GetUnresolvedPeers(),
	}, nil
}
