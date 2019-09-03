package blockchainsync

import (
	"bytes"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
	utils "github.com/zoobc/zoobc-core/core/util"
)

type (
	ForkingProcessInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block) error
		PopOffToBlock(commonBlock *model.Block) []*model.Block
		popLastBlock() (*model.Block, error)
		SetLastBlock(block *model.Block) error
		HasBlock(id int64) bool
		LoadTransactions(block *model.Block) *model.Block
		scheduleScan(height uint32, validate bool)
		getMinRollbackHeight() (uint32, error)
		SetIsScanning(isScanning bool)
	}
)

//The main function to process the forked blocks
func (bss *Service) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {
	var forkBlocksID []int64
	for _, block := range forkBlocks {
		forkBlocksID = append(forkBlocksID, block.ID)
	}

	lastBlockBeforeProcess, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	myPoppedOffBlocks, err := bss.PopOffToBlock(commonBlock)
	if err != nil {
		return err
	}

	pushedForkBlocks := 0

	lastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	if lastBlock.ID == commonBlock.ID {
		// rebuilding the chain
		for _, block := range forkBlocks {
			lastBlock, err := bss.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			lastBlockHash, err := utils.GetBlockHash(lastBlock)
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				err := bss.BlockService.PushBlock(lastBlock, block, false)
				if err != nil {
					// bss.P2pService.Blacklist(feederPeer)
					log.Warnf("\n\nPushBlock err %v\n\n", err)
					break
				}
				pushedForkBlocks++
			}
		}
	}

	currentLastBlock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	currentCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	cumulativeDifficultyOriginalBefore, _ := new(big.Int).SetString(beforeApplyCumulativeDifficulty, 10)

	// if after applying the fork blocks the cumulative difficulty is still less than current one
	// only take the transactions to be processed, but later will get back to our own fork
	if pushedForkBlocks > 0 && currentCumulativeDifficulty.Cmp(cumulativeDifficultyOriginalBefore) < 0 {
		peerPoppedOffBlocks, err := bss.PopOffToBlock(commonBlock)
		if err != nil {
			return err
		}
		pushedForkBlocks = 0
		for _, block := range peerPoppedOffBlocks {
			bss.ProcessLater(block.Transactions)
		}
	}

	// if no fork blocks succesfully applied, go back to our fork
	// other wise, just take the transactions of our popped blocks to be processed later
	if pushedForkBlocks == 0 {
		log.Println("Did not accept any blocks from peer, pushing back my blocks")
		for _, block := range myPoppedOffBlocks {
			lastBlock, err := bss.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			errPushBlock := bss.BlockService.PushBlock(lastBlock, block, false)
			if errPushBlock != nil {
				return blocker.NewBlocker(blocker.BlockErr, "Popped off block no longer acceptable")
			}
		}
	} else {
		for _, block := range myPoppedOffBlocks {
			bss.ProcessLater(block.Transactions)
		}
	}

	return nil
}

// PopOffToBlock will remove the block in current Chain until commonBlock is reached
func (bss *Service) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	// blockchain lock has been implemented by the Download Blockchain, so no additional lock is needed
	var err error

	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	minRollbackHeight, err := bss.getMinRollbackHeight()
	if err != nil {
		return []*model.Block{}, err
	}
	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork pocess
		return []*model.Block{}, nil
	}

	if !bss.HasBlock(commonBlock.GetID()) {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, "the common block is not found")
	}

	poppedBlocks := []*model.Block{}
	block, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockNotFoundErr,
			"Last block is not found",
		)
	}
	block = bss.LoadTransactions(block)

	genesisBlockID := bss.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID && block.Height-1 > 0 {
		poppedBlocks = append(poppedBlocks, block)

		block, err = bss.BlockService.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}

		// if block.Height == 1 {
		// 	break
		// }
		block = bss.LoadTransactions(block)
	}

	derivedTables := query.GetDerivedQuery(bss.ChainType)
	errTx := bss.QueryExecutor.BeginTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}

	for _, dTable := range derivedTables {
		if commonBlock.Height == 0 {
			break
		}
		queries, _ := dTable.Rollback(commonBlock.Height)
		for _, query := range queries {
			errTx = bss.QueryExecutor.ExecuteTransaction(query)
			if errTx != nil {
				_ = bss.QueryExecutor.RollbackTx()
				return []*model.Block{}, errTx
			}
		}
	}
	errTx = bss.QueryExecutor.CommitTx()
	if errTx != nil {
		return []*model.Block{}, errTx
	}
	//	TODO:
	//	NEED TO IMPLEMENT DERIVED TABLES ROLLBACK
	// if err == nil {
	// 	err = service.service.RollbackDerivedTables(commonBlock.GetHeight())
	// 	// _ = DbTransactionalService(chaintype).ClearCache() //need to implement ClearCache
	// 	err = service.CommitTransaction()
	// }

	blockIds := []int64{}
	for _, block := range poppedBlocks {
		blockIds = append(blockIds, block.ID)
	}
	return poppedBlocks, nil
}

func (bss *Service) HasBlock(id int64) bool {
	block, _ := bss.BlockService.GetBlockByID(id)
	if block == nil {
		return false
	}
	return true
}

func (bss *Service) LoadTransactions(block *model.Block) *model.Block {
	if block.Transactions == nil {
		txs, _ := bss.BlockService.GetTransactionsByBlockID(block.ID)
		block.Transactions = txs
	}
	return block
}

func (bss *Service) scheduleScan(height uint32, validate bool) {
	bss.ScheduleScan(height, validate)
}

func (bss *Service) getMinRollbackHeight() (uint32, error) {
	lastblock, err := bss.BlockService.GetLastBlock()
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

func (bss *Service) ProcessLater(transaction []*model.Transaction) {}

func (bss *Service) ScheduleScan(height uint32, validate bool) {}
