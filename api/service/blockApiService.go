// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType chaintype.ChainType, ID int64) (*model.Block, error)
		GetBlockByHeight(chainType chaintype.ChainType, Height uint32) (*model.Block, error)
		GetBlocks(chainType chaintype.ChainType, Count uint32, Height uint32) (*model.GetBlocksResponse, error)
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
func (bs *BlockService) GetBlockByID(chainType chaintype.ChainType, id int64) (*model.Block, error) {
	var (
		err  error
		bl   []*model.Block
		nr   []*model.NodeRegistration
		rows *sql.Rows
	)
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByID(id), false)
	if err != nil {
		fmt.Printf("GetBlockByID fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	bl = blockQuery.BuildModel(bl, rows)
	if len(bl) == 0 {
		return nil, errors.New("BlockNotFound")
	}

	// get node registration related to current block's BlockSmith
	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	rows, err = bs.Query.ExecuteSelect(nodeRegistrationQuery.GetNodeRegistrationByNodePublicKeyVersioned(bl.BlockSmithPublicKey, bl.Height), false)
	if err != nil {
		fmt.Printf("GetBlockByID fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	nr = nodeRegistrationQuery.BuildModel(nr, rows)
	if len(nr) == 0 {
		return nil, errors.New("BlockNotFound")
	}
	bl.BlocksmithAccountAddress = nr.AccountAddress

	return bl[0], nil

}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType chaintype.ChainType, height uint32) (*model.Block, error) {
	var (
		err  error
		bl   []*model.Block
		rows *sql.Rows
	)

	blockQuery := query.NewBlockQuery(chainType)

	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlockByHeight(height), false)
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
func (bs *BlockService) GetBlocks(chainType chaintype.ChainType, blockSize, height uint32) (*model.GetBlocksResponse, error) {
	var rows *sql.Rows
	var err error
	var blocks []*model.Block
	blockQuery := query.NewBlockQuery(chainType)
	rows, err = bs.Query.ExecuteSelect(blockQuery.GetBlocks(height, blockSize), false)

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
