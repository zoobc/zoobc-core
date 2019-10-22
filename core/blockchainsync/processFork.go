package blockchainsync

import (
	"bytes"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	utils "github.com/zoobc/zoobc-core/core/util"

	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	ForkingProcessorInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error
	}
	ForkingProcessor struct {
		ChainType          chaintype.ChainType
		BlockService       service.BlockServiceInterface
		BlockPopper        *BlockPopper
		QueryExecutor      query.ExecutorInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
	}
)

// ProcessFork processes the forked blocks
func (fp *ForkingProcessor) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {
	var (
		err                                                 error
		myPoppedOffBlocks, peerPoppedOffBlocks              []*model.Block
		lastBlockBeforeProcess, lastBlock, currentLastBlock *model.Block
	)

	lastBlockBeforeProcess, err = fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	myPoppedOffBlocks, err = fp.BlockPopper.PopOffToBlock(commonBlock)
	if err != nil {
		return err
	}

	pushedForkBlocks := 0

	lastBlock, err = fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}

	if lastBlock.ID == commonBlock.ID {
		// rebuilding the chain
		for _, block := range forkBlocks {
			lastBlock, err = fp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			lastBlockHash, _ := utils.GetBlockHash(lastBlock)
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				err := fp.BlockService.ValidateBlock(block, lastBlock, time.Now().Unix())
				if err != nil {
					// TODO: analyze the mechanism of blacklisting peer here
					// bd.P2pService.Blacklist(peer)
					log.Warnf("[pushing fork block] failed to verify block %v from peer: %s\nwith previous: %v\n", block.ID, err, lastBlock.ID)
				}
				err = fp.BlockService.PushBlock(lastBlock, block, false, false)
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

	currentLastBlock, err = fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	currentCumulativeDifficulty, _ := new(big.Int).SetString(currentLastBlock.CumulativeDifficulty, 10)
	cumulativeDifficultyOriginalBefore, _ := new(big.Int).SetString(beforeApplyCumulativeDifficulty, 10)

	// if after applying the fork blocks the cumulative difficulty is still less than current one
	// only take the transactions to be processed, but later will get back to our own fork
	if pushedForkBlocks > 0 && currentCumulativeDifficulty.Cmp(cumulativeDifficultyOriginalBefore) < 0 {
		peerPoppedOffBlocks, err = fp.BlockPopper.PopOffToBlock(commonBlock)
		if err != nil {
			return err
		}
		pushedForkBlocks = 0
		for _, block := range peerPoppedOffBlocks {
			_ = fp.ProcessLater(block.Transactions)
		}
	}

	// if no fork blocks successfully applied, go back to our fork
	// other wise, just take the transactions of our popped blocks to be processed later
	if pushedForkBlocks == 0 {
		log.Println("Did not accept any blocks from peer, pushing back my blocks")
		for _, block := range myPoppedOffBlocks {
			lastBlock, err = fp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			err = fp.BlockService.ValidateBlock(block, lastBlock, time.Now().Unix())
			if err != nil {
				// TODO: analyze the mechanism of blacklisting peer here
				// bd.P2pService.Blacklist(peer)
				log.Warnf("[pushing back own block] failed to verify block %v from peer: %s\n with previous: %v\n", block.ID, err, lastBlock.ID)
				return err
			}
			err = fp.BlockService.PushBlock(lastBlock, block, false, false)
			if err != nil {
				return blocker.NewBlocker(blocker.BlockErr, "Popped off block no longer acceptable")
			}
		}
	} else {
		for _, block := range myPoppedOffBlocks {
			_ = fp.ProcessLater(block.Transactions)
		}
	}

	return nil
}

func (fp *ForkingProcessor) ProcessLater(txs []*model.Transaction) error {
	var (
		err     error
		txBytes []byte
		txType  transaction.TypeAction
	)
	for _, tx := range txs {
		// Validate Tx
		txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return err
		}
		txBytes, err = commonUtil.GetTransactionBytes(tx, true)

		if err != nil {
			return err
		}

		// Save to mempool
		mpTx := &model.MempoolTransaction{
			FeePerByte:              constant.TxFeePerByte,
			ID:                      tx.ID,
			TransactionBytes:        txBytes,
			ArrivalTimestamp:        time.Now().Unix(),
			SenderAccountAddress:    tx.SenderAccountAddress,
			RecipientAccountAddress: tx.RecipientAccountAddress,
		}

		err = fp.MempoolService.ValidateMempoolTransaction(mpTx)
		if err != nil {
			return err
		}
		// Apply Unconfirmed
		err = fp.QueryExecutor.BeginTx()
		if err != nil {
			return err
		}
		err = txType.ApplyUnconfirmed()
		if err != nil {
			errRollback := fp.QueryExecutor.RollbackTx()
			if errRollback != nil {
				return errRollback
			}
			return err
		}
		err = fp.MempoolService.AddMempoolTransaction(mpTx)
		if err != nil {
			errRollback := fp.QueryExecutor.RollbackTx()
			if errRollback != nil {
				return err
			}
			return err
		}
		err = fp.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fp *ForkingProcessor) ScheduleScan(height uint32, validate bool) {
	// TODO: analyze if this mechanism is necessary
}
