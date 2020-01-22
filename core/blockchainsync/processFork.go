package blockchainsync

import (
	"bytes"
	"math/big"
	"time"

	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/p2p/strategy"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	ForkingProcessorInterface interface {
		ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error
	}
	ForkingProcessor struct {
		ChainType          chaintype.ChainType
		BlockService       service.BlockServiceInterface
		QueryExecutor      query.ExecutorInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		KVExecutor         kvdb.KVExecutorInterface
		Logger             *log.Logger
		PeerExplorer       strategy.PeerExplorerStrategyInterface
	}
)

// ProcessFork processes the forked blocks
func (fp *ForkingProcessor) ProcessFork(forkBlocks []*model.Block, commonBlock *model.Block, feederPeer *model.Peer) error {

	var (
		lastBlockBeforeProcess, lastBlock, currentLastBlock *model.Block
		myPoppedOffBlocks, peerPoppedOffBlocks              []*model.Block
		pushedForkBlocks                                    int
		err                                                 error
	)

	lastBlockBeforeProcess, err = fp.BlockService.GetLastBlock()
	if err != nil {
		return err
	}
	beforeApplyCumulativeDifficulty := lastBlockBeforeProcess.CumulativeDifficulty
	myPoppedOffBlocks, err = fp.BlockService.PopOffToBlock(commonBlock)
	if err != nil {
		return err
	}

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
			lastBlockHash, err := commonUtil.GetBlockHash(lastBlock, fp.ChainType)
			if err != nil {
				return err
			}
			if bytes.Equal(lastBlockHash, block.PreviousBlockHash) {
				err := fp.BlockService.ValidateBlock(block, lastBlock, time.Now().Unix())
				if err != nil {
					err := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
					if err != nil {
						fp.Logger.Errorf("Failed to add blacklist: %v\n", err)
					}
					fp.Logger.Warnf("[pushing fork block] failed to verify block %v from peer: %s\nwith previous: %v\n", block.ID, err, lastBlock.ID)
				}
				err = fp.BlockService.PushBlock(lastBlock, block, false, true)
				if err != nil {
					err := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
					if err != nil {
						fp.Logger.Errorf("Failed to add blacklist: %v\n", err)
					}
					fp.Logger.Warnf("\n\nPushBlock err %v\n\n", err)
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
		peerPoppedOffBlocks, err = fp.BlockService.PopOffToBlock(commonBlock)
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
		fp.Logger.Println("Did not accept any blocks from peer, pushing back my blocks")
		for _, block := range myPoppedOffBlocks {
			lastBlock, err = fp.BlockService.GetLastBlock()
			if err != nil {
				return err
			}
			err = fp.BlockService.ValidateBlock(block, lastBlock, time.Now().Unix())
			if err != nil {
				err := fp.PeerExplorer.PeerBlacklist(feederPeer, err.Error())
				if err != nil {
					fp.Logger.Errorf("Failed to add blacklist: %v\n", err)
				}
				fp.Logger.Warnf("[pushing back own block] failed to verify block %v from peer: %s\n with previous: %v\n", block.ID, err, lastBlock.ID)
				return err
			}
			err = fp.BlockService.PushBlock(lastBlock, block, false, true)
			if err != nil {
				return blocker.NewBlocker(blocker.BlockErr, "Popped off block no longer acceptable")
			}
		}
	} else {
		for _, block := range myPoppedOffBlocks {
			_ = fp.ProcessLater(block.Transactions)
		}
	}

	if fp.ChainType.HasTransactions() {
		// start restoring mempool from badgerDB
		err = fp.restoreMempoolsBackup()
		if err != nil {
			fp.Logger.Errorf("RestoreBackupFail: %s", err.Error())
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
		txBytes, err = transaction.GetTransactionBytes(tx, true)

		if err != nil {
			return err
		}

		// Save to mempool
		mpTx := &model.MempoolTransaction{
			FeePerByte:              commonUtil.FeePerByteTransaction(tx.GetFee(), txBytes),
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

// restoreMempoolsBackup will restore transaction from badgerDB and try to re-ApplyUnconfirmed
func (fp *ForkingProcessor) restoreMempoolsBackup() error {

	var (
		mempoolsBackupBytes []byte
		prev                uint32
		err                 error
	)

	kvdbMempoolsBackupKey := commonUtil.GetKvDbMempoolDBKey(fp.ChainType)
	mempoolsBackupBytes, err = fp.KVExecutor.Get(kvdbMempoolsBackupKey)
	if err != nil {
		return err
	}

	for int(prev) < len(mempoolsBackupBytes) {
		var (
			transactionBytes []byte
			mempoolTX        *model.MempoolTransaction
			txType           transaction.TypeAction
			tx               *model.Transaction
			size             uint32
		)

		prev += constant.TransactionBodyLength // initiate length of size
		size = commonUtil.ConvertBytesToUint32(mempoolsBackupBytes[:prev])
		transactionBytes = mempoolsBackupBytes[prev:][:size]
		prev += size

		tx, err = transaction.ParseTransactionBytes(transactionBytes, true)
		if err != nil {
			return err
		}
		mempoolTX = &model.MempoolTransaction{
			FeePerByte:              commonUtil.FeePerByteTransaction(tx.GetFee(), transactionBytes),
			ID:                      tx.ID,
			TransactionBytes:        transactionBytes,
			ArrivalTimestamp:        time.Now().Unix(),
			SenderAccountAddress:    tx.SenderAccountAddress,
			RecipientAccountAddress: tx.RecipientAccountAddress,
		}
		err = fp.MempoolService.ValidateMempoolTransaction(mempoolTX)
		if err != nil {
			return err
		}

		txType, err = fp.ActionTypeSwitcher.GetTransactionType(tx)
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
			err = fp.QueryExecutor.RollbackTx()
			if err != nil {
				return err
			}
			return err
		}
		err = fp.MempoolService.AddMempoolTransaction(mempoolTX)
		if err != nil {
			err = fp.QueryExecutor.RollbackTx()
			if err != nil {
				return err
			}
		}
		err = fp.QueryExecutor.CommitTx()
		if err != nil {
			return err
		}
	}
	return nil
}
