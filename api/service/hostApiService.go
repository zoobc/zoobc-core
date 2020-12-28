// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
