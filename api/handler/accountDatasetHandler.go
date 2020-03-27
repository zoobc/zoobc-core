package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type AccountDatasetHandler struct {
	Service service.AccountDatasetServiceInterface
}

func (adh *AccountDatasetHandler) GetAccountDatasets(
	_ context.Context,
	request *model.GetAccountDatasetsRequest,
) (*model.GetAccountDatasetsResponse, error) {

	return adh.Service.GetAccountDatasets(request)
}

func (adh *AccountDatasetHandler) GetAccountDataset(
	_ context.Context,
	request *model.GetAccountDatasetRequest,
) (*model.AccountDataset, error) {

	return adh.Service.GetAccountDataset(request)
}
