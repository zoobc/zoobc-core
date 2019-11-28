package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/common/model"
)

// SmithingHandler handles requests related to smithing statuses
type SmithingHandler struct {
	SmithingStatus *model.SmithingStatuses
}

func (sh *SmithingHandler) GetSmithingStatus(
	ctx context.Context,
	req *model.Empty,
) (*model.GetSmithingStatusResponse, error) {
	return &model.GetSmithingStatusResponse{
		SmithingStatus: *sh.SmithingStatus,
	}, nil
}
