package blockchainsync

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	utils "github.com/zoobc/zoobc-core/core/util"
)

type (
	ForkingProcess interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block) error
		PopOff(commonBlock *model.Block) []*model.Block
		popLastBlock() (*model.Block, error)
		SetLastBlock(block *model.Block) error
		HasBlock(id int64) bool
		LoadTransactions()
		popOffWithRescan(height uint32)
		scan(height uint32, validate bool) error
		scheduleScan(height uint32, validate bool)
		getMinRollbackHeight() uint32
		SetIsScanning(isScanning bool)
	}

	blockchainService struct {
		GetMoreBlocks        bool
		IsDownloading        bool // only for status
		LastBlockchainFeeder *model.Peer

		PeerHasMore bool

		IsRestoring          bool // for restoring prunabledata
		PrunableTransactions []int64

		IsScanning           bool
		Chaintype            chaintype.ChainType
		isScanningBlockchain bool

		LastBlock model.Block
	}
)

//The main function to process the forked blocks
func (bss *Service) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block) error {
	var forkBlocksID []int64
	for _, block := range forkBlocks {
		forkBlocksID = append(forkBlocksID, block.ID)
	}

	lastblocktemp, err := bss.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastblocktemp.GetCumulativeDifficulty()

	myPoppedOffBlocks := bss.ForkingProcess.PopOff(commonBlock)

	pushedForkBlocks := 0

	if lastblocktemp.GetID() == commonBlock.GetID() {
		// rebuilding the chain
		for _, block := range forkBlocks {
			lastBlock, err := bss.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			lastBlockHash, err := utils.GetBlockHash(lastBlock)
			// fmt.Printf("fork block to push %v with previous block %v, the current last block %v prev %v\n", block.GetID(), block.GetPreviousBlockID(), lastBlock.GetID(), lastBlock.GetPreviousBlockID())
			if !bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				fmt.Printf("fork block to push2\n")
				err := bss.BlockService.PushBlock(lastBlock, block, false)
				if err != nil {
					// feederPeer.Blacklist("error in pushing peer's block in fork")
					//TODO:
					//RESPONSE BACK IF PUSHING BLOCK ENCOUNTERING ERROR
				}
				pushedForkBlocks = pushedForkBlocks + 1
			}
		}

		peerCumulativeDifficulty, _ := new(big.Int).SetString("", 10)
		peerCumulativeDifficultyOriginalBefore, _ := new(big.Int).SetString(beforeApplyCumulativeDifficulty, 10)

		if pushedForkBlocks > 0 && peerCumulativeDifficulty.Cmp(peerCumulativeDifficultyOriginalBefore) < 0 {
			peerPoppedOffBlocks := bss.ForkingProcess.PopOff(commonBlock)
			pushedForkBlocks = 0
			for _, block := range peerPoppedOffBlocks {
				blockTransaction := block.GetTransactions()
				bss.TransactionService.ProcessLater(blockTransaction)
			}
		}

		if pushedForkBlocks == 0 {
			// HostService(chaintype).Host.Log("Did not accept any blocks from peer, pushing back my blocks")
			// feederPeer.Blacklist("Did not accept any blocks from peer, pushing back my blocks")
			for _, block := range myPoppedOffBlocks {
				lastBlock, err := bss.BlockService.GetLastBlock()
				if err != nil {
					return err
				}
				bss.BlockService.PushBlock(lastBlock, block, false)
			}
		} else {
			for _, block := range myPoppedOffBlocks {
				bss.TransactionService.ProcessLater(block.GetTransactions())
			}
		}
	}

	return nil

}

