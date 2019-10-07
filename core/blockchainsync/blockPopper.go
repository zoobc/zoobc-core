package blockchainsync

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/chaintype"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
)

type BlockPopper struct {
	BlockService  service.BlockServiceInterface
	QueryExecutor query.ExecutorInterface
	ChainType     chaintype.ChainType
}

// PopOffToBlock will remove the block in current Chain until commonBlock is reached
func (bp *BlockPopper) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return []*model.Block{}, err
	}
	minRollbackHeight := getMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork pocess
		log.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return []*model.Block{}, nil
	}

	_, err = bp.BlockService.GetBlockByID(commonBlock.ID)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	poppedBlocks := []*model.Block{}
	block := lastBlock
	txs, _ := bp.BlockService.GetTransactionsByBlockID(block.ID)
	block.Transactions = txs

	genesisBlockID := bp.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID {
		if block.Height == 0 {
			log.Fatal("Genesis is found to be corrupted or incorrect")
		}
		poppedBlocks = append(poppedBlocks, block)

		block, err = bp.BlockService.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		txs, _ := bp.BlockService.GetTransactionsByBlockID(block.ID)
		block.Transactions = txs
	}

	derivedQueries := query.GetDerivedQuery(bp.ChainType)
	errTx := bp.QueryExecutor.BeginTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}

	for _, dQuery := range derivedQueries {
		queries := dQuery.Rollback(commonBlock.Height)
		errTx = bp.QueryExecutor.ExecuteTransactions(queries)
		if errTx != nil {
			return []*model.Block{}, errTx
		}
	}
	errTx = bp.QueryExecutor.CommitTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}
	return poppedBlocks, nil
}
