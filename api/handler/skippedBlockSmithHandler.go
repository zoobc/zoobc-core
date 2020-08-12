package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SkippedBlockSmithHandler to handle request related to Skipped block smiths from client
type SkippedBlockSmithHandler struct {
	Service service.SkippedBlockSmithServiceInterface
}

func (sbh *SkippedBlockSmithHandler) GetSkippedBlockSmiths(
	ctx context.Context,
	request *model.GetSkippedBlocksmithsRequest,
) (*model.GetSkippedBlocksmithsResponse, error) {
	if request.GetBlockHeightStart() > request.GetBlockHeightEnd() {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"BlockHeightEnd should bigger than BlockHeightStart",
		)
	}
	if request.GetBlockHeightEnd()-request.GetBlockHeightStart() > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}
	return sbh.Service.GetSkippedBlockSmiths(request)
}
