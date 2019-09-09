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
func (fp *BlockPopper) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	// blockchain lock has been implemented by the Download Blockchain, so no additional lock is needed
	var err error

	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return []*model.Block{}, err
	}
	minRollbackHeight, err := getMinRollbackHeight(lastBlock.Height)
	if err != nil {
		return []*model.Block{}, err
	}
	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork pocess
		log.Warn("the node blockchain is experiencing hardfork, please manually delete the database to ")
		return []*model.Block{}, nil
	}

	_, err = fp.BlockService.GetBlockByID(commonBlock.ID)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %s"))
	}

	poppedBlocks := []*model.Block{}
	block := lastBlock
	txs, _ := fp.BlockService.GetTransactionsByBlockID(block.ID)
	block.Transactions = txs

	genesisBlockID := fp.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID && block.Height-1 > 0 {
		poppedBlocks = append(poppedBlocks, block)

		block, err = fp.BlockService.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		txs, _ := fp.BlockService.GetTransactionsByBlockID(block.ID)
		block.Transactions = txs
	}

	derivedTables := query.GetDerivedQuery(fp.ChainType)
	errTx := fp.QueryExecutor.BeginTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}

	for _, dTable := range derivedTables {
		if commonBlock.Height == 0 {
			break
		}
		queries := dTable.Rollback(commonBlock.Height)
		errTx = fp.QueryExecutor.ExecuteTransactions(queries)
	}
	errTx = fp.QueryExecutor.CommitTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}

	blockIds := []int64{}
	for _, block := range poppedBlocks {
		blockIds = append(blockIds, block.ID)
	}
	return poppedBlocks, nil
}
