package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	// ParticipationScoreInterface represents interface for ParticipationScoreService
	ParticipationScoreInterface interface {
		GetParticipationScores(params *model.GetParticipationScoresRequest) (*model.GetParticipationScoresResponse, error)
	}

	// ParticipationScoreService represents struct of ParticipationScoreService
	ParticipationScoreService struct {
		ParticipationScoreService coreService.ParticipationScoreServiceInterface
	}
)

var participationScoreServiceInstance *ParticipationScoreService

// NewParticipationScoreService creates a singleton instance of ParticipationScoreService
func NewParticipationScoreService(
	participationScoreService coreService.ParticipationScoreServiceInterface) *ParticipationScoreService {
	if participationScoreServiceInstance == nil {
		participationScoreServiceInstance = &ParticipationScoreService{
			ParticipationScoreService: participationScoreService,
		}
	}
	return participationScoreServiceInstance
}

// GetParticipationScores fetches participation scores for given height range
func (pss *ParticipationScoreService) GetParticipationScores(
	params *model.GetParticipationScoresRequest,
) (*model.GetParticipationScoresResponse, error) {
	participationScores, err := pss.ParticipationScoreService.GetParticipationScoreByBlockHeightRange(params.FromHeight, params.ToHeight)
	return &model.GetParticipationScoresResponse{ParticipationScores: participationScores}, err
}
