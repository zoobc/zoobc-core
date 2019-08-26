package blockchainsync

import (
	"bytes"
	"fmt"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	utils "github.com/zoobc/zoobc-core/core/util"
)

type (
	ForkingProcess interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block) error
		PopOffTo(commonBlock *model.Block) []*model.Block
		popLastBlock() (*model.Block, error)
		SetLastBlock(block *model.Block) error
		HasBlock(id int64) bool
		LoadTransactions(blockID int64)
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
	myPoppedOffBlocks, err := bss.PopOffTo(commonBlock)
	if err != nil {
		return err
	}

	pushedForkBlocks := 0

	if lastBlockBeforeProcess.ID == commonBlock.ID {
		// rebuilding the chain
		for _, block := range forkBlocks {
			lastBlock, err := bss.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			lastBlockHash, err := utils.GetBlockHash(lastBlock)
			// fmt.Printf("fork block to push %v with previous block %v, the current last block %v prev %v\n", block.GetID(), block.GetPreviousBlockID(), lastBlock.GetID(), lastBlock.GetPreviousBlockID())
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				err := bss.BlockService.PushBlock(lastBlock, block, false)
				if err != nil {
					// bss.P2pService.Blacklist(feederPeer)
					break
				}
				pushedForkBlocks = pushedForkBlocks + 1
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
		peerPoppedOffBlocks, err := bss.PopOffTo(commonBlock)
		if err != nil {
			return err
		}
		pushedForkBlocks = 0
		for _, block := range peerPoppedOffBlocks {
			bss.TransactionService.ProcessLater(block.Transactions)
		}
	}

	// if no fork blocks succesfully applied, go back to our fork
	// other wise, just take the transactions of our popped blocks to be processed later
	if pushedForkBlocks == 0 {
		log.Println("Did not accept any blocks from peer, pushing back my blocks")
		// bss.P2pService.Blacklist(feederPeer, "Did not accept any blocks from peer, pushing back my blocks")
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
			bss.TransactionService.ProcessLater(block.GetTransactions())
		}
	}

	return nil

}

// PopOffTo will remove the block in current Chain until commonBlock is reached
func (bss *Service) PopOffTo(commonBlock *model.Block) ([]*model.Block, error) {
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
	block, _ := bss.BlockService.GetLastBlock()
	bss.LoadTransactions(block.ID)
	fmt.Sprintf("Rollback from block %v at height %v, withCommonBlock %v at height %v", block.GetID(), block.GetHeight(), commonBlock.GetID(), commonBlock.GetHeight())

	genesisBlockID := bss.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID {
		poppedBlocks = append(poppedBlocks, block)

		block, err = bss.popLastBlock()
		if err != nil {
			break
		}
	}

	//	TODO:
	//	NEED TO IMPLEMENT DERIVED TABLES ROLLBACK
	// if err == nil {
	// 	err = service.service.TransactionService.RollbackDerivedTables(commonBlock.GetHeight())
	// 	// _ = DbTransactionalService(chaintype).ClearCache() //need to implement ClearCache
	// 	err = service.CommitTransaction()
	// }

	if err != nil {
		// fmt.Sprintf("Error popping off to %v, %v", commonBlock.GetHeight(), err)
		bss.TransactionService.RollbackTransaction()
		lastBlock, _ := bss.BlockService.GetLastBlock()
		bss.SetLastBlock(lastBlock)
		bss.PopOffTo(lastBlock)
		return []*model.Block{}, err
	}

	return poppedBlocks, nil
}

func (bss *Service) popLastBlock() (*model.Block, error) {
	block, _ := bss.BlockService.GetLastBlock()

	if block.ID == bss.ChainType.GetGenesisBlockID() {
		return nil, blocker.NewBlocker(
			blocker.AuthErr,
			"Failed to pop off because it's Genesis Block",
		)
	}

	previousBlock, _ := bss.TransactionService.DeleteBlocksFrom(block.ID)
	bss.SetLastBlock(previousBlock)

	return previousBlock, nil
}

// SetLastBlock sets the latest block according to inputed block
func (bss *Service) SetLastBlock(block *model.Block) error {
	bss.LastBlock = *block
	return nil
}

func (bss *Service) HasBlock(id int64) bool {
	block, _ := bss.BlockService.GetBlockByID(id)
	if block == nil {
		return false
	}
	return true
}

func (bss *Service) LoadTransactions(blockID int64) {
	if bss.LastBlock.Transactions == nil {
		transactionQ, transactionArg := bss.TransactionQuery.GetTransactionsByBlockID(blockID)
		rows, err := bss.QueryExecutor.ExecuteSelect(transactionQ, transactionArg...)
		if err != nil {
			blocker.NewBlocker(
				blocker.AuthErr,
				"Error when getting transaction to loaded",
			)
		}

		var txs []*model.Transaction
		for rows.Next() {
			txs = bss.TransactionQuery.BuildModel(txs, rows)
		}
		bss.LastBlock.Transactions = txs
	}
}

func (bss *Service) scheduleScan(height uint32, validate bool) {
	bss.TransactionService.ScheduleScan(height, validate)
}

func (bss *Service) getMinRollbackHeight() (uint32, error) {
	lastblock, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return 0, err
	}
	currentHeight := lastblock.Height
	return util.MaxUint32(currentHeight-constant.MinRollbackBlocks, 0), nil
}

func (bs *Service) SetIsScanning(isScanning bool) {
	bs.isScanningBlockchain = isScanning
}
