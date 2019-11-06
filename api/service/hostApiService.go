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
		Query                   query.ExecutorInterface
		P2pService              p2p.Peer2PeerServiceInterface
		BlockServices           map[int32]coreService.BlockServiceInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
	}
)

var hostServiceInstance *HostService

// NewHostService create a singleton instance of PeerExplorer
func NewHostService(queryExecutor query.ExecutorInterface, p2pService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface) HostServiceInterface {
	if hostServiceInstance == nil {
		hostServiceInstance = &HostService{
			Query:                   queryExecutor,
			P2pService:              p2pService,
			BlockServices:           blockServices,
			NodeRegistrationService: nodeRegistrationService,
		}
	}
	return hostServiceInstance
}

func (hs *HostService) GetHostInfo() (*model.HostInfo, error) {
	var chainStatuses []*model.ChainStatus
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

	scrambledNodes := hs.NodeRegistrationService.GetScrambledNodes()

	return &model.HostInfo{
		Host:           hs.P2pService.GetHostInfo(),
		ChainStatuses:  chainStatuses,
		ScrambledNodes: scrambledNodes.AddressNodes,
	}, nil
}

func (hs *HostService) GetHostPeers() (*model.GetHostPeersResponse, error) {
	return &model.GetHostPeersResponse{
		ResolvedPeers:   hs.P2pService.GetResolvedPeers(),
		UnresolvedPeers: hs.P2pService.GetUnresolvedPeers(),
	}, nil
}
