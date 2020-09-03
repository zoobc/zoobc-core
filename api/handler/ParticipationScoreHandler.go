package handler

import (
	"context"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ParticipationScoreHandler handles requests related to ParticipationScore
type ParticipationScoreHandler struct {
	Service service.ParticipationScoreInterface
}

// GetParticipationScores handles request to get data of a single Transaction
func (psh *ParticipationScoreHandler) GetParticipationScores(
	ctx context.Context,
	req *model.GetParticipationScoresRequest,
) (*model.GetParticipationScoresResponse, error) {
	var (
		response       *model.GetParticipationScoresResponse
		err            error
		totalRequested = req.GetToHeight() - req.GetFromHeight()
	)

	if req.GetFromHeight() > req.GetToHeight() {
		return nil, status.Errorf(codes.InvalidArgument, "ToHeight can not be less than FromHeight.")
	}

	if totalRequested+1 > constant.MaxAPILimitPerPage {
		return nil, status.Errorf(codes.OutOfRange, "Limit exceeded, max. %d", constant.MaxAPILimitPerPage)
	}

	response, err = psh.Service.GetParticipationScores(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetLatestParticipationScoreByNodeID return latest participation score of a node
func (psh *ParticipationScoreHandler) GetLatestParticipationScoreByNodeID(
	ctx context.Context,
	req *model.GetLatestParticipationScoreByNodeIDRequest,
) (*model.ParticipationScore, error) {
	var (
		response *model.ParticipationScore
		err      error
	)
	response, err = psh.Service.GetLatestParticipationScoreByNodeID(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
