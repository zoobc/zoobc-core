// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"fmt"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// BlockServiceInterface represents interface for BlockService
	BlockServiceInterface interface {
		GetBlockByID(chainType contract.ChainType, ID int64) (*model.Block, error)
		GetBlockByHeight(chainType contract.ChainType, BlockHeight uint32) (*model.Block, error)
		GetBlocks(chainType contract.ChainType, BlockSize uint32, BlockHeight uint32) (*model.GetBlocksResponse, error)
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

// GetBlockByID fetch a single block from Blockchain by providing block ID
func (bs *BlockService) GetBlockByID(chainType contract.ChainType, id int64) (*model.Block, error) {
	var (
		err  error
		bl   model.Block
		rows *sql.Rows
	)

	rows, err = bs.Query.ExecuteSelect(query.NewBlockQuery(chainType).GetBlockByID(id))
	if err != nil {
		fmt.Printf("GetBlockByID fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&bl.ID,
			&bl.PreviousBlockHash,
			&bl.Height,
			&bl.Timestamp,
			&bl.BlockSeed,
			&bl.BlockSignature,
			&bl.CumulativeDifficulty,
			&bl.SmithScale,
			&bl.PayloadLength,
			&bl.PayloadHash,
			&bl.BlocksmithID,
			&bl.TotalAmount,
			&bl.TotalFee,
			&bl.TotalCoinBase,
			&bl.Version,
		)
		if err != nil {
			fmt.Printf("GetBlockByID fails scan %v\n", err)
			return nil, err
		}
	}

	return &bl, nil

}

// GetBlockByHeight fetches a single block from Blockchain by providing block size
func (bs *BlockService) GetBlockByHeight(chainType contract.ChainType, blockHeight uint32) (*model.Block, error) {
	var (
		err  error
		bl   model.Block
		rows *sql.Rows
	)

	rows, err = bs.Query.ExecuteSelect(query.NewBlockQuery(chainType).GetBlockByHeight(blockHeight))
	if err != nil {
		fmt.Printf("GetBlockByHeight fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&bl.ID,
			&bl.PreviousBlockHash,
			&bl.Height,
			&bl.Timestamp,
			&bl.BlockSeed,
			&bl.BlockSignature,
			&bl.CumulativeDifficulty,
			&bl.SmithScale,
			&bl.PayloadLength,
			&bl.PayloadHash,
			&bl.BlocksmithID,
			&bl.TotalAmount,
			&bl.TotalFee,
			&bl.TotalCoinBase,
			&bl.Version,
		)
		if err != nil {
			fmt.Printf("GetBlockByHeight fails scan %v\n", err)
			return nil, err
		}
	}
	return &bl, nil
}

// GetBlocks fetches multiple blocks from Blockchain system
func (bs *BlockService) GetBlocks(chainType contract.ChainType, blockSize, blockHeight uint32) (*model.GetBlocksResponse, error) {
	var rows *sql.Rows
	var err error
	blocks := []*model.Block{}
	rows, err = bs.Query.ExecuteSelect(query.NewBlockQuery(chainType).GetBlocks(blockHeight))

	if err != nil {
		fmt.Printf("GetBlocks fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bl model.Block
		err = rows.Scan(
			&bl.ID,
			&bl.PreviousBlockHash,
			&bl.Height,
			&bl.Timestamp,
			&bl.BlockSeed,
			&bl.BlockSignature,
			&bl.CumulativeDifficulty,
			&bl.SmithScale,
			&bl.PayloadLength,
			&bl.PayloadHash,
			&bl.BlocksmithID,
			&bl.TotalAmount,
			&bl.TotalFee,
			&bl.TotalCoinBase,
			&bl.Version,
		)
		if err != nil {
			fmt.Printf("GetBlocks fails scan %v\n", err)
			return nil, err
		}
		blocks = append(blocks, &bl)
	}

	blocksResponse := &model.GetBlocksResponse{
		Blocks:      blocks,
		BlockHeight: blockHeight,
		BlockSize:   uint32(len(blocks)),
	}
	return blocksResponse, nil
}
