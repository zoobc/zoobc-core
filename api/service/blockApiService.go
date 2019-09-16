// Package service serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"
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
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	bl = blockQuery.BuildModel(bl, rows)
	if len(bl) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "BlockNotFound")
	}

	// get node registration related to current block's BlockSmith
	block := bl[0]
	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	qry, args := nodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(block.BlocksmithPublicKey, block.Height)
	rows, err = bs.Query.ExecuteSelect(qry, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nr = nodeRegistrationQuery.BuildModel(nr, rows)
	if len(nr) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "VersionedNodeRegistrationNotFound")
	}
	nodeRegistration := nr[0]
	block.BlocksmithAccountAddress = nodeRegistration.AccountAddress
	//FIXME: return mocked data, until underlying logic is implemented
	block.TotalReward = block.TotalFee + 50 ^ 10*8
	// ???
	block.TotalReceipts = 99
	// ???
	block.ReceiptValue = 99
	// once we have the receipt for this block we should be able to calculate this using util.CalculateParticipationScore
	block.PopChange = -200

	return block, nil

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
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	bl = blockQuery.BuildModel(bl, rows)
	if len(bl) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "BlockNotFound")
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
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
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
