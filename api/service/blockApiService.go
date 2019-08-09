// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType contract.ChainType, ID int64) (*model.Block, error)
		GetBlockByHeight(chainType contract.ChainType, Height uint32) (*model.Block, error)
		GetBlocks(chainType contract.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error)
	}

	// BlockService represents struct of BlockService
	BlockService struct {
		Query query.ExecutorInterface
	}
)

var blockServiceInstance *BlockService

// NewBlockService create a singleton instance of BlockService
func NewBlockService(queryExecutor query.ExecutorInterface) *BlockService {
	if blockServiceInstance == nil {
		blockServiceInstance = &BlockService{Query: queryExecutor}
	}
	return blockServiceInstance
}

// GetBlockByID fetch a single block from Blockchain by providing block ID
func (bs *BlockService) GetBlockByID(chainType contract.ChainType, id int64) (*model.Block, error) {
	var (
		err  error
		bl   []*model.Block
		rows *sql.Rows
	)
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByID(id))
	if err != nil {
		fmt.Printf("GetBlockByID fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	bl = blockQuery.BuildModel(bl, rows)
	if len(bl) == 0 {
		return nil, errors.New("BlockNotFound")
	}

	return bl[0], nil

}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType contract.ChainType, height uint32) (*model.Block, error) {
	var (
		err  error
		bl   []*model.Block
		rows *sql.Rows
	)

	blockQuery := query.NewBlockQuery(chainType)

	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByHeight(height))
	if err != nil {
		fmt.Printf("GetBlockByHeight fails %v\n", err)
		return nil, err
	}
	defer rows.Close()
	bl = blockQuery.BuildModel(bl, rows)
	if len(bl) == 0 {
		return nil, errors.New("BlockNotFound")
	}
	return bl[0], nil
}

// GetBlocks fetches multiple blocks from Blockchain system
func (bs *BlockService) GetBlocks(chainType contract.ChainType, blockSize, height uint32) (*model.GetBlocksResponse, error) {
	var rows *sql.Rows
	var err error
	var blocks []*model.Block
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlocks(height, blockSize))

	if err != nil {
		fmt.Printf("GetBlocks fails %v\n", err)
		return nil, err
	}
	defer rows.Close()
	blocks = blockQuery.BuildModel(blocks, rows)
	blocksResponse := &model.GetBlocksResponse{
		Blocks: blocks,
		Height: height,
		Count:  uint32(len(blocks)),
	}
	return blocksResponse, nil
}