//This function will remove the block in current Chain until commonBlock is reached
func (bss *Service) PopOff(commonBlock *model.Block) []*model.Block {

	if !bss.TransactionService.IsInTransaction() {
		bss.TransactionService.BeginTransaction()
		defer bss.TransactionService.EndTransaction()
		return bss.ForkingProcess.PopOff(commonBlock)
	}

	lastblock, err := bss.BlockService.GetLastBlock()
	currentHeight := lastblock.GetHeight()

	//If currentHigh is lower than minimal height required for rollback then
	if currentHeight < bss.ForkingProcess.getMinRollbackHeight() {
		bss.ForkingProcess.popOffWithRescan(commonBlock.GetHeight() + 1)
		return []*model.Block{}
	}

	if !bss.ForkingProcess.HasBlock(commonBlock.GetID()) {
		return []*model.Block{}
	}

	poppedBlocks := []*model.Block{}
	block, _ := bss.BlockService.GetLastBlock()
	bss.ForkingProcess.LoadTransactions()
	fmt.Sprintf("Rollback from block %v at height %v, withCommonBlock %v at height %v", block.GetID(), block.GetHeight(), commonBlock.GetID(), commonBlock.GetHeight())

	// var err error
	genesisBlockid := bss.ChainType.GetGenesisBlockID()
	for block.GetID() != commonBlock.GetID() && block.GetID() != genesisBlockid {
		poppedBlocks = append(poppedBlocks, block)

		block, err = bss.ForkingProcess.popLastBlock()
		if err != nil {
			break
		}
	}

	// if err == nil {
	// 	err = service.service.TransactionService.RollbackDerivedTables(commonBlock.GetHeight()) //ask about how derived table works
	// 	// _ = DbTransactionalService(chaintype).ClearCache() //need to implement ClearCache
	// 	err = service.CommitTransaction()
	// }

	if err != nil {
		// fmt.Sprintf("Error popping off to %v, %v", commonBlock.GetHeight(), err)
		bss.TransactionService.RollbackTransaction()
		lastBlock, _ := bss.BlockService.GetLastBlock()
		bss.ForkingProcess.SetLastBlock(lastBlock)
		bss.ForkingProcess.PopOff(lastBlock)
		return []*model.Block{}
	}

	return poppedBlocks
}

func (bss *Service) popLastBlock() (*model.Block, error) {
	block, _ := bss.BlockService.GetLastBlock()
	blockid := block.GetID()

	if block.GetID() == bss.ChainType.GetGenesisBlockID() {
		return nil, errors.New("Cannot pop off Genesis block")
	}

	previousBlock, _ := bss.TransactionService.DeleteBlocksFrom(blockid)
	bss.ForkingProcess.SetLastBlock(previousBlock)

	return previousBlock, nil
}

//Set the latest block according to inputed block
func (bss *Service) SetLastBlock(block *model.Block) error {
	bss.LastBlock = *block
	return nil
}

func (bss *Service) HasBlock(id int64) bool {
	block, _ := bss.BlockService.GetBlockByID(id)
	if block.GetID() == -1 {
		return false
	}
	return true
}

func (bss *Service) LoadTransactions() {
	if bss.LastBlock.Transactions == nil {
		// txs, _ := TransactionRepository(bss.ChainType).GetTransactionByBlockId(b.LastBlock.GetID())
		txs, _ := bss.TransactionQuery.GetTransactionByBlockId(bss.LastBlock.GetID())
		bss.LastBlock.Transactions = txs
	}
}

func (bss *Service) popOffWithRescan(height uint32) {
	bss.ForkingProcess.scheduleScan(0, false)
	currentBlock, _ := bss.BlockService.GetBlockByHeight(height)
	currentID := currentBlock.GetID()
	lastBlock, _ := bss.TransactionService.DeleteBlocksFrom(currentID)
	bss.ForkingProcess.SetLastBlock(lastBlock)
	bss.ForkingProcess.scan(0, false)
}

