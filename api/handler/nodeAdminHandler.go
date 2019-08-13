package handler

import (
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

// PoownHandler handles requests related to poown
type GeneratePoownHandler struct {
	Service service.NodeAdminAPIServiceInterface
}

// GetPoown handles request to get data of a proof of ownership
func (gp *GeneratePoownHandler) GetProofOfOwnership(accountAddress string) (*model.ProofOfOwnership, error) {
	var response *model.ProofOfOwnership
	var err error

	response, err = gp.Service.GetProofOfOwnership(accountAddress)
	if err != nil {
		return nil, err
	}

	return response, nil
}
