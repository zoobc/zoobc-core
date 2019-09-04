package blockchainsync

import (
	"bytes"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	utils "github.com/zoobc/zoobc-core/core/util"

	"github.com/zoobc/zoobc-core/core/service"
)

type (
	ForkingProcessorInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error
		PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error)
		HasBlock(id int64) bool
		LoadTransactions(block *model.Block) *model.Block
		scheduleScan(height uint32, validate bool)
		getMinRollbackHeight() (uint32, error)
	}
	ForkingProcessor struct {
		ChainType     chaintype.ChainType
		BlockService  service.BlockServiceInterface
		QueryExecutor query.ExecutorInterface
	}
)

//The main function to process the forked blocks
func (fp *ForkingProcessor) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {
	var forkBlocksID []int64
	for _, block := range forkBlocks {
		forkBlocksID = append(forkBlocksID, block.ID)
	}

	lastBlockBeforeProcess, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	myPoppedOffBlocks, err := fp.PopOffToBlock(commonBlock)
	if err != nil {
		return err
	}

	pushedForkBlocks := 0

	lastBlock, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	if lastBlock.ID == commonBlock.ID {
		// rebuilding the chain
		for _, block := range forkBlocks {
			lastBlock, err := fp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			lastBlockHash, err := utils.GetBlockHash(lastBlock)
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				err := fp.BlockService.PushBlock(lastBlock, block, false)
				if err != nil {
					// fp.P2pService.Blacklist(feederPeer)
					log.Warnf("\n\nPushBlock err %v\n\n", err)
					break
				}
				pushedForkBlocks++
			}
		}
	}

	currentLastBlock, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	currentCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	cumulativeDifficultyOriginalBefore, _ := new(big.Int).SetString(beforeApplyCumulativeDifficulty, 10)

	// if after applying the fork blocks the cumulative difficulty is still less than current one
	// only take the transactions to be processed, but later will get back to our own fork
	if pushedForkBlocks > 0 && currentCumulativeDifficulty.Cmp(cumulativeDifficultyOriginalBefore) < 0 {
		peerPoppedOffBlocks, err := fp.PopOffToBlock(commonBlock)
		if err != nil {
			return err
		}
		pushedForkBlocks = 0
		for _, block := range peerPoppedOffBlocks {
			fp.ProcessLater(block.Transactions)
		}
	}

	// if no fork blocks succesfully applied, go back to our fork
	// other wise, just take the transactions of our popped blocks to be processed later
	if pushedForkBlocks == 0 {
		log.Println("Did not accept any blocks from peer, pushing back my blocks")
		for _, block := range myPoppedOffBlocks {
			lastBlock, err := fp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			errPushBlock := fp.BlockService.PushBlock(lastBlock, block, false)
			if errPushBlock != nil {
				return blocker.NewBlocker(blocker.BlockErr, "Popped off block no longer acceptable")
			}
		}
	} else {
		for _, block := range myPoppedOffBlocks {
			fp.ProcessLater(block.Transactions)
		}
	}

	return nil

}

// PopOffToBlock will remove the block in current Chain until commonBlock is reached
func (fp *ForkingProcessor) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	// blockchain lock has been implemented by the Download Blockchain, so no additional lock is needed
	var err error

	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	minRollbackHeight, err := fp.getMinRollbackHeight()
	if err != nil {
		return []*model.Block{}, err
	}
	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork pocess
		log.Warn("the node blockchain is experiencing hardfork, please manually delete the database to ")
		return []*model.Block{}, nil
	}

	if !fp.HasBlock(commonBlock.GetID()) {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, "the common block is not found")
	}

	poppedBlocks := []*model.Block{}
	block, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockNotFoundErr,
			"Last block is not found",
		)
	}
	block = fp.LoadTransactions(block)

	genesisBlockID := fp.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID && block.Height-1 > 0 {
		poppedBlocks = append(poppedBlocks, block)

		block, err = fp.BlockService.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		block = fp.LoadTransactions(block)
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
		queries, _ := dTable.Rollback(commonBlock.Height)
		for _, query := range queries {
			errTx = fp.QueryExecutor.ExecuteTransaction(query)
			if errTx != nil {
				_ = fp.QueryExecutor.RollbackTx()
				return []*model.Block{}, errTx
			}
		}
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

func (fp *ForkingProcessor) HasBlock(id int64) bool {
	block, _ := fp.BlockService.GetBlockByID(id)
	if block == nil {
		return false
	}
	return true
}

func (fp *ForkingProcessor) LoadTransactions(block *model.Block) *model.Block {
	if block.Transactions == nil {
		txs, _ := fp.BlockService.GetTransactionsByBlockID(block.ID)
		block.Transactions = txs
	}
	return block
}

func (fp *ForkingProcessor) scheduleScan(height uint32, validate bool) {
	fp.ScheduleScan(height, validate)
}

func (fp *ForkingProcessor) getMinRollbackHeight() (uint32, error) {
	lastblock, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return 0, err
	}
	currentHeight := lastblock.Height
	if currentHeight < constant.MinRollbackBlocks {
		return 0, nil
	}
	return currentHeight - constant.MinRollbackBlocks, nil
}

func (bs *Service) SetIsScanning(isScanning bool) {
	bs.isScanningBlockchain = isScanning
}

func (fp *ForkingProcessor) ProcessLater(transaction []*model.Transaction) {}

func (fp *ForkingProcessor) ScheduleScan(height uint32, validate bool) {}