func (bss *Service) scan(height uint32, validate bool) error {
	// bss.ChainType.WriteLock()
	// defer bss.ChainType.WriteUnlock()

	if !bss.TransactionService.IsInTransaction() {
		bss.TransactionService.BeginTransaction()
		bss.TransactionService.EndTransaction()
		// TODO:
		// defer BlockListener().RemoveListener(checksumListener(), Event.BLOCK_SCANNED)

		if validate {
			// TODO:
			// BlockListener().AddListener(checksumListener(), Event.BLOCK_SCANNED)
			bss.ForkingProcess.scan(height, validate)
		}
		return nil
	}

	bss.ForkingProcess.scheduleScan(height, validate)

	if height < 0 || (height > 0 && height < bss.ForkingProcess.getMinRollbackHeight()) {
		if height > 0 && height < bss.ForkingProcess.getMinRollbackHeight() {
			// HostService(chaintype).Host.Log(fmt.Sprintf("Rollback to height less than %v is not supported, doing a full scan", height))
		}
		height = 0
	}
	// HostService(chaintype).Host.Log(fmt.Sprintf("Scanning the blockchain starting from height %v", height))

	bss.ForkingProcess.SetIsScanning(true)
	defer bss.ForkingProcess.SetIsScanning(false)

	bcHeight := bss.LastBlock.GetHeight()
	// Question: why +1?
	if height > bcHeight+1 {
		// HostService(chaintype).Host.Log(fmt.Sprintf("Rollback height %v exceeds the blockchain height of %v, no scan needed", height-1, bcHeight))
		bss.TransactionService.ScanFinish()
		bss.TransactionService.CommitTransaction()
		return nil
	}

	// service.TransactionService(bss.ChainType).RollbackDerivedTables(height - 1)
	bss.TransactionService.ClearCache()
	bss.TransactionService.CommitTransaction()
	// HostService(chaintype).Host.Log("The derived tables has been rolled back as part of scan")

	currentBlock, _ := bss.TransactionService.GetBlockAtHeight(height)
	// TODO:
	// BlockListener().Notify(currentBlock, Event.RESCAN_BEGIN)

	// TODO: confirm this logic
	if height == 0 {
		bss.ForkingProcess.SetLastBlock(currentBlock)
		// TODO:
		// Evaluate the use of this line
		// models.AccountRepository(chaintype).AddOrGetAccount(constant.GENESIS_PUBLIC_KEY)
	} else {
		bss.ForkingProcess.SetLastBlock(currentBlock)
	}

	hasMore := true
	heightToRetrieve := height
	for hasMore == true {
		blocksToIterate, err := bss.BlockService.GetBlocksFromHeight(height, 50000)
		heightToRetrieve = heightToRetrieve + uint32(len(blocksToIterate))
		if len(blocksToIterate) < 1 {
			hasMore = false
		}

		var blockError error
		for _, block := range blocksToIterate {
			if blockError != nil {
				bss.ForkingProcess.LoadTransactions()
				bss.TransactionService.ProcessLater(currentBlock.GetTransactions())
				continue
			}

			// TODO:
			// loading block's transaction if needed

			if block.GetID() != currentBlock.GetID() || block.GetHeight() > bss.LastBlock.GetHeight()+1 {
				return errors.New("Database blocks in the wrong order!")
			}
			currentBlock = block
			// TODO:
			// BlockListener().Notify(block, event.BEFORE_BLOCK_ACCEPT)
			bss.ForkingProcess.SetLastBlock(block)
			// TODO:
			// evaluate change of accept() to pushBlock()
			// accept(block)
			// BlockService(bss.ChainType).PushBlock(&block)
			bss.TransactionService.ClearCache()
			bss.TransactionService.CommitTransaction()
			// TODO:
			// BlockListener().Notify(block, event.AFTER_BLOCK_ACCEPT)

			if err != nil {
				bss.TransactionService.RollbackTransaction()
				// HostService(chaintype).Host.Log("Failure in the scan procedure")
				// HostService(chaintype).Host.Log(fmt.Sprintf("%v", err))
				if currentBlock.GetID() != -1 {
					bss.ForkingProcess.LoadTransactions()
					bss.TransactionService.ProcessLater(currentBlock.GetTransactions())
				}
				blockError = err
			}
		}

		if blockError != nil {
			lastBlock, _ := bss.TransactionService.DeleteBlocksFrom(currentBlock.GetID())
			bss.ForkingProcess.SetLastBlock(lastBlock)
			bss.ForkingProcess.PopOff(lastBlock)
			break
		}

		// TODO:
		// BlockListener().Notify(currentBlock, event.BLOCK_SCANNED)
	}

	bss.TransactionService.ScanFinish()
	bss.TransactionService.CommitTransaction()
	// TODO:
	// BlockListener().Notify(currentBlock, event.RESCAN_END)
	// HostService(chaintype).Host.Log(fmt.Sprintf("...done at height %v", BlockchainService(chaintype).GetHeight()))

	if height == 0 && validate {
		// HostService(chaintype).Host.Log("Successfully performed full rescan with validation")
	}

	return nil

	// TODO:
	// analyze
	// lastRestoreTime = 0
}

func (bss *Service) scheduleScan(height uint32, validate bool) {
	bss.TransactionService.ScheduleScan(height, validate)
}

func (bss *Service) getMinRollbackHeight() uint32 {
	// TODO:
	// perform the correct calculation
	lastblock, _ := bss.BlockService.GetLastBlock() //need to add handling error if failed
	currentHeight := lastblock.GetHeight()
	return util.MaxUint32(currentHeight-720, 0) //NEED TO ADD MAX_ROLLBACK VARIABLE TO CONSTANT LATER AND CHANGE THE VALUE TO VARIABLE
}

func (bs blockchainService) SetIsScanning(isScanning bool) {
	bs.isScanningBlockchain = isScanning
}
