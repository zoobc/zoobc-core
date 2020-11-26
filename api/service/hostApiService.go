package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		ScrambleNodeService     coreService.ScrambleNodeServiceInterface
		BlockStateStorages      map[int32]storage.CacheStorageInterface
	}
)

var hostServiceInstance *HostService

// NewHostService create a singleton instance of PeerExplorer
func NewHostService(
	queryExecutor query.ExecutorInterface,
	p2pService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	scrambleNodeService coreService.ScrambleNodeServiceInterface,
	blockStateStorages map[int32]storage.CacheStorageInterface,
) HostServiceInterface {
	if hostServiceInstance == nil {
		hostServiceInstance = &HostService{
			Query:                   queryExecutor,
			P2pService:              p2pService,
			BlockServices:           blockServices,
			NodeRegistrationService: nodeRegistrationService,
			ScrambleNodeService:     scrambleNodeService,
			BlockStateStorages:      blockStateStorages,
		}
	}
	return hostServiceInstance
}

func (hs *HostService) GetHostInfo() (*model.HostInfo, error) {
	var (
		chainStatuses = make([]*model.ChainStatus, len(hs.BlockServices))
		err           error
	)
	for chainType := range hs.BlockServices {
		var lastBlock model.Block
		err = hs.BlockStateStorages[chainType].GetItem(nil, &lastBlock)
		if err != nil {
			continue
		}
		chainStatuses[chainType] = &model.ChainStatus{
			ChainType: chainType,
			Height:    lastBlock.Height,
			LastBlock: &lastBlock,
		}
	}

	// check existing main chaintype
	if len(chainStatuses) == 0 || chainStatuses[(&chaintype.MainChain{}).GetTypeInt()] == nil {
		return nil, status.Error(codes.InvalidArgument, "mainLastBlockIsNil")
	}
	scrambledNodes, err := hs.ScrambleNodeService.GetScrambleNodesByHeight(chainStatuses[0].GetHeight())
	if err != nil {
		return nil, err
	}

	return &model.HostInfo{
		Host:                 hs.P2pService.GetHostInfo(),
		ChainStatuses:        chainStatuses,
		ScrambledNodes:       scrambledNodes.AddressNodes,
		ScrambledNodesHeight: scrambledNodes.BlockHeight,
		PriorityPeers:        hs.P2pService.GetPriorityPeers(),
	}, nil
}

func (hs *HostService) GetHostPeers() (*model.GetHostPeersResponse, error) {
	return &model.GetHostPeersResponse{
		ResolvedPeers:   hs.P2pService.GetResolvedPeers(),
		UnresolvedPeers: hs.P2pService.GetUnresolvedPeers(),
	}, nil
}
