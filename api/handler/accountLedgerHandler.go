package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// AccountLedgerHandler struct fields of AccountLedgerService
	AccountLedgerHandler struct {
		Service service.AccountLedgerServiceInterface
	}
)

// GetAccountLedgers api handler of account ledger service that return account ledgers collection
func (al *AccountLedgerHandler) GetAccountLedgers(
	ctx context.Context,
	request *model.GetAccountLedgersRequest,
) (*model.GetAccountLedgersResponse, error) {
	return al.Service.GetAccountLedgers(request)
}
