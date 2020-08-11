package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// SkippedBlockSmithServiceInterface represents interface for BlockService
	SkippedBlockSmithServiceInterface interface {
		GetSkippedBlockSmiths(*model.GetSkippedBlocksmithsRequest) (*model.GetSkippedBlocksmithsResponse, error)
	}

	// SkippedBlockSmithService represents struct of SkippedBlockSmithService
	SkippedBlockSmithService struct {
		QueryExecutor          query.ExecutorInterface
		SkippedBlocksmithQuery *query.SkippedBlocksmithQuery
	}
)

func NewSkippedBlockSmithService(
	skippedBlocksmithQuery *query.SkippedBlocksmithQuery,
	queryExecutor query.ExecutorInterface,
) SkippedBlockSmithServiceInterface {
	return &SkippedBlockSmithService{
		SkippedBlocksmithQuery: skippedBlocksmithQuery,
		QueryExecutor:          queryExecutor,
	}
}

func (sbs *SkippedBlockSmithService) GetSkippedBlockSmiths(
	req *model.GetSkippedBlocksmithsRequest,
) (*model.GetSkippedBlocksmithsResponse, error) {
	var (
		err                error
		rowCount           *sql.Row
		count              uint64
		caseQ              = query.NewCaseQuery()
		skippedBlockSmiths []*model.SkippedBlocksmith
	)
	caseQ.Select(sbs.SkippedBlocksmithQuery.TableName, sbs.SkippedBlocksmithQuery.Fields...)
	caseQ.Where(caseQ.Between("block_height", req.BlockHeightStart, req.BlockHeightEnd))
	caseQ.OrderBy("block_height", model.OrderBy_ASC)

	selectQ, args := caseQ.Build()
	rowCount, _ = sbs.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(selectQ), false, args...)
	if err = rowCount.Scan(&count); err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, "Record not found")
	}

	rows, err := sbs.QueryExecutor.ExecuteSelect(selectQ, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	skippedBlockSmiths, err = sbs.SkippedBlocksmithQuery.BuildModel([]*model.SkippedBlocksmith{}, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &model.GetSkippedBlocksmithsResponse{
		Total:              count,
		SkippedBlocksmiths: skippedBlockSmiths,
	}, nil
}
