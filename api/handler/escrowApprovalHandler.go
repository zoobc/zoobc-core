package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

// EscrowApprovalHandler handles requests for escrow approval
type EscrowApprovalHandler struct {
	Service service.EscrowApprovalServiceInterface
}

// PostApprovalEscrowTransaction handle request for escrow approval
func (eah *EscrowApprovalHandler) PostApprovalEscrowTransaction(
	_ context.Context,
	req *model.PostEscrowApprovalRequest,
) (*model.Transaction, error) {
	chainType := chaintype.GetChainType(0)
	transaction, err := eah.Service.PostApprovalEscrowTransaction(chainType, req)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}
