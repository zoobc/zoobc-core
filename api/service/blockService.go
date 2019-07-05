// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"fmt"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/schema/model"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType contract.ChainType, ID int64) (*model.Block, error)
		GetBlockByHeight(chainType contract.ChainType, BlockHeight int32) (*model.Block, error)
		GetBlocks(chainType contract.ChainType, BlockSize int32, BlockHeight int32) (*model.GetBlocksResponse, error)
	}

	// BlockService represents struct of BlockService
	BlockService struct {
		Query *query.Executor
	}
)

var blockServiceInstance *BlockService

// NewBlockService create a singleton instance of BlockService
func NewBlockService(queryExecutor *query.Executor) *BlockService {
	if blockServiceInstance == nil {
		blockServiceInstance = &BlockService{Query: queryExecutor}
	}
	return blockServiceInstance
}

// ResetBlockService resets the singleton back to nil, used in test case teardown
func ResetBlockService() {
	blockServiceInstance = nil
}

// GetBlockByID fetch a single block from Blockchain by providing block ID
func (bs *BlockService) GetBlockByID(chainType contract.ChainType, ID int64) (*model.Block, error) {
	var err error
	rows, err := bs.Query.ExecuteSelect(query.NewBlockQuery().GetBlockByID(chainType, ID))
	if err != nil {
		fmt.Printf("GetBlockByID fails %v\n", err)
		return nil, err
	}

	var bl model.Block
	if rows.Next() {
		err = rows.Scan(&bl.ID, &bl.Timestamp, &bl.TotalAmount, &bl.TotalFee, &bl.PayloadLength, &bl.PayloadHash, &bl.PreviousBlockHash, &bl.PreviousBlockID, &bl.Height, &bl.GeneratorPublicKey, &bl.GenerationSignature, &bl.BlockSignature, &bl.Version)
		if err != nil {
			fmt.Printf("GetBlockByID fails scan %v\n", err)
			return nil, err
		}
	}

	return &bl, nil

}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType contract.ChainType, BlockHeight int32) (*model.Block, error) {
	var err error
	rows, err := bs.Query.ExecuteSelect(query.NewBlockQuery().GetBlockByHeight(chainType, BlockHeight))
	if err != nil {
		fmt.Printf("GetBlockByHeight fails %v\n", err)
		return nil, err
	}

	var bl model.Block
	if rows.Next() {
		err = rows.Scan(&bl.ID, &bl.Timestamp, &bl.TotalAmount, &bl.TotalFee, &bl.PayloadLength, &bl.PayloadHash, &bl.PreviousBlockHash, &bl.PreviousBlockID, &bl.Height, &bl.GeneratorPublicKey, &bl.GenerationSignature, &bl.BlockSignature, &bl.Version)
		if err != nil {
			fmt.Printf("GetBlockByHeight fails scan %v\n", err)
			return nil, err
		}
	}
	return &bl, nil
}

// GetBlocks fetches multiple blocks from Blockchain system
func (bs *BlockService) GetBlocks(chainType contract.ChainType, BlockSize int32, BlockHeight int32) (*model.GetBlocksResponse, error) {
	var rows *sql.Rows
	var err error
	blocks := []*model.Block{}
	rows, err = bs.Query.ExecuteSelect(query.NewBlockQuery().GetBlocks(chainType, BlockHeight))
	if err != nil {
		fmt.Printf("GetBlocks fails %v\n", err)
		return nil, err
	}

	for rows.Next() {
		var bl model.Block
		err = rows.Scan(&bl.ID, &bl.Timestamp, &bl.TotalAmount, &bl.TotalFee, &bl.PayloadLength, &bl.PayloadHash, &bl.PreviousBlockHash, &bl.PreviousBlockID, &bl.Height, &bl.GeneratorPublicKey, &bl.GenerationSignature, &bl.BlockSignature, &bl.Version)
		if err != nil {
			fmt.Printf("GetBlocks fails scan %v\n", err)
			return nil, err
		}
		blocks = append(blocks, &bl)
	}

	blocksResponse := &model.GetBlocksResponse{
		Blocks:      blocks,
		BlockHeight: BlockHeight,
		BlockSize:   int32(len(blocks)),
	}
	return blocksResponse, nil
}
