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
	coreService "github.com/zoobc/zoobc-core/core/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	NodeAddressInfoAPIServiceInterface interface {
		GetNodeAddressesInfo(request *model.GetNodeAddressesInfoRequest) (*model.GetNodeAddressesInfoResponse, error)
	}

	NodeAddressInfoAPIService struct {
		NodeAddressInfoService coreService.NodeAddressInfoServiceInterface
	}
)

func NewNodeAddressInfoAPIService(
	nodeAddressInfoService coreService.NodeAddressInfoServiceInterface,
) *NodeAddressInfoAPIService {
	return &NodeAddressInfoAPIService{
		NodeAddressInfoService: nodeAddressInfoService,
	}
}

// GetNodeAddressesInfo client api to get one, many or all address info from db
// note: if request.NodeIDs is an empty array, the whole address info table will be returned
// note2: only one address per (registered) node is returned. if a node has two addresses for the same nodeID (pending and confirmed),
// confirmed address is chosen over the pending one
func (nhs *NodeAddressInfoAPIService) GetNodeAddressesInfo(request *model.GetNodeAddressesInfoRequest,
) (*model.GetNodeAddressesInfoResponse, error) {
	nais, err := nhs.NodeAddressInfoService.GetAddressInfoByNodeIDs(
		request.NodeIDs,
		[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressConfirmed, model.NodeAddressStatus_NodeAddressPending},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// remove duplicates
	var naisMap = make(map[int64]*model.NodeAddressInfo)
	for _, nai := range nais {
		if _, ok := naisMap[nai.NodeID]; !ok {
			naisMap[nai.NodeID] = nai
			// always prefer confirmed addresses over pending
		} else if nai.Status == model.NodeAddressStatus_NodeAddressConfirmed &&
			naisMap[nai.NodeID].Status == model.NodeAddressStatus_NodeAddressPending {
			naisMap[nai.NodeID] = nai
		}
	}
	// rebuild the array
	var res []*model.NodeAddressInfo
	for _, naiMap := range naisMap {
		res = append(res, naiMap)
	}

	return &model.GetNodeAddressesInfoResponse{
		NodeAddressesInfo: res,
	}, nil
}
