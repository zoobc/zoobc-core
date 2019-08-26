package util

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func IsBlockIDExist(blockIds []int64, expectedBlockID int64) bool {
	for _, blockID := range blockIds {
		if blockID == expectedBlockID {
			return true
		}
	}
	return false
}

// GetLastBlock TODO: this should be used by services instead of blockService.GetLastBlock
func GetLastBlock(queryExecutor query.ExecutorInterface, blockQuery query.BlockQueryInterface) (*model.Block, error) {
	qry := blockQuery.GetLastBlock()
	rows, err := queryExecutor.ExecuteSelect(qry)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var (
		blocks []*model.Block
	)
	blocks = blockQuery.BuildModel(blocks, rows)
	if len(blocks) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "LastBlockNotFound")
	}
	return blocks[0], nil
}

// GetBlockByHeight TODO: this should be used by services instead of blockService.GetLastBlock
func GetBlockByHeight(height uint32, queryExecutor query.ExecutorInterface, blockQuery query.BlockQueryInterface) (*model.Block, error) {
	qry := blockQuery.GetBlockByHeight(height)
	rows, err := queryExecutor.ExecuteSelect(qry)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks = blockQuery.BuildModel(blocks, rows)
	if len(blocks) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "BlockNotFound")
	}
	return blocks[0], nil
}
