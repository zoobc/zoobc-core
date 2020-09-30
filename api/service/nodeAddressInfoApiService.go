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
