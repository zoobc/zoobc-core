package service

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	ParticipationScoreServiceInterface interface {
		GetLatestParticipationScoreByNodeID(nodeID int64) (*model.ParticipationScore, error)
		GetParticipationScoreByBlockHeightRange(fromBlockHeight, toBlockHeight uint32) ([]*model.ParticipationScore, error)
	}

	ParticipationScoreService struct {
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
)

func NewParticipationScoreService(
	participationScoreQuery query.ParticipationScoreQueryInterface,
	queryExecutor query.ExecutorInterface,
) *ParticipationScoreService {
	return &ParticipationScoreService{
		ParticipationScoreQuery: participationScoreQuery,
		QueryExecutor:           queryExecutor,
	}
}

// GetParticipationScoreByNodeID get latest participation score of a node
func (pss *ParticipationScoreService) GetLatestParticipationScoreByNodeID(nodeID int64) (*model.ParticipationScore, error) {
	var (
		participationScore model.ParticipationScore
	)
	participationScoreQ, args := pss.ParticipationScoreQuery.GetParticipationScoreByNodeID(nodeID)
	row, err := pss.QueryExecutor.ExecuteSelectRow(participationScoreQ, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = pss.ParticipationScoreQuery.Scan(&participationScore, row)
	// if there aren't participation scores for this address/node, return 0
	if err != nil {
		return nil, nil
	}
	return &participationScore, nil
}

// GetParticipationScoreByBlockHeightRange get list of participation score change in the range Heights
func (pss *ParticipationScoreService) GetParticipationScoreByBlockHeightRange(fromBlockHeight,
	toBlockHeight uint32) ([]*model.ParticipationScore, error) {
	var (
		participationScores []*model.ParticipationScore
	)
	participationScoreQ, args := pss.ParticipationScoreQuery.GetParticipationScoresByBlockHeightRange(fromBlockHeight, toBlockHeight)
	rows, err := pss.QueryExecutor.ExecuteSelect(participationScoreQ, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	participationScores, err = pss.ParticipationScoreQuery.BuildModel(participationScores, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.ParserErr, err.Error())
	}
	// if there aren't participation scores for this address/node, return 0
	if len(participationScores) == 0 {
		return nil, nil
	}
	return participationScores, nil
}
