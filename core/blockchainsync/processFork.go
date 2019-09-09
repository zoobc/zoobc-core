package blockchainsync

import (
	"bytes"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	utils "github.com/zoobc/zoobc-core/core/util"

	"github.com/zoobc/zoobc-core/core/service"
)

type (
	ForkingProcessorInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error
	}
	ForkingProcessor struct {
		ChainType    chaintype.ChainType
		BlockService service.BlockServiceInterface
		BlockPopper  *BlockPopper
	}
)

//The main function to process the forked blocks
func (fp *ForkingProcessor) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {
	log.Info("processing %d fork blocks...\n", len(forkBlocks))
	var forkBlocksID []int64
	for _, block := range forkBlocks {
		forkBlocksID = append(forkBlocksID, block.ID)
	}

	lastBlockBeforeProcess, err := fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	myPoppedOffBlocks, err := fp.BlockPopper.PopOffToBlock(commonBlock)
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
					// TODO: blacklist the wrong peer
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
		peerPoppedOffBlocks, err := fp.BlockPopper.PopOffToBlock(commonBlock)
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

func (fp *ForkingProcessor) ProcessLater(transaction []*model.Transaction) {}

func (fp *ForkingProcessor) ScheduleScan(height uint32, validate bool) {}
