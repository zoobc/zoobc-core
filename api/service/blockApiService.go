// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType chaintype.ChainType, ID int64) (*model.BlockExtendedInfo, error)
		GetBlockByHeight(chainType chaintype.ChainType, Height uint32) (*model.BlockExtendedInfo, error)
		GetBlocks(chainType chaintype.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error)
	}

	// BlockService represents struct of BlockService
	BlockService struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
)

var blockServiceInstance *BlockService

// NewBlockService create a singleton instance of BlockService
func NewBlockService(queryExecutor query.ExecutorInterface, blockCoreServices map[int32]coreService.BlockServiceInterface) *BlockService {
	if blockServiceInstance == nil {
		blockServiceInstance = &BlockService{Query: queryExecutor}
	}
	blockServiceInstance.BlockCoreServices = blockCoreServices
	return blockServiceInstance
}

// GetBlockByID fetch a single block from Blockchain by providing block ID
func (bs *BlockService) GetBlockByID(chainType chaintype.ChainType, id int64) (*model.BlockExtendedInfo, error) {
	var (
		err  error
		bl   []*model.Block
		rows *sql.Rows
	)
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByID(id), false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	bl, err = blockQuery.BuildModel(bl, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed build block into model")
	}
	if len(bl) == 0 {
		return nil, status.Error(codes.NotFound, "block not found")
	}

	// Get block extended info
	blExt, err := bs.BlockCoreServices[0].GetBlockExtendedInfo(bl[0], false)
	if err != nil {
		return nil, status.Error(codes.Internal, "fail to get block extended information")
	}

	return blExt, nil

}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType chaintype.ChainType, height uint32) (*model.BlockExtendedInfo, error) {
	var (
		err  error
		bl   []*model.Block
		rows *sql.Rows
	)

	blockQuery := query.NewBlockQuery(chainType)

	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByHeight(height), false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	bl, err = blockQuery.BuildModel(bl, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed build block into model")
	}

	if len(bl) == 0 {
		return nil, status.Error(codes.NotFound, "block not found")
	}
	return bs.BlockCoreServices[0].GetBlockExtendedInfo(bl[0], true)
}

// GetBlocks fetches multiple blocks from Blockchain system
func (bs *BlockService) GetBlocks(chainType chaintype.ChainType, blockSize, height uint32) (*model.GetBlocksResponse, error) {
	var (
		rows   *sql.Rows
		err    error
		blocks []*model.Block
	)
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlocks(height, blockSize), false)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	blocks, err = blockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed build block into model")
	}

	blocksExt := make([]*model.BlockExtendedInfo, 0)
	for _, block := range blocks {
		blExt, err := bs.BlockCoreServices[0].GetBlockExtendedInfo(block, false)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		blocksExt = append(blocksExt, blExt)
	}
	blocksResponse := &model.GetBlocksResponse{
		Blocks: blocksExt,
		Height: height,
		Count:  uint32(len(blocks)),
	}
	return blocksResponse, nil
}
