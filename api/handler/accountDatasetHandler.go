package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountDatasetHandler struct {
	Service service.AccountDatasetServiceInterface
}

func (adh *AccountDatasetHandler) GetAccountDatasets(
	_ context.Context,
	request *model.GetAccountDatasetsRequest,
) (*model.GetAccountDatasetsResponse, error) {

	pagination := request.GetPagination()
	if pagination == nil {
		pagination = &model.Pagination{
			OrderField: "height",
			OrderBy:    model.OrderBy_ASC,
			Page:       0,
			Limit:      constant.MaxAPILimitPerPage,
		}
	}
	if pagination.GetLimit() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	return adh.Service.GetAccountDatasets(request)
}

func (adh *AccountDatasetHandler) GetAccountDataset(
	_ context.Context,
	request *model.GetAccountDatasetRequest,
) (*model.AccountDataset, error) {

	if request.GetRecipientAccountAddress() == nil && request.GetProperty() == "" {
		return nil, status.Error(codes.InvalidArgument, "Request must have Property or RecipientAccountAddress")
	}

	return adh.Service.GetAccountDataset(request)
}
